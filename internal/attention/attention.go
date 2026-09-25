// Package attention reduces per-repo git state into a glanceable verdict:
// an ordinal tier (Clean/Watch/Action) plus a within-tier ranking. It is pure
// and depends only on internal/model, so it carries no TUI dependency and can
// back both the dashboard and future CLI/CI consumers.
package attention

import (
	"time"

	"github.com/Bharath-code/git-scope/internal/model"
)

// StaleThreshold is how long a dirty repo may sit untouched before it is
// considered stale and nudged up within its tier.
const StaleThreshold = 7 * 24 * time.Hour

// Tier is the ordinal attention bucket for a repo. Higher is more urgent.
type Tier int

const (
	Clean Tier = iota
	Watch
	Action
)

func (t Tier) String() string {
	switch t {
	case Action:
		return "ACTION"
	case Watch:
		return "WATCH"
	default:
		return "CLEAN"
	}
}

// Glyph returns the single-rune indicator shown in the dashboard Status column.
func (t Tier) Glyph() string {
	switch t {
	case Action:
		return "▲"
	case Watch:
		return "●"
	default:
		return "✓"
	}
}

// Classify buckets a repo by its highest-priority signal. A scan error or any
// divergence from the remote (ahead/behind) is ACTION; uncommitted work that is
// in sync with the remote is WATCH; everything else is CLEAN.
func Classify(s model.RepoStatus, now time.Time) Tier {
	if s.ScanError != "" || s.Ahead > 0 || s.Behind > 0 {
		return Action
	}
	if s.IsDirty {
		return Watch
	}
	return Clean
}

// Stale reports whether a dirty repo has been left untouched past StaleThreshold.
// It never promotes a repo to a higher tier; it only nudges within-tier order.
func Stale(s model.RepoStatus, now time.Time) bool {
	return s.IsDirty && now.Sub(s.LastCommit) > StaleThreshold
}
