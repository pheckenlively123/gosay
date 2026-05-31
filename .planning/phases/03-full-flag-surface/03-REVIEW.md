---
phase: 03-full-flag-surface
reviewed: 2026-05-31T00:00:00Z
depth: standard
files_reviewed: 9
files_reviewed_list:
  - cmd/gosay/main.go
  - cmd/gosay/main_test.go
  - internal/cowsay/balloon.go
  - internal/cowsay/balloon_test.go
  - internal/cowsay/golden_test.go
  - internal/cowsay/renderer.go
  - internal/cowsay/renderer_test.go
  - internal/cowsay/wrap.go
  - internal/cowsay/wrap_test.go
findings:
  critical: 1
  warning: 4
  info: 3
  total: 8
status: issues_found
---

# Phase 03: Code Review Report

**Reviewed:** 2026-05-31T00:00:00Z
**Depth:** standard
**Files Reviewed:** 9
**Status:** issues_found

## Summary

Reviewed the Phase 03 source changes: Unicode display-width handling (runewidth),
display-width-aware word wrapping, think-mode rendering, and the CLI flag surface
(`-e`, `-T`, `--think`, `-W`, `-n`, plus `-h`/`--help`). The flag wiring, conflict
detection (`-f`+`--random`, `-l` combinations), help-to-stdout routing, and CJK
balloon alignment are implemented correctly and are well covered by tests.

The blocking issue is in `hardBreak`: when the wrap width is smaller than a single
rune's display width (reachable via `gosay -W 1 漢字`), the function silently
discards the remainder of the message rather than emitting it, producing an empty
or truncated bubble. The "safety break" comment claims this skips one rune to avoid
an infinite loop, but the implementation aborts the entire loop and drops the rest
of the input — a data-loss bug, not a skip. Several warnings cover related
robustness and consistency gaps.

## Critical Issues

### CR-01: `hardBreak` silently drops message when width < rune display width

**File:** `internal/cowsay/wrap.go:102-104`
**Issue:** The inner loop builds a chunk rune-by-rune and breaks once `chunkW+rw > width`.
When the very first rune is wider than `width` (e.g. a CJK/emoji rune of display width 2
with `width == 1`), `chunk.Len() == 0` and the code executes `break`, which exits the
**outer** `for len(s) > 0` loop entirely. Every remaining rune in `s` is silently
discarded. The comment ("skip to prevent infinite loop") describes skipping a single
rune, but the code abandons the whole remaining string.

This is reachable from the CLI: `gosay -W 1 漢字`. Trace: `wrapMessage("漢字", 1)` →
`wrapWords("漢字", 1)`; word width 2 > 1 → `hardBreak("漢字", 1)` returns an empty
slice → the `for i, chunk := range chunks` loop in `wrapWords` (line 41-48) never
executes → `current` stays empty, `currentW` stays 0 → `wrapWords` returns `nil` →
`wrapMessage` joins to `""`. The user's entire message vanishes and an empty bubble is
rendered with exit code 0. The same loss applies to any single oversized word mid-line.

`TestHardBreak_RuneSafe` only tests width 2 (where each CJK rune fits), so this path is
untested.

**Fix:** Force progress by emitting at least one rune per outer iteration instead of
aborting. This preserves the message (the over-wide rune simply overflows its line,
which is the least-surprising behavior) and still guarantees termination:
```go
func hardBreak(s string, width int) []string {
	var result []string
	for len(s) > 0 {
		var chunk strings.Builder
		chunkW := 0
		remaining := s
		for len(remaining) > 0 {
			r, size := utf8.DecodeRuneInString(remaining)
			rw := displayWidth(string(r))
			// Always take at least one rune so we make progress even when a
			// single rune is wider than width; otherwise stop at the boundary.
			if chunk.Len() > 0 && chunkW+rw > width {
				break
			}
			chunk.WriteRune(r)
			chunkW += rw
			remaining = remaining[size:]
		}
		result = append(result, chunk.String())
		s = s[len(s)-len(remaining):]
	}
	return result
}
```
Add a test for `hardBreak("漢字", 1)` asserting both chars survive, and a `run`-level
test for `gosay -W 1 漢字` asserting the message is not lost.

## Warnings

### WR-01: `--random` can select a non-renderable cow and fail with no fallback

**File:** `cmd/gosay/main.go:140-147`
**Issue:** `--random` picks `names[rand.Intn(len(names))]` from `ListCows()` and passes it
straight to `Render`. `ListCows()` enumerates every embedded `.cow`. If any embedded cow
fails to parse (or `ListCows()` ever returns a name that `LoadCow` cannot resolve), the
subsequent `Render` returns an error and `gosay` exits 1 with a confusing
"unknown cowfile" message for a name the user never typed. There is also no guard that
`len(names) > 0`; an empty embed set would panic at `rand.Intn(0)`.

