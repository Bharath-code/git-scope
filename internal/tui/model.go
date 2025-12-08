package tui

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/Bharath-code/git-scope/internal/config"
	"github.com/Bharath-code/git-scope/internal/model"
)

// State represents the current UI state
type State int

const (
	StateLoading State = iota
	StateReady
	StateError
)

// Model is the Bubbletea model for the TUI
type Model struct {
	cfg       *config.Config
	table     table.Model
	repos     []model.Repo
	state     State
	err       error
	statusMsg string
	width     int
	height    int
}

// NewModel creates a new TUI model
func NewModel(cfg *config.Config) Model {
	columns := []table.Column{
		{Title: "Repo", Width: 20},
		{Title: "Path", Width: 35},
		{Title: "Branch", Width: 12},
		{Title: "Stg", Width: 4},
		{Title: "Unst", Width: 5},
		{Title: "Untrk", Width: 5},
		{Title: "Last Commit", Width: 19},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	// Apply styles
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)
	t.SetStyles(s)

	return Model{
		cfg:   cfg,
		table: t,
		state: StateLoading,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return scanReposCmd(m.cfg)
}

// reposToRows converts repos to table rows
func reposToRows(repos []model.Repo) []table.Row {
	// Sort by dirty first, then by name
	sorted := make([]model.Repo, len(repos))
	copy(sorted, repos)
	sort.Slice(sorted, func(i, j int) bool {
		// Dirty repos first
		if sorted[i].Status.IsDirty != sorted[j].Status.IsDirty {
			return sorted[i].Status.IsDirty
		}
		// Then by name
		return sorted[i].Name < sorted[j].Name
	})

	rows := make([]table.Row, 0, len(sorted))
	for _, r := range sorted {
		lastCommit := "N/A"
		if !r.Status.LastCommit.IsZero() {
			lastCommit = r.Status.LastCommit.Format("2006-01-02 15:04")
		}

		rows = append(rows, table.Row{
			r.Name,
			truncatePath(r.Path, 35),
			r.Status.Branch,
			fmt.Sprintf("%d", r.Status.Staged),
			fmt.Sprintf("%d", r.Status.Unstaged),
			fmt.Sprintf("%d", r.Status.Untracked),
			lastCommit,
		})
	}
	return rows
}

// truncatePath shortens a path to fit in the given width
func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "..." + path[len(path)-maxLen+3:]
}
