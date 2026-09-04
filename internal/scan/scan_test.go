package scan

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"

	"github.com/Bharath-code/git-scope/internal/model"
)

func TestShouldIgnore(t *testing.T) {
	ignoreSet := map[string]struct{}{
		"node_modules": {},
		".cache":       {},
	}
	tests := []struct {
		name string
		dir  string
		want bool
	}{
		{"exact match", "node_modules", true},
		{"dotfile exact", ".cache", true},
		{"suffix match", "my.cache", true},
		{"no match", "src", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldIgnore(tt.dir, ignoreSet); got != tt.want {
				t.Errorf("shouldIgnore(%q) = %v, want %v", tt.dir, got, tt.want)
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	if got := expandPath("~/projects"); got != filepath.Join(home, "projects") {
		t.Errorf("expandPath(~/projects) = %q, want %q", got, filepath.Join(home, "projects"))
	}

	t.Setenv("GS_TEST_VAR", "/expanded")
	if got := expandPath("$GS_TEST_VAR/x"); got != "/expanded/x" {
		t.Errorf("expandPath env = %q, want /expanded/x", got)
	}

	if got := expandPath("/absolute"); got != "/absolute" {
		t.Errorf("expandPath(/absolute) = %q, want unchanged", got)
	}
}

func TestPrintJSON(t *testing.T) {
	repos := []model.Repo{
		{Name: "alpha", Path: "/code/alpha", Status: model.RepoStatus{Branch: "main"}},
	}
	var buf bytes.Buffer
	if err := PrintJSON(&buf, repos); err != nil {
		t.Fatalf("PrintJSON: %v", err)
	}

	var decoded []model.Repo
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if len(decoded) != 1 || decoded[0].Name != "alpha" {
		t.Errorf("decoded = %+v, want one repo named alpha", decoded)
	}
}

// --- Integration test against real git repos ---

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func makeRepo(t *testing.T, parent, name string) {
	t.Helper()
	dir := filepath.Join(parent, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "init", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "add", "f.txt")
	gitRun(t, dir, "commit", "-m", "init")
}

func TestScanRoots_FindsReposAndRespectsIgnore(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	root := t.TempDir()

	makeRepo(t, root, "service-a")
	makeRepo(t, root, "service-b")

	// A plain directory with no .git -> should not be counted.
	if err := os.MkdirAll(filepath.Join(root, "plain-dir"), 0755); err != nil {
		t.Fatal(err)
	}

	// A repo nested inside an ignored directory -> should be skipped.
	makeRepo(t, filepath.Join(root, "node_modules"), "ignored-repo")

	repos, err := ScanRoots([]string{root}, []string{"node_modules"})
	if err != nil {
		t.Fatalf("ScanRoots: %v", err)
	}

	names := make([]string, 0, len(repos))
	for _, r := range repos {
		names = append(names, r.Name)
	}
	sort.Strings(names)

	want := []string{"service-a", "service-b"}
	if len(names) != len(want) {
		t.Fatalf("found repos %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Errorf("repo[%d] = %q, want %q", i, names[i], want[i])
		}
	}
}

func TestScanRoots_NonexistentRootIsSkipped(t *testing.T) {
	repos, err := ScanRoots([]string{filepath.Join(t.TempDir(), "missing")}, nil)
	if err != nil {
		t.Fatalf("ScanRoots: %v", err)
	}
	if len(repos) != 0 {
		t.Errorf("expected no repos for missing root, got %d", len(repos))
	}
}
