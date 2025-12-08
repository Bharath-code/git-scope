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

	// Get branch and status info using porcelain v2 format
	cmd := exec.Command("git", "status", "--porcelain=v2", "-b")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return status, fmt.Errorf("git status: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		// Parse branch name
		if strings.HasPrefix(line, "# branch.head ") {
			status.Branch = strings.TrimPrefix(line, "# branch.head ")
		}

		// Parse ahead/behind
		if strings.HasPrefix(line, "# branch.ab ") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				aheadStr := strings.TrimPrefix(parts[2], "+")
				behindStr := strings.TrimPrefix(parts[3], "-")
				status.Ahead, _ = strconv.Atoi(aheadStr)
				status.Behind, _ = strconv.Atoi(behindStr)
			}
		}

		// Count file statuses (non-header lines)
		if len(line) > 0 && !strings.HasPrefix(line, "#") {
			// Porcelain v2 format:
			// 1 = Changed entries (staged or unstaged)
			// 2 = Renamed/copied entries
			// ? = Untracked files
			// ! = Ignored files

			if strings.HasPrefix(line, "1 ") || strings.HasPrefix(line, "2 ") {
				// Parse the XY status
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					xy := parts[1]
					if len(xy) >= 2 {
						// X = staged status, Y = unstaged status
						if xy[0] != '.' {
							status.Staged++
						}
						if xy[1] != '.' {
							status.Unstaged++
						}
					}
				}
			} else if strings.HasPrefix(line, "? ") {
				status.Untracked++
			}
		}
	}

	status.IsDirty = status.Staged > 0 || status.Unstaged > 0 || status.Untracked > 0

	// Get last commit time
	t, err := lastCommitTime(repoPath)
	if err == nil {
		status.LastCommit = t
	}

	return status, nil
}

// lastCommitTime retrieves the timestamp of the most recent commit
func lastCommitTime(repoPath string) (time.Time, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%ct")
	cmd.Dir = repoPath
	out, err := cmd.Output()
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
