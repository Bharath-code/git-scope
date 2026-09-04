package gitops

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Bharath-code/git-scope/internal/model"
)

func TestSummaryString(t *testing.T) {
	tests := []struct {
		name string
		sum  Summary
		want string
	}{
		{"all ok", Summary{Succeeded: 8}, "✓ 8 fetched"},
		{"with skips", Summary{Succeeded: 8, Skipped: 2}, "✓ 8 fetched · 2 skipped (no remote)"},
		{"with failures", Summary{Succeeded: 8, Skipped: 2, Failed: 1}, "✓ 8 fetched · 2 skipped (no remote) · 1 failed"},
		{"none", Summary{}, "✓ 0 fetched"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sum.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSummarize(t *testing.T) {
	results := []ActionResult{
		{Status: StatusSuccess},
		{Status: StatusSuccess},
		{Status: StatusSkipped},
		{Status: StatusFailed},
	}
	s := summarize(results)
	if s.Succeeded != 2 || s.Skipped != 1 || s.Failed != 1 {
		t.Errorf("summarize = %+v, want 2/1/1", s)
	}
	if len(s.Results) != 4 {
		t.Errorf("Results length = %d, want 4", len(s.Results))
	}
}

func TestFirstLine(t *testing.T) {
	if got := firstLine("one\ntwo"); got != "one" {
		t.Errorf("firstLine multi = %q, want one", got)
	}
	if got := firstLine("only"); got != "only" {
		t.Errorf("firstLine single = %q, want only", got)
	}
}

// --- Integration tests against real git repos ---

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

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

// makeCloneWithRemote creates a bare "origin" plus a working clone of it,
// returning the path to the working clone (which has a configured remote).
func makeCloneWithRemote(t *testing.T) string {
	t.Helper()
	base := t.TempDir()

	origin := filepath.Join(base, "origin.git")
	gitRun(t, base, "init", "--bare", origin)

	seed := filepath.Join(base, "seed")
	if err := os.MkdirAll(seed, 0755); err != nil {
		t.Fatal(err)
	}
	gitRun(t, seed, "init", "-b", "main")
	if err := os.WriteFile(filepath.Join(seed, "f.txt"), []byte("x\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, seed, "add", "f.txt")
	gitRun(t, seed, "commit", "-m", "init")
	gitRun(t, seed, "remote", "add", "origin", origin)
	gitRun(t, seed, "push", "origin", "main")

	clone := filepath.Join(base, "clone")
	gitRun(t, base, "clone", origin, clone)
	return clone
}

func makeRepoNoRemote(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitRun(t, dir, "init", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("y\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "add", "f.txt")
	gitRun(t, dir, "commit", "-m", "init")
	return dir
}

func TestHasRemote(t *testing.T) {
	requireGit(t)
	if !hasRemote(makeCloneWithRemote(t)) {
		t.Error("clone with origin should report a remote")
	}
	if hasRemote(makeRepoNoRemote(t)) {
		t.Error("local-only repo should report no remote")
	}
	if hasRemote(t.TempDir()) {
		t.Error("non-git dir should report no remote")
	}
}

func TestFetchAll_MixedRepos(t *testing.T) {
	requireGit(t)

	withRemote := makeCloneWithRemote(t)
	noRemote := makeRepoNoRemote(t)

	repos := []model.Repo{
		{Name: "cloned", Path: withRemote},
		{Name: "local-only", Path: noRemote},
	}

	sum := FetchAll(repos)

	if sum.Succeeded != 1 {
		t.Errorf("Succeeded = %d, want 1", sum.Succeeded)
	}
	if sum.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", sum.Skipped)
	}
	if sum.Failed != 0 {
		t.Errorf("Failed = %d, want 0", sum.Failed)
	}

	byName := map[string]ActionResult{}
	for _, r := range sum.Results {
		byName[r.Repo] = r
	}
	if byName["cloned"].Status != StatusSuccess {
		t.Errorf("cloned status = %v, want success", byName["cloned"].Status)
	}
	if byName["local-only"].Status != StatusSkipped {
		t.Errorf("local-only status = %v, want skipped", byName["local-only"].Status)
	}
}

func TestFetchAll_Empty(t *testing.T) {
	sum := FetchAll(nil)
	if sum.Succeeded != 0 || sum.Failed != 0 || sum.Skipped != 0 {
		t.Errorf("empty FetchAll should be all zero, got %+v", sum)
	}
}
