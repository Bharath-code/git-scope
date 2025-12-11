package tui

import (
	"fmt"
	"os/exec"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/Bharath-code/git-scope/internal/model"
	"github.com/Bharath-code/git-scope/internal/stats"
)

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeTable()

	case scanCompleteMsg:
		m.repos = msg.repos
		m.state = StateReady
		m.updateTable()
		
		// Show helpful message if no repos found
		if len(msg.repos) == 0 {
			m.statusMsg = fmt.Sprintf("⚠️  No git repos found in configured directories. Press 'r' to rescan or run 'git-scope init' to configure.")
		} else if msg.fromCache {
			m.statusMsg = fmt.Sprintf("✓ Loaded %d repos from cache", len(msg.repos))
		} else {
			m.statusMsg = fmt.Sprintf("✓ Found %d repos", len(msg.repos))
		}
		return m, nil

	case scanErrorMsg:
		m.state = StateError
		m.err = msg.err
		return m, nil

	case openEditorMsg:
		// Check if editor exists in PATH before trying to launch
		_, err := exec.LookPath(m.cfg.Editor)
		if err != nil {
			m.statusMsg = fmt.Sprintf("❌ Editor '%s' not found. Press 'e' to change editor or install it first.", m.cfg.Editor)
			return m, nil
		}
		
		c := exec.Command(m.cfg.Editor, msg.path)
		return m, tea.ExecProcess(c, func(err error) tea.Msg {
			if err != nil {
				return editorClosedMsg{err: err}
			}
			return editorClosedMsg{}
		})

	case editorClosedMsg:
		if msg.err != nil {
			m.statusMsg = "Error: " + msg.err.Error()
		} else {
			m.statusMsg = ""
		}
		return m, scanReposCmd(m.cfg)

	case grassDataLoadedMsg:
		m.grassData = msg.data
		if msg.data != nil {
			m.statusMsg = fmt.Sprintf("🌿 %d commits in %d weeks", msg.data.TotalCommits, msg.data.WeeksCount)
		}
		return m, nil

	case diskDataLoadedMsg:
		m.diskData = msg.data
		if msg.data != nil {
			m.statusMsg = fmt.Sprintf("💾 %s total across %d repos", stats.FormatBytes(msg.data.TotalSize), msg.data.RepoCount)
		}
		return m, nil

	case timelineDataLoadedMsg:
		m.timelineData = msg.data
		if msg.data != nil {
			m.statusMsg = fmt.Sprintf("⏰ %d repos with recent activity", len(msg.data.Entries))
		}
		return m, nil

	case tea.KeyMsg:
		// Handle search mode separately
		if m.state == StateSearching {
			return m.handleSearchMode(msg)
		}
		
		// Normal mode key handling
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "/":
			// Enter search mode
			if m.state == StateReady {
				m.state = StateSearching
				m.resizeTable()
				m.textInput.Focus()
				m.textInput.SetValue(m.searchQuery)
				return m, textinput.Blink
			}

		case "enter":
			if m.state == StateReady {
				repo := m.GetSelectedRepo()
				if repo != nil {
					m.statusMsg = "Opening " + repo.Name + " in " + m.cfg.Editor + "..."
					return m, func() tea.Msg {
						return openEditorMsg{path: repo.Path}
					}
				}
			}

		case "r":
			m.state = StateLoading
			m.statusMsg = "Rescanning..."
			return m, scanReposCmd(m.cfg)

		case "f":
			// Cycle through filter modes
			if m.state == StateReady {
				m.filterMode = (m.filterMode + 1) % 3
				m.updateTable()
				m.statusMsg = "Filter: " + m.GetFilterModeName()
				return m, nil
			}

		case "s":
			if m.state == StateReady {
				m.sortMode = (m.sortMode + 1) % 4
				m.updateTable()
				m.statusMsg = "Sorted by: " + m.GetSortModeName()
				return m, nil
			}

		case "1":
			if m.state == StateReady {
				m.sortMode = SortByDirty
				m.updateTable()
				m.statusMsg = "Sorted by: Dirty First"
				return m, nil
			}

		case "2":
			if m.state == StateReady {
				m.sortMode = SortByName
				m.updateTable()
				m.statusMsg = "Sorted by: Name"
				return m, nil
			}

		case "3":
			if m.state == StateReady {
				m.sortMode = SortByBranch
				m.updateTable()
				m.statusMsg = "Sorted by: Branch"
				return m, nil
			}

		case "4":
			if m.state == StateReady {
				m.sortMode = SortByLastCommit
				m.updateTable()
				m.statusMsg = "Sorted by: Recent"
				return m, nil
			}

		case "c":
			// Clear search and filters
			if m.state == StateReady {
				m.searchQuery = ""
				m.textInput.SetValue("") // Also reset the text input
				m.filterMode = FilterAll
				m.resizeTable()
				m.updateTable()
				m.statusMsg = "Filters cleared"
				return m, nil
			}

		case "e":
			if m.state == StateReady {
				// Check if editor exists
				_, err := exec.LookPath(m.cfg.Editor)
				if err != nil {
					m.statusMsg = fmt.Sprintf("❌ Editor '%s' not found in PATH. Install it or edit ~/.config/git-scope/config.yml", m.cfg.Editor)
				} else {
					m.statusMsg = fmt.Sprintf("✓ Editor: %s (edit config at ~/.config/git-scope/config.yml)", m.cfg.Editor)
				}
				return m, nil
			}

		case "g":
			// Toggle grass panel
			if m.state == StateReady {
				if m.activePanel == PanelGrass {
					m.activePanel = PanelNone
					m.statusMsg = ""
				} else {
					m.activePanel = PanelGrass
					m.statusMsg = "🌿 Loading contribution graph..."
					return m, loadGrassDataCmd(m.repos)
				}
				return m, nil
			}

		case "d":
			// Toggle disk usage panel
			if m.state == StateReady {
				if m.activePanel == PanelDisk {
					m.activePanel = PanelNone
					m.statusMsg = ""
				} else {
					m.activePanel = PanelDisk
					m.statusMsg = "💾 Calculating disk usage..."
					return m, loadDiskDataCmd(m.repos)
				}
				return m, nil
			}

		case "t":
			// Toggle timeline panel
			if m.state == StateReady {
				if m.activePanel == PanelTimeline {
					m.activePanel = PanelNone
					m.statusMsg = ""
				} else {
					m.activePanel = PanelTimeline
					m.statusMsg = "⏰ Loading timeline..."
					return m, loadTimelineDataCmd(m.repos)
				}
				return m, nil
			}

		case "esc":
			// Close panel if open
			if m.activePanel != PanelNone {
				m.activePanel = PanelNone
				m.statusMsg = ""
				return m, nil
			}
		}
	}

	// Update the table
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// handleSearchMode handles key events when in search mode
func (m Model) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel search, keep previous query
		m.state = StateReady
		m.resizeTable()
		m.textInput.Blur()
		return m, nil
		
	case "enter":
		// Apply search
		m.searchQuery = m.textInput.Value()
		m.state = StateReady
		m.resizeTable()
		m.textInput.Blur()
		m.updateTable()
		if m.searchQuery != "" {
			m.statusMsg = "Searching: " + m.searchQuery
		} else {
			m.statusMsg = "Search cleared"
		}
		return m, nil
		
	case "ctrl+c":
		return m, tea.Quit
	}
	
	// Update text input
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	
	// Live search as you type
	m.searchQuery = m.textInput.Value()
	m.updateTable()
	
	return m, cmd
}

