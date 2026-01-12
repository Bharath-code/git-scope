package gitstatus

import (
	"fmt"
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

	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}

		// header lines -> branch metadata
		if strings.HasPrefix(line, "#") {
			applyBranchHeader(&status, line)
			continue
		}

		// non-header lines -> file status records
		applyFileLine(&status, line)
	}

	status.IsDirty = status.Staged > 0 || status.Unstaged > 0 || status.Untracked > 0

	if t, err := lastCommitTime(repoPath); err == nil {
		status.LastCommit = t
	}

	return status, nil
}

// runGit is a helper that executes a git command with the given arguments
// in the specified directory and returns its stdout output
func runGit(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.Output()
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
