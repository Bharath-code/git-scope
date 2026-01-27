package tui

import (
	"time"

	"github.com/Bharath-code/git-scope/internal/cache"
	"github.com/Bharath-code/git-scope/internal/config"
	"github.com/Bharath-code/git-scope/internal/model"
	"github.com/Bharath-code/git-scope/internal/scan"
	tea "github.com/charmbracelet/bubbletea"
)

// Cache max age - use cached data if less than 5 minutes old
const cacheMaxAge = 5 * time.Minute

// Run starts the Bubbletea TUI application
func Run(cfg *config.Config) error {
	m := NewModel(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// scanReposCmd is a command that scans for repositories
// If forceRefresh is true, bypass cache and scan fresh
func scanReposCmd(cfg *config.Config, forceRefresh bool) tea.Cmd {
	return func() tea.Msg {
		cacheStore := cache.NewFileStore()

		// Try to load from cache first (unless forcing refresh)
		if !forceRefresh {
			cached, err := cacheStore.Load()
			if err == nil && cacheStore.IsValid(cacheMaxAge) && cacheStore.IsSameRoots(cfg.Roots) {
				return scanCompleteMsg{
					repos:     cached.Repos,
					fromCache: true,
				}
			}
		}

		// Scan fresh
		repos, err := scan.ScanRoots(cfg.Roots, cfg.Ignore)
		if err != nil {
			return scanErrorMsg{err: err}
		}

		// Save to cache
		_ = cacheStore.Save(repos, cfg.Roots)

		return scanCompleteMsg{
			repos:     repos,
			fromCache: false,
		}
	}
}

// scanCompleteMsg is sent when scanning is complete
type scanCompleteMsg struct {
	repos     []model.Repo
	fromCache bool
}

// scanErrorMsg is sent when scanning fails
type scanErrorMsg struct {
	err error
}

// openEditorMsg is sent to trigger opening an editor
type openEditorMsg struct {
	path string
}
