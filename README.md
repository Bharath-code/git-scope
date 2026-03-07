# git-scope

> **A fast TUI dashboard to view the git status of *all your repositories* in one place.** > Stop the `cd` → `git status` loop.

[![Go Report Card](https://goreportcard.com/badge/github.com/Bharath-code/git-scope)](https://goreportcard.com/report/github.com/Bharath-code/git-scope)
[![GitHub Release](https://img.shields.io/github/v/release/Bharath-code/git-scope?color=8B5CF6)](https://github.com/Bharath-code/git-scope/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GitHub stars](https://img.shields.io/github/stars/Bharath-code/git-scope)](https://github.com/Bharath-code/git-scope/stargazers)
[![Buy Me A Coffee](https://img.shields.io/badge/Buy%20Me%20A%20Coffee-FFDD00?style=flat&logo=buy-me-a-coffee&logoColor=black)](https://buymeacoffee.com/iam_pbk)

🌐 **Website:** [bharath-code.github.io/git-scope](https://bharath-code.github.io/git-scope/)

![git-scope Demo](docs/git-scope-demo-1.webp)

---

## 🛡️ Trust & Safety

git-scope is designed to be safe to run frequently and safe to recommend.

- **Read-only by design** — it does not modify repositories
- **Local-first** — no network access, no telemetry, no accounts
- **Predictable** — no background services or daemons
- **Conservative scope** — focused on visibility, not automation

These choices make git-scope suitable for daily use and work environments where reliability and transparency matter.

---

## ⚡ Installation

Get started in seconds.

### Homebrew (macOS/Linux)
```bash
brew tap Bharath-code/tap && brew install git-scope
````
### Update
```bash
brew upgrade git-scope
````

### Universal Installer (macOS/Linux)
```bash
curl -sSL https://raw.githubusercontent.com/Bharath-code/git-scope/main/scripts/install.sh | sh
```

### Scoop (Windows)
```powershell
scoop bucket add git-scope https://github.com/Bharath-code/git-scope
scoop install git-scope
```

### From Source (Windows)

```bash
go install github.com/Bharath-code/git-scope/cmd/git-scope@latest
```

*If you find this useful, please consider giving it a ⭐ star to help others find it\!*

-----

## 🚀 Usage

Simply run it in any directory containing your projects:

```bash
git-scope
```

#### Commands
```bash
git-scope              # Launch TUI dashboard
git-scope init         # Create config file interactively
git-scope scan         # Scan and print repos (JSON)
git-scope scan-all     # Full system scan from home directory
git-scope issue        # Open GitHub issues page in browser
git-scope -h           # Show help
```

*By default, it recursively scans the current directory. You can configure permanent root paths later.*

-----

## 🆚 git-scope vs. lazygit


  * **git-scope** is for your **workspace** (bird's-eye view).
  * **lazygit** is for a specific **repository** (deep dive).

| Feature | **git-scope** | **lazygit** |
| :--- | :--- | :--- |
| **Scope** | **All repos at once** | One repo at a time |
| **Primary Goal** | Find what needs attention | Stage/Commit/Diff |
| **Fuzzy Search** | Find repo by name/path | ❌ |
| **Integration** | Press `Enter` to open editor | Press `Enter` to stage files |
| **Performance** | \~10ms startup (cached) | Slower on large monorepos |

-----

## ✨ Features

  * **📁 Workspace Switch** — Switch root directories without quitting (`w`). Supports `~`, relative paths, and **symlinks**.
  * **🔍 Fuzzy Search** — Find any repo by name, path, or branch (`/`).
  * **🛡️ Dirty Filter** — Instantly show only repos with uncommitted changes (`f`).
  * **📄 Pagination** — Navigate large repo lists with page-by-page browsing (`[` / `]`). Shows 15 repos per page with a dynamic page indicator.
  * **🚀 Editor Jump** — Open the selected repo in VSCode, Neovim, Vim, or Helix (`Enter`).
  * **⚡ Blazing Fast** — JSON caching ensures \~10ms launch time even with 50+ repos.
  * **📊 Dashboard Stats** — See branch name, staged/unstaged counts, and last commit time.
  * **🌿 Contribution Graph** — GitHub-style local heatmap for your activity (`g`).
  * **💾 Disk Usage** — Visualize `.git` vs `node_modules` size (`d`).
  * **⏰ Timeline** — View recent activity across all projects (`t`).
  * **🔗 Symlink Support** — Symlinked directories resolve transparently (great for Codespaces/devcontainers).

-----

## 🎯 Use Cases

git-scope excels in environments where multi-repo complexity is a daily burden:

*   **Microservices Management** — Quickly verify if all your services are on the correct branch and have no unpushed changes.
*   **OSS Contribution Tracking** — Keep tabs on various upstream forks and personal branches in one view.
*   **Infrastructure as Code (IaC)** — Monitor multiple Terraform/CloudFormation repos for configuration drift or uncommitted edits.
*   **Context Recovery** — Instantly see where you left off after a weekend or a holiday without running `git status` 20 times.

-----

## 🏆 The git-scope Advantage

While many Git tools focus on the *micro* (committing, staging, diffing), git-scope is built for the *macro*.

Typical git workflows involve "tunnel vision"—working deep inside one repository. git-scope provides the "command center" view. It is intentionally **read-only**, **local-first**, and **blazing fast** (<10ms), making it the perfect companion to tools like `lazygit`, `gitui`, or VS Code.

-----

## ⌨️ Keyboard Shortcuts

| Key | Action |
| :--- | :--- |
| `w` | **Switch Workspace** (with Tab completion) |
| `/` | **Search** repositories (Fuzzy) |
| `f` | **Filter** (Cycle: All / Dirty / Clean) |
| `s` | Cycle **Sort** Mode |
| `1`–`4` | Sort by: Dirty / Name / Branch / Recent |
| `[` / `]` | **Page Navigation** (Previous / Next) |
| `Enter` | **Open** repo in Editor |
| `c` | **Clear** search & filters |
| `r` | **Rescan** directories |
| `g` | Toggle **Contribution Graph** |
| `d` | Toggle **Disk Usage** view |
| `t` | Toggle **Timeline** view |
| `q` | Quit |

-----

## ⚙️ Configuration

Edit workspace location and code editor of your choice in `~/.config/git-scope/config.yml`:


```yaml
# ~/.config/git-scope/config.yml
roots:
  - ~/code
  - ~/work/microservices
  - ~/personal/experiments

ignore:
  - node_modules
  - .venv
  - dist

editor: code # options: code,nvim,lazygit,vim,cursor
```

-----

## 💡 Why I Built This

I work across dozens of small repositories—microservices, dotfiles, and side projects. I kept forgetting which repos had uncommitted changes or unpushed commits.

My mornings used to look like this:

```bash
cd repo-1 && git status
cd ../repo-2 && git status
# ... repeat for 20 repos
```

I built `git-scope` to solve the **"Multi-Repo Blindness"** problem. It gives me a single screen to see what is dirty, what is ahead/behind, and where I left off yesterday.

-----

## 🗺️ Roadmap

  - [x] In-app workspace switching with Tab completion
  - [x] Symlink resolution for devcontainers/Codespaces
  - [x] Background file watcher (real-time updates)
  - [ ] Quick actions (bulk pull/fetch)
  - [ ] Repo grouping (Service / Team / Stack)
  - [ ] Custom team dashboards

## ❓ FAQ

### How do you manage multiple Git repositories locally?

git-scope provides a fast, read-only terminal dashboard that shows the status of many local Git repositories at once. It helps developers regain context across projects without switching directories or running commands repeatedly.

### What problem does git-scope solve?

git-scope reduces context switching when working across many Git repositories, such as microservices, tools, or configuration repos. It gives a single overview of repository state so developers can quickly see what needs attention.

### Is git-scope safe to use at work?

Yes. git-scope is read-only by design, runs entirely locally, has no network access or telemetry, and does not modify repositories. This makes it safe for daily and work use.

### Does git-scope replace git commands?

No. git-scope does not perform git actions like commit, push, or pull. It focuses on visibility and orientation, not automation.

### Is git-scope suitable for monorepos?

git-scope is designed for multi-repo (polyrepo) workflows. It is not intended to manage monorepos.

### What platforms does git-scope support?

git-scope runs on macOS, Linux, and Windows.

### How is git-scope different from other git TUIs?

Most git TUIs focus on interacting with a single repository. git-scope focuses on visibility across many repositories and stays intentionally read-only.

---

## 📄 License

MIT © [Bharath-code](https://github.com/Bharath-code)

---

## 🙏 Acknowledgements

Built with these amazing open-source projects:

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — The TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Style definitions
- [Bubbles](https://github.com/charmbracelet/bubbles) — TUI components (table, spinner, text input)

---

## ⭐ Star History

<a href="https://star-history.com/#Bharath-code/git-scope&Date">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=Bharath-code/git-scope&type=Date&theme=dark" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=Bharath-code/git-scope&type=Date" />
   <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=Bharath-code/git-scope&type=Date" />
 </picture>
</a>

---

## 👥 Contributors

<a href="https://github.com/Bharath-code/git-scope/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=Bharath-code/git-scope" />
</a>

Made with [contrib.rocks](https://contrib.rocks).