package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	Roots    []string `yaml:"roots"`
	Ignore   []string `yaml:"ignore"`
	Editor   string   `yaml:"editor"`
	PageSize int      `yaml:"pageSize,omitempty"`
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
