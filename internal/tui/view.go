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
	case StateWorkspaceSwitch:
		b.WriteString(m.renderWorkspaceModal())
	}

	return b.String()
}

func (m Model) renderLoading() string {
	var b strings.Builder

	b.WriteString(compactLogo())
	b.WriteString("  ")
	b.WriteString(m.spinner.View())
	b.WriteString(" ")
	b.WriteString(loadingStyle.Render("Scanning repositories..."))
	b.WriteString("\n\n")

	b.WriteString(subtitleStyle.Render("Searching for git repos in:"))
	b.WriteString("\n")
	
	// Show workspace path if switching, otherwise show config roots
	if m.activeWorkspace != "" {
		b.WriteString(pathBulletStyle.Render("  → "))
		b.WriteString(pathStyle.Render(m.activeWorkspace))
		b.WriteString("\n")
	} else {
		for _, root := range m.cfg.Roots {
			b.WriteString(pathBulletStyle.Render("  → "))
			b.WriteString(pathStyle.Render(root))
			b.WriteString("\n")
		}
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
	b.WriteString("\n\n")

	errContent := ""
	if m.err != nil {
		errContent = m.err.Error()
	} else {
		errContent = "Unknown error occurred"
	}
	b.WriteString(errorBoxStyle.Render(errContent))
	b.WriteString("\n\n")

	// Actionable suggestions
	b.WriteString(subtitleStyle.Render("💡 Suggestions:"))
	b.WriteString("\n")
	b.WriteString(pathBulletStyle.Render("  → "))
	b.WriteString(pathStyle.Render("Check your config at ~/.config/git-scope/config.yml"))
	b.WriteString("\n")
	b.WriteString(pathBulletStyle.Render("  → "))
	b.WriteString(pathStyle.Render("Run 'git-scope init' to reconfigure"))
	b.WriteString("\n")
	b.WriteString(pathBulletStyle.Render("  → "))
	b.WriteString(pathStyle.Render("Make sure git is installed and in PATH"))
	b.WriteString("\n\n")

	b.WriteString(helpItem("r", "retry"))
	b.WriteString("  •  ")
	b.WriteString(helpItem("q", "quit"))

	return b.String()
}

func (m Model) renderDashboard() string {
	var b strings.Builder

	// Header with logo on its own line
	logo := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A78BFA")).Render("git-scope")
	version := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render(" v1.3.0")
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

	// Main content area - split pane if panel is active
	if m.activePanel != PanelNone {
		// Render table content
		tableContent := m.table.View()
		
		// Render panel content based on active panel
		var panelContent string
		switch m.activePanel {
		case PanelGrass:
			panelContent = renderGrassPanel(m.grassData, m.width/2, m.height-15)
		case PanelDisk:
			panelContent = renderDiskPanel(m.diskData, m.width/2, m.height-15)
		case PanelTimeline:
			panelContent = renderTimelinePanel(m.timelineData, m.width/2, m.height-15)
		}
		
		b.WriteString(renderSplitPane(tableContent, panelContent, m.width-4))
	} else {
		// Full-width table
		b.WriteString(m.table.View())
	}
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

	// Filter indicator with inline hint
	if m.filterMode != FilterAll {
		filterBadge := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#60A5FA")).
			Padding(0, 1).
			Bold(true).
			Render("⚡ " + m.GetFilterModeName())
		filterHint := hintStyle.Render(" (f)")
		stats = append(stats, filterBadge+filterHint)
	}

	// Sort indicator with inline hint
	sortBadge := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7C3AED")).
		Padding(0, 1).
		Render("⇅ " + m.GetSortModeName())
	sortHint := hintStyle.Render(" (s)")
	stats = append(stats, sortBadge+sortHint)

	return lipgloss.JoinHorizontal(lipgloss.Center, stats...)
}

// renderLegend renders a compact single-line legend (Tuimorphic style)
func (m Model) renderLegend() string {
	dirty := dirtyDotStyle.Render("●") + legendStyle.Render(" dirty")
	clean := cleanDotStyle.Render("○") + legendStyle.Render(" clean")
	editor := legendStyle.Render(fmt.Sprintf("  Editor: %s", m.cfg.Editor))

	return legendStyle.Render(dirty + "  " + clean + editor)
}

// renderHelp renders a Tuimorphic keybindings bar with box-drawing separators
func (m Model) renderHelp() string {
	sep := keyBindingSepStyle.Render(" │ ")
	var items []string

	if m.state == StateSearching {
		// Search mode help
		items = []string{
			keyBinding("type", "search"),
			keyBinding("enter", "apply"),
			keyBinding("esc", "cancel"),
		}
	} else if m.state == StateWorkspaceSwitch {
		// Workspace switch mode help
		items = []string{
			keyBinding("type", "path"),
			keyBinding("tab", "complete"),
			keyBinding("enter", "switch"),
			keyBinding("esc", "cancel"),
		}
	} else if m.activePanel != PanelNone {
		// Panel active help
		items = []string{
			keyBinding("↑↓", "nav"),
			keyBinding("esc", "close"),
			keyBinding("g", "grass"),
			keyBinding("d", "disk"),
			keyBinding("t", "time"),
			keyBinding("q", "quit"),
		}
	} else {
		// Normal mode help - Tuimorphic style
		items = []string{
			keyBinding("↑↓", "nav"),
			keyBinding("enter", "open"),
			keyBinding("/", "search"),
			keyBinding("a", "workspace"),
			keyBinding("f", "filter"),
			keyBinding("s", "sort"),
			keyBinding("g", "grass"),
			keyBinding("d", "disk"),
			keyBinding("t", "time"),
			keyBinding("r", "rescan"),
			keyBinding("q", "quit"),
		}
	}

	return keyBindingsBarStyle.Render(strings.Join(items, sep))
}

// keyBinding creates a styled key-action pair for the keybindings bar
func keyBinding(key, action string) string {
	return keyBindingKeyStyle.Render(key) + " " + action
}

// renderWorkspaceModal renders the workspace switch modal
func (m Model) renderWorkspaceModal() string {
	var b strings.Builder

	// Header with logo
	b.WriteString(compactLogo())
	b.WriteString("\n\n")

	// Modal box
	modalStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Padding(1, 2).
		Width(50)

	// Modal title
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A78BFA")).
		Bold(true).
		Render("📁 Switch Workspace")

	// Path input
	label := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true).
		Render("Path: ")

	// Error message if any
	errorLine := ""
	if m.workspaceError != "" {
		errorLine = "\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Render("❌ " + m.workspaceError)
	}

	// Footer hints
	footer := lipgloss.NewStyle().
		Foreground(mutedColor).
		Render("\n\nTab = complete   Enter = scan   Esc = cancel")

	modalContent := title + "\n\n" + label + m.workspaceInput.View() + errorLine + footer
	b.WriteString(modalStyle.Render(modalContent))

	// Help bar
	b.WriteString("\n\n")
	b.WriteString(m.renderHelp())

	return b.String()
}
