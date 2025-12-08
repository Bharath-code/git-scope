package scan

import (
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

// ScanRoots recursively scans the given root directories for git repositories
// It skips directories matching the ignore patterns
func ScanRoots(roots, ignore []string) ([]model.Repo, error) {
	ignoreSet := make(map[string]struct{}, len(ignore))
	for _, pattern := range ignore {
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

				// Found a .git directory
				if d.IsDir() && d.Name() == ".git" {
					repoPath := filepath.Dir(path)
					repoName := filepath.Base(repoPath)

					status, serr := gitstatus.Status(repoPath)

					repo := model.Repo{
						Name:   repoName,
						Path:   repoPath,
						Status: status,
					}
					if serr != nil {
						repo.Status.ScanError = serr.Error()
					}

					mu.Lock()
					repos = append(repos, repo)
					mu.Unlock()

					// Don't walk into .git directory
					return filepath.SkipDir
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
