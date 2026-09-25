# Tasks: Attention Summary + Scoring

Status: **Accepted**
Spec: `docs/specs/attention-summary-scoring.md`
Plan: `docs/specs/attention-summary-scoring-plan.md`

Tasks are dependency-ordered. Each is a single focused session, touches ≤5 files,
and has explicit acceptance + verification.

---

- [ ] **T1: `attention` package — Tier classification (TDD)**
  - Build: `internal/attention/attention.go` with `Tier` (Clean<Watch<Action),
    `String()`, `Glyph()`, `Classify(model.RepoStatus, now) Tier`, and
    `Stale(model.RepoStatus, now) bool` + `StaleThreshold` const.
  - Write `attention_test.go` first (failing): each tier; boundaries
    (Ahead 0→1, Behind 0→1); dirty combinations; ScanError→Action;
    diverged (Ahead>0 && Behind>0)→Action; Stale 7-day boundary with injected now.
  - Acceptance: tier rules match spec table exactly; scan error → ACTION;
    stale never changes tier.
  - Verify: `go test ./internal/attention/... -cover` passes, ≥90%.
  - Files: `internal/attention/attention.go`, `internal/attention/attention_test.go`.

- [ ] **T2: `attention` package — SubScore, Less, Summarize (TDD)**
  - Build: `SubScore(status, now) int` (caps + weights per spec:
    1000·ahead, 100·behind, 10·dirtyCount, 5·scanErr, 1·stale),
    `Less(a, b model.Repo, now) bool` (tier, then subScore desc, then name asc),
    `Summary{ToPush,Behind,Dirty,Clean}` + `Summarize([]model.Repo, now) Summary`.
  - Tests first (failing): unpushed outranks behind outranks dirty; stale nudge
    breaks a tie; caps prevent band overflow (ahead=200 stays in unpushed band);
    `Less` total-order sanity on a mixed slice; `Summarize` on mixed slice incl.
    diverged overlap and empty slice.
  - Acceptance: ordering priority unpushed→behind→dirty→stale holds; Summarize
    counts are per-signal (diverged repo counted in both ToPush and Behind).
  - Verify: `go test ./internal/attention/... -cover` ≥90%; `go list -deps
    ./internal/attention` shows **no** `internal/tui` import.
  - Files: `internal/attention/attention.go`, `internal/attention/attention_test.go`.

- [ ] **T3: Default sort = Attention**
  - Build: add `SortByAttention` to the `SortMode` iota; init `sortMode` to it;
    add `case SortByAttention` in `sortRepos()` using `attention.Less(.., time.Now())`;
    include Attention in the `s` cycle; keep `1`–`4` mappings unchanged.
  - Acceptance: on launch ACTION repos sort to top; `s` cycles through Attention +
    the 4 legacy modes; `1`–`4` unchanged.
  - Verify: `go build ./...`; `go run ./cmd/git-scope` in a multi-repo dir →
    ACTION repos on top; press `1`–`4` and `s` to confirm legacy sorts still work.
  - Files: `internal/tui/model.go` (+ `internal/tui/update.go` if `s`/help text).

- [ ] **T4: Status-column tier glyph**
  - Build: render `Classify(...).Glyph()` in the `Status` column of the table row
    builder; add tier colors in `styles.go` (Action/Watch/Clean).
  - Acceptance: each row's `Status` cell shows `▲`/`●`/`✓` matching its tier,
    column alignment intact (8-wide).
  - Verify: `go build ./...`; `go run ./cmd/git-scope` → glyphs present, aligned,
    colored; a known-dirty repo shows `●`, an ahead/behind repo shows `▲`.
  - Files: `internal/tui/model.go`, `internal/tui/styles.go`.

- [ ] **T5: Summary rollup in header**
  - Build: extend `renderStats()` to render
    `▲ N to push · ▼ N behind · ● N dirty · ✓ N clean` from
    `attention.Summarize(m.filteredRepos, time.Now())`; reuse existing badge styles.
  - Acceptance: counts reflect the **filtered** set and update on filter (`f`),
    rescan (`r`), and bulk fetch (`F`).
  - Verify: `go build ./...`; `go run ./cmd/git-scope` → header shows rollup;
    toggle `f` and observe counts change; `F` updates behind count.
  - Files: `internal/tui/view.go`.

- [ ] **T6: Docs + final gate**
  - Build: flip spec/plan/tasks Status → Accepted; update README features list and
    note in the keyboard table that the **default sort is Attention** (and `s`
    includes it).
  - Acceptance: spec success criteria 1–7 all checked; README accurate.
  - Verify: `go test ./...` clean; `golangci-lint run` clean; `go list -deps
    ./internal/attention | grep -q internal/tui` returns nothing (no TUI dep);
    re-read spec §Success Criteria and tick each.
  - Files: `README.md`, the three `docs/specs/attention-*` files.

---

## Sequencing notes

- T1 → T2 are the pure core; merge-able and reviewable before any TUI change.
- T3 and T4 both edit `model.go` → do T3 then T4, not in parallel.
- T5 is independent of T3/T4 (only `view.go`) but logically follows so the demo
  shows glyph + header together.
- T6 closes the spec gate.

## Definition of done (whole feature)

All of spec §Success Criteria (1–7) pass; `go test ./...` and `golangci-lint run`
clean; default launch sorts by attention; header rollup + per-repo glyph render
correctly; `internal/attention` carries no TUI dependency and is ≥90% covered.
