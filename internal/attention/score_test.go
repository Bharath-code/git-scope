package attention

import (
	"sort"
	"testing"

	"github.com/Bharath-code/git-scope/internal/model"
)

func repo(name string, mut func(*model.RepoStatus)) model.Repo {
	return model.Repo{Name: name, Status: status(mut)}
}

func TestSubScorePriority(t *testing.T) {
	unpushed := SubScore(status(func(s *model.RepoStatus) { s.Ahead = 1 }), now)
	behind := SubScore(status(func(s *model.RepoStatus) { s.Behind = 99 }), now)
	dirty := SubScore(status(func(s *model.RepoStatus) { s.IsDirty = true; s.Unstaged = 99 }), now)

	if unpushed <= behind {
		t.Errorf("unpushed (%d) must outrank behind (%d)", unpushed, behind)
	}
	if behind <= dirty {
		t.Errorf("behind (%d) must outrank dirty (%d)", behind, dirty)
	}
}

func TestSubScoreCapsPreventBandOverflow(t *testing.T) {
	// A single ahead commit must always outrank any amount of behind/dirty,
	// even pathological counts, because of the per-band caps.
	oneAhead := SubScore(status(func(s *model.RepoStatus) { s.Ahead = 1 }), now)
	hugeBehindDirty := SubScore(status(func(s *model.RepoStatus) {
		s.Behind = 200
		s.IsDirty = true
		s.Staged = 200
		s.Unstaged = 200
		s.Untracked = 200
	}), now)
	if oneAhead <= hugeBehindDirty {
		t.Errorf("one ahead (%d) must outrank huge behind+dirty (%d)", oneAhead, hugeBehindDirty)
	}
}

func TestSubScoreStaleNudge(t *testing.T) {
	old := now.Add(-(StaleThreshold + 1))
	fresh := repo("fresh", func(s *model.RepoStatus) { s.IsDirty = true; s.Unstaged = 1 })
	staleR := repo("stale", func(s *model.RepoStatus) { s.IsDirty = true; s.Unstaged = 1; s.LastCommit = old })

	if SubScore(staleR.Status, now) <= SubScore(fresh.Status, now) {
		t.Error("stale dirty repo must score above an equally-dirty fresh repo")
	}
}

func TestSubScoreStaleNeverDominates(t *testing.T) {
	// Stale only nudges; it must never lift a WATCH repo above an ACTION repo's band.
	staleDirty := SubScore(status(func(s *model.RepoStatus) {
		s.IsDirty = true
		s.Unstaged = 50
		s.LastCommit = now.Add(-1000 * StaleThreshold)
	}), now)
	oneBehind := SubScore(status(func(s *model.RepoStatus) { s.Behind = 1 }), now)
	if staleDirty >= oneBehind {
		t.Errorf("stale dirty (%d) must not reach the behind band (%d)", staleDirty, oneBehind)
	}
}

func TestLessOrdersByTierThenScoreThenName(t *testing.T) {
	repos := []model.Repo{
		repo("zzz-clean", nil),
		repo("dirty-b", func(s *model.RepoStatus) { s.IsDirty = true; s.Unstaged = 1 }),
		repo("dirty-a", func(s *model.RepoStatus) { s.IsDirty = true; s.Unstaged = 1 }),
		repo("behind", func(s *model.RepoStatus) { s.Behind = 5 }),
		repo("ahead", func(s *model.RepoStatus) { s.Ahead = 1 }),
	}
	sort.SliceStable(repos, func(i, j int) bool { return Less(repos[i], repos[j], now) })

	want := []string{"ahead", "behind", "dirty-a", "dirty-b", "zzz-clean"}
	for i, n := range want {
		if repos[i].Name != n {
			t.Fatalf("position %d = %q, want %q (order: %v)", i, repos[i].Name, n, names(repos))
		}
	}
}

func names(repos []model.Repo) []string {
	out := make([]string, len(repos))
	for i, r := range repos {
		out[i] = r.Name
	}
	return out
}

func TestSummarize(t *testing.T) {
	repos := []model.Repo{
		repo("clean1", nil),
		repo("clean2", nil),
		repo("ahead", func(s *model.RepoStatus) { s.Ahead = 2 }),
		repo("behind", func(s *model.RepoStatus) { s.Behind = 1 }),
		repo("diverged", func(s *model.RepoStatus) { s.Ahead = 1; s.Behind = 1 }),
		repo("dirty", func(s *model.RepoStatus) { s.IsDirty = true; s.Staged = 1 }),
	}
	got := Summarize(repos, now)
	want := Summary{ToPush: 2, Behind: 2, Dirty: 1, Clean: 2}
	if got != want {
		t.Fatalf("Summarize() = %+v, want %+v", got, want)
	}
}

func TestSummarizeEmpty(t *testing.T) {
	if got := Summarize(nil, now); got != (Summary{}) {
		t.Fatalf("Summarize(nil) = %+v, want zero", got)
	}
}
