package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// NormalizeWorkspacePath normalizes a workspace path input from the user.
// It expands ~, converts relative paths to absolute, resolves symlinks,
// and validates that the path exists and is a directory.
func NormalizeWorkspacePath(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	path := input

	// Step 1: Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot expand ~: %w", err)
		}
		path = filepath.Join(home, path[2:])
	} else if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot expand ~: %w", err)
		}
		path = home
	}

	// Step 2: Convert relative paths to absolute
	if !filepath.IsAbs(path) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("cannot resolve path: %w", err)
		}
		path = absPath
	}

	// Step 3: Check if path exists before resolving symlinks
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("path does not exist: %s", input)
		}
		return "", fmt.Errorf("cannot access path: %w", err)
	}

	// Step 4: Validate it's a directory
	if !info.IsDir() {
		return "", fmt.Errorf("path is not a directory: %s", input)
	}

	// Step 5: Resolve symlinks
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		// If symlink resolution fails, use the original path
		// (might happen with broken symlinks)
		return path, nil
	}

	return resolved, nil
}

// expandTilde expands ~ to home directory without validation
func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	} else if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home
	}
	return path
}

// CompleteDirectoryPath attempts to complete a partial directory path.
// It returns the completed path if a unique match is found, or the longest
// common prefix if multiple matches exist. Returns the original input if
// no matches are found.
func CompleteDirectoryPath(input string) string {
	if input == "" {
		return input
	}

	// Remember if input started with ~
	hadTilde := strings.HasPrefix(input, "~")

	// Expand tilde for processing
	path := expandTilde(input)

	// Handle relative paths
	if !filepath.IsAbs(path) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return input
		}
		path = absPath
	}

	// Get the directory and prefix to match
	dir := filepath.Dir(path)
	prefix := filepath.Base(path)

	// If the path exists as-is and is a directory, just add trailing slash
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		// Path is already a complete directory
		if hadTilde {
			home, _ := os.UserHomeDir()
			if strings.HasPrefix(path, home) {
				return "~" + strings.TrimPrefix(path, home) + "/"
			}
		}
		if !strings.HasSuffix(path, "/") {
			return path + "/"
		}
		return path
	}

	// Read the parent directory to find matches
	entries, err := os.ReadDir(dir)
	if err != nil {
		return input
	}

	// Find directories that start with the prefix
	var matches []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			matches = append(matches, entry.Name())
		}
	}

	if len(matches) == 0 {
		return input
	}

	if len(matches) == 1 {
		// Unique match - complete it
		completedPath := filepath.Join(dir, matches[0])

		// Convert back to ~ format if it started with ~
		if hadTilde {
			home, _ := os.UserHomeDir()
			if strings.HasPrefix(completedPath, home) {
				return "~" + strings.TrimPrefix(completedPath, home) + "/"
			}
		}
		return completedPath + "/"
	}

	// Multiple matches - find longest common prefix
	commonPrefix := matches[0]
	for _, match := range matches[1:] {
		for i := 0; i < len(commonPrefix) && i < len(match); i++ {
			if commonPrefix[i] != match[i] {
				commonPrefix = commonPrefix[:i]
				break
			}
		}
		if len(match) < len(commonPrefix) {
			commonPrefix = match
		}
	}

	if len(commonPrefix) > len(prefix) {
		completedPath := filepath.Join(dir, commonPrefix)
		if hadTilde {
			home, _ := os.UserHomeDir()
			if strings.HasPrefix(completedPath, home) {
				return "~" + strings.TrimPrefix(completedPath, home)
			}
		}
		return completedPath
	}

	return input
}
