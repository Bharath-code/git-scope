package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_MissingFileReturnsDefaults(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "does-not-exist.yml"))
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if cfg.Editor != "code" {
		t.Errorf("default editor = %q, want %q", cfg.Editor, "code")
	}
	if cfg.PageSize != 15 {
		t.Errorf("default pageSize = %d, want 15", cfg.PageSize)
	}
	if len(cfg.Roots) == 0 {
		t.Error("default roots should not be empty")
	}
	if !containsString(cfg.Ignore, "node_modules") {
		t.Errorf("default ignore should contain node_modules, got %v", cfg.Ignore)
	}
}

func TestLoad_ValidFileMergesOverDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	body := "roots:\n  - /tmp/projects\neditor: nvim\n"
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Editor != "nvim" {
		t.Errorf("editor = %q, want nvim", cfg.Editor)
	}
	if len(cfg.Roots) != 1 || cfg.Roots[0] != "/tmp/projects" {
		t.Errorf("roots = %v, want [/tmp/projects]", cfg.Roots)
	}
	// Fields absent from the file keep their defaults.
	if !containsString(cfg.Ignore, "node_modules") {
		t.Errorf("ignore should fall back to defaults, got %v", cfg.Ignore)
	}
	if cfg.PageSize != 15 {
		t.Errorf("pageSize should fall back to 15, got %d", cfg.PageSize)
	}
}

func TestLoad_NonPositivePageSizeCoercedTo15(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte("pageSize: 0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.PageSize != 15 {
		t.Errorf("pageSize = %d, want 15 (coerced)", cfg.PageSize)
	}
}

func TestLoad_InvalidYAMLReturnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte("roots: [unterminated\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for malformed YAML, got nil")
	}
}

func TestLoad_ExpandsTildeInRoots(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home directory available")
	}
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte("roots:\n  - ~/code\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := filepath.Join(home, "code")
	if cfg.Roots[0] != want {
		t.Errorf("expanded root = %q, want %q", cfg.Roots[0], want)
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()
	cwd, _ := os.Getwd()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"tilde", "~/foo", filepath.Join(home, "foo")},
		{"absolute unchanged", "/var/log", "/var/log"},
		{"dot becomes cwd", ".", cwd},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := expandPath(tt.in); got != tt.want {
				t.Errorf("expandPath(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestExpandPath_RelativeBecomesAbsolute(t *testing.T) {
	got := expandPath("some/relative/dir")
	if !filepath.IsAbs(got) {
		t.Errorf("expandPath returned non-absolute path %q", got)
	}
}

func TestConfigExists(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(existing, []byte("editor: code\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if !ConfigExists(existing) {
		t.Error("ConfigExists = false for existing file, want true")
	}
	if ConfigExists(filepath.Join(dir, "nope.yml")) {
		t.Error("ConfigExists = true for missing file, want false")
	}
}

func TestCreateConfig_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.yml")
	roots := []string{"/a", "/b"}

	if err := CreateConfig(path, roots, "vim"); err != nil {
		t.Fatalf("CreateConfig: %v", err)
	}
	if !ConfigExists(path) {
		t.Fatal("config file was not created")
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load after CreateConfig: %v", err)
	}
	if cfg.Editor != "vim" {
		t.Errorf("editor = %q, want vim", cfg.Editor)
	}
	if len(cfg.Roots) != 2 {
		t.Errorf("roots = %v, want 2 entries", cfg.Roots)
	}
}

func TestDefaultConfigPath(t *testing.T) {
	got := DefaultConfigPath()
	if got == "" {
		t.Fatal("DefaultConfigPath returned empty string")
	}
	if !strings.HasSuffix(filepath.ToSlash(got), "git-scope/config.yml") {
		t.Errorf("DefaultConfigPath = %q, want suffix git-scope/config.yml", got)
	}
}

func containsString(s []string, want string) bool {
	for _, v := range s {
		if v == want {
			return true
		}
	}
	return false
}
