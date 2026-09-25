# Plan: Attention Summary + Scoring

Status: **Accepted**
Spec: `docs/specs/attention-summary-scoring.md`

## Components & dependencies

```
┌─────────────────────────────┐
│ internal/attention (NEW)    │  pure logic, no TUI import
│  Tier, Classify, SubScore,  │  depends only on: internal/model, time
│  Summary, Summarize, Less   │
└──────────────┬──────────────┘
               │ consumed by
       ┌───────┴────────────────────────────┐
       ▼                                     ▼
┌───────────────────┐                ┌──────────────────────┐
│ tui/model.go      │                │ tui/view.go          │
│  SortByAttention  │                │  renderStats() rollup│
│  default sortMode │                │  Status-col glyph    │
│  sortRepos() case │                │ (+ tui/styles.go)    │
└───────────────────┘                └──────────────────────┘
```

Dependency direction is one-way: TUI → attention → model. `attention` never
imports `tui`. This is what keeps it unit-testable and reusable for a future
CLI/`--json` exit-code path.

## Build order (dependency-ordered)

1. **`internal/attention` package + tests first** (TDD). Nothing depends on TUI
   here, so it's fully verifiable in isolation before any UI wiring.
2. **Sort integration** in `tui/model.go` — add `SortByAttention`, flip default,
   add the `sortRepos` case delegating to `attention.Less`.
3. **Status-column glyph** in `tui/model.go` row builder + `tui/styles.go` colors.
4. **Summary rollup** in `tui/view.go` `renderStats()`.
5. **Docs** — flip spec/plan Status to Accepted; update README (features +
   keyboard table note that default sort is Attention).

Steps 2–4 all depend on step 1. Steps 3 and 4 are independent of each other
(could be parallel) but both depend on 2 being merged-in mentally (shared file
`model.go` for 2 and 3 → keep sequential to avoid edit conflicts).

## API surface of `internal/attention`

```go
type Tier int                  // Clean < Watch < Action (ordinal)
func (t Tier) String() string  // "CLEAN" | "WATCH" | "ACTION"
func (t Tier) Glyph() string   // "✓" | "●" | "▲"

func Classify(s model.RepoStatus, now time.Time) Tier
func SubScore(s model.RepoStatus, now time.Time) int   // within-tier ordering
func Stale(s model.RepoStatus, now time.Time) bool     // IsDirty && age>7d

// Less reports whether repo a should sort before repo b (tier, then subscore desc,
// then name asc). Used by tui sortRepos.
func Less(a, b model.Repo, now time.Time) bool

type Summary struct{ ToPush, Behind, Dirty, Clean int }
func Summarize(repos []model.Repo, now time.Time) Summary
```

`StaleThreshold = 7 * 24 * time.Hour` exported const (eases the future config knob
and the test boundary).

## Risks & mitigations

| Risk | Mitigation |
| :--- | :--- |
| Changing the default sort surprises existing users | It's a refinement of Dirty-first, not a reorder of unrelated data; legacy sorts stay on `s`/`1`–`4`; README note. |
| `now` via hidden `time.Now()` makes stale tests flaky | `now` is an injected param throughout; only the TUI call site passes `time.Now()`. |
| Sub-score integer overflow on huge ahead/behind | Counts capped (`min(.,99)`) per spec before weighting. |
| Glyph width breaks the 8-wide `Status` column alignment | Use single-rune glyphs already used elsewhere (`▲▼●✓`); verify by manual run. |
| Summary double-counts diverged repos | Intentional per spec (per-signal, not a partition); documented in header semantics. |
| Scan-error repos sorting/classification | Covered: ScanError → ACTION tier + sub-score bump; explicit test. |

## Verification checkpoints

- **After step 1:** `go test ./internal/attention/... -cover` ≥ 90%; package
  builds with no `tui` import (verify `go list -deps`).
- **After step 2:** `go build ./...`; manual run shows ACTION repos on top at
  launch; `1`–`4` still switch sorts.
- **After step 3:** manual run shows tier glyph in `Status` column, aligned.
- **After step 4:** manual run shows `▲ N to push · ▼ N behind · ● N dirty ·
  ✓ N clean` updating on filter/`r`/`F`.
- **Before PR:** `go test ./...` + `golangci-lint run` clean; spec success
  criteria 1–7 each checked off.

## Out of scope (deferred)

- Config keys for weights / stale threshold (constants in v1).
- CLI/`--json`/CI exit-code consumption of `attention` (separate feature; the
  package is designed to enable it later).
- Stale visual marker and diverged glyph beyond the spec's proposed defaults.
