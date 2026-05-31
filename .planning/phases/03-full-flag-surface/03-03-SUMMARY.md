---
phase: 03-full-flag-surface
plan: "03"
subsystem: internal/cowsay
tags: [think-mode, balloon, renderer, golden, tdd, RENDER-07]
dependency_graph:
  requires: ["03-02"]
  provides: ["RENDER-07", "think-mode-balloon", "think-mode-thoughts-trail"]
  affects: ["internal/cowsay/balloon.go", "internal/cowsay/renderer.go"]
tech_stack:
  added: []
  patterns: ["TDD RED/GREEN cycle", "goldie golden fixture generation"]
key_files:
  created:
    - internal/cowsay/testdata/golden/think_say_hello.golden
  modified:
    - internal/cowsay/balloon.go
    - internal/cowsay/renderer.go
    - internal/cowsay/renderer_test.go
    - internal/cowsay/golden_test.go
    - internal/cowsay/balloon_test.go
decisions:
  - "Think branch uses ( ) on every line regardless of line count (D-10 confirmed against upstream Perl construct_balloon @border qw[ ( ) ( ) ( ) ])"
  - "Thoughts='o' is filled only when opts.Think && opts.Thoughts == '' — explicit override is respected"
  - "Thoughts fill happens before substituteVars to ensure the correct character reaches the replacer"
metrics:
  duration: "~6 minutes"
  completed: "2026-05-31"
  tasks: 3
  files: 5
---

# Phase 3 Plan 3: Think Mode (RENDER-07) Summary

Implemented think-mode thought bubble: `( )` borders on every content line and `o` thought-trail character, filling the RENDER-07 requirement. `Render("gopher", "hello", RenderOpts{Think: true})` now produces a complete thought bubble with an `o` trail connecting bubble to gopher.

## Tasks Completed

| Task | Name | Commits | Files |
|------|------|---------|-------|
| 1 | Implement think-mode ( ) border branch in buildBalloon | 3d200ed (RED), 87ec6be (GREEN) | balloon.go, balloon_test.go |
| 2 | Add RenderOpts.Think and thread think through Render | a9e0730 (RED), db9bc78 (GREEN) | renderer.go, renderer_test.go |
| 3 | Add think-mode golden test and fixture | 5bf8f79 | golden_test.go, think_say_hello.golden |

## What Was Built

**`buildBalloon` think branch (balloon.go):** Replaced the Plan-02 placeholder that fell through to say-mode with the real implementation. When `think=true`, every line (regardless of count) is wrapped in `( )` borders using `padRight` for display-width alignment. Say-mode single-line (`< >`) and multi-line (`/ | \`) branches are untouched.

**`Render` think threading (renderer.go):** Added `if opts.Think && opts.Thoughts == "" { opts.Thoughts = "o" }` before `substituteVars`. This sets the `$thoughts` trail character to `o` for think mode (D-11). Explicit `Thoughts` overrides are preserved. `RenderOpts.Think bool` was already present from Plan 02.

**Golden fixture (think_say_hello.golden):** Generated via `go test -update`. Shows `( hello )` thought bubble with `o` trail connecting to gopher. No backslash trail present.

## Verification

- `go test ./internal/cowsay/ -run TestBuildBalloon` — 10 cases pass (7 say + 3 new think cases)
- `go test ./internal/cowsay/ -run TestRender` — all renderer tests pass including new `TestRender_ThinkMode`
- `go test ./internal/cowsay/...` — full suite green (includes CJK golden, wrap golden, all existing goldens)
- `go build ./...` — clean build

## Deviations from Plan

**None — plan executed exactly as written.**

The test for "o trail in think output" required careful scoping: `strings.Contains(thinkOut, "o")` would produce a false positive because `(oo)` appears in the cow body. The test was written to assert the specific trail pattern `"o   ^__^"` from the default.cow template. This was not a deviation — just careful test authoring within the plan's intent.

## TDD Gate Compliance

Task 1:
- RED commit: `3d200ed` — `test(03-03): add failing think-mode balloon tests (RED)`
- GREEN commit: `87ec6be` — `feat(03-03): implement think-mode ( ) border branch in buildBalloon`

Task 2:
- RED commit: `a9e0730` — `test(03-03): add failing TestRender_ThinkMode test (RED)`
- GREEN commit: `db9bc78` — `feat(03-03): add Thoughts='o' threading in Render for think mode (D-11)`

## Known Stubs

None. The `Think` field is fully wired: `buildBalloon` think branch is implemented and `Render` sets `Thoughts="o"`. The `--think` CLI flag (Plan 04) will pass `RenderOpts{Think: true}` to `Render` — no stub in the render path.

## Threat Flags

None. The think branch reuses `padRight` (same display-width safety as say branches). No new trust boundaries introduced.

## Self-Check: PASSED

All created/modified files verified present on disk. All 5 task commits found in git log.

| Check | Result |
|-------|--------|
| internal/cowsay/balloon.go | FOUND |
| internal/cowsay/balloon_test.go | FOUND |
| internal/cowsay/renderer.go | FOUND |
| internal/cowsay/renderer_test.go | FOUND |
| internal/cowsay/golden_test.go | FOUND |
| internal/cowsay/testdata/golden/think_say_hello.golden | FOUND |
| .planning/phases/03-full-flag-surface/03-03-SUMMARY.md | FOUND |
| Commit 3d200ed (RED balloon tests) | FOUND |
| Commit 87ec6be (GREEN balloon impl) | FOUND |
| Commit a9e0730 (RED renderer test) | FOUND |
| Commit db9bc78 (GREEN renderer impl) | FOUND |
| Commit 5bf8f79 (golden test + fixture) | FOUND |
