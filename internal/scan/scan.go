package scan

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/Bharath-code/git-scope/internal/gitstatus"
	"github.com/Bharath-code/git-scope/internal/model"
)

// smartIgnorePatterns are always-ignored directories for performance
// These are system/tool directories that should never contain user repos
var smartIgnorePatterns = []string{
	// macOS/Linux system directories
	"Library", ".Trash", ".cache", ".local",
	// Package managers & runtimes
	".npm", ".yarn", ".pnpm", ".bun", ".cargo", ".rustup", ".go",
	".venv", ".pyenv", ".rbenv", ".nvm", ".sdkman",
	// IDE extensions (contain third-party repos, not your code)
	".vscode", ".vscode-server", ".cursor", ".zed", ".idea", ".atom",
	// Shell & tools configs
	".oh-my-zsh", ".tmux", ".vim", ".emacs.d", ".gemini",
	// Docker/Cloud
	".docker", ".kube", ".ssh", ".gnupg",
	// Cloud sync (slow and likely duplicates)
	"Google Drive", "OneDrive", "Dropbox", "iCloud",
}

// ScanRoots recursively scans the given root directories for git repositories.
// Skips directories matching the ignore patterns. Worktrees are excluded.
func ScanRoots(roots, ignore []string) ([]model.Repo, error) {
	return ScanRootsWithOptions(roots, ignore, false)
}

// repoFinding is a repo discovered by the directory walk, before its git
// status has been resolved.
type repoFinding struct {
	path       string
	isWorktree bool
}

// ScanRootsWithOptions is like ScanRoots but also accepts toggles. When
// includeWorktrees is true, linked worktrees (.git is a regular file
// containing a "gitdir:" pointer) are returned alongside regular repos.
//
// The scan runs in two phases: a parallel directory walk discovers repos,
// then a fixed worker pool resolves git status for each finding. The status
// phase is the bottleneck on large trees (every repo forks `git`), so
// parallelism here scales nearly linearly with CPU count on cold scans.
func ScanRootsWithOptions(roots, ignore []string, includeWorktrees bool) ([]model.Repo, error) {
	// Build ignore set from user config + smart defaults
	ignoreSet := make(map[string]struct{}, len(ignore)+len(smartIgnorePatterns))
	for _, pattern := range ignore {
		ignoreSet[pattern] = struct{}{}
	}
	for _, pattern := range smartIgnorePatterns {
		ignoreSet[pattern] = struct{}{}
	}

	// Phase 1: discover repos in parallel by root.
	findings := discoverRepos(roots, ignoreSet, includeWorktrees)

	// Phase 2: resolve git status concurrently across a worker pool.
	return resolveStatuses(findings), nil
}

// discoverRepos walks each root in parallel and returns repo findings.
// Walks share an ignore set; results are deduplicated implicitly because
// `.git` directories are pruned from further traversal.
func discoverRepos(roots []string, ignoreSet map[string]struct{}, includeWorktrees bool) []repoFinding {
	var mu sync.Mutex
	var findings []repoFinding
	var wg sync.WaitGroup

	for _, root := range roots {
		root = expandPath(root)
		if _, err := os.Stat(root); os.IsNotExist(err) {
			continue
		}

		wg.Add(1)
		go func(r string) {
			defer wg.Done()
			local := walkRoot(r, ignoreSet, includeWorktrees)
			if len(local) == 0 {
				return
			}
			mu.Lock()
			findings = append(findings, local...)
			mu.Unlock()
		}(root)
	}

	wg.Wait()
	return findings
}

// walkRoot walks one root and returns the repos it found.
func walkRoot(root string, ignoreSet map[string]struct{}, includeWorktrees bool) []repoFinding {
	var findings []repoFinding
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && shouldIgnore(d.Name(), ignoreSet) {
			return filepath.SkipDir
		}
		if d.IsDir() && d.Name() == ".git" {
			findings = append(findings, repoFinding{path: filepath.Dir(path), isWorktree: false})
			return filepath.SkipDir
		}
		if includeWorktrees && !d.IsDir() && d.Name() == ".git" && isWorktreeGitfile(path) {
			findings = append(findings, repoFinding{path: filepath.Dir(path), isWorktree: true})
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: scan error in %s: %v\n", root, err)
	}
	return findings
}

// resolveStatuses runs gitstatus.Status across a worker pool. Order is not
// preserved; the TUI sorts results separately.
func resolveStatuses(findings []repoFinding) []model.Repo {
	if len(findings) == 0 {
		return nil
	}

	workers := runtime.NumCPU() * 2
	if workers > len(findings) {
		workers = len(findings)
	}

	jobs := make(chan repoFinding, len(findings))
	for _, f := range findings {
		jobs <- f
	}
	close(jobs)

	results := make(chan model.Repo, len(findings))
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range jobs {
				results <- buildRepo(f.path, f.isWorktree)
			}
		}()
	}
	wg.Wait()
	close(results)

	repos := make([]model.Repo, 0, len(findings))
	for r := range results {
		repos = append(repos, r)
	}
	return repos
}

// buildRepo constructs a Repo from a working-tree path. Resolves the path to
// absolute form so the repo name uses the real basename even when the input
// was relative.
func buildRepo(repoPath string, isWorktree bool) model.Repo {
	if abs, err := filepath.Abs(repoPath); err == nil {
		repoPath = abs
	}
	status, serr := gitstatus.Status(repoPath)
	repo := model.Repo{
		Name:       filepath.Base(repoPath),
		Path:       repoPath,
		Status:     status,
		IsWorktree: isWorktree,
	}
	if serr != nil {
		repo.Status.ScanError = serr.Error()
	}
	return repo
}

// isWorktreeGitfile checks whether a .git file is a linked-worktree pointer.
// Worktrees: `.git` is a file whose first line is `gitdir: <main-repo>/.git/worktrees/<name>`.
// Submodules use the same gitdir-pointer format but point at `.git/modules/<name>`,
// so we filter on the path segment to exclude them.
func isWorktreeGitfile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return false
	}
	line := scanner.Text()
	if !strings.HasPrefix(line, "gitdir:") {
		return false
	}
	gitdir := strings.TrimSpace(strings.TrimPrefix(line, "gitdir:"))
	return strings.Contains(gitdir, "/worktrees/") || strings.Contains(gitdir, `\worktrees\`)
}

// shouldIgnore checks if a directory name matches any ignore pattern
func shouldIgnore(name string, ignoreSet map[string]struct{}) bool {
	// Exact match
	if _, ok := ignoreSet[name]; ok {
		return true
	}
	// Suffix match
	for pat := range ignoreSet {
		if strings.HasSuffix(name, pat) {
			return true
		}
	}
	return false
}

// expandPath expands ~ and environment variables in a path
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		path = filepath.Join(home, path[2:])
	}
	return os.ExpandEnv(path)
}

// PrintJSON outputs the repos as formatted JSON
func PrintJSON(w io.Writer, repos []model.Repo) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(repos); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}
