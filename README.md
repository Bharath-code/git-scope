# git-scope

> A fast TUI to see the status of all git repositories on your machine.

![Status](https://img.shields.io/badge/status-MVP-blue)
![Go Version](https://img.shields.io/badge/go-1.22-00ADD8)

## Overview

**git-scope** is a terminal-based dashboard that helps you manage multiple git repositories. It scans your configured directories, shows you which repos have uncommitted changes, and lets you jump into your editor with a single keystroke.

### Features

- 🔍 **Scan** configured directories for git repos
- 📊 **Dashboard** showing branch, dirty files, and last commit
- ⌨️ **Keyboard-driven** navigation
- 🚀 **Jump** into your editor from the TUI
- ⚡ **Fast** concurrent scanning with goroutines

## Installation

### From Source

```bash
go install github.com/Bharath-code/git-scope/cmd/git-scope@latest
```

Or clone and build:

```bash
git clone https://github.com/Bharath-code/git-scope.git
cd git-scope
go build -o git-scope ./cmd/git-scope
```

## Usage

### Launch TUI Dashboard

```bash
git-scope
# or
git-scope tui
```

### Scan and Output JSON

```bash
git-scope scan
```

### Configuration

Create a config file at `~/.config/git-scope/config.yml`:

```yaml
# Directories to scan for git repos
roots:
  - ~/code
  - ~/projects
  - ~/work

# Directories to ignore
ignore:
  - node_modules
  - .next
  - dist
  - build
  - target
  - .venv
  - vendor

# Editor to open repos (default: code)
editor: code
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate repos |
| `Enter` | Open repo in editor |
| `r` | Rescan directories |
| `q` | Quit |

## Dashboard Columns

| Column | Description |
|--------|-------------|
| Repo | Repository name |
| Path | File path (truncated) |
| Branch | Current branch |
| Stg | Staged file count |
| Unst | Unstaged file count |
| Untrk | Untracked file count |
| Last Commit | Last commit timestamp |

## Roadmap

- [ ] Caching for faster startup
- [ ] Fuzzy search filter
- [ ] Sort by different columns
- [ ] Quick actions (pull, push)
- [ ] Background file watcher

## Tech Stack

- **Go** - Fast, compiled binary
- **Bubbletea** - TUI framework
- **Lipgloss** - Terminal styling
- **Bubbles** - TUI components

## License

MIT
