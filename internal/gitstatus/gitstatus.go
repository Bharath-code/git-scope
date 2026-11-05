package gitstatus

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Bharath-code/git-scope/internal/model"
)

// Status retrieves the git status for a repository at the given path
func Status(repoPath string) (model.RepoStatus, error) {
	status := model.RepoStatus{}

	out, err := runGit(repoPath, "status", "--porcelain=v2", "-b")
	if err != nil {
		return status, fmt.Errorf("git status: %w", err)
	}

	hasUpstream, unborn := false, false
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "# branch.upstream ") {
			hasUpstream = true
		}
		if line == "# branch.oid (initial)" {
			unborn = true
		}

		// header lines -> branch metadata
		if strings.HasPrefix(line, "#") {
			applyBranchHeader(&status, line)
			continue
		}

		// non-header lines -> file status records
		applyFileLine(&status, line)
	}

	hiddenErr := applyHiddenWork(repoPath, &status, hasUpstream, unborn)

	status.IsDirty = status.Staged > 0 || status.Unstaged > 0 || status.Untracked > 0 ||
		status.Ahead > 0 || status.Behind > 0 || status.Stashes > 0 || status.OtherBranchCommits > 0

	if t, err := lastCommitTime(repoPath); err == nil {
		status.LastCommit = t
	}

	if hiddenErr != nil {
		return status, hiddenErr
	}
	return status, nil
}

// safeArgs neutralise repo config that would make a read turn into a write or
// an exec: core.fsmonitor names a program git runs during status,
// log.showSignature makes log run gpg.program, and --no-optional-locks stops
// status from rewriting .git/index.
var safeArgs = []string{"-c", "core.fsmonitor=false", "-c", "log.showSignature=false", "--no-optional-locks"}

// Command builds a git command that is safe to run inside repositories
// git-scope did not create. Every git invocation should go through it.
func Command(ctx context.Context, dir string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "git", append(append([]string{}, safeArgs...), args...)...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	return cmd
}

// gitTimeout bounds each read so one hung repo (network mount, stuck lock)
// cannot stall a whole scan or a long-running MCP server.
const gitTimeout = 10 * time.Second

// runGit runs a git command in dir and returns its stdout.
func runGit(dir string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()
	return Command(ctx, dir, args...).Output()
}

// applyHiddenWork fills in the work `git status` doesn't show. A failure is
// returned, never counted as 0: "couldn't tell" must not read as "nothing at risk".
func applyHiddenWork(repoPath string, status *model.RepoStatus, hasUpstream, unborn bool) error {
	stashes, err := runGit(repoPath, "stash", "list", "--format=%gd")
	if err != nil {
		return fmt.Errorf("git stash list: %w", err)
	}
	status.Stashes = len(strings.Fields(string(stashes)))

	remotes, err := runGit(repoPath, "remote")
	if err != nil {
		return fmt.Errorf("git remote: %w", err)
	}
	// With nowhere to push, "unpushed" would flag every local-only project.
	if len(strings.TrimSpace(string(remotes))) == 0 {
		return nil
	}

	// Never-pushed branch or detached HEAD: no `branch.ab` line to read.
	if !hasUpstream && !unborn && status.Branch != "" {
		if status.Ahead, err = countRevs(repoPath, "HEAD", "--not", "--remotes"); err != nil {
			return err
		}
	}

	args := []string{"--branches", "--not", "--remotes"}
	if !unborn {
		// HEAD's unpushed commits are already in Ahead.
		args = append(args, "HEAD")
	}
	status.OtherBranchCommits, err = countRevs(repoPath, args...)
	return err
}

func countRevs(repoPath string, args ...string) (int, error) {
	out, err := runGit(repoPath, append([]string{"rev-list", "--count"}, args...)...)
	if err != nil {
		return 0, fmt.Errorf("git rev-list: %w", err)
	}
	return strconv.Atoi(strings.TrimSpace(string(out)))
}

// applyBranchHeader parses porcelain v2 branch metadata lines and updates
// the repository status with branch name and ahead/behind information
func applyBranchHeader(status *model.RepoStatus, line string) {
	if strings.HasPrefix(line, "# branch.head ") {
		status.Branch = strings.TrimPrefix(line, "# branch.head ")
		return
	}

	if strings.HasPrefix(line, "# branch.ab ") {
		ahead, behind, ok := parseAheadBehind(line)
		if ok {
			status.Ahead = ahead
			status.Behind = behind
		}
		return
	}
}

// parseAheadBehind extracts ahead/behind commit counts from a
// `# branch.ab +N -M` porcelain v2 header lien.
// It returns ok = false if the line cannot be parsed.
func parseAheadBehind(line string) (ahead int, behind int, ok bool) {
	parts := strings.Fields(line)
	// Expected: ["#", "branch.ab", "+N", "-M"]
	if len(parts) < 4 {
		return 0, 0, false
	}

	aheadStr := strings.TrimPrefix(parts[2], "+")
	behindStr := strings.TrimPrefix(parts[3], "-")

	a, err1 := strconv.Atoi(aheadStr)
	b, err2 := strconv.Atoi(behindStr)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}

	return a, b, true
}

func applyFileLine(status *model.RepoStatus, line string) {
	// Porcelain v2 format:
	// 1 = Changed entries (staged or unstaged)
	// 2 = Renamed/copied entries
	// ? = Untracked files
	// ! = Ignored files

	switch {
	case strings.HasPrefix(line, "1 "), strings.HasPrefix(line, "2 "):
		staged, unstaged := parseXY(line)
		if staged {
			status.Staged++
		}
		if unstaged {
			status.Unstaged++
		}

	// u = unmerged (conflict) entry; count as unstaged work so it is never CLEAN
	case strings.HasPrefix(line, "u "):
		status.Unstaged++

	case strings.HasPrefix(line, "? "):
		status.Untracked++
	}
}

// parseXY extracts staged (X) and unstaged (Y) change indicators from a
// porcelain v2 file status line and reports whether each side is dirty
func parseXY(line string) (staged bool, unstaged bool) {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return false, false
	}

	xy := parts[1]
	if len(xy) < 2 {
		return false, false
	}

	// X = staged status, Y = unstaged status. '.' means clean.
	return xy[0] != '.', xy[1] != '.'
}

// lastCommitTime retrieves the timestamp of the most recent commit
func lastCommitTime(repoPath string) (time.Time, error) {
	out, err := runGit(repoPath, "log", "-1", "--format=%ct")
	if err != nil {
		return time.Time{}, fmt.Errorf("git log: %w", err)
	}

	ts := strings.TrimSpace(string(out))
	if ts == "" {
		return time.Time{}, fmt.Errorf("no commits found")
	}

	sec, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse timestamp: %w", err)
	}

	return time.Unix(sec, 0), nil
}
