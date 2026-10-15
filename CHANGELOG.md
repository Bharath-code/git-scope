# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [Unreleased]

### Added
- `git-scope status --json`: every repo's branch, uncommitted files,
  ahead/behind, unpushed commits, last commit, and attention tier
  (ACTION/WATCH/CLEAN) with a reason, sorted most urgent first.
- `git-scope mcp`: a read-only, local-only MCP server for AI coding agents
  (Claude Code, Cursor) with `list_repos`, `needs_attention` and
  `repo_status` tools. See "Use with Claude Code / Cursor" in the README.
- Releases now include an SPDX SBOM per archive and signed build provenance
  (`gh attestation verify`).
- Editor commands support a `{path}` placeholder for the repo path,
  e.g. `editor: "lazygit --path {path}"` (#41).

### Fixed
- Linked git worktrees are now scanned (their `.git` is a file, so they were
  skipped entirely).
- Commits on a branch that was never pushed, or on a detached HEAD, now count
  as unpushed instead of showing the repo as clean. Repos with no remote are
  unaffected.
- Merge conflicts now count as uncommitted work instead of a clean tree.
- Each git call is bounded by a 10s timeout, so one hung repo cannot stall
  a scan.
- `editor: lazygit` (and `gitui`, `tig`) now launches correctly instead of
  treating the repo path as a subcommand. All editors start with the working
  directory set to the selected repo (#23, #41).

### Security
- Config, cache and nudge-state files are now created owner-only
  (`0700` dirs, `0600` files) instead of world-readable; the cache lists every
  scanned repo path.

### Changed
- Homebrew: published as a cask instead of the deprecated GoReleaser formula.
- CI tests on Linux, macOS and Windows, checks `go.mod` tidiness, pins every
  GitHub Action to a commit SHA, and runs stricter linters (gosec, errorlint,
  revive, misspell). OpenSSF Scorecard runs weekly.
- Bug and feature reports use issue forms; blank issues are disabled.
- Upgraded TUI libraries: bubbletea 0.26 → 1.3, bubbles 0.18 → 1.0,
  lipgloss 0.11 → 1.1; mvdan.cc/sh 3.7 → 3.14 (#38).
- Bumped GitHub Actions dependencies (#39).

### Removed
- The `pageSize` config option. Page size has followed terminal height since
  1.4.0, so the setting had no effect; existing configs that still set it load
  fine and the key is ignored. Thanks @jimmckeeth for flagging it (#32).

## [1.4.0] - 2026-09-25

### Security
- Scanning no longer runs programs configured in a repository's `.git/config`
  (`core.fsmonitor`). Previously, scanning a folder containing a crafted repo
  (e.g. an unzipped archive) could execute arbitrary code.
- Status checks run with `--no-optional-locks`, so git-scope no longer
  rewrites `.git/index` and cannot cause `index.lock` errors in your own git
  commands. git-scope is now strictly read-only during scans.
- Git never prompts for credentials during bulk fetch (`GIT_TERMINAL_PROMPT=0`).
- The install script verifies the archive's SHA-256 against the release's
  `checksums.txt` and fails on HTTP errors.

### Added
- Attention summary and default Attention sort: repos ranked into
  Action / Watch / Clean tiers (#36).
- Bulk fetch across all repos with `F` (#31).
- Ahead/behind commit columns (#16); repos ahead or behind count as dirty (#15).
- Scoop installation on Windows (#22).
- Page size adapts to terminal height (#21).

### Fixed
- `git-scope --version` reported `1.0.1` for every release: the version was
  a `const`, which `-ldflags -X` cannot override.
- The star nudge now tracks the real running version.
