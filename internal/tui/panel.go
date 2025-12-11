package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/Bharath-code/git-scope/internal/stats"
)

// PanelType represents which panel is currently active
type PanelType int

const (
	PanelNone PanelType = iota
	PanelGrass
	PanelDisk
	PanelTimeline
)

// Heatmap color palette (GitHub-style green gradient)
var (
	heatmapLevel0 = lipgloss.NewStyle().Foreground(lipgloss.Color("#161b22")) // No commits (dark)
	heatmapLevel1 = lipgloss.NewStyle().Foreground(lipgloss.Color("#0e4429")) // Low
	heatmapLevel2 = lipgloss.NewStyle().Foreground(lipgloss.Color("#006d32")) // Medium-Low
	heatmapLevel3 = lipgloss.NewStyle().Foreground(lipgloss.Color("#26a641")) // Medium-High
	heatmapLevel4 = lipgloss.NewStyle().Foreground(lipgloss.Color("#39d353")) // High

	// Panel styling
	panelBorderStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7C3AED")).
				Padding(0, 1)

	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#A78BFA")).
			MarginBottom(1)

	panelMutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))
)

// renderSplitPane renders a split-pane layout with table on left and panel on right
func renderSplitPane(leftContent, rightContent string, totalWidth int) string {
	// 60% for table, 40% for panel
	leftWidth := int(float64(totalWidth) * 0.58)
	rightWidth := totalWidth - leftWidth - 3 // Account for borders/gaps

	if rightWidth < 20 {
		rightWidth = 20
		leftWidth = totalWidth - rightWidth - 3
	}

	leftPane := lipgloss.NewStyle().
		Width(leftWidth).
		Render(leftContent)

	rightPane := panelBorderStyle.
		Width(rightWidth).
		Render(rightContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, " ", rightPane)
}

