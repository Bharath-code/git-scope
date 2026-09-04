// Package gitops performs bulk git actions across many repositories.
//
// It is deliberately separate from the read-only gitstatus package: gitstatus
// observes repositories, gitops acts on them. Actions here are restricted to
// operations that are safe to run unattended across a whole workspace.
package gitops

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/Bharath-code/git-scope/internal/model"
)

// Status is the outcome of an action on a single repository.
type Status int

const (
	StatusSuccess Status = iota
	StatusFailed
	StatusSkipped // e.g. no remote configured
)

// ActionResult is the outcome of an action on one repository.
type ActionResult struct {
	Repo   string
	Path   string
	Status Status
	Err    error
}

// Summary aggregates the results of a bulk action.
type Summary struct {
	Results   []ActionResult
	Succeeded int
	Failed    int
	Skipped   int
}

// String renders a one-line human summary, e.g.
// "✓ 8 fetched · 2 skipped (no remote) · 1 failed".
func (s Summary) String() string {
	parts := []string{fmt.Sprintf("✓ %d fetched", s.Succeeded)}
	if s.Skipped > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped (no remote)", s.Skipped))
	}
	if s.Failed > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", s.Failed))
	}
	return strings.Join(parts, " · ")
}

const (
	fetchTimeout   = 30 * time.Second
	maxConcurrency = 8
)

// FetchAll runs `git fetch` concurrently across the given repositories.
//
// Fetch is network-only: it updates remote-tracking refs but never modifies the
// working tree, local branches, or commits — which is why it is safe to run in
// bulk without confirmation. Repositories with no configured remote are skipped
// rather than reported as failures. Each fetch is bounded by a timeout so a
// single unreachable remote cannot stall the batch.
func FetchAll(repos []model.Repo) Summary {
	results := make([]ActionResult, len(repos))
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for i := range repos {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[i] = fetchOne(repos[i])
		}(i)
	}
	wg.Wait()

	return summarize(results)
}

// summarize tallies per-repo results into aggregate counts.
func summarize(results []ActionResult) Summary {
	s := Summary{Results: results}
	for _, r := range results {
		switch r.Status {
		case StatusSuccess:
			s.Succeeded++
		case StatusFailed:
			s.Failed++
		case StatusSkipped:
			s.Skipped++
		}
	}
	return s
}

func fetchOne(repo model.Repo) ActionResult {
	res := ActionResult{Repo: repo.Name, Path: repo.Path}

	if !hasRemote(repo.Path) {
		res.Status = StatusSkipped
		return res
	}

	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "fetch", "--all", "--quiet")
	cmd.Dir = repo.Path
	out, err := cmd.CombinedOutput()
	if err != nil {
		res.Status = StatusFailed
		res.Err = fetchError(ctx, err, out)
		return res
	}

	res.Status = StatusSuccess
	return res
}

// hasRemote reports whether the repository has at least one configured remote.
func hasRemote(path string) bool {
	cmd := exec.Command("git", "remote")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

// fetchError produces a concise error, preferring git's own message and
// surfacing timeouts explicitly.
func fetchError(ctx context.Context, err error, out []byte) error {
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("timed out after %s", fetchTimeout)
	}
	if msg := firstLine(strings.TrimSpace(string(out))); msg != "" {
		return fmt.Errorf("%s", msg)
	}
	return err
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}
