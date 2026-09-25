# Spec: Attention Summary + Scoring

Status: **Accepted**
Branch: `feat/attention-summary-scoring`

## Objective

git-scope promises a "command center… find what needs attention," but today the
dashboard delivers **views** (search/sort/filter/graph/disk/timeline) and leaves
the user to synthesize the verdict. This feature reduces the already-collected
per-repo git state into a single, glanceable answer to: **"What do I need to act
on right now?"**

Two deliverables:

1. **Attention summary header** — an always-visible one-line rollup of the
   workspace (e.g. `▲3 to push · ▼2 behind · ●5 dirty · 12 clean`).
2. **Per-repo attention tier + ranking** — each repo is bucketed into an ordinal
   tier (`ACTION` / `WATCH` / `CLEAN`), shown in the existing `Status` column,
   and the dashboard sorts by attention so the work queue lands at the top.

Pure synthesis and ranking. **Zero new git calls. Zero repository mutations.**
The read-only / local-first / no-network trust posture is fully preserved.

### User stories

- As a polyrepo dev, on launch I see at a glance how many repos need a push, are
  behind remote, and are dirty — without scanning every row.
- As a polyrepo dev, the repos needing action are sorted to the top by default.
- As a polyrepo dev, a repo I left dirty and untouched for a week is nudged up so
  it stops being forgotten.

## Tech Stack

- Go 1.26, existing TUI on Bubble Tea / Lip Gloss / Bubbles (unchanged versions).
- New pure package `internal/attention` (no Bubble Tea dependency).
- Input type: existing `model.RepoStatus` (`internal/model/repo.go`). No schema
  change required for v1.

## Commands

```
Build:  go build ./...
Test:   go test ./...
Test 1: go test ./internal/attention/...
Lint:   golangci-lint run        # v2.x, as configured in CI
Run:    go run ./cmd/git-scope
```

## Project Structure

```
internal/attention/           → NEW: pure scoring/tier logic (no TUI deps)
  attention.go                →   Tier type, Classify(), Summarize(), Less()
  attention_test.go           →   table-driven unit tests
internal/model/repo.go        →   RepoStatus (input; unchanged for v1)
internal/tui/model.go         →   add SortByAttention; default to it; row Status glyph
internal/tui/view.go          →   extend renderStats() with the summary rollup
internal/tui/styles.go        →   tier colors/glyph styles
docs/specs/attention-summary-scoring.md  → this spec
```

## Scoring Model (decisions locked)

Ordinal **tiers** (chosen over a weighted numeric score for predictability), with
a derived numeric **sub-score** used only to order repos *within* a tier.

### Tier classification (primary, user-facing)

| Tier     | Condition                                              | Glyph |
| :------- | :----------------------------------------------------- | :---- |
| `ACTION` | `ScanError != ""` OR `Ahead > 0` OR `Behind > 0`       | `▲`   |
| `WATCH`  | not ACTION, AND `IsDirty` (staged/unstaged/untracked)  | `●`   |
| `CLEAN`  | none of the above                                      | `✓`   |

Rationale: ACTION = "git history out of sync with remote (push/pull needed)" or
"something is broken." WATCH = local uncommitted work, in sync with remote.
A scan error is surfaced as ACTION so failures are never hidden by a CLEAN look.

### Sub-score (within-tier ordering only; never changes tier)

Priority order (most → least urgent): **unpushed → behind → dirty → stale**.
Top priority is *unpushed* because it is the only state where data loss is
possible (local-only commits lost if the machine dies); behind/dirty are
recoverable or already on disk.

```
subScore = 1000 * min(Ahead, 99)          # unpushed dominates
         +  100 * min(Behind, 99)
         +   10 * dirtyCount               # Staged+Unstaged+Untracked, capped
         +    5 * boolToInt(scanError)     # broken repos float up within ACTION
         +    1 * boolToInt(stale)         # low-weight nudge, never dominates
```

`stale = IsDirty && (now - LastCommit) > 7*24h`. Staleness only nudges
within-tier order and may show a subtle marker; it **never** promotes a repo to a
higher tier (per decision: "low weight, never dominates"). The 7-day threshold is
a constant in v1 (config later).

### Summary rollup (header)

