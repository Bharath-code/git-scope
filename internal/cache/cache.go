package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/Bharath-code/git-scope/internal/model"
)

// CacheData represents the cached scan results
type CacheData struct {
	Repos            []model.Repo `json:"repos"`
	Timestamp        time.Time    `json:"timestamp"`
	Roots            []string     `json:"roots"`
	IncludeWorktrees bool         `json:"include_worktrees,omitempty"`
}

// Store interface for caching repo data
type Store interface {
	Load() (*CacheData, error)
	Save(repos []model.Repo, roots []string, includeWorktrees bool) error
	IsValid(maxAge time.Duration) bool
}

// FileStore implements Store using a JSON file
type FileStore struct {
	path string
	data *CacheData
}

// NewFileStore creates a new file-based cache store
func NewFileStore() *FileStore {
	return &FileStore{
		path: getCachePath(),
	}
}

// getCachePath returns the path to the cache file
func getCachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".cache", "git-scope", "repos.json")
}

// Load reads cached data from disk
func (s *FileStore) Load() (*CacheData, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}

	var cache CacheData
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	s.data = &cache
	return &cache, nil
}

// Save writes repos to cache file
func (s *FileStore) Save(repos []model.Repo, roots []string, includeWorktrees bool) error {
	cache := CacheData{
		Repos:            repos,
		Timestamp:        time.Now(),
		Roots:            roots,
		IncludeWorktrees: includeWorktrees,
	}

	// Ensure cache directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

// IsValid checks if cache is still valid based on max age
func (s *FileStore) IsValid(maxAge time.Duration) bool {
	if s.data == nil {
		return false
	}
	return time.Since(s.data.Timestamp) < maxAge
}

// IsSameRoots checks if cached roots match current roots
func (s *FileStore) IsSameRoots(roots []string) bool {
	if s.data == nil || len(s.data.Roots) != len(roots) {
		return false
	}
	for i, r := range roots {
		if s.data.Roots[i] != r {
			return false
		}
	}
	return true
}

// IsSameIncludeWorktrees checks whether the cache was produced with the same
// worktree-inclusion setting. Toggling forces a rescan.
func (s *FileStore) IsSameIncludeWorktrees(includeWorktrees bool) bool {
	if s.data == nil {
		return false
	}
	return s.data.IncludeWorktrees == includeWorktrees
}

// GetTimestamp returns the cache timestamp
func (s *FileStore) GetTimestamp() time.Time {
	if s.data == nil {
		return time.Time{}
	}
	return s.data.Timestamp
}

// Clear removes the cache file
func (s *FileStore) Clear() error {
	return os.Remove(s.path)
}
