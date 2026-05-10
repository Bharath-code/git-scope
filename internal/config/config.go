package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	Roots            []string `yaml:"roots"`
	Ignore           []string `yaml:"ignore"`
	Editor           string   `yaml:"editor"`
	PageSize         int      `yaml:"pageSize,omitempty"`
	IncludeWorktrees bool     `yaml:"includeWorktrees,omitempty"`
}

// defaultConfig returns sensible defaults
// By default, scan the current directory so git-scope works out of the box
func defaultConfig() *Config {
	// Get current working directory as default root
	cwd, err := os.Getwd()
	if err != nil {
		// Fallback to home directory if cwd fails
		cwd, _ = os.UserHomeDir()
	}

	return &Config{
		Roots: []string{cwd},
		Ignore: []string{
			"node_modules",
			".next",
			"dist",
			"build",
			"target",
			".venv",
			"vendor",
		},
		Editor:   "code",
		PageSize: 15,
	}
}

// Load reads configuration from a YAML file
// Returns default config if file doesn't exist
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// If file does not exist, return defaults (no error)
		if os.IsNotExist(err) {
			return defaultConfig(), nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := defaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Expand ~ in paths
	for i, root := range cfg.Roots {
		cfg.Roots[i] = expandPath(root)
	}

	// Ensure pageSize has a sensible value
	if cfg.PageSize <= 0 {
		cfg.PageSize = 15
	}

	return cfg, nil
}

// expandPath expands ~ to user home directory and resolves relative paths
func expandPath(path string) string {
	// Handle ~ prefix
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		path = filepath.Join(home, path[2:])
	}

	// Handle "." or relative paths - convert to absolute
	if path == "." || !filepath.IsAbs(path) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return path
		}
		path = absPath
	}

	return path
}

// DefaultConfigPath returns the default config file path
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "./config.yml"
	}
	return filepath.Join(home, ".config", "git-scope", "config.yml")
}

// State holds user-toggled preferences that should persist across runs but
// don't belong in the human-edited YAML config (and would clobber its
// comments on rewrite).
type State struct {
	IncludeWorktrees bool `json:"include_worktrees"`
}

// DefaultStatePath returns the default location for the state file.
func DefaultStatePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "./state.json"
	}
	return filepath.Join(home, ".config", "git-scope", "state.json")
}

// LoadState reads the persisted state file. Returns a zero-value State and
// no error when the file is missing — first-run is not an error.
func LoadState(path string) (State, error) {
	var s State
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return s, fmt.Errorf("read state: %w", err)
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return s, fmt.Errorf("parse state: %w", err)
	}
	return s, nil
}

// SaveState writes the state file, creating the parent directory as needed.
func SaveState(path string, s State) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write state: %w", err)
	}
	return nil
}

// ConfigExists checks if a config file exists at the given path
func ConfigExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// CreateConfig creates a new config file at the given path
func CreateConfig(path string, roots []string, editor string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	cfg := &Config{
		Roots: roots,
		Ignore: []string{
			"node_modules",
			".next",
			"dist",
			"build",
			"target",
			".venv",
			"vendor",
		},
		Editor: editor,
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	// Add header comment
	content := "# git-scope configuration\n# Edit this file to customize scanning behavior\n\n" + string(data)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
