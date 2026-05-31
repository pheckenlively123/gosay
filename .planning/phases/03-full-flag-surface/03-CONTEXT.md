# Phase 3: Full Flag Surface - Context

**Gathered:** 2026-05-31
**Status:** Ready for planning

<domain>
## Phase Boundary

Complete the gosay flag surface on top of the working Phase 1 render engine and Phase 2 input/selection layer. Phase 3 delivers the remaining user-facing flags and the Unicode-correctness fix:

- **Word wrap** — wrap the message to a column width; `-W <n>` overrides the default 40; `-n` disables wrapping (RENDER-05)
- **Display-width sizing** — bubble width measured in display columns via `runewidth.StringWidth`, replacing the Phase 1 rune-count seam, so CJK/emoji/combining characters align (RENDER-06)
- **Think mode** — `--think` swaps the speech bubble for a thought bubble (`( )` borders, `o` thought trail) (RENDER-07)
- **Eyes/tongue** — `-e <xx>` overrides eyes (default `oo`), `-T <xx>` overrides tongue (default two spaces) (RENDER-08)
- **Help** — `-h`/`--help` prints full usage with every flag documented plus example invocations (HELP-01)

`main.go` stays thin and continues to route all rendering through `cowsay.Render`. The wrap/think/eyes/tongue options flow into `cowsay.RenderOpts` (which already carries `Eyes`/`Tongue`/`Thoughts`); new option fields are added for width/no-wrap/think as needed.

**Out of Phase 3** (deferred to later phases by ROADMAP):
- Release pipeline, GoReleaser, GitHub Actions, `--version`, `go install` → Phase 4

</domain>

<decisions>
## Implementation Decisions

