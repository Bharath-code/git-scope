package attention

import (
	"time"

	"github.com/Bharath-code/git-scope/internal/model"
)

// Per-band weights. Counts are capped at countCap so each higher-priority band
// always outranks any amount of lower-priority signal (a single unpushed commit
// outranks any number of behind/dirty). Priority: unpushed > behind > dirty > stale.
const (
	countCap    = 99
	weightAhead = 100000
	weightBehd  = 1000
	weightDirty = 10
	weightErr   = 5
	weightStale = 1
)

func capCount(n int) int {
	if n > countCap {
		return countCap
	}
	if n < 0 {
		return 0
	}
	return n
}

// SubScore ranks repos within a tier. It is not meaningful across tiers (tier is
// the primary key); Less applies tier first, then SubScore.
func SubScore(s model.RepoStatus, now time.Time) int {
	score := weightAhead*capCount(s.Ahead) + weightBehd*capCount(s.Behind)
	score += weightDirty * capCount(s.Staged+s.Unstaged+s.Untracked)
	if s.ScanError != "" {
		score += weightErr
	}
	if Stale(s, now) {
		score += weightStale
	}
	return score
}

// Less reports whether repo a should sort before repo b: higher tier first, then
// higher sub-score, then name ascending for a stable, deterministic order.
func Less(a, b model.Repo, now time.Time) bool {
	ta, tb := Classify(a.Status, now), Classify(b.Status, now)
	if ta != tb {
		return ta > tb
	}
	sa, sb := SubScore(a.Status, now), SubScore(b.Status, now)
	if sa != sb {
		return sa > sb
	}
	return a.Name < b.Name
}

// Summary is the workspace-level rollup shown in the dashboard header. Counts are
// per-signal, not a partition: a diverged repo is counted in both ToPush and Behind.
type Summary struct {
	ToPush int
	Behind int
	Dirty  int
	Clean  int
}

// Summarize aggregates attention signals across a set of repos.
func Summarize(repos []model.Repo, now time.Time) Summary {
	var s Summary
	for _, r := range repos {
		if r.Status.Ahead > 0 {
			s.ToPush++
		}
		if r.Status.Behind > 0 {
			s.Behind++
		}
		if r.Status.IsDirty {
			s.Dirty++
		}
		if Classify(r.Status, now) == Clean {
			s.Clean++
		}
	}
	return s
}