Aggregate counts across the **currently filtered** repo set:

```
▲ N to push   = count(Ahead > 0)
▼ N behind    = count(Behind > 0)
● N dirty     = count(IsDirty)        # already shown today
✓ N clean     = count(CLEAN tier)
```

Note: "to push" and "behind" can overlap one repo (diverged) — counts are per
signal, not partitions, and that is intentional/honest.

## Sort Behavior (decision: Attention becomes the default)

- Add `SortByAttention` as a new `SortMode`.
- Initialize `sortMode = SortByAttention` (replaces `SortByDirty` default).
  Attention is a strict refinement of Dirty-first, so this is an evolution, not a
  jarring reorder.
- Ordering: tier rank (ACTION < WATCH < CLEAN), then sub-score desc, then name asc
  for stability.
- All existing sorts (`s` cycle, `1`–`4`) remain reachable and unchanged. The `s`
  cycle includes Attention.

## Code Style

Pure, dependency-free, table-driven-tested. Example shape:

```go
package attention

type Tier int

const (
    Clean Tier = iota
    Watch
    Action
)

// Classify returns the attention tier for a repo's status.
func Classify(s model.RepoStatus, now time.Time) Tier {
    if s.ScanError != "" || s.Ahead > 0 || s.Behind > 0 {
        return Action
    }
    if s.IsDirty {
        return Watch
    }
    return Clean
}
```

Conventions: no comments except exported-symbol doc lines; small functions;
injected `now time.Time` for deterministic tests (no hidden `time.Now()`).

## Testing Strategy

- Framework: stdlib `testing`, table-driven (matches existing repo style in
  `internal/scan`, `internal/gitstatus`).
- New `internal/attention/attention_test.go` covers:
  - Tier classification: each tier, boundary (Ahead=0/1, Behind=0/1, dirty combos),
    scan-error → ACTION, diverged (Ahead>0 && Behind>0).
  - Sub-score ordering: unpushed outranks behind outranks dirty; stale nudge;
    caps (Ahead=200 doesn't overflow into next priority band).
  - `Summarize` counts on a mixed slice, including overlap (diverged) and an
    empty slice.
  - Stale boundary around the 7-day threshold using an injected `now`.
- Coverage expectation: `internal/attention` ≥ 90% (it's pure logic).
- TUI wiring verified by `go build ./...` + manual run; no new TUI test harness
  introduced in this feature.

## Boundaries

- **Always:** keep scoring in `internal/attention` (pure, no TUI import); inject
  `now`; run `go test ./...` and `golangci-lint run` before commit; preserve
  read-only behavior.
- **Ask first:** adding a field to `model.RepoStatus`; adding a config schema key;
  changing existing default keybindings or removing a sort mode.
- **Never:** introduce a git mutation; add network/telemetry; remove or weaken
  existing tests; hide a scan error behind a CLEAN verdict.

## Success Criteria

1. `go test ./internal/attention/...` passes with ≥90% coverage.
2. `go build ./...` and `golangci-lint run` are clean.
3. On launch, the dashboard sorts ACTION repos to the top by default; `s`/`1`–`4`
   still switch to the legacy sorts.
4. The header shows the live `▲ to push · ▼ behind · ● dirty · ✓ clean` rollup,
   updating with filter/rescan/`F`-fetch.
5. The `Status` column shows a per-repo tier glyph (`▲`/`●`/`✓`).
6. A dirty repo with `LastCommit` > 7 days sorts above an equally-dirty fresh repo.
7. No new git subprocess calls are introduced (verify: scoring reads only
   existing `RepoStatus`).

## Open Questions

1. **Glyph vs short label** in the 8-wide `Status` column — glyph-only (`▲`) keeps
   width, a short label (`ACT`) is more legible. Proposed: glyph + existing color.
2. **Stale marker** — show a distinct `◷`/`*` for stale repos, or rely on order
   only? Proposed: order only in v1 to avoid column clutter.
3. **Diverged repo glyph** (both ahead & behind) — `⇅`, or just `▲`? Proposed:
   `▲` (ACTION tier already captures it; ahead/behind columns show detail).

These are low-stakes and can be resolved at Plan time unless you have a preference.
