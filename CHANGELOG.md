# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Opt-in support for linked git worktrees ([#24](https://github.com/Bharath-code/git-scope/issues/24)).
  - Config: `includeWorktrees: false` (default) in `~/.config/git-scope/config.yml`.
  - CLI flag: `--worktrees` for one-shot enable.
  - TUI: press `W` to toggle live; the toggle controls both visibility and totals.
  - Scan/JSON: each repo carries an `is_worktree` field; worktrees show a `⎇` marker in the TUI table.
  - Submodules are excluded — only `gitdir:` pointers under `.git/worktrees/` are recognised.

