package gitstatus

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

// A repo's own config must not be able to make git-scope execute code.
// core.fsmonitor names a program git runs during `git status`.
func TestStatus_IgnoresRepoFsmonitor(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script hook")
	}
	dir := initRepoWithCommit(t)
	marker := filepath.Join(t.TempDir(), "executed")
	hook := filepath.Join(dir, "hook.sh")
	if err := os.WriteFile(hook, []byte("#!/bin/sh\ntouch "+marker+"\n"), 0755); err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "config", "core.fsmonitor", hook)

	if _, err := Status(dir); err != nil {
		t.Fatalf("Status: %v", err)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("core.fsmonitor from repo config was executed")
	}
}

// log.showSignature makes `git log` and `git stash list` verify signatures by
// running gpg.program, which a repo's own config can point anywhere.
func TestStatus_IgnoresRepoGpgProgram(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script hook")
	}
	dir := initRepoWithCommit(t)
	marker := filepath.Join(t.TempDir(), "executed")
	hook := filepath.Join(t.TempDir(), "gpg.sh")
	if err := os.WriteFile(hook, []byte("#!/bin/sh\ntouch "+marker+"\n"), 0755); err != nil {
		t.Fatal(err)
	}

	// Sign HEAD and a stash with a fake signature; the content never needs to verify.
	sign := func(ref string) string {
		obj := gitOutput(t, dir, "cat-file", "commit", ref)
		head, body, _ := strings.Cut(obj, "\n\n")
		signed := head + "\ngpgsig -----BEGIN PGP SIGNATURE-----\n x\n -----END PGP SIGNATURE-----\n\n" + body
		file := filepath.Join(t.TempDir(), "commit")
		if err := os.WriteFile(file, []byte(signed), 0644); err != nil {
			t.Fatal(err)
		}
		return strings.TrimSpace(gitOutput(t, dir, "hash-object", "-t", "commit", "-w", file))
	}
	gitRun(t, dir, "update-ref", "HEAD", sign("HEAD"))
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("changed\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "stash", "-q")
	gitRun(t, dir, "update-ref", "refs/stash", sign("refs/stash"))

	gitRun(t, dir, "config", "log.showSignature", "true")
	gitRun(t, dir, "config", "gpg.program", hook)

	if _, err := Status(dir); err != nil {
		t.Fatalf("Status: %v", err)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("gpg.program from repo config was executed")
	}
}

func TestApplyFileLine_ConflictCountsAsUnstaged(t *testing.T) {
	var st model.RepoStatus
	applyFileLine(&st, "u UU N... 100644 100644 100644 100644 aa bb cc f.go")
	if st.Unstaged != 1 {
		t.Errorf("Unstaged = %d, want 1 for a conflicted file", st.Unstaged)
	}
}