// renderGrassPanel renders the contribution heatmap panel
func renderGrassPanel(data *stats.ContributionData, width, height int) string {
	if data == nil {
		return panelMutedStyle.Render("Loading contribution data...")
	}

	var b strings.Builder

	// Title
	b.WriteString(panelTitleStyle.Render("🌿 Contribution Graph"))
	b.WriteString("\n\n")

	// Month labels
	months := data.GetMonthLabels()
	b.WriteString("    ")
	for i, month := range months {
		if i > 0 {
			b.WriteString("  ")
		}
		b.WriteString(panelMutedStyle.Render(month))
	}
	b.WriteString("\n")

	// Day labels and heatmap grid
	dayLabels := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	weeks := data.GetWeeksData()

	// Limit weeks to fit in panel (show last N weeks that fit)
	maxWeeks := (width - 6) / 2 // Each cell is 2 chars wide
	if maxWeeks < 1 {
		maxWeeks = 1
	}
	if len(weeks) > maxWeeks {
		weeks = weeks[len(weeks)-maxWeeks:]
	}

	// Render each row (day of week)
	for day := 0; day < 7; day++ {
		// Only show Mon, Wed, Fri labels to save space
		if day == 1 || day == 3 || day == 5 {
			b.WriteString(panelMutedStyle.Render(dayLabels[day][:3]))
		} else {
			b.WriteString("   ")
		}
		b.WriteString(" ")

		for _, week := range weeks {
			if day < len(week) {
				dateStr := week[day]
				date, _ := stats.ParseDate(dateStr)
				
				// Don't show future dates
				if date.After(time.Now()) {
					b.WriteString("  ")
					continue
				}

				level := data.GetIntensityLevel(dateStr)
				block := getHeatmapBlock(level)
				b.WriteString(block)
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Legend
	b.WriteString(panelMutedStyle.Render("Less "))
	b.WriteString(getHeatmapBlock(0))
	b.WriteString(getHeatmapBlock(1))
	b.WriteString(getHeatmapBlock(2))
	b.WriteString(getHeatmapBlock(3))
	b.WriteString(getHeatmapBlock(4))
	b.WriteString(panelMutedStyle.Render(" More"))
	b.WriteString("\n\n")

	// Stats
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render(
		fmt.Sprintf("%d", data.TotalCommits)))
	b.WriteString(panelMutedStyle.Render(" commits in the last "))
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Render(
		fmt.Sprintf("%d", data.WeeksCount)))
	b.WriteString(panelMutedStyle.Render(" weeks"))

	return b.String()
}

// getHeatmapBlock returns a colored block for the heatmap based on intensity level
func getHeatmapBlock(level int) string {
	block := "██" // Full block character (2 chars wide for visibility)

	switch level {
	case 0:
		return heatmapLevel0.Render(block)
	case 1:
		return heatmapLevel1.Render(block)
	case 2:
		return heatmapLevel2.Render(block)
	case 3:
		return heatmapLevel3.Render(block)
	case 4:
		return heatmapLevel4.Render(block)
	default:
		return heatmapLevel0.Render(block)
	}
}

// getPanelHelp returns help text for the active panel
func getPanelHelp(panel PanelType) string {
	switch panel {
	case PanelGrass:
		return helpItem("g", "close") + " • " + helpItem("esc", "close")
	case PanelDisk:
		return helpItem("d", "close") + " • " + helpItem("esc", "close")
	case PanelTimeline:
		return helpItem("t", "close") + " • " + helpItem("esc", "close")
	default:
		return ""
	}
}

// Disk usage color palette (warm gradient for size visualization)
var (
	diskBarLow    = lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")) // Green - small
	diskBarMed    = lipgloss.NewStyle().Foreground(lipgloss.Color("#eab308")) // Yellow - medium
	diskBarHigh   = lipgloss.NewStyle().Foreground(lipgloss.Color("#f97316")) // Orange - large
	diskBarMax    = lipgloss.NewStyle().Foreground(lipgloss.Color("#ef4444")) // Red - huge
	diskNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	diskSizeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#A78BFA")).Bold(true)
	diskNodeSizeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F97316")).Bold(true) // Orange for node_modules value
	
	// Separate colors for git and node_modules
	diskBarGit  = lipgloss.NewStyle().Foreground(lipgloss.Color("#8B5CF6")) // Purple for .git
	diskBarNode = lipgloss.NewStyle().Foreground(lipgloss.Color("#F97316")) // Orange for node_modules
)

// renderDiskPanel renders the disk usage panel with bar chart
func renderDiskPanel(data *stats.DiskUsageData, width, height int) string {
	if data == nil {
		return panelMutedStyle.Render("Loading disk usage data...")
	}

	var b strings.Builder

	// Title - update based on what we're showing
	if data.HasNodeModules {
		b.WriteString(panelTitleStyle.Render("💾 Disk Usage (.git + node_modules)"))
	} else {
		b.WriteString(panelTitleStyle.Render("💾 Disk Usage (.git folders)"))
	}
	b.WriteString("\n\n")

	// Summary with breakdown
	b.WriteString(diskSizeStyle.Render(stats.FormatBytes(data.TotalSize)))
	b.WriteString(panelMutedStyle.Render(" total"))
	b.WriteString("\n")
	
	// Show breakdown if we have node_modules
	if data.HasNodeModules {
		b.WriteString(diskBarGit.Render("█"))
		b.WriteString(panelMutedStyle.Render(" .git: "))
		b.WriteString(diskSizeStyle.Render(stats.FormatBytes(data.TotalGitSize)))
		b.WriteString("  ")
		b.WriteString(diskBarNode.Render("█"))
		b.WriteString(panelMutedStyle.Render(" node_modules: "))
		b.WriteString(diskNodeSizeStyle.Render(stats.FormatBytes(data.TotalNodeSize)))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Bar chart (show top repos that fit)
	maxRows := height - 10
	if maxRows < 3 {
		maxRows = 3
	}
	if maxRows > 12 {
		maxRows = 12
	}
	
	barWidth := width - 35 // Leave more room for name and size
	if barWidth < 8 {
		barWidth = 8
	}
	if barWidth > 25 {
		barWidth = 25 // Cap max width to prevent overflow
	}

	for i, repo := range data.Repos {
		if i >= maxRows {
			remaining := len(data.Repos) - maxRows
			if remaining > 0 {
				b.WriteString(panelMutedStyle.Render(fmt.Sprintf("  ... and %d more\n", remaining)))
			}
			break
		}

		// Truncate name
		name := repo.Name
		if len(name) > 10 {
			name = name[:9] + "…"
		}
		name = fmt.Sprintf("%-10s", name)

		// Calculate bar lengths
		gitBarLen := 0
		nodeBarLen := 0
		
		if data.MaxSize > 0 {
			gitBarLen = int(float64(repo.GitSize) / float64(data.MaxSize) * float64(barWidth))
			nodeBarLen = int(float64(repo.NodeModulesSize) / float64(data.MaxSize) * float64(barWidth))
		}
		
		if gitBarLen < 1 && repo.GitSize > 0 {
			gitBarLen = 1
		}
		if nodeBarLen < 1 && repo.NodeModulesSize > 0 {
			nodeBarLen = 1
		}

		// Create stacked bar (git + node_modules)
		gitBar := strings.Repeat("█", gitBarLen)
		nodeBar := strings.Repeat("█", nodeBarLen)

		b.WriteString(diskNameStyle.Render(name))
		b.WriteString(" ")
		b.WriteString(diskBarGit.Render(gitBar))
		b.WriteString(diskBarNode.Render(nodeBar))
		b.WriteString(" ")
		b.WriteString(panelMutedStyle.Render(stats.FormatBytes(repo.TotalSize)))
		b.WriteString("\n")
	}

	// Legend
	b.WriteString("\n")
	b.WriteString(diskBarGit.Render("█"))
	b.WriteString(panelMutedStyle.Render(" .git "))
	if data.HasNodeModules {
		b.WriteString(diskBarNode.Render("█"))
		b.WriteString(panelMutedStyle.Render(" node_modules"))
	}

	return b.String()
}

// Timeline styling
var (
	timelineTodayStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Bold(true) // Green
	timelineYesterdayStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#eab308")).Bold(true) // Yellow
	timelineOlderStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))            // Gray
	timelineRepoStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	timelineBranchStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#A78BFA"))
	timelineMessageStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Italic(true)
	timelineTimeStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
)

