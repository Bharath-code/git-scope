package gitstatus

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Bharath-code/git-scope/internal/model"
)

func TestParseAheadBehind(t *testing.T) {
	tests := []struct {
		name          string
		line          string
		ahead, behind int
		ok            bool
	}{
		{"ahead and behind", "# branch.ab +2 -3", 2, 3, true},
		{"up to date", "# branch.ab +0 -0", 0, 0, true},
		{"ahead only", "# branch.ab +5 -0", 5, 0, true},
		{"too few fields", "# branch.ab +1", 0, 0, false},
		{"non-numeric", "# branch.ab +x -y", 0, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, b, ok := parseAheadBehind(tt.line)
			if a != tt.ahead || b != tt.behind || ok != tt.ok {
				t.Errorf("parseAheadBehind(%q) = (%d, %d, %v), want (%d, %d, %v)",
					tt.line, a, b, ok, tt.ahead, tt.behind, tt.ok)
			}
		})
	}
}

func TestParseXY(t *testing.T) {
	tests := []struct {
		name             string
		line             string
		staged, unstaged bool
	}{
		{"staged only", "1 M. N... 100644 100644 100644 aa bb file", true, false},
		{"unstaged only", "1 .M N... 100644 100644 100644 aa bb file", false, true},
		{"both", "1 MM N... 100644 100644 100644 aa bb file", true, true},
		{"clean", "1 .. N... 100644 100644 100644 aa bb file", false, false},
		{"malformed", "1", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, u := parseXY(tt.line)
			if s != tt.staged || u != tt.unstaged {
				t.Errorf("parseXY(%q) = (%v, %v), want (%v, %v)", tt.line, s, u, tt.staged, tt.unstaged)
			}
		})
	}
}

func TestApplyFileLine(t *testing.T) {
	var st model.RepoStatus
	applyFileLine(&st, "1 M. N... 100644 100644 100644 aa bb staged.go")
	applyFileLine(&st, "1 .M N... 100644 100644 100644 aa bb unstaged.go")
	applyFileLine(&st, "? newfile.txt")
	applyFileLine(&st, "? another.txt")
	applyFileLine(&st, "! ignored.txt") // ignored lines are no-ops

	if st.Staged != 1 {
		t.Errorf("Staged = %d, want 1", st.Staged)
	}
	if st.Unstaged != 1 {
		t.Errorf("Unstaged = %d, want 1", st.Unstaged)
	}
	if st.Untracked != 2 {
		t.Errorf("Untracked = %d, want 2", st.Untracked)
	}
}

func TestApplyBranchHeader(t *testing.T) {
	var st model.RepoStatus
	applyBranchHeader(&st, "# branch.head feature/login")
	applyBranchHeader(&st, "# branch.ab +1 -4")

	if st.Branch != "feature/login" {
		t.Errorf("Branch = %q, want feature/login", st.Branch)
	}
	if st.Ahead != 1 || st.Behind != 4 {
		t.Errorf("ahead/behind = %d/%d, want 1/4", st.Ahead, st.Behind)
	}
}

// --- Integration test against a real git repo ---

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

func initRepoWithCommit(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	gitRun(t, dir, "init", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "add", "README.md")
	gitRun(t, dir, "commit", "-m", "initial")
	return dir
}

func TestStatus_CleanRepo(t *testing.T) {
	dir := initRepoWithCommit(t)

	st, err := Status(dir)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if st.IsDirty {
		t.Errorf("clean repo reported dirty: %+v", st)
	}
	if st.Branch != "main" {
		t.Errorf("branch = %q, want main", st.Branch)
	}
	if st.LastCommit.IsZero() {
		t.Error("LastCommit should be set after a commit")
	}
}

func TestStatus_DirtyRepo(t *testing.T) {
	dir := initRepoWithCommit(t)

	// One staged change and one untracked file.
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("changed\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "add", "README.md")
	if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new\n"), 0644); err != nil {
		t.Fatal(err)
	}

	st, err := Status(dir)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !st.IsDirty {
		t.Error("repo with changes should be dirty")
	}
	if st.Staged != 1 {
		t.Errorf("Staged = %d, want 1", st.Staged)
	}
	if st.Untracked != 1 {
		t.Errorf("Untracked = %d, want 1", st.Untracked)
	}
}

func TestStatus_NonRepoReturnsError(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	if _, err := Status(t.TempDir()); err == nil {
		t.Error("expected error for non-git directory, got nil")
	}
}
