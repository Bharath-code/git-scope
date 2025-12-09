# Marketing & Submission Copy

## Hacker News (Show HN)

**Title:** Show HN: git-scope — A fast TUI to manage all your git repos

**Comment:**
Hi HN,

I built `git-scope` because I often work on 10+ microservices at once and lose track of which ones have uncommitted changes or are behind origin. existing tools were either too slow or too complex.

`git-scope` is a minimal TUI written in Go (using Bubble Tea) that:
- recursively finds all git repos in your projects folder
- shows a dashboard with dirty/clean status, branch, and sync status
- fuzzy search (`/`), filter by status (`f`), and sort columns
- opens the repo in your editor with one keypress
- caches results for instant startup

It's open source and available via Homebrew.

Repo: https://github.com/Bharath-code/git-scope
Landing Page: https://bharath-code.github.io/git-scope/

Would love your feedback!

---

## Reddit (r/golang, r/commandline, r/git)

**Title:** I built a TUI to manage all my local git repositories (Go + Bubble Tea)

**Body:**
Hey everyone,

I wanted to share a tool I built to solve "repo chaos". It's called **git-scope**.

It scans your dev directories and gives you a bird's-eye view of all your repositories. You can instantly see which ones have uncommitted changes, are on the wrong branch, or have unsaved files.

**Features:**
- ⚡️ Fast concurrent scanning + caching
- 🔍 Fuzzy search for repo name/path
- 🛡️ Filter by "Dirty" status to find lost work
- 🎨 Modern TUI with Bubble Tea & Lipgloss
- ⌨️ Vim-like navigation (j/k)

**Homebrew (macOS/Linux):**
```bash
brew tap Bharath-code/tap
brew install git-scope
```

**Go Install:**
```bash
go install github.com/Bharath-code/git-scope/cmd/git-scope@latest
```

**Repo:** https://github.com/Bharath-code/git-scope

Feedback welcome!

---

## Awesome-Go Submission

**Entry:**
[git-scope](https://github.com/Bharath-code/git-scope) - A fast TUI to visualize and manage multiple git repositories with fuzzy search and filtering.

**Section:**
Development Tools or Version Control
