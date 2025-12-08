package cache

import "github.com/Bharath-code/git-scope/internal/model"

// Store defines the interface for caching repo data
type Store interface {
	Load() ([]model.Repo, error)
	Save([]model.Repo) error
}

// TODO: Implement JSON or SQLite-backed store for faster startup
// This is a stub for the MVP - caching will be added in a future version
