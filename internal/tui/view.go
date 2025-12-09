package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the TUI
func (m Model) View() string {
	content := m.renderContent()
	return appStyle.Render(content)
}

func (m Model) renderContent() string {
	var b strings.Builder

	switch m.state {
	case StateLoading:
		b.WriteString(m.renderLoading())
	case StateError:
		b.WriteString(m.renderError())
	case StateReady, StateSearching:
		b.WriteString(m.renderDashboard())
	}

	return b.String()
}

func (m Model) renderLoading() string {
	var b strings.Builder

	b.WriteString(compactLogo())
	b.WriteString("  ")
	b.WriteString(loadingStyle.Render("⏳ Scanning repositories..."))
	b.WriteString("\n\n")

	b.WriteString(subtitleStyle.Render("Searching for git repos in:"))
	b.WriteString("\n")
	for _, root := range m.cfg.Roots {
		b.WriteString(pathBulletStyle.Render("  → "))
		b.WriteString(pathStyle.Render(root))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	b.WriteString(helpStyle.Render("Press " + helpKeyStyle.Render("q") + " to quit"))

	return b.String()
}

func (m Model) renderError() string {
	var b strings.Builder

	b.WriteString(compactLogo())
	b.WriteString("  ")
	b.WriteString(errorTitleStyle.Render("✗ Error"))
	b.WriteString("\n")

	errContent := ""
	if m.err != nil {
		errContent = m.err.Error()
	} else {
		errContent = "Unknown error occurred"
	}
	b.WriteString(errorBoxStyle.Render(errContent))
	b.WriteString("\n\n")

	b.WriteString(helpItem("q", "quit"))
	b.WriteString("  •  ")
	b.WriteString(helpItem("r", "retry"))

	return b.String()
}

func (m Model) renderDashboard() string {
	var b strings.Builder

	// Header with logo on its own line
	logo := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A78BFA")).Render("git-scope")
	version := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render(" v0.3.1")
	b.WriteString(logo + version)
	b.WriteString("\n\n")

	// Stats bar (always show first for consistent layout)
	b.WriteString(m.renderStats())
	b.WriteString("\n")

	// Search bar (show when searching or has active search)
	if m.state == StateSearching {
		b.WriteString(m.renderSearchBar())
		b.WriteString("\n")
	} else if m.searchQuery != "" {
		// Show search badge only if searchQuery is actually set
		b.WriteString(m.renderSearchBadge())
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Table
	b.WriteString(m.table.View())
	b.WriteString("\n")

	// Status message if any
	if m.statusMsg != "" {
		b.WriteString(statusStyle.Render("→ " + m.statusMsg))
		b.WriteString("\n")
	}

	// Legend
	b.WriteString(m.renderLegend())
	b.WriteString("\n")

	// Help footer
	b.WriteString(m.renderHelp())

	return b.String()
}

func (m Model) renderSearchBar() string {
	searchStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Padding(0, 1)

	// Show active search input
	label := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true).
		Render("🔍 Search: ")
	return searchStyle.Render(label + m.textInput.View())
}

func (m Model) renderSearchBadge() string {
	// Guard: don't render empty badge
	if m.searchQuery == "" {
		return ""
	}
	
	// Show current search query as badge
	searchBadge := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7C3AED")).
		Padding(0, 1).
		Render("🔍 " + m.searchQuery)
	
	clearHint := lipgloss.NewStyle().
		Foreground(mutedColor).
		Render(" (press c to clear)")
	
	return searchBadge + clearHint
}

func (m Model) renderStats() string {
	total := len(m.repos)
	shown := len(m.sortedRepos)
	dirty := 0
	clean := 0
	for _, r := range m.repos {
		if r.Status.IsDirty {
			dirty++
		} else {
			clean++
		}
	}

	stats := []string{}
	
	// Show count with filter info
	if shown == total {
		stats = append(stats, statsBadgeStyle.Render(fmt.Sprintf("📁 %d repos", total)))
	} else {
		stats = append(stats, statsBadgeStyle.Render(fmt.Sprintf("📁 %d/%d repos", shown, total)))
	}
	
	if dirty > 0 {
		stats = append(stats, dirtyBadgeStyle.Render(fmt.Sprintf("● %d dirty", dirty)))
	}
	if clean > 0 {
		stats = append(stats, cleanBadgeStyle.Render(fmt.Sprintf("✓ %d clean", clean)))
	}
	
	// Filter indicator
	if m.filterMode != FilterAll {
		filterBadge := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#60A5FA")).
			Padding(0, 1).
			Bold(true).
			Render("⚡ " + m.GetFilterModeName())
		stats = append(stats, filterBadge)
	}
	
	// Sort indicator
	sortBadge := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7C3AED")).
		Padding(0, 1).
		Render("⇅ " + m.GetSortModeName())
	stats = append(stats, sortBadge)

	return lipgloss.JoinHorizontal(lipgloss.Center, stats...)
}

func (m Model) renderLegend() string {
	legend := lipgloss.NewStyle().
		Foreground(mutedColor).
		MarginTop(1)
	
	dirtyLegend := lipgloss.NewStyle().
		Foreground(dirtyColor).
		Bold(true).
		Render("● Dirty")
	
	cleanLegend := lipgloss.NewStyle().
		Foreground(cleanColor).
		Bold(true).
		Render("✓ Clean")
	
	editorInfo := lipgloss.NewStyle().
		Foreground(mutedColor).
		Render(fmt.Sprintf("Editor: %s", m.cfg.Editor))

	return legend.Render(
		dirtyLegend + "  " + cleanLegend + "     " + editorInfo,
	)
}

func (m Model) renderHelp() string {
	var items []string

	if m.state == StateSearching {
		// Search mode help
		items = append(items, helpItem("type", "search"))
		items = append(items, helpItem("enter", "apply"))
		items = append(items, helpItem("esc", "cancel"))
	} else {
		// Normal mode help
		items = append(items, helpItem("↑↓", "nav"))
		items = append(items, helpItem("enter", "open"))
		items = append(items, helpItem("/", "search"))
		items = append(items, helpItem("f", "filter"))
		items = append(items, helpItem("s", "sort"))
		items = append(items, helpItem("c", "clear"))
		items = append(items, helpItem("r", "rescan"))
		items = append(items, helpItem("q", "quit"))
	}

	return helpStyle.Render(strings.Join(items, " • "))
}
