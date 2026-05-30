# Phase 2: Input and Cow Selection - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-30
**Phase:** 02-input-and-cow-selection
**Areas discussed:** Input resolution, --random scope & testing, -l list format, Unknown-cow error UX

---

## Input resolution

### No-input case (no args, nothing piped, interactive terminal)

| Option | Description | Selected |
|--------|-------------|----------|
| Block on stdin (cowsay) | Wait for typed input, end on Ctrl-D, then render. Faithful but surprising. | |
| Print usage to stderr, exit 1 | Keep Phase 1 behavior; friendlier for accidental runs. | ✓ |
| Empty bubble, exit 0 | Treat no input as empty input; never blocks. | |

**User's choice:** Print usage to stderr, exit 1.
**Notes:** Implies a stdin TTY check — piped stdin (INPUT-02) is still read; only a bare interactive run prints usage. Real-cowsay blocking behavior explicitly rejected.

### Empty bubble shape (INPUT-04)

| Option | Description | Selected |
|--------|-------------|----------|
| cowsay-style empty bubble | Top/bottom borders with single-space `<  >` interior. | ✓ |
| Minimal collapsed bubble | Tightest `<>` with no interior space. | |

**User's choice:** cowsay-style empty bubble.
**Notes:** Covers both `gosay ""` and empty piped stdin; exit 0.

---

## --random scope & testing

### Random pool

| Option | Description | Selected |
|--------|-------------|----------|
| All embedded (full 51) | Every cow incl. gopher and daemon.cow; reuse ListCows(). | ✓ |
| All except gopher | Always surprise with a non-default animal. | |
| Exclude provenance-flagged | Drop daemon.cow via a denylist. | |

**User's choice:** All embedded (full 51).

### RNG testability

| Option | Description | Selected |
|--------|-------------|----------|
| Testable seam | Injectable RNG/seed for deterministic tests. | |
| No seam | Package-level math/rand; tests assert result is a valid cow name. | ✓ |

**User's choice:** No seam.

### -f + --random conflict

| Option | Description | Selected |
|--------|-------------|----------|
| Error, exit non-zero | Mutually exclusive; clear message. | ✓ |
| --random wins | Ignore -f silently. | |
| -f wins | Ignore --random silently. | |

**User's choice:** Error, exit non-zero.

---

## -l list format

### Output format

| Option | Description | Selected |
|--------|-------------|----------|
| One name per line, sorted | Plain, script-friendly. | |
| Upstream columnar | `Cow files:` header + wrapped space-separated columns. | ✓ |

**User's choice:** Upstream columnar.
**Notes:** Conflicted with locked ROADMAP success criterion #3 ("one per line"). Surfaced to user; user chose to amend ROADMAP/REQUIREMENTS to the columnar format rather than revert. Edit landed with this discuss-phase commit.

### Gopher in listing

| Option | Description | Selected |
|--------|-------------|----------|
| Plain, alphabetical | gopher is just another name (sorts under 'g'), unmarked. | ✓ |
| Marked as default | Annotate e.g. `gopher (default)`. | |

**User's choice:** Plain, alphabetical.

### -l combined with message / other flags

| Option | Description | Selected |
|--------|-------------|----------|
| Short-circuit, exit 0 | Print list, ignore message and selection flags. | |
| Error if combined | Reject -l mixed with message/-f/--random as usage error. | ✓ |

**User's choice:** Error if combined.

---

## Unknown-cow error UX

### Error wording

| Option | Description | Selected |
|--------|-------------|----------|
| Clean + hint | `gosay: unknown cowfile "x"` plus `(try 'gosay -l')`. | |
| Clean, no hint | `gosay: unknown cowfile "x"`, exit non-zero. | ✓ |
| Keep wrapped error | Surface Render()'s error; leaks `cows/x.cow` path. | |

**User's choice:** Clean, no hint.

### Check location

| Option | Description | Selected |
|--------|-------------|----------|
| Pre-check in main.go | Validate -f against ListCows() before Render. | |
| Translate Render's error | Map Render's not-found error (sentinel + errors.Is) into the clean message. | ✓ |

**User's choice:** Translate Render's error.
**Notes:** Implies exporting a sentinel error (e.g. `cowsay.ErrUnknownCow`) wrapped by LoadCow on `fs.ErrNotExist`.

---

## Claude's Discretion

- `flag` stdlib mechanics (short/long flag registration, `flag.Usage` override for the usage string)
- stdin read via `io.ReadAll` and trailing-newline trimming to match upstream
- TTY detection via `os.Stdin.Stat()` / `os.ModeCharDevice`
- Exact conflict-message wording and `-l` column width/wrap logic
- Sentinel error name and any new unexported helpers

## Deferred Ideas

None — discussion stayed within Phase 2 scope.
</content>
