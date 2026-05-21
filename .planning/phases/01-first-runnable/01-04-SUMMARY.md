---
phase: 01-first-runnable
plan: 04
subsystem: cli

tags: [go, embed, cowsay, goldie, golden-tests, cli]

requires:
  - phase: 01-first-runnable/01-01
    provides: Walking-skeleton main.go and module declaration; this plan replaces the stub body with a real renderer dispatch.
  - phase: 01-first-runnable/01-02
    provides: embed.FS via //go:embed cows/*.cow plus the unexported readCowFile() seam; gopher.cow drops into cows/ and is picked up automatically.
  - phase: 01-first-runnable/01-03
    provides: Render(animal, message, opts) (string, error) and RenderOpts{Eyes, Tongue, Thoughts}; main.go now calls this directly.
provides:
  - Hand-authored gopher.cow (user-pick Variant B) installed as the default animal.
  - cmd/gosay/main.go dispatches to cowsay.Render and prints the result to stdout — the binary now actually says hello as a gopher.
  - Goldie-driven golden test suite (6 active goldens + 1 skipped CJK) covering all four Phase 1 pitfall classes.
  - sebdah/goldie/v2 v2.8.0 added as the project's first (test-only) third-party dependency.
affects: [phase-02-input-and-cow-selection, phase-03-full-flag-surface, phase-04-release-pipeline]

tech-stack:
  added:
    - github.com/sebdah/goldie/v2 v2.8.0 (test-only)
  patterns:
    - Goldie pattern: `goldie.New(t, goldie.WithFixtureDir("testdata/golden")).Assert(t, name, []byte(out))`; regenerate fixtures with `go test -update ./...`.
    - Single CLI dispatch line in main.go: `out, err := cowsay.Render("gopher", message, cowsay.RenderOpts{})` — no flag.Parse() in Phase 1; that lands in Phase 3.
    - Skipped-test convention: `t.Skip("Phase 3 / RENDER-04 — go-runewidth required for CJK width")` keeps a placeholder for the deferred CJK golden without failing CI.

key-files:
  created:
    - internal/cowsay/cows/gopher.cow (hand-authored, user-picked, 20 lines)
    - internal/cowsay/golden_test.go (133 lines; 6 active golden tests + 1 skipped CJK)
    - internal/cowsay/testdata/golden/gopher_say_hello.golden
    - internal/cowsay/testdata/golden/gopher_say_multiline.golden
    - internal/cowsay/testdata/golden/default_say_hello.golden
    - internal/cowsay/testdata/golden/dragon_and_cow_say_hello.golden
    - internal/cowsay/testdata/golden/three_eyes_say_hello.golden
    - internal/cowsay/testdata/golden/non_eoc_say_hello.golden
    - go.sum (introduced by adding goldie)
  modified:
    - cmd/gosay/main.go (stub print → cowsay.Render dispatch)
    - go.mod (added require github.com/sebdah/goldie/v2 v2.8.0)
    - internal/cowsay/cows/NOTICE (appended gopher.cow attribution paragraph)
    - .gitignore (anchored `gosay` pattern to `/gosay` so it no longer matches cmd/gosay/)

key-decisions:
  - "Gopher variant: user picked Variant B ('tall standing, arms out') from 3 hand-drafted candidates surfaced via AskUserQuestion preview. D-04 user-pick gate honored."
  - "Goldie v2.8.0 chosen per CLAUDE.md and 01-RESEARCH.md §F-Q16; pinned exact version; go.sum committed for module-checksum integrity."
  - "CJK golden left skipped, not deleted — keeps a visible reminder that Phase 3 / RENDER-04 (runewidth) is the closure point."

patterns-established:
  - "Pitfall coverage by golden: default.cow exercises `\\\\` → `\\`, dragon-and-cow exercises `\\@` → `@`, three-eyes documents that gosay does NOT replicate Perl `chop()` eye-manipulation, non-eoc (synthetic fixture) exercises the dynamic heredoc terminator."
  - "Multi-line speech bubble convention: 1 line uses `< … >`, 2+ lines use `/ … \\` (top), `| … |` (middle), `\\ … /` (bottom). For exactly 2 lines only top and bottom rows are emitted (no middle `|` row)."

requirements-completed:
  - INPUT-01
  - COW-01
  - COWS-04
  - RENDER-01
  - RENDER-02
  - RENDER-03
  - RENDER-04

duration: 7m
completed: 2026-05-21
---

# Phase 1 / Plan 01-04: Wire-up + Golden Tests Summary

**Hand-authored gopher.cow (user-picked) wired through cowsay.Render in main.go, with goldie-driven golden coverage of the four Phase 1 pitfall classes (dynamic terminator, backslash unescape, `\@` unescape, three-eyes Perl divergence).**

## Performance

- **Duration:** ~7 min (excluding both user-checkpoint pauses)
- **Started:** 2026-05-21
- **Completed:** 2026-05-21
- **Tasks:** 5 (Tasks 1, 2, 3, 4 executed; Task 5 was the final human-verify checkpoint, satisfied by user approval)
- **Files created:** 9
- **Files modified:** 4