// editorClosedMsg is sent when the editor process closes
type editorClosedMsg struct {
	err error
}

// grassDataLoadedMsg is sent when contribution data is loaded
type grassDataLoadedMsg struct {
	data *stats.ContributionData
}

// loadGrassDataCmd loads contribution data from all repos
func loadGrassDataCmd(repos []model.Repo) tea.Cmd {
	return func() tea.Msg {
		data, _ := stats.GetContributions(repos, 12) // Last 12 weeks
		return grassDataLoadedMsg{data: data}
	}
}

// diskDataLoadedMsg is sent when disk usage data is loaded
type diskDataLoadedMsg struct {
	data *stats.DiskUsageData
}

// loadDiskDataCmd loads disk usage data from all repos
func loadDiskDataCmd(repos []model.Repo) tea.Cmd {
	return func() tea.Msg {
		data, _ := stats.GetDiskUsage(repos)
		return diskDataLoadedMsg{data: data}
	}
}

// timelineDataLoadedMsg is sent when timeline data is loaded
type timelineDataLoadedMsg struct {
	data *stats.TimelineData
}

// loadTimelineDataCmd loads timeline data from all repos
func loadTimelineDataCmd(repos []model.Repo) tea.Cmd {
	return func() tea.Msg {
		data, _ := stats.GetTimeline(repos)
		return timelineDataLoadedMsg{data: data}
	}
}
