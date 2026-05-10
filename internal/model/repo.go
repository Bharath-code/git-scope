package model

import "time"

// RepoStatus contains the git status information for a repository
type RepoStatus struct {
	Branch     string    `json:"branch"`
	Ahead      int       `json:"ahead"`
	Behind     int       `json:"behind"`
	Staged     int       `json:"staged"`
	Unstaged   int       `json:"unstaged"`
	Untracked  int       `json:"untracked"`
	LastCommit time.Time `json:"last_commit"`
	IsDirty    bool      `json:"is_dirty"`
	ScanError  string    `json:"scan_error,omitempty"`
}

// Repo represents a git repository with its metadata and status
type Repo struct {
	Name       string     `json:"name"`
	Path       string     `json:"path"`
	Status     RepoStatus `json:"status"`
	IsWorktree bool       `json:"is_worktree,omitempty"`
}
