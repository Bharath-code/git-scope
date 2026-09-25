# Roadmap Plan: Items 2–7

Status: planned, not started. Order below is the implementation order.
Prerequisite (item 1, not covered here): release v1.4.0 from current main.

---

## 2. Fix bugs #33 (ahead/behind not visible) and #23 (lazygit launch)

### 2a. #33 — Ahead/behind status not visible

**Root cause:** Reporter is on v1.3.1; ahead/behind columns landed after that
release (#16, unreleased). Residual product gap: `behind` is only accurate
after a fetch, so a stale repo silently shows "clean" — exactly the
reporter's complaint.

**Plan**
1. Ship v1.4.0 (columns become visible) and comment on the issue.
2. Show remote-freshness in the UI: track last-fetch time per repo (readable
   from `.git/FETCH_HEAD` mtime; no network needed) and render a stale
   indicator (e.g. dim `?` in behind column or `fetched 3d ago`) when older
   than a threshold (default 24h, configurable).
3. Point users at the existing bulk-fetch quick action (F) in the help bar
   and README so the "why is behind wrong" loop has an in-app answer.

**Acceptance criteria**
- [ ] A repo 1+ commits behind its upstream shows a non-zero behind count in
      the TUI after a fetch.
- [ ] A repo with no upstream branch shows a distinct marker (e.g. `-`), not
      `0`, in ahead/behind columns.
- [ ] A repo whose last fetch is older than the staleness threshold shows a
      visible stale indicator; after pressing F it clears.
- [ ] No network calls happen during scan/render (fetch only on explicit F).
- [ ] README "ahead/behind" claim matches actual behavior, including the
      fetch requirement, in one short paragraph.
- [ ] Issue #33 closed with a comment linking the release.

### 2b. #23 — `editor: lazygit` doesn't launch / relaunch

**Root causes (three separate defects):**
1. Editor invoked as `<editor> <repo-path>`; lazygit rejects a bare path
   (needs `--path <repo>` or cwd set to the repo).
2. After the child editor exits once, subsequent Enter presses do nothing —
   the bubbletea `tea.ExecProcess` return isn't restoring/re-arming state.
3. `git-scope --version` prints `v1.0.1` on Scoop builds — version ldflags
   not injected in that build path.

**Plan**
1. Change editor launch to run the command with `cmd.Dir = repoPath` and no
   positional path arg when the configured editor has no `{path}`
   placeholder issues — simplest fix that works for lazygit, vim, nvim, etc.
   Support an optional `{path}` token in the `editor` config for tools that
   need the path as an explicit arg (e.g. `code {path}`); if absent and the
   editor is a known GUI editor pattern (`code`, `subl`, ...), append path
   as today. Keep it to: token present → substitute; token absent → append
   path AND set cwd. Document in README.
2. Fix the relaunch bug: audit the `tea.ExecProcess` callback path in
   internal/tui — ensure terminal state and key handling are restored and
   the editor can be launched repeatedly in one session.
3. Fix version stamping for all release artifacts (goreleaser/Scoop
   manifest) so `--version` matches the installed tag.

**Acceptance criteria**
- [ ] With `editor: lazygit`, pressing Enter on a repo opens lazygit *in
      that repo*; quitting lazygit returns to a fully functional TUI.
- [ ] Pressing Enter on a second repo immediately after opens the editor
      again (repeatable ≥3 times in one session).
- [ ] `editor: code {path}` opens VS Code at the repo; `editor: vim` opens
      vim with cwd = repo.
- [ ] Works on Windows (reporter's platform) — verified via CI build +
      reporter confirmation or a Windows smoke test.
- [ ] `git-scope --version` prints the release tag on brew, Scoop, and
      install-script builds.
- [ ] Unit test covers the command-construction logic ({path} token,
      cwd fallback).
- [ ] Issue #23 closed with reporter confirmation requested.

---

## 3. Ship attention scoring (current branch)

**State:** implemented on `feat/attention-summary-scoring`
(internal/attention, 95.5% coverage; TUI wiring done; specs in this folder).

**Plan**
1. Rebase on main, run full test suite + lint.
2. Self-review the diff against the spec (attention-summary-scoring.md);
   run the TUI against a real workspace with mixed repo states.
3. Update README: new hero framing ("You have 40 repos. Three need you."),
   document verdicts (ACTION/WATCH/CLEAN), glyphs (▲●✓), default sort, and
   how to revert to name sort.
4. New demo GIF showing the summary header + sort.
5. PR → merge → release v1.5.0 (its own release; don't bundle with bug
   fixes so the changelog tells one story).

**Acceptance criteria**
- [ ] Default sort is Attention; toggling back to name sort works and the
      choice is persistable via config.
- [ ] Header rollup shows counts per verdict and matches the row glyphs.
- [ ] A repo with unpushed commits always ranks above a merely-dirty repo
      (per spec's unpushed-first scoring) — covered by a unit test.
- [ ] Empty workspace / all-clean workspace render sensibly (no divide-by-
      zero, header reads "all clean").
- [ ] `go test ./... -race` and golangci-lint pass in CI.
- [ ] README + demo GIF updated; website hero copy updated to match.
- [ ] Released as v1.5.0 with changelog entry explaining the new default
      sort (behavior change — call it out prominently).

---

## 4. MCP server / agent integration

**Goal:** make workspace state legible to AI agents without breaking the
read-only/local-first promise. Build on the existing `scan` JSON output.

**Plan**
1. Stabilize the JSON contract first: define the output schema of
   `git-scope scan` (fields incl. ahead/behind, verdict, score), add
   `--format json` flag explicitly, and treat it as a versioned API
   (add `"schemaVersion": 1`).
2. New subcommand `git-scope mcp` — stdio MCP server exposing tools:
   - `list_repos` (filters: verdict, dirty, unpushed)
   - `repo_status` (single repo detail)
   - `attention_summary` (the header rollup as structured data)
   Use the official Go MCP SDK (`modelcontextprotocol/go-sdk`); verify
   current API from docs before implementing.
3. Read-only guarantee: the MCP server exposes zero mutating tools; no
   fetch, no network. State this in the tool descriptions.
4. Docs: README section "Use with Claude Code / AI agents" with the
   one-line `claude mcp add` setup; short blog-able writeup.

**Acceptance criteria**
- [ ] `git-scope scan --format json` output documented (schema in
      docs/), includes schemaVersion, stable field names.
- [ ] `git-scope mcp` speaks MCP over stdio; `claude mcp add git-scope --
      git-scope mcp` works and Claude Code can answer "which repos have
      unpushed work?" end-to-end (manually verified, steps recorded).
- [ ] All three tools return within 2s on a 50-repo workspace using cache;
      no tool performs writes or network I/O.
- [ ] Malformed tool input returns an MCP error, not a crash.
- [ ] Unit tests for tool handlers (table-driven, using fake scan data).
- [ ] README section + demo (screenshot or asciinema of Claude querying it).
- [ ] Released as v1.6.0; announcement post drafted ("MCP server for your
      git workspace").

---

## 5. Watch mode

**Goal:** `git-scope --watch` (or key `w` conflict check — `w` is taken by
workspace switch; use flag + maybe `W`) — auto-refresh so the TUI works as
a persistent dashboard.

**Plan**
1. Timer-based rescan (default every 30s, `--watch[=interval]`), reusing
   the existing rescan path; no fsnotify in v1 (watching hundreds of repos
   via fsnotify is complexity + fd-limit risk; ponytail: timer first,
   fsnotify only if users ask).
2. Preserve UI state across refresh: cursor position, page, sort, filter.
3. Visible "last scanned 12s ago / scanning…" indicator; pause/resume key.
4. Scan runs async (already bubbletea-friendly) so UI never blocks.

**Acceptance criteria**
- [ ] `git-scope --watch` refreshes at the interval; default interval
      documented; `--watch=10s` style override works.
- [ ] During refresh the UI stays responsive and cursor/page/sort/filter
      are preserved.
- [ ] Status bar shows time since last scan and an active-scan spinner.
- [ ] A repo changed externally (touch a file, make a commit) is reflected
      within one interval — covered by an integration test or scripted
      manual check.
- [ ] CPU usage at idle with 50 repos stays negligible between ticks (no
      busy loop); no memory growth over 1h run (manual soak check).
- [ ] Works combined with attention sort: rows re-rank after refresh.

---

## 6. AUR + nixpkgs packaging

**Plan**
1. AUR: publish `git-scope-bin` (from goreleaser artifacts) — maintain a
   PKGBUILD repo; optionally automate bump via goreleaser AUR support.
2. Nix: submit derivation to nixpkgs (`buildGoModule`); follow nixpkgs
   contribution flow (PR against NixOS/nixpkgs, respond to review).
3. Add both to README install section; add release-checklist note so
   future tags bump AUR automatically.

**Acceptance criteria**
- [ ] `yay -S git-scope-bin` installs a working binary with correct
      `--version` on Arch (VM or Docker check).
- [ ] AUR package updates automatically (or via one scripted step) on the
      next tagged release.
- [ ] nixpkgs PR submitted; `nix run nixpkgs#git-scope` works once merged
      (merge timing is out of our control — criterion is submitted +
      review feedback addressed within a week of each round).
- [ ] README install section lists AUR and Nix with copy-paste commands.

---

## 7. #18 — mrconfig support (community-built)

**Goal:** repos list from `~/.mrconfig` as an alternative source of truth.
Deliberately delegated: this is contributor-bait, not core work.

**Plan (maintainer effort only)**
1. Design the seam so a contribution is easy: comment on #18 with the
   agreed approach — a `repoSource` concept where config gains
   `mrconfig: <path>` (or CLI `--mrconfig`); parser reads section headers
   (INI-style paths) from .mrconfig and feeds the existing scan list;
   scanning and mrconfig can merge (union, dedup by path).
2. Label `good first issue` + `help wanted`; include pointers to the
   files to touch (internal/config, internal/scan) and the test pattern
   to follow.
3. Review the PR when it comes; if no takers in ~2 months, implement it
   ourselves in an afternoon (parser is ~50 lines).

**Acceptance criteria (for the eventual PR, ours or theirs)**
- [ ] With `mrconfig: ~/.mrconfig` in config, the TUI lists exactly the
      repos named in the file (plus scanned roots if both configured).
- [ ] Missing/unreadable mrconfig → clear error message, not a crash;
      non-repo paths in mrconfig are skipped with a warning.
- [ ] Parser has unit tests incl. comments, chain/checkout lines ignored,
      `~` expansion.
- [ ] README documents the option with a myrepos interop example.
- [ ] Issue #18 has the design comment + labels within this milestone,
      regardless of who implements.

---

## Sequencing & releases

| Release | Contents | Why |
|---------|----------|-----|
| v1.4.0 | existing unreleased main | unblock #33, restore cadence |
| v1.4.1 | items 2a residual + 2b | bug-fix release, closes both bugs |
| v1.5.0 | item 3 (attention scoring) | flagship feature, own story, HN post |
| v1.6.0 | item 4 (MCP) | second announcement beat |
| v1.6.x | item 5 (watch mode) | quality-of-life |
| ongoing | items 6, 7 | packaging + community, no release coupling |
