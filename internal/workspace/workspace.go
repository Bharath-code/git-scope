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
