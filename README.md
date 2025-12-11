# git-scope

> A fast TUI to see the status of all git repositories on your machine.

![Status](https://img.shields.io/badge/status-v2.0_Released-green)
![Go Version](https://img.shields.io/badge/go-1.20+-00ADD8)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

<p align="center">
  <img src="docs/git-scope-demo-1.webp" alt="git-scope demo" width="600">
</p>

**[🌐 Visit the Landing Page](https://bharath-code.github.io/git-scope/)**

## Overview

**git-scope** is a terminal-based dashboard that helps you manage multiple git repositories. It scans your configured directories, shows you which repos have uncommitted changes, and lets you jump into your editor with a single keystroke.

### Features

- 🔍 **Fuzzy Search**: Find any repo by name, path, or branch with `/`
- 🛡️ **Dirty Filter**: Press `f` to see only repos with uncommitted changes
- ⚡ **Fast caching**: Instant startup using JSON caching
- 📊 **Dashboard**: Shows branch, staged/unstaged counts, and last commit
- ⌨️ **Keyboard-driven**: Vim-like navigation (`j`/`k`) and sorting (`s`)
- 🚀 **Quick Jump**: Open any repo in your editor (VSCode, nvim, etc.) with `Enter`
- 🌿 **Contribution Graph**: GitHub-style heatmap of your local commits (`g`)
- 💾 **Disk Usage**: See `.git` and `node_modules` sizes at a glance (`d`)
- ⏰ **Timeline**: What were you working on? See recent activity (`t`)

## Installation

### Homebrew (macOS/Linux)
```bash
brew tap Bharath-code/tap
brew install git-scope
```

**To Update:**
```bash
brew upgrade git-scope
```

### From Source
```bash
go install github.com/Bharath-code/git-scope/cmd/git-scope@latest
```

**To Update:**
Run the install command again to fetch the latest version.

## Usage

### Launch TUI
```bash
git-scope
```

### Configuration
On first run, `git-scope` will auto-detect your directories. To customize, edit `~/.config/git-scope/config.yml`:

```yaml
roots:
  - ~/code
  - ~/work

ignore:
  - node_modules
  - .venv

editor: code  # or nvim, vim, helix
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `/` | **Search** repos |
| `f` | Toggle **Filter** (All/Dirty/Clean) |
| `s` | Cycle **Sort** mode |
| `1-4` | Sort by Dirty/Name/Branch/Recent |
| `c` | **Clear** search & filters |
| `Enter` | **Open** repo in editor |
| `r` | **Rescan** directories |
| `g` | Toggle **Grass** (contribution graph) |
| `d` | Toggle **Disk** usage panel |
| `t` | Toggle **Timeline** panel |
| `Esc` | Close panel |
| `q` | **Quit** |

## Roadmap

- [ ] Background file watcher
- [ ] Quick actions (pull, push)
- [ ] Team dashboards

## License

MIT
