package tui

import (
	"github.com/Bharath-code/git-scope/internal/cache"
	"github.com/Bharath-code/git-scope/internal/config"
	"github.com/Bharath-code/git-scope/internal/gitops"
	"github.com/Bharath-code/git-scope/internal/gitstatus"
	"github.com/Bharath-code/git-scope/internal/model"
	tea "github.com/charmbracelet/bubbletea"
)

// fetchCompleteMsg is sent when a bulk fetch finishes. It carries both the
// outcome summary and repos with refreshed status so ahead/behind counts update.
type fetchCompleteMsg struct {
	summary gitops.Summary
	repos   []model.Repo
}

// fetchAllCmd fetches every repo's remote, then refreshes their git status so
// the dashboard's ahead/behind counts reflect what was just fetched.
func fetchAllCmd(cfg *config.Config, repos []model.Repo) tea.Cmd {
	return func() tea.Msg {
		summary := gitops.FetchAll(repos)

		// Re-read status for each repo so ahead/behind reflect the new refs.
		refreshed := make([]model.Repo, len(repos))
		copy(refreshed, repos)
		for i := range refreshed {
			if st, err := gitstatus.Status(refreshed[i].Path); err == nil {
				refreshed[i].Status = st
			}
		}

		// Keep the cache consistent with the refreshed statuses.
		_ = cache.NewFileStore().Save(refreshed, cfg.Roots)

		return fetchCompleteMsg{summary: summary, repos: refreshed}
	}
}
