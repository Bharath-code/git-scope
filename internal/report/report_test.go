package report

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/Bharath-code/git-scope/internal/model"
)

func TestJSONShape(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	recent := now.Add(-time.Hour)
	old := now.Add(-30 * 24 * time.Hour)

	tests := []struct {
		name string
		repo model.Repo
		want map[string]any
	}{
		{
			name: "clean",
			repo: model.Repo{Name: "a", Path: "/r/a", Status: model.RepoStatus{Branch: "main", LastCommit: recent}},
			want: map[string]any{"path": "/r/a", "name": "a", "branch": "main", "dirty": 0.0, "ahead": 0.0, "behind": 0.0,
				"unpushed": 0.0, "last_commit": recent.Format(time.RFC3339), "tier": "CLEAN", "reason": "clean"},
		},
		{
			name: "unpushed",
			repo: model.Repo{Name: "b", Path: "/r/b", Status: model.RepoStatus{Branch: "feat", Ahead: 2, IsDirty: true, LastCommit: recent}},
			want: map[string]any{"path": "/r/b", "name": "b", "branch": "feat", "dirty": 0.0, "ahead": 2.0, "behind": 0.0,
				"unpushed": 2.0, "last_commit": recent.Format(time.RFC3339), "tier": "ACTION", "reason": "2 unpushed commits"},
		},
		{
			name: "stale dirty",
			repo: model.Repo{Name: "c", Path: "/r/c", Status: model.RepoStatus{Branch: "main", Unstaged: 1, Untracked: 2, IsDirty: true, LastCommit: old}},
			want: map[string]any{"path": "/r/c", "name": "c", "branch": "main", "dirty": 3.0, "ahead": 0.0, "behind": 0.0,
				"unpushed": 0.0, "last_commit": old.Format(time.RFC3339), "tier": "WATCH", "reason": "3 uncommitted files, stale"},
		},
		{
			name: "scan error",
			repo: model.Repo{Name: "d", Path: "/r/d", Status: model.RepoStatus{ScanError: "boom"}},
			want: map[string]any{"path": "/r/d", "name": "d", "branch": "", "dirty": 0.0, "ahead": 0.0, "behind": 0.0,
				"unpushed": 0.0, "tier": "ACTION", "reason": "scan error", "error": "boom"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintJSON(&buf, Build([]model.Repo{tt.repo}, now)); err != nil {
				t.Fatal(err)
			}
			var got []map[string]any
			if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
				t.Fatalf("not a JSON array: %v\n%s", err, buf.String())
			}
			if len(got) != 1 {
				t.Fatalf("got %d items, want 1", len(got))
			}
			if len(got[0]) != len(tt.want) {
				t.Errorf("keys = %v, want %v", got[0], tt.want)
			}
			for k, v := range tt.want {
				if got[0][k] != v {
					t.Errorf("%s = %v, want %v", k, got[0][k], v)
				}
			}
		})
	}
}

func TestBuildSortsByAttentionAndEmptyIsArray(t *testing.T) {
	now := time.Now()
	got := Build([]model.Repo{
		{Name: "clean"},
		{Name: "dirty", Status: model.RepoStatus{Unstaged: 1, IsDirty: true, LastCommit: now}},
		{Name: "ahead", Status: model.RepoStatus{Ahead: 1, IsDirty: true}},
	}, now)
	if got[0].Name != "ahead" || got[1].Name != "dirty" || got[2].Name != "clean" {
		t.Errorf("order = %s,%s,%s", got[0].Name, got[1].Name, got[2].Name)
	}

	var buf bytes.Buffer
	if err := PrintJSON(&buf, nil); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "[]\n" {
		t.Errorf("empty = %q, want []", buf.String())
	}
}
