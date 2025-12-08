package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/Bharath-code/git-scope/internal/config"
	"github.com/Bharath-code/git-scope/internal/model"
	"github.com/Bharath-code/git-scope/internal/scan"
)

// Run starts the Bubbletea TUI application
func Run(cfg *config.Config) error {
	m := NewModel(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// scanReposCmd is a command that scans for repositories
func scanReposCmd(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		repos, err := scan.ScanRoots(cfg.Roots, cfg.Ignore)
		if err != nil {
			return scanErrorMsg{err: err}
		}
		return scanCompleteMsg{repos: repos}
	}
}

// scanCompleteMsg is sent when scanning is complete
type scanCompleteMsg struct {
	repos []model.Repo
}

// scanErrorMsg is sent when scanning fails
type scanErrorMsg struct {
	err error
}

// openEditorMsg is sent to trigger opening an editor
type openEditorMsg struct {
	path string
}
