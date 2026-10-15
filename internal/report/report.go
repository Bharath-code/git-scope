// Package report flattens scanned repos into the machine-readable shape shared
// by `git-scope status --json` and the MCP server, so agents see one schema.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/Bharath-code/git-scope/internal/attention"
	"github.com/Bharath-code/git-scope/internal/model"
)

// Repo is one repository's state as reported to scripts and agents.
type Repo struct {
	Path       string    `json:"path" jsonschema:"absolute path of the repository"`
	Name       string    `json:"name"`
	Branch     string    `json:"branch"`
	Dirty      int       `json:"dirty" jsonschema:"staged + unstaged + untracked file count"`
	Ahead      int       `json:"ahead"`
	Behind     int       `json:"behind"`
	Unpushed   int       `json:"unpushed" jsonschema:"local commits not on the upstream branch"`
	LastCommit time.Time `json:"last_commit,omitzero" jsonschema:"omitted when the repo has no commits"`
	Tier       string    `json:"tier" jsonschema:"ACTION, WATCH or CLEAN"`
	Reason     string    `json:"reason"`
	Error      string    `json:"error,omitempty"`
}

// Build sorts repos by attention (most urgent first) and converts them.
func Build(repos []model.Repo, now time.Time) []Repo {
	sorted := append([]model.Repo(nil), repos...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return attention.Less(sorted[i], sorted[j], now)
	})
	out := make([]Repo, 0, len(sorted))
	for _, r := range sorted {
		s := r.Status
		out = append(out, Repo{
			Path:       r.Path,
			Name:       r.Name,
			Branch:     s.Branch,
			Dirty:      s.Staged + s.Unstaged + s.Untracked,
			Ahead:      s.Ahead,
			Behind:     s.Behind,
			Unpushed:   s.Ahead,
			LastCommit: s.LastCommit,
			Tier:       attention.Classify(s, now).String(),
			Reason:     Reason(s, now),
			Error:      s.ScanError,
		})
	}
	return out
}

// Reason explains the tier in a short phrase, listing signals in attention
// priority order (error, unpushed, behind, dirty, stale).
func Reason(s model.RepoStatus, now time.Time) string {
	var parts []string
	if s.ScanError != "" {
		parts = append(parts, "scan error")
	}
	if s.Ahead > 0 {
		parts = append(parts, fmt.Sprintf("%d unpushed commit%s", s.Ahead, plural(s.Ahead)))
	}
	if s.Behind > 0 {
		parts = append(parts, fmt.Sprintf("%d behind upstream", s.Behind))
	}
	if n := s.Staged + s.Unstaged + s.Untracked; n > 0 {
		parts = append(parts, fmt.Sprintf("%d uncommitted file%s", n, plural(n)))
	}
	if attention.Stale(s, now) {
		parts = append(parts, "stale")
	}
	if len(parts) == 0 {
		return "clean"
	}
	return strings.Join(parts, ", ")
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// PrintJSON writes repos as an indented JSON array (never null).
func PrintJSON(w io.Writer, repos []Repo) error {
	if repos == nil {
		repos = []Repo{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(repos); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}
