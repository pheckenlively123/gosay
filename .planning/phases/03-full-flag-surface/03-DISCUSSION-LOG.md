# Phase 3: Full Flag Surface - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-31
**Phase:** 3-full-flag-surface
**Areas discussed:** Word-wrap semantics, Eyes/tongue validation, Help output, Think-mode rendering

---

## Word-wrap semantics

### Long words exceeding the wrap width

| Option | Description | Selected |
|--------|-------------|----------|
| Hard-break mid-word | Split the long word across lines so the bubble never exceeds the width. Matches Perl cowsay; bubble stays a guaranteed rectangle. | ✓ |
| Let it overflow | Keep the long word intact; the bubble widens to fit it. Simpler but bubble can blow past `-W`. | |

**User's choice:** Hard-break mid-word
**Notes:** Hard-break is display-width aware (counts columns, not bytes), consistent with the runewidth work.

### Existing newlines in input

| Option | Description | Selected |
|--------|-------------|----------|
| Preserve, then wrap each line | Honor existing `\n` as hard breaks, then word-wrap within each line. Builds on current `strings.Split(message, "\n")`. `-n` keeps breaks, skips wrap. | ✓ |
| Reflow everything | Collapse input into one stream and re-wrap by width, ignoring original line breaks. | |

**User's choice:** Preserve, then wrap each line

---

## Eyes/tongue validation

| Option | Description | Selected |
|--------|-------------|----------|
| Pass through verbatim | Substitute whatever the user gives, any length. Matches upstream cowsay; zero validation. | ✓ |
| Normalize to 2 chars | Pad/truncate to exactly 2. Cleaner alignment but diverges from upstream. | |
| Reject non-2-char with error | Exit non-zero if not exactly 2 chars. Strictest; least playful. | |

**User's choice:** Pass through verbatim
**Notes:** Existing `RenderOpts`/`substituteVars` already substitutes verbatim — just wire the flags. Flagged a planner edge: `RenderOpts` treats empty string as "use default," so explicit `-e ""` is indistinguishable from unset (pathological; planner to decide consciously).

---

## Help output (-h/--help)

### Destination + exit code

| Option | Description | Selected |
|--------|-------------|----------|
| stdout, exit 0 | Explicit help is a success — stdout + exit 0, pipeable. Error usage stays stderr + non-zero. Requires intercepting `flag.ErrHelp`. | ✓ |
| stderr, exit 2 (Go default) | Let `flag` handle it: stderr, exit 2. Less work; `-h \| less` doesn't work. | |

**User's choice:** stdout, exit 0

### Help text shape

| Option | Description | Selected |
|--------|-------------|----------|
| Go-native block + examples | Synopsis line, each flag described, 2-3 example invocations. Reads naturally; satisfies the examples criterion. | ✓ |
| Mirror upstream cowsay -h | Reproduce Perl cowsay's help layout. Familiar but terse/dated, no examples. | |

**User's choice:** Go-native block + examples

---

## Think-mode rendering

### Border characters

| Option | Description | Selected |
|--------|-------------|----------|
| `( )` on every line | All bubble lines use `(`/`)` — single and multi-line — replacing `< >` and `/ | \`. Upstream cowthink behavior. | ✓ |
| `( )` single, keep `/ | \` multi | Use `( )` only for single-line; keep angled borders for multi-line. Diverges from cowthink. | |

**User's choice:** `( )` on every line

### Short alias

| Option | Description | Selected |
|--------|-------------|----------|
| Long-only `--think` | Keep just `--think`. `-t` reserved/excluded in REQUIREMENTS; upstream uses a separate `cowthink` command. | ✓ |
| Add `-t` alias | Also accept `-t`. Contradicts REQUIREMENTS Out-of-Scope note. | |

**User's choice:** Long-only `--think`

---

## Claude's Discretion

- Exact `RenderOpts` field shape for width/no-wrap/think and where the wrap logic lives (renderer/balloon/new helper).
- Exact greedy wrap algorithm and mid-word hard-break mechanics.
- `-W` edge cases (`-W 0`, negative, very large).
- East-Asian ambiguous-width handling (default to go-runewidth standard, ambiguous = 1).
- Exact help-text wording and which 2-3 examples to show.
- Whether to promote `displayWidth` to its own `width.go` file.

## Deferred Ideas

None — discussion stayed within Phase 3 scope. Phase 4 (release pipeline) and the Out-of-Scope mood/`-t` flags remain as scoped.