// Stashes and commits on branches that aren't checked out are invisible to
// `git status`, so a repo holding only those must not look clean.
func TestStatus_HiddenWork(t *testing.T) {
	origin := initRepoWithCommit(t)
	clone := t.TempDir()
	gitRun(t, clone, "clone", "-q", origin, ".")

	commit := func(dir, name string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0644); err != nil {
			t.Fatal(err)
		}
		gitRun(t, dir, "add", name)
		gitRun(t, dir, "commit", "-qm", name)
	}
	check := func(t *testing.T, dir string, stashes, other int, dirty bool) {
		t.Helper()
		st, err := Status(dir)
		if err != nil {
			t.Fatal(err)
		}
		if st.Stashes != stashes || st.OtherBranchCommits != other || st.IsDirty != dirty {
			t.Errorf("Stashes=%d OtherBranchCommits=%d IsDirty=%v, want %d, %d, %v",
				st.Stashes, st.OtherBranchCommits, st.IsDirty, stashes, other, dirty)
		}
	}

	check(t, clone, 0, 0, false)
	main := strings.TrimSpace(gitOutput(t, clone, "branch", "--show-current"))

	gitRun(t, clone, "checkout", "-qb", "agent/done")
	commit(clone, "a")
	commit(clone, "b")
	gitRun(t, clone, "checkout", "-q", main)
	check(t, clone, 0, 2, true)

	// Once pushed, the branch is safe.
	gitRun(t, clone, "push", "-q", "origin", "agent/done")
	check(t, clone, 0, 0, false)

	// Merged into HEAD: the commits are reported as Ahead, not twice.
	gitRun(t, clone, "checkout", "-qb", "agent/merged")
	commit(clone, "c")
	gitRun(t, clone, "checkout", "-q", main)
	gitRun(t, clone, "merge", "-q", "--ff-only", "agent/merged")
	if st, _ := Status(clone); st.OtherBranchCommits != 0 || st.Ahead != 1 {
		t.Errorf("merged branch: OtherBranchCommits=%d Ahead=%d, want 0, 1", st.OtherBranchCommits, st.Ahead)
	}

	// A stash counts even with no remote; other branches don't, because there
	// is nowhere to push them.
	local := initRepoWithCommit(t)
	branch := strings.TrimSpace(gitOutput(t, local, "branch", "--show-current"))
	gitRun(t, local, "checkout", "-qb", "side")
	commit(local, "s")
	gitRun(t, local, "checkout", "-q", branch)
	check(t, local, 0, 0, false)

	commit(local, "tracked")
	if err := os.WriteFile(filepath.Join(local, "tracked"), []byte("changed"), 0644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, local, "stash", "-q")
	check(t, local, 1, 0, true)
}

// A repo with no commits must not error just because HEAD doesn't resolve yet,
// but a failed check elsewhere must surface rather than read as "nothing hidden".
func TestStatus_HiddenWorkErrors(t *testing.T) {
	unborn := t.TempDir()
	gitRun(t, unborn, "init", "-q")
	gitRun(t, unborn, "remote", "add", "origin", "/nonexistent")
	if _, err := Status(unborn); err != nil {
		t.Errorf("unborn HEAD with a remote: %v", err)
	}

	var st model.RepoStatus
	if err := applyHiddenWork(t.TempDir(), &st, false, false); err == nil {
		t.Error("applyHiddenWork outside a repo returned no error")
	}
	if _, err := countRevs(initRepoWithCommit(t), "refs/heads/missing"); err == nil {
		t.Error("countRevs on a missing ref returned no error")
	}
}

func gitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return string(out)
}

// Work with no upstream (a never-pushed agent branch, a detached HEAD) has no
// `branch.ab` line, so it must be counted against the remotes directly.
func TestStatus_UnpublishedWork(t *testing.T) {
	origin := initRepoWithCommit(t)
	clone := t.TempDir()
	gitRun(t, clone, "clone", "-q", origin, ".")

	commit := func(name string) {
		if err := os.WriteFile(filepath.Join(clone, name), []byte(name), 0644); err != nil {
			t.Fatal(err)
		}
		gitRun(t, clone, "add", name)
		gitRun(t, clone, "commit", "-qm", name)
	}

	tests := []struct {
		name  string
		setup func()
		ahead int
	}{
		{"tracked and in sync", func() {}, 0},
		{"new branch, never pushed", func() { gitRun(t, clone, "checkout", "-qb", "agent/x"); commit("a"); commit("b") }, 2},
		{"detached head", func() { gitRun(t, clone, "checkout", "-q", "--detach"); commit("c") }, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			st, err := Status(clone)
			if err != nil {
				t.Fatal(err)
			}
			if st.Ahead != tt.ahead {
				t.Errorf("Ahead = %d, want %d", st.Ahead, tt.ahead)
			}
		})
	}

	// A repo with no remote has nowhere to push: never report unpushed.
	if st, _ := Status(origin); st.Ahead != 0 {
		t.Errorf("no-remote repo Ahead = %d, want 0", st.Ahead)
	}
}
