package stats

import (
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Bharath-code/git-scope/internal/model"
)

// ContributionData holds commit counts per day for the heatmap
type ContributionData struct {
	Days         map[string]int // "2024-01-15" -> 5 commits
	TotalCommits int
	WeeksCount   int
	StartDate    time.Time
	EndDate      time.Time
	MaxDaily     int // Max commits in a single day (for scaling)
}

// GetContributions aggregates commits from all repos for the last N weeks
func GetContributions(repos []model.Repo, weeks int) (*ContributionData, error) {
	data := &ContributionData{
		Days:       make(map[string]int),
		WeeksCount: weeks,
		EndDate:    time.Now(),
		StartDate:  time.Now().AddDate(0, 0, -7*weeks),
	}

	sinceDate := data.StartDate.Format("2006-01-02")

	for _, repo := range repos {
		commits, err := getRepoCommits(repo.Path, sinceDate)
		if err != nil {
			continue // Skip repos with errors
		}

		for _, date := range commits {
			data.Days[date]++
			data.TotalCommits++
			if data.Days[date] > data.MaxDaily {
				data.MaxDaily = data.Days[date]
			}
		}
	}

	return data, nil
}

// getRepoCommits returns a list of commit dates (YYYY-MM-DD) from a repo
func getRepoCommits(repoPath, sinceDate string) ([]string, error) {
	cmd := exec.Command("git", "log", "--since="+sinceDate, "--format=%ad", "--date=short")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	dates := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			dates = append(dates, line)
		}
	}

	return dates, nil
}

// GetIntensityLevel returns 0-4 based on commit count relative to max
func (d *ContributionData) GetIntensityLevel(date string) int {
	count := d.Days[date]
	if count == 0 {
		return 0
	}
	if d.MaxDaily == 0 {
		return 1
	}

	// Scale to 1-4 based on percentage of max
	ratio := float64(count) / float64(d.MaxDaily)
	switch {
	case ratio >= 0.75:
		return 4
	case ratio >= 0.5:
		return 3
	case ratio >= 0.25:
		return 2
	default:
		return 1
	}
}

// GetDayCount returns commit count for a specific date
func (d *ContributionData) GetDayCount(date string) int {
	return d.Days[date]
}

// FormatDate formats a time.Time to the key format
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// ParseDate parses a date string
func ParseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

// GetWeeksData returns contribution data organized by weeks for rendering
// Returns a slice of weeks, each containing 7 days (Sun-Sat)
func (d *ContributionData) GetWeeksData() [][]string {
	weeks := make([][]string, 0, d.WeeksCount)

	// Find the Sunday before or on the start date
	current := d.StartDate
	for current.Weekday() != time.Sunday {
		current = current.AddDate(0, 0, -1)
	}

	for current.Before(d.EndDate) || current.Equal(d.EndDate) {
		week := make([]string, 7)
		for i := 0; i < 7; i++ {
			week[i] = FormatDate(current)
			current = current.AddDate(0, 0, 1)
		}
		weeks = append(weeks, week)
	}

	return weeks
}

// GetMonthLabels returns month labels for the heatmap header
func (d *ContributionData) GetMonthLabels() []string {
	months := make([]string, 0, 12)
	current := d.StartDate
	lastMonth := ""

	for current.Before(d.EndDate) || current.Equal(d.EndDate) {
		monthLabel := current.Format("Jan")
		if monthLabel != lastMonth {
			months = append(months, monthLabel)
			lastMonth = monthLabel
		}
		current = current.AddDate(0, 0, 7) // Skip by week
	}

	return months
}

// FormatCount returns a formatted string for display
func FormatCount(n int) string {
	if n == 0 {
		return "0"
	}
	return strconv.Itoa(n)
}