## Accomplishments

- `gosay "hello"` now prints a real gopher with a `< hello >` speech bubble — Phase 1 Goal #1 satisfied.
- `gopher.cow` is the default animal — Phase 1 Goal #2 satisfied.
- Heredoc parser + backslash unescape pitfalls are pinned by goldens — Phase 1 Goal #3 satisfied (`default`, `dragon-and-cow`, `non-eoc` goldens).
- Multi-line input renders with `/ \` / `\ /` borders (single-line uses `< >`) — Phase 1 Goal #4 satisfied (`gopher_say_multiline` golden).
- `cows/NOTICE` carries a gopher.cow attribution paragraph alongside the upstream cowsay attribution; `.gitattributes` LF rule was already in place from Plan 01-01.

## Task Commits

Each task was committed atomically (the 3-commit form below is the canonical sequence; Task 3 was verification-only with no file changes):

1. **Task 1: gopher.cow (user-pick Variant B) + NOTICE update** — `4909b66` (feat)
2. **Task 2: main.go → cowsay.Render dispatch + goldie dep + .gitignore anchor fix** — `7544f2b` (feat)
3. **Task 3: end-to-end smoke run** — no commit (verification only)
4. **Task 4: goldie golden tests + 6 seeded fixtures + 1 skipped CJK test** — `b9e6f54` (feat)

**Merge commit (orchestrator):** `9162a2f` — chore: merge executor worktree
**Plan SUMMARY commit:** to follow this write.

## Files Created/Modified

**Created:**
- `internal/cowsay/cows/gopher.cow` — hand-authored 20-line Variant B (round head outline, side ears, arms-out body); standard `$thoughts` / `$eyes` / `$tongue` placeholder positions.
- `internal/cowsay/golden_test.go` — `goldie.New(t, goldie.WithFixtureDir("testdata/golden"))` driver covering 6 active scenarios + 1 skipped CJK.
- `internal/cowsay/testdata/golden/*.golden` — 6 fixture files (gopher × 2, default, dragon-and-cow, three-eyes, non-eoc).
- `go.sum` — first appearance; locks goldie/v2 v2.8.0 and its transitive set.

**Modified:**
- `cmd/gosay/main.go` — replaced the walking-skeleton stub with `out, err := cowsay.Render("gopher", message, cowsay.RenderOpts{})` + `os.Stdout.WriteString(out)`. Argument shape unchanged (`os.Args[1:]` join, no `flag.Parse()`).
- `go.mod` — added `require github.com/sebdah/goldie/v2 v2.8.0`.
- `internal/cowsay/cows/NOTICE` — appended gopher.cow attribution paragraph.
- `.gitignore` — anchored bare `gosay` pattern to `/gosay` (root only), so it no longer matches the `cmd/gosay/` package directory.

## Decisions Made

- **Gopher art (Variant B):** user picked from 3 hand-drafted candidates presented as fenced previews. The other two variants (round oval, minimal compact) were discarded; no commits made for them.
- **Goldie API surface:** `goldie.New(t, goldie.WithFixtureDir("testdata/golden")).Assert(t, "scenario_name", []byte(out))`. Regenerate with `go test -update ./internal/cowsay/...`.
- **CJK test deliberately skipped (not deleted):** `t.Skip("Phase 3 / RENDER-04 — needs go-runewidth")` keeps the placeholder visible until Phase 3.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 — Pre-existing bug] `.gitignore` pattern `gosay` matched too aggressively**
- **Found during:** Task 2 (wire-up)
- **Issue:** The bare `gosay` line in `.gitignore` (added by Plan 01-01 to ignore the compiled binary) was treating `cmd/gosay/` as ignored too — `git add cmd/gosay/main.go` was rejected.
- **Fix:** Re-anchored to `/gosay` so only the root-level compiled binary is ignored.
- **Files modified:** `.gitignore`
- **Verification:** `git add cmd/gosay/main.go` succeeds; root-level `./gosay` is still ignored.
- **Committed in:** `7544f2b` (alongside Task 2)

---

**Total deviations:** 1 auto-fixed (Rule 1 — pre-existing bug fix).
**Impact on plan:** Necessary for the plan to commit at all. No scope creep.

## Issues Encountered

- None during this plan. The single checkpoint pause (gopher pick) and the final approval pause were both expected per the plan's `autonomous: false` + 2× `checkpoint:human-verify` design.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- **Phase 1 Goals 1–4 (per ROADMAP.md):** all four success criteria verified at the live binary level and locked by goldens.
- **Ready for Phase 2 (Input and Cow Selection):** `cowsay.Render` is the stable interface Phase 2 will call from new code paths (stdin reader, `-f` flag, `-l` listing, `--random`).
- **No blockers.** Phase 3's `displayWidth` upgrade (`utf8.RuneCountInString` → `go-runewidth.StringWidth`) is documented as a skipped CJK golden; not a Phase 2 prerequisite.

---
*Phase: 01-first-runnable*
*Completed: 2026-05-21*