**Fix:** Guard the empty case explicitly, and surface a clear internal-error message
rather than the user-facing "unknown cowfile" path:
```go
names, err := cowsay.ListCows()
if err != nil { /* ... */ }
if len(names) == 0 {
	fmt.Fprintln(stderr, "gosay: no embedded animals available")
	return 1
}
animal = names[rand.Intn(len(names))]
```

### WR-02: Explicit `-W 0` and negative widths silently coerce to the default 40

**File:** `cmd/gosay/main.go:87`, `internal/cowsay/renderer.go:78-84`
**Issue:** `-W` defaults to `0`, and `Render` treats `Width <= 0 && !NoWrap` as "use 40".
Consequently an explicit `gosay -W 0 ...` (a plausible user request for "no wrap") and any
negative `-W` value both silently become width 40 rather than being rejected or honored.
There is no way to distinguish "unset" from "explicitly 0" here, and no validation rejects
nonsensical negatives. Behavior is surprising and undocumented.

**Fix:** Either validate `-W` (reject `< 1` with a usage error) or document that `0`/negative
means "default 40" in the help text. If `-W 0` should mean "no wrap", detect explicit
setting via `fs.Visit` (as already done for `-f`) and map it to `NoWrap`.

### WR-03: `formatCowList` wraps on byte length, not display width

**File:** `cmd/gosay/main.go:35`
**Issue:** The listing wrap test `len(line)+1+len(name) > wrapWidth` uses `len` (bytes). The
rest of the phase deliberately migrated to display-column measurement via `displayWidth`
(runewidth) for correctness with multi-byte names. Cow names are currently ASCII so this is
latent, but it is inconsistent with the phase's stated goal and will misbehave if a
non-ASCII cow name is ever embedded.

**Fix:** Use `displayWidth(line)+1+displayWidth(name)` for consistency with the rest of the
codebase, or add a comment noting names are guaranteed ASCII and the byte count is intentional.

### WR-04: Stdin path uses `os.Stdin` directly, bypassing the injected writers and untestable

**File:** `cmd/gosay/main.go:156,162`
**Issue:** `run` accepts injected `stdout`/`stderr` for testability but reads stdin from the
package-global `os.Stdin` (both `isTTY(os.Stdin)` and `io.ReadAll(os.Stdin)`). As the
`TestRun_NoArgs` comment acknowledges, this leaves the entire stdin branch (TTY-vs-pipe
detection, read-error handling, single-trailing-newline trim) un-unit-testable, and makes
`TestRun_NoArgs` non-deterministic (it accepts either exit 0 or 1 depending on the harness's
stdin). A test asserting "either of two outcomes" provides weak regression protection.

**Fix:** Add an `io.Reader` (and a TTY predicate) parameter to `run`, defaulting to `os.Stdin`
in `main`, so the stdin assembly and newline-trim logic can be tested deterministically.

## Info

### IN-01: Help text and actual flag set drift (`--random` and `-h` documented inconsistently)

**File:** `cmd/gosay/main.go:50-67` vs `84-91`
**Issue:** `helpText` is a hand-maintained string separate from the registered flags. It lists
`--random` and `--think` but the registered defaults differ in description wording (e.g. the
`-W` default note "(default 40)" lives only in the help text and the `IntVar` usage string,
not enforced anywhere). Hand-maintained help risks drifting from the real flag surface as
flags change. This is acceptable for a 5-flag tool but worth a note.

**Fix:** Consider a test that asserts every registered flag name appears in `helpText` to
catch drift.

### IN-02: `RenderOpts.Thoughts` is undocumented at the CLI and only reachable internally

**File:** `internal/cowsay/renderer.go:21`, `cmd/gosay/main.go:171-177`
**Issue:** `Render` honors an explicit `Thoughts` override (and tests cover it), but `main`
never sets it, so the "explicit Thoughts override" branch (`renderer.go:88`) is dead from the
CLI's perspective. Not a bug, but the field's only live producer is tests.

**Fix:** None required; optionally note in the struct doc that `Thoughts` is library-only and
not wired to a flag in Phase 03.

### IN-03: Magic constant `40` duplicated across help text and renderer

**File:** `cmd/gosay/main.go:60`, `internal/cowsay/renderer.go:80`
**Issue:** The default wrap width `40` appears as a literal in the renderer and as text "(default
40)" in two help strings. If the default changes, three sites must be updated in lockstep.

**Fix:** Define an exported `const DefaultWidth = 40` in the `cowsay` package and reference it
from both the renderer and (via `fmt`) the help/usage text.

---

_Reviewed: 2026-05-31T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
