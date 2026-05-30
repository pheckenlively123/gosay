# Phase 2: Input and Cow Selection - Context

**Gathered:** 2026-05-30
**Status:** Ready for planning

<domain>
## Phase Boundary

Build the input and animal-selection layer on top of the Phase 1 rendering engine. This is the phase where `flag` parsing enters `main.go` (Phase 1 read `os.Args[1:]` directly). Phase 2 delivers:

- **Input** — message from positional args OR piped stdin; positional args win when both present (INPUT-03); empty input renders a valid empty bubble (INPUT-04)
- **Selection** — `-f <name>` picks an embedded animal (COW-02); `-l` lists all animals (COW-03); `--random` picks one at random (COW-04); unknown `-f` exits non-zero with a clean error (COW-05)
- `main.go` migrates to `flag` (stdlib) and stays thin — all rendering still goes through the existing `cowsay.Render` / `cowsay.ListCows` seam
- One small addition to the `internal/cowsay` API: an exported sentinel error so `main.go` can recognize "unknown cow" cleanly

**Out of Phase 2** (deferred to later phases by ROADMAP):
- `-W`/`-n` wrap, `--think`, `-e`/`-T`, runewidth display-width fix, `-h`/`--help` → Phase 3
- Release pipeline, `--version`, `go install` → Phase 4

</domain>

<decisions>
## Implementation Decisions

### Input Resolution
- **D-01:** Input precedence (INPUT-03): when positional args are present, use them and do **not** read stdin at all. Only read stdin when there are no positional args.
- **D-02:** Bare `gosay` with no args **and** nothing piped (interactive terminal) → print usage to stderr and exit 1. Do **not** block waiting on terminal stdin (the real-cowsay behavior was explicitly rejected). This requires detecting whether stdin is a TTY vs a pipe — e.g. `os.Stdin.Stat()` and checking `os.ModeCharDevice`. If stdin is piped, read it (INPUT-02 must still work).
- **D-03:** Empty message (INPUT-04) → render the **cowsay-style empty bubble**: top/bottom borders with a single-space `<  >` interior line, then the gopher; exit 0. This path covers both an explicitly empty positional arg (`gosay ""`) and empty piped stdin (`echo -n "" | gosay`). The existing balloon builder should produce this naturally; verify it does not panic or collapse on the empty string.

### Cow Selection — `--random`
- **D-04:** Random pool = **all 51 embedded cows**, including `gopher` and `daemon.cow`. No exclusion list. Reuse `cowsay.ListCows()` as the source of truth for the pool.
- **D-05:** No injectable/seedable RNG seam. Use package-level `math/rand`. Tests assert only that the chosen name is a member of `ListCows()`, not which specific animal was picked.
- **D-06:** `-f <name>` and `--random` together are **mutually exclusive** → print a clear error (e.g. `gosay: cannot combine -f and --random`) and exit non-zero. No silent precedence.

### Cow Selection — `-l` listing
- **D-07:** `-l` output format = **upstream cowsay columnar**: a `Cow files:` header followed by all embedded animal names wrapped into space-separated columns. (Chosen over plain one-per-line.) Exact column count / wrap width is Claude's discretion — match upstream cowsay's general shape; keep it deterministic and testable via golden file.
- **D-08:** The default `gopher` appears **plain and alphabetical** in the listing (sorts under `g`). No `(default)` marker — keep it a uniform list, consistent with upstream cowsay.
- **D-09:** `-l` is **not** a short-circuit query that ignores other input. Combining `-l` with a message and/or `-f`/`--random` is a **usage error** → exit non-zero. `-l` alone prints the list and exits 0.
- **D-10:** ⚠ **ROADMAP/REQUIREMENTS edit required.** This contradicts the locked wording. ROADMAP.md §"Phase 2" success criterion #3 currently says `-l` lists names "one per line"; COW-03 in REQUIREMENTS.md implies the same. Both must be amended to the columnar format so the verifier checks the right behavior. This edit is made as part of this discuss-phase commit (see canonical refs). User explicitly approved the change.

### Unknown-Cow Error UX (COW-05)
- **D-11:** Error message = `gosay: unknown cowfile "nosuchcow"` (animal name quoted), exit non-zero. No `-l` hint, no leaked internal `cows/x.cow` path. (The current Phase 1 error `load cow "x": open cows/x.cow: file does not exist` leaks the path and must not surface to users.)
- **D-12:** Implementation: the `internal/cowsay` package exports a **sentinel error** (e.g. `var ErrUnknownCow = errors.New("unknown cowfile")`). `LoadCow` wraps it when the embedded read returns `fs.ErrNotExist`. `main.go` calls `Render` and does `errors.Is(err, cowsay.ErrUnknownCow)` to translate into the clean user-facing message; any other (non-not-found) render error is printed generically. Validation lives via Render's error rather than a pre-check against `ListCows()`.

