package attention

import (
	"testing"
	"time"

	"github.com/Bharath-code/git-scope/internal/model"
)

var now = time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)

func status(mut func(*model.RepoStatus)) model.RepoStatus {
	s := model.RepoStatus{LastCommit: now}
	if mut != nil {
		mut(&s)
	}
	return s
}

func TestClassify(t *testing.T) {
	tests := []struct {
		name string
		s    model.RepoStatus
		want Tier
	}{
		{"empty is clean", status(nil), Clean},
		{"ahead is action", status(func(s *model.RepoStatus) { s.Ahead = 1 }), Action},
		{"behind is action", status(func(s *model.RepoStatus) { s.Behind = 1 }), Action},
		{"diverged is action", status(func(s *model.RepoStatus) { s.Ahead = 2; s.Behind = 3 }), Action},
		{"scan error is action", status(func(s *model.RepoStatus) { s.ScanError = "boom" }), Action},
		{"dirty in sync is watch", status(func(s *model.RepoStatus) { s.IsDirty = true; s.Unstaged = 1 }), Watch},
		{"staged in sync is watch", status(func(s *model.RepoStatus) { s.IsDirty = true; s.Staged = 2 }), Watch},
		{"ahead beats dirty -> action", status(func(s *model.RepoStatus) { s.IsDirty = true; s.Ahead = 1 }), Action},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Classify(tc.s, now); got != tc.want {
				t.Fatalf("Classify() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestClassifyAheadBehindBoundary(t *testing.T) {
	if Classify(status(func(s *model.RepoStatus) { s.Ahead = 0; s.Behind = 0 }), now) != Clean {
		t.Fatal("0/0 should be clean")
	}
	if Classify(status(func(s *model.RepoStatus) { s.Ahead = 1 }), now) != Action {
		t.Fatal("ahead=1 should be action")
	}
}

func TestStale(t *testing.T) {
	old := now.Add(-(StaleThreshold + time.Hour))
	fresh := now.Add(-time.Hour)
	tests := []struct {
		name string
		s    model.RepoStatus
		want bool
	}{
		{"dirty and old is stale", status(func(s *model.RepoStatus) { s.IsDirty = true; s.LastCommit = old }), true},
		{"dirty and fresh not stale", status(func(s *model.RepoStatus) { s.IsDirty = true; s.LastCommit = fresh }), false},
		{"clean and old not stale", status(func(s *model.RepoStatus) { s.LastCommit = old }), false},
		{"exactly at threshold not stale", status(func(s *model.RepoStatus) {
			s.IsDirty = true
			s.LastCommit = now.Add(-StaleThreshold)
		}), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Stale(tc.s, now); got != tc.want {
				t.Fatalf("Stale() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTierStringAndGlyph(t *testing.T) {
	cases := []struct {
		t     Tier
		str   string
		glyph string
	}{
		{Clean, "CLEAN", "✓"},
		{Watch, "WATCH", "●"},
		{Action, "ACTION", "▲"},
	}
	for _, c := range cases {
		if c.t.String() != c.str {
			t.Errorf("String() = %q, want %q", c.t.String(), c.str)
		}
		if c.t.Glyph() != c.glyph {
			t.Errorf("Glyph() = %q, want %q", c.t.Glyph(), c.glyph)
		}
	}
}

func TestTierOrdinal(t *testing.T) {
	if Clean >= Watch || Watch >= Action {
		t.Fatal("tier ordinal must be Clean < Watch < Action")
	}
}
