---
phase: 03-full-flag-surface
plan: "04"
subsystem: cmd/gosay
tags: [cli-flags, help, eyes, tongue, wrap, think, golden, RENDER-08, HELP-01]
dependency_graph:
  requires: ["03-03"]
  provides: ["RENDER-08", "HELP-01", "full-flag-surface-CLI"]
  affects: ["cmd/gosay/main.go", "cmd/gosay/main_test.go", "internal/cowsay/golden_test.go", "internal/cowsay/testdata/golden/custom_eyes_tongue.golden"]
tech_stack:
  added: []
  patterns: ["flag.ErrHelp interception with no-op fs.Usage", "verbatim flag pass-through via RenderOpts", "goldie golden fixture generation"]
key_files:
  created:
    - internal/cowsay/testdata/golden/custom_eyes_tongue.golden
  modified:
    - cmd/gosay/main.go
    - cmd/gosay/main_test.go
    - internal/cowsay/golden_test.go
decisions:
  - "flag.ErrHelp interception: set fs.Usage=func(){} no-op BEFORE fs.Parse to suppress auto stderr print (D-14 / Pitfall 2)"
  - "Help output goes to stdout with exit 0; parse errors go to stderr with exit 1 (D-13)"
  - "No -t alias for --think (D-12); think is long-only flag"
  - "-e '' is indistinguishable from 'not passed' — empty string maps to default 'oo' in substituteVars; D-06 sentinel tracking not needed (D-06 accepted)"
  - "Verbatim -e/-T pass-through: no validation, no truncation (D-05)"
metrics:
  duration: "~10 minutes"
  completed: "2026-05-31"
  tasks: 3
  files: 4
---

# Phase 3 Plan 4: Full Flag Surface CLI (RENDER-08, HELP-01) Summary

Wired the complete Phase 3 flag surface into the CLI: five new flags (-W, -n, --think, -e, -T), flag.ErrHelp interception printing full help to stdout with exit 0, and a custom eyes/tongue golden fixture confirming verbatim substitution.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Register -W/-n/--think/-e/-T, intercept ErrHelp, add helpText, thread RenderOpts | abe8a9d | cmd/gosay/main.go |
| 2 | Add CLI tests for -W/-n/--think/-e/-T/-h/--help | d21b76a | cmd/gosay/main_test.go |
| 3 | Add a custom eyes/tongue golden test and fixture | 3530094 | internal/cowsay/golden_test.go, custom_eyes_tongue.golden |

## What Was Built

**`cmd/gosay/main.go` flag surface (Task 1):**
- Registered five new flags: `-W` (int wrap width, default 0 → engine uses 40), `-n` (bool no-wrap), `--think` (bool think mode, long-only per D-12), `-e` (string eyes), `-T` (string tongue)
- No `-t` alias (D-12)
- Added `const helpText` package-level constant with synopsis, all flag descriptions, and 3 example invocations (`gosay hello`, `echo hi | gosay -f tux`, `gosay --think -e "^^" "thinking..."`)
- Set `fs.Usage = func(){}` no-op BEFORE `fs.Parse` to suppress automatic stderr print when `-h`/`--help` is passed (D-14 / Pitfall 2)
- Replaced parse-error path: `errors.Is(err, flag.ErrHelp)` → print `helpText` to stdout, return 0; other errors → one-line usage to stderr, return 1
- Replaced final `Render` call with populated `cowsay.RenderOpts{Eyes, Tongue, Width, NoWrap, Think}`

**`cmd/gosay/main_test.go` CLI tests (Task 2):**
9 new test functions covering all new flags plus the ErrHelp help path:
- `TestRun_Help_ExitsZero` / `TestRun_LongHelp_ExitsZero`: -h and --help exit 0, help on stdout, empty stderr
- `TestRun_Think_UsesBubble`: --think produces `( hello )` in output
- `TestRun_WFlag_CustomWidth`: -W 10 wraps "hello world" to multi-row output
- `TestRun_NFlag_DisablesWrap`: -n preserves 50-char line unwrapped
- `TestRun_EyesFlag_Custom`: -e ^^ shows verbatim `^^` in output
- `TestRun_TongueFlag_Custom`: -T exits 0 with non-empty output
- `TestRun_EyesAndTongue_DashTongue`: `-e XX -T=-- -- hello` passes XX eyes and -- tongue (D-05 verbatim)
- `TestRun_RandomThink_NotConflict`: --random --think composes freely (not a conflict)

**`internal/cowsay/golden_test.go` + fixture (Task 3):**
- Added `TestGolden_GopherCustomEyesTongue`: `Render("gopher", "hello", RenderOpts{Eyes: "^^", Tongue: "U "})` captured as golden
- Generated `custom_eyes_tongue.golden` via `go test ./internal/cowsay/ -args -update`; fixture shows `(^^)` eyes and `| U |` tongue in gopher body (replacing defaults `oo` and `  `)

## Verification

- `go test ./...` — all 21 cmd/gosay tests + all internal/cowsay tests pass
- `go build ./...` — clean build
- `gosay -h` / `gosay --help` → full help to stdout, exit 0 (asserted by TestRun_Help_ExitsZero, TestRun_LongHelp_ExitsZero)
- `gosay -e XX -T=-- -- hello` → XX eyes, -- tongue in output
- `gosay --think hello` → `( hello )` thought bubble
- `gosay -W 10 "hello world"` → two-row balloon at width 10
- `gosay -n xxxx...x(50)` → full 50-char line in one balloon row

## Deviations from Plan

None - plan executed exactly as written.

The plan lists Task 2 as `tdd="true"`. Since Task 1 (implementation) was completed first and Task 2 adds the tests, the tests passed immediately on the GREEN path without a RED phase. This is structurally expected: the plan's tasks are sequenced as impl-first (Task 1), test-second (Task 2), golden-third (Task 3) — not a TDD cycle in the traditional sense but a test-coverage addition after implementation.

## Known Stubs

None. All five flags are fully wired through to `RenderOpts` and the engine. The `-e ""` edge case (empty string maps to default "oo") is intentional per D-06 and documented in the decisions table.

## Threat Flags

None. The five new flag values flow into `RenderOpts` fields already covered by the threat model (T-03-08 through T-03-10). No new trust boundaries introduced.

## Self-Check: PASSED

| Check | Result |
|-------|--------|
| cmd/gosay/main.go | FOUND |
| cmd/gosay/main_test.go | FOUND |
| internal/cowsay/golden_test.go | FOUND |
| internal/cowsay/testdata/golden/custom_eyes_tongue.golden | FOUND |
| Commit abe8a9d (Task 1 feat) | FOUND |
| Commit d21b76a (Task 2 test) | FOUND |
| Commit 3530094 (Task 3 golden) | FOUND |
