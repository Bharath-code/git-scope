# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [Unreleased] — planned as 1.4.0

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