// renderTimelinePanel renders the activity timeline panel
func renderTimelinePanel(data *stats.TimelineData, width, height int) string {
	if data == nil {
		return panelMutedStyle.Render("Loading timeline...")
	}

	var b strings.Builder

	// Title
	b.WriteString(panelTitleStyle.Render("⏰ Recent Activity"))
	b.WriteString("\n\n")

	if len(data.Entries) == 0 {
		b.WriteString(panelMutedStyle.Render("No recent commits found."))
		return b.String()
	}

	// Show entries grouped by day
	maxRows := height - 6
	if maxRows < 5 {
		maxRows = 5
	}

	currentDayLabel := ""
	rowCount := 0

	for _, entry := range data.Entries {
		if rowCount >= maxRows {
			remaining := len(data.Entries) - rowCount
			if remaining > 0 {
				b.WriteString(panelMutedStyle.Render(fmt.Sprintf("\n  ... and %d more\n", remaining)))
			}
			break
		}

		// Day header
		if entry.DayLabel != currentDayLabel {
			if currentDayLabel != "" {
				b.WriteString("\n")
			}
			
			var dayStyle lipgloss.Style
			switch entry.DayLabel {
			case "Today":
				dayStyle = timelineTodayStyle
			case "Yesterday":
				dayStyle = timelineYesterdayStyle
			default:
				dayStyle = timelineOlderStyle
			}
			
			b.WriteString(dayStyle.Render("● " + entry.DayLabel))
			b.WriteString("\n")
			currentDayLabel = entry.DayLabel
			rowCount++
		}

		// Entry
		name := entry.Name
		if len(name) > 15 {
			name = name[:14] + "…"
		}

		b.WriteString("  ")
		b.WriteString(timelineRepoStyle.Render(name))
		b.WriteString(" ")
		
		branch := entry.Branch
		if len(branch) > 10 {
			branch = branch[:9] + "…"
		}
		b.WriteString(timelineBranchStyle.Render("(" + branch + ")"))
		b.WriteString("\n")

		// Commit message
		if entry.Message != "" {
			msg := entry.Message
			maxMsgLen := width - 8
			if maxMsgLen < 20 {
				maxMsgLen = 20
			}
			if len(msg) > maxMsgLen {
				msg = msg[:maxMsgLen-3] + "..."
			}
			b.WriteString("    ")
			b.WriteString(timelineMessageStyle.Render("\"" + msg + "\""))
			b.WriteString("\n")
			rowCount++
		}

		// Time ago
		b.WriteString("    ")
		b.WriteString(timelineTimeStyle.Render(entry.TimeAgo))
		b.WriteString("\n")

		rowCount += 2
	}

	return b.String()
}