### Claude's Discretion
- Flag mechanics with `flag` stdlib: short vs long flag registration (`--random` is long; `flag` stdlib accepts both single- and double-dash, so no special handling needed), flag var wiring, and `flag.Usage` override for the D-02 usage string.
- Reading stdin via `io.ReadAll(os.Stdin)`; trimming of a single trailing newline from piped input (match upstream cowsay's behavior — a trailing `\n` from `echo` should not produce a spurious blank line).
- Exact wording of the `-f`+`--random` conflict and `-l`+input conflict messages (within D-06/D-09 intent).
- Exact column width / wrap logic for the `-l` columnar output (D-07).
- Naming of the sentinel error and any new unexported helpers.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project context
- `.planning/PROJECT.md` — vision, scope, constraints, Key Decisions table
- `.planning/REQUIREMENTS.md` — v1 requirements; Phase 2 covers INPUT-02, INPUT-03, INPUT-04, COW-02, COW-03, COW-04, COW-05. ⚠ COW-03 wording to be amended to columnar (D-10).
- `.planning/ROADMAP.md` §"Phase 2: Input and Cow Selection" — phase goal and success criteria. ⚠ Criterion #3 to be amended from "one per line" to columnar (D-10).

### Prior phase context (carry-forward)
- `.planning/phases/01-first-runnable/01-CONTEXT.md` — Phase 1 decisions; establishes the tiny `internal/cowsay` API surface (`Render`, `ListCows`) Phase 2 must build on, not bypass.

### Domain research (HIGH-confidence)
- `.planning/research/ARCHITECTURE.md` — canonical package layout and component boundaries; the thin-`main.go` / `internal/cowsay`-seam pattern
- `.planning/research/STACK.md` — `flag` stdlib chosen over Cobra/Kong; `sebdah/goldie` v2 for golden tests
- `.planning/research/FEATURES.md` — feature dependency order (Phase 2 = input + selection)
- `.planning/research/PITFALLS.md` — input/parser traps

### Existing code (the integration surface)
- `cmd/gosay/main.go` — current thin CLI (`run(args, stdout, stderr) int`); Phase 2 extends this to add `flag` parsing + stdin + selection. Keep the testable `run`-returns-exit-code shape.
- `internal/cowsay/embed.go` — `ListCows() ([]string, error)` (already sorted, `.cow`-stripped) is the source for `-l` and the `--random` pool.
- `internal/cowsay/renderer.go` — `Render(animal, message string, opts RenderOpts) (string, error)`; the call site for selection.
- `internal/cowsay/cowfile.go` — `LoadCow`; the place to wrap the new `ErrUnknownCow` sentinel on `fs.ErrNotExist`.

### Upstream reference
- `github.com/cowsay-org/cowsay` — reference for `-l` columnar `Cow files:` format, empty-bubble shape, and stdin/arg precedence behavior

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `cowsay.ListCows()` — already returns sorted, `.cow`-stripped names; directly powers both `-l` (columnar formatting on top) and the `--random` pool (D-04, D-07).
- `cowsay.Render(animal, message, opts)` — unchanged signature; Phase 2 just chooses `animal` (gopher / `-f` value / random pick) and `message` (args or stdin) before calling it.
- `run(args []string, stdout, stderr io.Writer) int` in `main.go` — testable harness already in place; Phase 2 tests drive it with crafted args + stdin without `os.Exit`.

### Established Patterns
- Thin `main.go` over an unexported-surface `internal/cowsay` package — Phase 2 must not grow the public API beyond a tiny, intentional addition (the `ErrUnknownCow` sentinel, D-12).
- Golden-file tests via `sebdah/goldie` under `internal/cowsay/testdata/golden/` — add goldens for empty-bubble and (if rendered in-package) `-l` output; CLI-level behavior (precedence, conflicts, exit codes) tested in `cmd/gosay/main_test.go`.

### Integration Points
- `main.go` ↔ `flag` stdlib: new flags `-f` (string), `-l` (bool), `--random` (bool). `flag.Usage` overridden to emit the D-02 usage string.
- `main.go` ↔ stdin: `io.ReadAll(os.Stdin)` only when no positional args; gated by a TTY check (`os.Stdin.Stat()` / `os.ModeCharDevice`) to avoid blocking on a terminal (D-02).
- `cowfile.go` `LoadCow` ↔ `main.go`: new `ErrUnknownCow` sentinel translated via `errors.Is` (D-12).

</code_context>

<specifics>
## Specific Ideas

- **Match real cowsay where it's cheap, diverge where it helps.** Kept: columnar `-l` (D-07), empty-bubble shape (D-03), args-win precedence (D-01). Deliberately diverged: no blocking on terminal stdin (D-02 — prints usage instead), and `-l`/`-f`/`--random` conflicts are hard errors rather than silent precedence (D-06, D-09).
- **Don't leak internals to users.** The Phase 1 error path exposes `cows/x.cow`; Phase 2's clean `unknown cowfile` message (D-11) is the user-visible contract for COW-05.
- **The columnar `-l` decision changes a locked success criterion** — this is a paper-trail item (D-10). The ROADMAP/REQUIREMENTS edit must land with this context so Phase 2 verification checks the columnar behavior, not the old "one per line" wording.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within Phase 2 scope. Phase 3 (wrap/think/eyes/tongue/runewidth/`-h`) and Phase 4 (release pipeline) remain as scoped in ROADMAP.md and REQUIREMENTS.md.

</deferred>

---

*Phase: 2-input-and-cow-selection*
*Context gathered: 2026-05-30*
</content>
</invoke>
