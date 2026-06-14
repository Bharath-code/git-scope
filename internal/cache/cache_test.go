package cache

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/Bharath-code/git-scope/internal/model"
)

// newTestStore returns a FileStore pointed at an isolated temp file so tests
// never touch the user's real ~/.cache directory.
func newTestStore(t *testing.T) *FileStore {
	t.Helper()
	return &FileStore{path: filepath.Join(t.TempDir(), "repos.json")}
}

func sampleRepos() []model.Repo {
	return []model.Repo{
		{Name: "alpha", Path: "/code/alpha", Status: model.RepoStatus{Branch: "main", IsDirty: true, Staged: 2}},
		{Name: "beta", Path: "/code/beta", Status: model.RepoStatus{Branch: "dev"}},
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	s := newTestStore(t)
	roots := []string{"/code"}

	if err := s.Save(sampleRepos(), roots); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := s.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Repos) != 2 {
		t.Fatalf("loaded %d repos, want 2", len(loaded.Repos))
	}
	if loaded.Repos[0].Name != "alpha" || !loaded.Repos[0].Status.IsDirty {
		t.Errorf("first repo not round-tripped correctly: %+v", loaded.Repos[0])
	}
	if len(loaded.Roots) != 1 || loaded.Roots[0] != "/code" {
		t.Errorf("roots = %v, want [/code]", loaded.Roots)
	}
}

func TestLoad_MissingFileReturnsError(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.Load(); err == nil {
		t.Fatal("expected error loading non-existent cache, got nil")
	}
}

func TestIsValid(t *testing.T) {
	s := newTestStore(t)

	// No data loaded yet -> invalid.
	if s.IsValid(time.Hour) {
		t.Error("IsValid = true before any data, want false")
	}

	if err := s.Save(sampleRepos(), []string{"/code"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Load(); err != nil {
		t.Fatal(err)
	}

	if !s.IsValid(time.Hour) {
		t.Error("freshly saved cache should be valid within 1h")
	}
	if s.IsValid(time.Nanosecond) {
		t.Error("cache should be stale against a 1ns max age")
	}
}

func TestIsSameRoots(t *testing.T) {
	s := newTestStore(t)
	if err := s.Save(sampleRepos(), []string{"/a", "/b"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Load(); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		roots []string
		want  bool
	}{
		{"identical", []string{"/a", "/b"}, true},
		{"different length", []string{"/a"}, false},
		{"different order", []string{"/b", "/a"}, false},
		{"different values", []string{"/a", "/c"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.IsSameRoots(tt.roots); got != tt.want {
				t.Errorf("IsSameRoots(%v) = %v, want %v", tt.roots, got, tt.want)
			}
		})
	}
}

func TestGetTimestamp(t *testing.T) {
	s := newTestStore(t)
	if !s.GetTimestamp().IsZero() {
		t.Error("GetTimestamp should be zero before load")
	}

	before := time.Now()
	if err := s.Save(sampleRepos(), nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Load(); err != nil {
		t.Fatal(err)
	}
	if ts := s.GetTimestamp(); ts.Before(before.Add(-time.Second)) {
		t.Errorf("timestamp %v is older than save time %v", ts, before)
	}
}

func TestClear(t *testing.T) {
	s := newTestStore(t)
	if err := s.Save(sampleRepos(), nil); err != nil {
		t.Fatal(err)
	}
	if err := s.Clear(); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	if _, err := s.Load(); err == nil {
		t.Error("expected error loading after Clear, got nil")
	}
}
