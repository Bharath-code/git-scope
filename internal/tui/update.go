package tui

import (
	"fmt"
	"os/exec"

	"github.com/Bharath-code/git-scope/internal/browser"
	"github.com/Bharath-code/git-scope/internal/model"
	"github.com/Bharath-code/git-scope/internal/nudge"
	"github.com/Bharath-code/git-scope/internal/scan"
	"github.com/Bharath-code/git-scope/internal/stats"
	"github.com/Bharath-code/git-scope/internal/workspace"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"mvdan.cc/sh/v3/shell"
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

	case spinner.TickMsg:
		// Update spinner during loading
		if m.state == StateLoading {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case scanCompleteMsg:
		m.repos = msg.repos
		m.state = StateReady
		m.resetPage()
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

	case workspaceScanCompleteMsg:
		m.repos = msg.repos
		m.state = StateReady
		m.resetPage()
		m.updateTable()

		// Show helpful message about switched workspace
		if len(msg.repos) == 0 {
			m.statusMsg = fmt.Sprintf("⚠️  No git repos found in %s", msg.workspacePath)
		} else {
			m.statusMsg = fmt.Sprintf("✓ Switched to %s (%d repos)", msg.workspacePath, len(msg.repos))

			// Trigger star nudge after successful workspace switch
			if nudge.ShouldShowNudge() && !m.nudgeShownThisSession {
				m.showStarNudge = true
				m.nudgeShownThisSession = true
				nudge.MarkShown()
			}
		}
		return m, nil

	case workspaceScanErrorMsg:
		m.state = StateError
		m.err = msg.err
		return m, nil

	case openEditorMsg:
		// Parse editor command (handles "editor --flag" style configs)
		fields, err := shell.Fields(m.cfg.Editor, nil)
		if err != nil || len(fields) == 0 {
			m.statusMsg = fmt.Sprintf("❌ Invalid editor command: '%s'", m.cfg.Editor)
			return m, nil
		}
		// Check if editor binary exists in PATH
		_, err = exec.LookPath(fields[0])
		if err != nil {
			m.statusMsg = fmt.Sprintf("❌ Editor '%s' not found. Press 'e' to change editor or install it first.", fields[0])
			return m, nil
		}

		args := append(fields[1:], msg.path)
		c := exec.Command(fields[0], args...)
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

		// Handle workspace switch mode
		if m.state == StateWorkspaceSwitch {
			return m.handleWorkspaceSwitchMode(msg)
		}

		// Normal mode key handling
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "S":
			// Open GitHub repo (Star nudge action)
			if m.showStarNudge {
				m.showStarNudge = false
				nudge.MarkCompleted()
				m.statusMsg = "⭐ Opening GitHub..."
				return m, openBrowserCmd(nudge.GitHubRepoURL)
			}

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
				m.resetPage()
				m.updateTable()
				m.statusMsg = "Filter: " + m.GetFilterModeName()
				return m, nil
			}

		case "s":
			if m.state == StateReady {
				m.sortMode = (m.sortMode + 1) % 4
				m.resetPage()
				m.updateTable()
				m.statusMsg = "Sorted by: " + m.GetSortModeName()
				return m, nil
			}

		case "1":
			if m.state == StateReady {
				m.sortMode = SortByDirty
				m.resetPage()
				m.updateTable()
				m.statusMsg = "Sorted by: Dirty First"
				return m, nil
			}

		case "2":
			if m.state == StateReady {
				m.sortMode = SortByName
				m.resetPage()
				m.updateTable()
				m.statusMsg = "Sorted by: Name"
				return m, nil
			}

		case "3":
			if m.state == StateReady {
				m.sortMode = SortByBranch
				m.resetPage()
				m.updateTable()
				m.statusMsg = "Sorted by: Branch"
				return m, nil
			}

		case "4":
			if m.state == StateReady {
				m.sortMode = SortByLastCommit
				m.resetPage()
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
				m.resetPage()
				m.resizeTable()
				m.updateTable()
				m.statusMsg = "Filters cleared"
				return m, nil
			}

		case "e":
			if m.state == StateReady {
				// Check if editor exists (parse command to get binary name)
				fields, err := shell.Fields(m.cfg.Editor, nil)
				if err != nil || len(fields) == 0 {
					m.statusMsg = fmt.Sprintf("❌ Invalid editor command: '%s'", m.cfg.Editor)
				} else if _, err := exec.LookPath(fields[0]); err != nil {
					m.statusMsg = fmt.Sprintf("❌ Editor '%s' not found in PATH. Install it or edit ~/.config/git-scope/config.yml", fields[0])
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

		case "w":
			// Open workspace switch modal
			if m.state == StateReady {
				m.state = StateWorkspaceSwitch
				m.workspaceInput.SetValue("")
				m.workspaceInput.Focus()
				m.workspaceError = ""
				return m, textinput.Blink
			}

		case "[":
			// Previous page
			if m.state == StateReady && m.canGoPrev() {
				m.currentPage--
				m.updateTable()
				m.statusMsg = fmt.Sprintf("Page %d of %d", m.currentPage+1, m.getTotalPages())
				return m, nil
			}

		case "]":
			// Next page
			if m.state == StateReady && m.canGoNext() {
				m.currentPage++
				m.updateTable()
				m.statusMsg = fmt.Sprintf("Page %d of %d", m.currentPage+1, m.getTotalPages())
				return m, nil
			}
		}
	}

	// Dismiss star nudge on any key (if not already handled)
	if m.showStarNudge {
		m.showStarNudge = false
		nudge.MarkDismissed()
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
		m.resetPage()
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

// handleWorkspaceSwitchMode handles key events when in workspace switch mode
func (m Model) handleWorkspaceSwitchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel workspace switch
		m.state = StateReady
		m.workspaceInput.Blur()
		m.workspaceError = ""
		return m, nil

	case "enter":
		// Validate and switch workspace
		inputPath := m.workspaceInput.Value()
		if inputPath == "" {
			m.workspaceError = "Please enter a path"
			return m, nil
		}

		// Normalize the path (expand ~, resolve symlinks, validate)
		normalizedPath, err := workspace.NormalizeWorkspacePath(inputPath)
		if err != nil {
			m.workspaceError = err.Error()
			return m, nil
		}

		// Switch to loading state and scan the new workspace
		m.state = StateLoading
		m.workspaceInput.Blur()
		m.workspaceError = ""
		m.activeWorkspace = normalizedPath
		m.statusMsg = "🔄 Switching to " + normalizedPath + "..."

		return m, scanWorkspaceCmd(normalizedPath, m.cfg.Ignore)

	case "tab":
		// Tab completion for directory paths
		currentPath := m.workspaceInput.Value()
		if currentPath != "" {
			completedPath := workspace.CompleteDirectoryPath(currentPath)
			if completedPath != currentPath {
				m.workspaceInput.SetValue(completedPath)
				// Move cursor to end
				m.workspaceInput.CursorEnd()
			}
		}
		return m, nil

	case "ctrl+c":
		return m, tea.Quit
	}

	// Update text input
	var cmd tea.Cmd
	m.workspaceInput, cmd = m.workspaceInput.Update(msg)

	// Clear error when typing
	if m.workspaceError != "" {
		m.workspaceError = ""
	}

	return m, cmd
}

// workspaceScanCompleteMsg is sent when workspace scanning is complete
type workspaceScanCompleteMsg struct {
	repos         []model.Repo
	workspacePath string
}

// workspaceScanErrorMsg is sent when workspace scanning fails
type workspaceScanErrorMsg struct {
	err error
}

// scanWorkspaceCmd scans a single workspace path for repositories
func scanWorkspaceCmd(workspacePath string, ignore []string) tea.Cmd {
	return func() tea.Msg {
		repos, err := scan.ScanRoots([]string{workspacePath}, ignore)
		if err != nil {
			return workspaceScanErrorMsg{err: err}
		}

		return workspaceScanCompleteMsg{
			repos:         repos,
			workspacePath: workspacePath,
		}
	}
}

// openBrowserCmd opens a URL in the default browser
func openBrowserCmd(url string) tea.Cmd {
	return func() tea.Msg {
		_ = browser.Open(url)
		return nil
	}
}