### Word Wrap (RENDER-05)
- **D-01:** Default wrap width is **40 columns**; `-W <n>` overrides it; `-n` disables word-wrapping entirely. (Per ROADMAP success criterion #1 and locked REQUIREMENTS RENDER-05.)
- **D-02:** **Long words that exceed the wrap width are hard-broken mid-word** (Perl-cowsay behavior) so the bubble is always a guaranteed rectangle no wider than the requested width. No overflow mode.
- **D-03:** Wrapping is **display-width aware** — both the wrap boundary and the mid-word hard-break count display columns (`runewidth.StringWidth`), not bytes or runes. This shares the same width primitive as D-06 so a CJK string wraps and sizes consistently.
- **D-04:** **Existing newlines are preserved, then each resulting line is wrapped to the width.** Input is split on `\n` first (building on the current `strings.Split(message, "\n")` foundation in `balloon.go`), then word-wrap applies within each line. `-n` keeps the explicit line breaks but skips the word-wrapping step.

### Eyes / Tongue (RENDER-08)
- **D-05:** `-e <eyes>` and `-T <tongue>` **pass through verbatim — any length, no validation, no truncation.** `-e XXX` widens the face, `-e X` narrows it; the user owns the visual consequences. Matches upstream cowsay exactly and requires zero validation code (the existing `substituteVars` already substitutes whatever it is handed).
- **D-06 (eyes/tongue edge):** ⚠ Planner note — `RenderOpts` currently treats an empty `Eyes`/`Tongue`/`Thoughts` string as "use the default" (`renderer.go` default-fills empties). This makes an *explicit* `-e ""` indistinguishable from "flag not passed." This is a pathological case; acceptable to leave as-is, but the planner should consciously decide (e.g., a sentinel/`*string` or "explicitly set" tracking) rather than discover it by accident.

### Display Width / runewidth (RENDER-06)
- **D-07:** Swap the body of the existing `displayWidth(s string) int` seam in `internal/cowsay/balloon.go` from `utf8.RuneCountInString(s)` to `runewidth.StringWidth(s)`. The seam was built in Phase 1 (D-16/D-18) precisely so this is a one-line body swap with zero call-site churn. `github.com/mattn/go-runewidth` is **the one approved external dependency** for this phase (ROADMAP decision).
- **D-08:** Fix the byte-padding bug alongside the width swap: `balloon.go` currently pads with `fmt.Fprintf("%-*s", maxWidth, line)`, and Go's `%-*s` pads by **byte** width, not display width. Padding must be computed in display columns (`maxWidth - displayWidth(line)` trailing spaces, or equivalent) so the right border aligns for CJK/emoji. Both the width measurement *and* the padding must move to display-width for the bubble to actually align.
- **D-09:** Remove the Phase 1 `t.Skip(...)`-marked CJK golden test and replace it with a real golden that asserts the correctly-aligned bubble for `漢字テスト` (per ROADMAP success criterion #4 and Phase 1 D-19).

### Think Mode (RENDER-07)
- **D-10:** `--think` renders a **thought bubble**: the bubble border uses `(` on the left and `)` on the right of **every** content line — both single-line and every line of multi-line input (replacing `< >` and the angled `/ | \`). Top/bottom borders remain underscores/dashes. This is upstream `cowthink` behavior.
- **D-11:** `--think` sets `RenderOpts.Thoughts = "o"` (vs the say-mode default `\`), driving the `$thoughts` substitution already wired in `renderer.go` so the trail of thought characters connects bubble to cow.
- **D-12:** `--think` is **long-only**; no `-t` alias. `-t` is explicitly reserved in REQUIREMENTS Out-of-Scope (mood-flag collision), and upstream ships a separate `cowthink` command rather than a short flag, so there is no canonical short form to honor.

### Help (HELP-01)
- **D-13:** Explicit `-h`/`--help` prints the **full usage to stdout and exits 0** (a help request is a success, and `gosay -h | less` should work). Error-triggered usage (bad flag, conflicting flags, bare interactive invocation) continues to go to **stderr with a non-zero exit**, as in Phase 2.
- **D-14:** Implementation: the `flag` package (in `ContinueOnError` mode) returns `flag.ErrHelp` when `-h`/`-help` is requested. `run()` must `errors.Is(err, flag.ErrHelp)` after `fs.Parse`, print the full help to stdout, and `return 0` — replacing the current blanket `return 1` on any parse error.
- **D-15:** Help text is a **clean Go-native usage block**: a synopsis line, each flag with a one-line description, then 2–3 example invocations (e.g. `gosay hello`, `echo hi | gosay -f tux`, `gosay --think -e ^^ "thinking"`). It does NOT reproduce Perl cowsay's help layout. Examples are mandatory (ROADMAP success criterion #5). The existing one-line `fs.Usage` stderr string (Phase 2) is expanded/reused for the error path; the full block is the `-h` path.

### Claude's Discretion
- Exact `RenderOpts` field shape for width/no-wrap/think (e.g. `Width int`, `NoWrap bool`, `Think bool`) and whether wrapping lives in the renderer, the balloon builder, or a small `wrap.go` helper — keep `main.go` thin and the public API minimal.
- Exact wrap algorithm (greedy word packing via `strings.Fields` + per-word display-width accounting) and how the mid-word hard-break splits a too-long word.
- `-W` edge cases (`-W 0`, negative, or absurdly large) — pick sane behavior; not worth a user decision on a toy.
- East-Asian *ambiguous*-width characters: default to `go-runewidth`'s standard (ambiguous = width 1; do not enable `EastAsianWidth`) unless a golden test reveals a concrete problem.
- Exact help-text wording, flag-description phrasing, and which 2–3 examples to show (within D-15 intent).
- Whether to promote `displayWidth` to its own `width.go` file now that a second consideration (wrap) shares it (Phase 1 D-17 left this open).

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project context
- `.planning/PROJECT.md` — vision, scope, constraints, Key Decisions table
- `.planning/REQUIREMENTS.md` — v1 requirements; Phase 3 covers RENDER-05, RENDER-06, RENDER-07, RENDER-08, HELP-01. Note the Out-of-Scope table excludes mood-preset flags and the `-t` alias (relevant to D-12).
- `.planning/ROADMAP.md` §"Phase 3: Full Flag Surface" — phase goal and the 5 success criteria the verifier will check.

### Prior phase context (carry-forward)
- `.planning/phases/01-first-runnable/01-CONTEXT.md` — establishes the `displayWidth` seam (D-16/D-17/D-18/D-19), the `$eyes`/`$tongue`/`$thoughts` placeholder layout in `gopher.cow`, and the golden-test convention Phase 3 extends.
- `.planning/phases/02-input-and-cow-selection/02-CONTEXT.md` — establishes the `flag`-based thin `main.go`, the `run(args, stdout, stderr) int` testable harness, and the usage/exit-code conventions Phase 3's help path must stay consistent with.

### Domain research (HIGH-confidence)
- `.planning/research/STACK.md` — `flag` stdlib + `sebdah/goldie` v2; confirms `mattn/go-runewidth` posture and what NOT to add (no Cobra/Kong, no go-wordwrap — hand-roll the ~30-line wrap).
- `.planning/research/ARCHITECTURE.md` — package layout and the thin-`main.go` / `internal/cowsay`-seam pattern the new options must respect.
- `.planning/research/PITFALLS.md` — rendering/width traps (the `%-*s` byte-vs-display padding pitfall behind D-08 lives here).
- `.planning/research/FEATURES.md` — feature dependency order (Phase 3 = full flag surface).

### Existing code (the integration surface)
- `internal/cowsay/balloon.go` — `buildBalloon(message string)` and the `displayWidth` seam; the site of D-03, D-07, D-08, D-10 (border swap for think mode) and the new wrap step.
- `internal/cowsay/renderer.go` — `Render(animal, message string, opts RenderOpts) (string, error)` and `RenderOpts{Eyes, Tongue, Thoughts}` + `substituteVars`; the place to add width/no-wrap/think option fields and thread think → `Thoughts="o"`.
- `cmd/gosay/main.go` — the thin `run()` CLI; the place to register `-W`/`-n`/`--think`/`-e`/`-T`/`-h`, intercept `flag.ErrHelp` (D-14), and expand the help block (D-15).
- `internal/cowsay/golden_test.go` / `internal/cowsay/testdata/golden/` — where the CJK golden (D-09) and any think/wrap/eyes goldens land.

### Upstream reference
- `github.com/cowsay-org/cowsay` — reference for wrap-at-40 + hard-break-long-word behavior, `cowthink` `( )` bubble shape, and `-e`/`-T` verbatim substitution.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- **`displayWidth(s) int` seam** in `balloon.go` — purpose-built in Phase 1 for exactly this swap (D-07). Both the bubble sizing and the new wrap logic (D-03) call it, so the runewidth upgrade fixes both at once.
- **`RenderOpts{Eyes, Tongue, Thoughts}`** in `renderer.go` — already substituted by `substituteVars` with default-filling. `-e`/`-T` (D-05) and `--think`→`Thoughts="o"` (D-11) just populate these fields; no substitution code changes needed.
- **`run(args, stdout, stderr) int`** in `main.go` — the testable harness; Phase 3 drives it with `-W`/`-n`/`--think`/`-e`/`-T`/`-h` arg combinations and asserts stdout/stderr/exit code without `os.Exit`.

### Established Patterns
- Thin `main.go` over a minimal `internal/cowsay` surface — Phase 3 adds option fields to the existing `RenderOpts`, not a new public API.
- Golden-file tests via `sebdah/goldie` under `internal/cowsay/testdata/golden/` for render output (think bubble, wrapped output, CJK alignment); CLI-level behavior (help to stdout+exit 0, flag conflicts, exit codes) tested in `cmd/gosay/main_test.go`.

### Integration Points
- `main.go` ↔ `flag`: new flags `-W` (int, default 40), `-n` (bool), `--think` (bool), `-e` (string), `-T` (string), `-h`/`--help` (intercept `flag.ErrHelp`).
- `main.go` ↔ `cowsay.RenderOpts`: thread width/no-wrap/think + eyes/tongue from parsed flags into the opts struct.
- `balloon.go` internal: wrap step (display-width aware, hard-break) runs before bubble sizing; think mode selects `( )` borders on all lines.

</code_context>

<specifics>
## Specific Ideas

- **Match upstream cowsay where it's the established behavior, diverge only where it helps the toy.** Kept faithful: wrap-at-40 + hard-break long words (D-02), `( )` thought bubble on every line (D-10), verbatim `-e`/`-T` (D-05). Deliberately Go-native: help to stdout + exit 0 with example invocations (D-13/D-15) instead of Perl cowsay's terse stderr help.
- **The runewidth fix is two changes, not one.** Measuring with `runewidth.StringWidth` (D-07) is necessary but insufficient — the `%-*s` padding also pads by bytes (D-08). Both must land together or the CJK golden (D-09) won't align.
- **The width primitive is shared.** Wrap boundary (D-03) and bubble sizing (D-07) must use the same `displayWidth` so a string wraps and then sizes to a consistent column count.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within Phase 3 scope. The release pipeline (GoReleaser, GitHub Actions, `--version`, `go install`) remains Phase 4 as scoped in ROADMAP.md and REQUIREMENTS.md. Mood-preset flags and a `-t` alias remain explicitly Out-of-Scope in REQUIREMENTS.md (reaffirmed by D-12).

</deferred>

---

*Phase: 3-full-flag-surface*
*Context gathered: 2026-05-31*
