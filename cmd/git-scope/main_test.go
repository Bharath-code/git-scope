package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestRunCheck(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	if err := os.Mkdir(repo, 0755); err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "init", "-q")
	gitIn(t, repo, "commit", "-q", "--allow-empty", "-m", "init")

	run := func(roots ...string) (int, string, string) {
		var stdout, stderr bytes.Buffer
		code := runCheck(&stdout, &stderr, roots, nil)
		return code, stdout.String(), stderr.String()
	}

	if code, out, _ := run(root); code != checkSafe || !strings.Contains(out, "nothing at risk in 1 repo") {
		t.Errorf("clean repo: code %d, output %q", code, out)
	}

	if err := os.WriteFile(filepath.Join(repo, "f"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if code, out, _ := run(root); code != checkAtRisk || !strings.Contains(out, repo+": 1 uncommitted file") {
		t.Errorf("dirty repo: code %d, output %q", code, out)
	}

	// A typo or an empty directory must never read as "safe".
	if code, _, errOut := run(filepath.Join(root, "typo")); code != checkUnsure || errOut == "" {
		t.Errorf("missing root: code %d, stderr %q", code, errOut)
	}
	if code, _, errOut := run(t.TempDir()); code != checkUnsure || errOut == "" {
		t.Errorf("no repos: code %d, stderr %q", code, errOut)
	}
	if code, _, _ := run(repo, filepath.Join(root, "typo")); code != checkUnsure {
		t.Errorf("one root missing: code %d, want %d", code, checkUnsure)
	}
}
