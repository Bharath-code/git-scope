package stats

import (
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/Bharath-code/git-scope/internal/model"
)

// TimelineData holds timeline information
type TimelineData struct {
	Entries    []TimelineEntry
	TotalRepos int
}

// TimelineEntry represents a repo activity entry
type TimelineEntry struct {
	Name       string
	Path       string
	Branch     string
	LastCommit time.Time
	Message    string // Last commit message
	TimeAgo    string // Human-readable time ago
	DayLabel   string // "Today", "Yesterday", "2 days ago", etc.
}

// GetTimeline gets recent activity timeline sorted by last commit
func GetTimeline(repos []model.Repo) (*TimelineData, error) {
	data := &TimelineData{
		Entries:    make([]TimelineEntry, 0, len(repos)),
		TotalRepos: len(repos),
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	for _, repo := range repos {
		// Get last commit info
		lastCommit := repo.Status.LastCommit
		if lastCommit.IsZero() {
			continue
		}

		// Get commit message
		message := getLastCommitMessage(repo.Path)

		entry := TimelineEntry{
			Name:       repo.Name,
			Path:       repo.Path,
			Branch:     repo.Status.Branch,
			LastCommit: lastCommit,
			Message:    message,
			TimeAgo:    formatTimeAgo(lastCommit, now),
			DayLabel:   formatDayLabel(lastCommit, today),
		}

		data.Entries = append(data.Entries, entry)
	}

	// Sort by last commit descending (most recent first)
	sort.Slice(data.Entries, func(i, j int) bool {
		return data.Entries[i].LastCommit.After(data.Entries[j].LastCommit)
	})

	return data, nil
}

// getLastCommitMessage gets the last commit message for a repo
func getLastCommitMessage(repoPath string) string {
	cmd := exec.Command("git", "log", "-1", "--format=%s")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	msg := strings.TrimSpace(string(out))
	if len(msg) > 50 {
		msg = msg[:47] + "..."
	}
	return msg
}

// formatTimeAgo formats a time as "2 hours ago", "3 days ago", etc.
func formatTimeAgo(t time.Time, now time.Time) string {
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return formatInt(int64(mins)) + " mins ago"
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return formatInt(int64(hours)) + " hours ago"
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return formatInt(int64(days)) + " days ago"
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return formatInt(int64(weeks)) + " weeks ago"
	default:
		return t.Format("Jan 2")
	}
}

// formatDayLabel returns "Today", "Yesterday", or the day name
func formatDayLabel(t time.Time, today time.Time) string {
	tDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	diff := today.Sub(tDay)
	days := int(diff.Hours() / 24)

	switch days {
	case 0:
		return "Today"
	case 1:
		return "Yesterday"
	case 2, 3, 4, 5, 6:
		return t.Weekday().String()
	default:
		return t.Format("Jan 2")
	}
}
