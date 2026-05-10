package scan

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

// ScanRootsWithOptions is like ScanRoots but also accepts toggles. When
// includeWorktrees is true, linked worktrees (.git is a regular file
// containing a "gitdir:" pointer) are returned alongside regular repos.
func ScanRootsWithOptions(roots, ignore []string, includeWorktrees bool) ([]model.Repo, error) {
	// Build ignore set from user config + smart defaults
	ignoreSet := make(map[string]struct{}, len(ignore)+len(smartIgnorePatterns))

	// Add user-defined ignores
	for _, pattern := range ignore {
		ignoreSet[pattern] = struct{}{}
	}

	// Add smart defaults (always apply for performance)
	for _, pattern := range smartIgnorePatterns {
		ignoreSet[pattern] = struct{}{}
	}

	var mu sync.Mutex
	var repos []model.Repo
	var wg sync.WaitGroup

	for _, root := range roots {
		// Expand ~ and environment variables
		root = expandPath(root)

		// Check if root exists
		if _, err := os.Stat(root); os.IsNotExist(err) {
			continue
		}

		wg.Add(1)
		go func(r string) {
			defer wg.Done()
			err := filepath.WalkDir(r, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					// Skip directories we can't access
					return nil
				}

				// Skip ignored directories
				if d.IsDir() && shouldIgnore(d.Name(), ignoreSet) {
					return filepath.SkipDir
				}

				// Found a .git directory — regular repository
				if d.IsDir() && d.Name() == ".git" {
					repo := buildRepo(path, false)
					mu.Lock()
					repos = append(repos, repo)
					mu.Unlock()
					return filepath.SkipDir
				}

				// Found a .git file — linked worktree (opt-in)
				if includeWorktrees && !d.IsDir() && d.Name() == ".git" {
					if isWorktreeGitfile(path) {
						repo := buildRepo(path, true)
						mu.Lock()
						repos = append(repos, repo)
						mu.Unlock()
					}
					return nil
				}

				return nil
			})
			if err != nil {
				// Log but don't fail
				fmt.Fprintf(os.Stderr, "warning: scan error in %s: %v\n", r, err)
			}
		}(root)
	}

	wg.Wait()
	return repos, nil
}

// buildRepo constructs a Repo from a .git path (directory for regular repos,
// file for worktrees). The repo's path is the parent of the .git entry.
func buildRepo(gitPath string, isWorktree bool) model.Repo {
	repoPath := filepath.Dir(gitPath)
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
