# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [Unreleased]

### Added
- Editor commands support a `{path}` placeholder for the repo path,
  e.g. `editor: "lazygit --path {path}"` (#41).

### Fixed
- `editor: lazygit` (and `gitui`, `tig`) now launches correctly instead of
  treating the repo path as a subcommand. All editors start with the working
  directory set to the selected repo (#23, #41).

### Changed
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
