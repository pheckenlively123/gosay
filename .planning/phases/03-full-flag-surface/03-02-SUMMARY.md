---
phase: 03-full-flag-surface
plan: "02"
subsystem: internal/cowsay
tags: [wrap, display-width, unicode, tdd, renderer]
dependency_graph:
  requires: ["03-01"]
  provides: ["wrapMessage", "wrapWords", "hardBreak", "RenderOpts.Width", "RenderOpts.NoWrap", "buildBalloon([]string, bool)"]
  affects: ["internal/cowsay/renderer.go", "internal/cowsay/balloon.go"]
tech_stack:
  added: []
  patterns: ["TDD RED-GREEN-REFACTOR", "display-width-aware greedy wrap", "hardBreak rune-boundary splitting"]
key_files:
  created:
    - internal/cowsay/wrap.go
    - internal/cowsay/wrap_test.go
    - internal/cowsay/testdata/golden/wrap_long_message.golden
  modified:
    - internal/cowsay/balloon.go
    - internal/cowsay/renderer.go
    - internal/cowsay/renderer_test.go
    - internal/cowsay/balloon_test.go
    - internal/cowsay/golden_test.go
decisions:
  - "buildBalloon signature changed to (lines []string, think bool) — think branch stubs to say-mode until Plan 03 fills it"
  - "Width<=0 with !NoWrap resolves to 40-col default in Render; NoWrap=true passes 0 to wrapMessage (passthrough)"
  - "hardBreak advances by utf8.DecodeRuneInString size, never by display width — rune-boundary safety guaranteed"
metrics:
  duration: "~20m"
  completed: "2026-05-31"
  tasks: 3
  files: 8
---

# Phase 3 Plan 02: Word Wrap (RENDER-05) Summary

**One-liner:** Display-width-aware greedy word wrap with hard-break mid-word splitting at rune boundaries, wired into Render with 40-col default and NoWrap override.

## What Was Built

### Task 1: wrap.go + wrap_test.go (TDD)
Created `internal/cowsay/wrap.go` with three functions in `package cowsay`:

- `wrapMessage(message string, width int) string` — passthrough when `width <= 0`; splits on `\n`, wraps each line via `wrapWords`, rejoins with `\n`
- `wrapWords(line string, width int) []string` — greedy word-packing using `strings.Fields` and `displayWidth`; words exceeding width are hard-broken via `hardBreak`; empty/whitespace-only lines return `[]string{""}` to preserve blank lines
- `hardBreak(s string, width int) []string` — splits into chunks of at most `width` display columns; advances by `utf8.DecodeRuneInString` byte size (never by display width) for guaranteed rune-boundary safety; single rune wider than width causes safety break to prevent infinite loop

All three functions reuse the package-level `displayWidth` from `balloon.go` — no duplication.

`wrap_test.go` covers: `width=0` passthrough, `width=-1` passthrough, fits-on-one-line, word-boundary wrap, hard-break, existing newline preservation, CJK display-col splitting, empty string. `TestHardBreak_RuneSafe` asserts `utf8.ValidString` on every chunk and `displayWidth <= 2` for `"漢字漢字"` at `width=2`.

### Task 2: buildBalloon signature change + Render wrap threading (TDD)
Changed `buildBalloon(message string) string` → `buildBalloon(lines []string, think bool) string`:
- Caller (Render) now provides pre-split, pre-wrapped lines
- `think bool` parameter added; placeholder branch for Plan 03's `( )` borders; falls through to say-mode for now

Added `Width int` and `NoWrap bool` to `RenderOpts`:
- `Width`: 0 = use default 40; ignored when `NoWrap` is true
- `NoWrap`: skip wrapping entirely, pass 0 to `wrapMessage`

Wired wrap step in `Render`: resolves effective `wrapWidth` (defaults to 40), calls `wrapMessage`, splits result on `\n`, passes `[]string` to `buildBalloon`.

Updated `balloon_test.go` — all `buildBalloon(tc.input)` calls changed to `buildBalloon(strings.Split(tc.input, "\n"), false)` (with `"strings"` import). Updated `golden_test.go`'s direct `buildBalloon("hello")` call similarly. All 7 existing balloon test cases pass unchanged.

### Task 3: Wrapped golden test and fixture
Added `TestGolden_GopherWrap` to `golden_test.go`. Fixture `wrap_long_message.golden` generated via `go test -update`. Shows a 2-line bubble breaking `"a long message that should wrap at forty columns"` across lines — first content line `"a long message that should wrap at forty"` (40 display cols), second line `"columns"`.

## Deviations from Plan

None - plan executed exactly as written.

## Commits

| Hash | Type | Description |
|------|------|-------------|
| 020f473 | test | add failing tests for wrapMessage and hardBreak (RED) |
| 80b1629 | feat | implement wrapMessage, wrapWords, hardBreak in wrap.go (GREEN) |
| 08f2a31 | test | add failing tests for Render wrap integration (RED) |
| a3f285e | feat | change buildBalloon signature and thread wrap step into Render (GREEN) |
| 4adea22 | feat | add wrap golden test and fixture for 40-col wrapped bubble |

## Verification

- `go test ./internal/cowsay/...` — 42 tests, all PASS
- `go build ./...` — CLEAN
- `go vet ./internal/cowsay/` — CLEAN
- `grep -n 'utf8.DecodeRuneInString' internal/cowsay/wrap.go` — found inside `hardBreak` (rune-boundary advance confirmed)
- `grep -n 'wrapMessage' internal/cowsay/renderer.go` precedes `grep -n 'buildBalloon'` — wrap-before-balloon ordering confirmed

## Known Stubs

- `buildBalloon` `think=true` branch: falls through to say-mode borders (no `( )` rendering yet). Plan 03 fills this branch. Intentional — stub exists so the signature is final and callers compile.

## Threat Flags

None — no new trust boundaries introduced. `wrapMessage`/`hardBreak` loop bounds are O(n) over message length (T-03-03 / T-03-04 mitigations confirmed in implementation).

## Self-Check: PASSED

Files created/exist:
- internal/cowsay/wrap.go: FOUND
- internal/cowsay/wrap_test.go: FOUND
- internal/cowsay/testdata/golden/wrap_long_message.golden: FOUND

Commits exist:
- 020f473: FOUND
- 80b1629: FOUND
- 08f2a31: FOUND
- a3f285e: FOUND
- 4adea22: FOUND
