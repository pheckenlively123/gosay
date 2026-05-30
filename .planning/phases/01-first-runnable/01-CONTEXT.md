# Phase 1: First Runnable - Context

**Gathered:** 2026-05-20
**Status:** Ready for planning

<domain>
## Phase Boundary

Build the gosay engine end-to-end so that `gosay "hello"` prints the default gopher animal inside a correctly-formatted speech bubble. Phase 1 delivers:

- `cmd/gosay/main.go` — thin CLI: reads positional args, dispatches to renderer, prints to stdout, exits non-zero on error
- `internal/cowsay/` — embed.go, cowfile.go, balloon.go, renderer.go
- `internal/cowsay/cows/` — full upstream cow set + hand-authored `gopher.cow` (the default) + `NOTICE` + `SOURCE.md`
- `//go:embed cows/*.cow` wired and listable via `fs.ReadDir`
- Heredoc parser with dynamic terminator + Perl-backslash unescape
- `$thoughts`/`$eyes`/`$tongue` variable substitution
- Balloon rendering for both single-line (`< >`) and multi-line (`/ | \`) inputs
- Golden tests covering gopher + 3-5 pitfall-targeted cows
- `.gitattributes` enforcing `cows/*.cow text eol=lf`

**Out of Phase 1** (deferred to later phases by ROADMAP):
- stdin reading, `-f`, `-l`, `--random`, empty-input handling, unknown-cow errors → Phase 2
- `-W`/`-n` wrap, `--think`, `-e`/`-T`, runewidth, `-h` → Phase 3
- Release pipeline → Phase 4

</domain>

<decisions>
## Implementation Decisions

### Gopher ASCII Art
- **D-01:** Detailed art (~12+ lines) in the proportions of upstream `koala.cow` / `default.cow` — recognisably the Renee French Go mascot
- **D-02:** Standing, facing-forward pose — classic mascot stance with buck teeth visible
- **D-03:** Standard `.cow` placeholder layout — `$eyes` substitutes the two-character eye glyphs on the face; `$tongue` substitutes the two-character tongue/teeth-gap (so `-e XX -T --` will work as expected once Phase 3 lands)
- **D-04:** During Phase 1 execution, draft 2-3 ASCII gopher candidates and present them to the user; the user picks one before the phase commits

### Cow-Vendoring Mechanism
- **D-05:** One-time snapshot copy — `internal/cowsay/cows/` populated by hand-copy from upstream
- **D-06:** Source: `cowsay-org/cowsay` at the latest tagged stable release (current v3.8.x family); pin the exact tag/SHA in `cows/SOURCE.md`
- **D-07:** Copy `.cow` files verbatim from upstream's `cows/` directory; normalize CRLF → LF during the copy
- **D-08:** `.gitattributes` at the repo root enforces `cows/*.cow text eol=lf` going forward
- **D-09:** `internal/cowsay/cows/SOURCE.md` records: upstream repo URL, exact tag/commit SHA, vendoring date, refresh procedure
- **D-10:** `internal/cowsay/cows/NOTICE` records: upstream attribution, per-file license variances (`apt.cow` GPL-only, `gnu.cow`/`suse.cow` WTFPL-2, `kangaroo.cow` GPL-2.0+, `daemon.cow` Fedora-removed provenance caveat — but included anyway)

### Golden-Test Coverage
- **D-11:** `sebdah/goldie` v2.8.0 used for golden-file snapshot testing
- **D-12:** Phase 1 covers the gopher plus 3-5 pitfall-targeted cows. Required pitfall coverage:
  - One cow that uses a non-`EOC` terminator (exercises dynamic-terminator capture)
  - One cow that has Perl backslash escape sequences in its body (`\\`, `\@`, or `\$` — e.g., `default.cow`)
  - Optionally `three-eyes.cow` or similar to verify `$eyes` substitution across non-standard layouts
- **D-13:** Golden files are hand-curated expected outputs — no Perl-cowsay dependency in CI; the user reviews each golden once during Phase 1 execution
- **D-14:** Golden test input variants: `hello` (single-line `< >` rendering) and `line1\nline2` (multi-line `/ | \` rendering). Empty-input + long-input golden tests deferred to Phase 2 / Phase 3 respectively when those code paths land
- **D-15:** Fixtures live in `internal/cowsay/testdata/golden/` (standard Go `testdata/` convention; `goldie` default)

### Display-Width Abstraction
- **D-16:** Build a `displayWidth(s string) int` seam in Phase 1 — Phase 3 will swap the body to `runewidth.StringWidth(s)` with zero call-site changes
- **D-17:** `displayWidth` lives as an unexported function in `internal/cowsay/balloon.go` next to its only caller; promote to a `width.go` file only if a second caller emerges in Phase 3
- **D-18:** Phase 1 body: `displayWidth(s) = utf8.RuneCountInString(s)` (correct for ASCII; wrong for CJK/emoji — documented gap)
- **D-19:** Phase 1 adds a `t.Skip(...)`-marked golden test for CJK input (e.g., `漢字テスト`) documenting the Phase 3 fix — Phase 3 removes the skip and replaces the displayWidth body

### Claude's Discretion
- Naming of internal helpers, exact error message wording, and detailed package documentation strings are at Claude's discretion within the layout constraints above
- The exact set of "3-5 pitfall-targeted cows" in D-12 is at Claude's discretion subject to the pitfall-coverage requirement (one non-EOC terminator, one with backslash escapes)
- Whether to use `bufio.Scanner` vs `bufio.Reader` for the heredoc parser is at Claude's discretion; ARCHITECTURE.md recommends `bufio.Scanner` and that should be the default unless a concrete reason emerges to switch

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project context
- `.planning/PROJECT.md` — vision, scope, constraints, Key Decisions table
- `.planning/REQUIREMENTS.md` — full v1 requirements list and Phase 1 traceability (INPUT-01, COW-01, COWS-01..05, RENDER-01..04)
- `.planning/ROADMAP.md` §"Phase 1: First Runnable" — phase goal and success criteria

### Domain research (HIGH-confidence; recently completed)
- `.planning/research/SUMMARY.md` — executive summary, key findings, roadmap implications
- `.planning/research/STACK.md` — `flag` stdlib + `embed.FS` + `sebdah/goldie` v2.8.0 + GoReleaser v2; what NOT to use
- `.planning/research/FEATURES.md` — feature categorization and dependency order (Phase 1 covers the engine half)
- `.planning/research/ARCHITECTURE.md` — package layout, component boundaries, build order (this is the canonical structure ref)
- `.planning/research/PITFALLS.md` — heredoc parser traps, backslash unescape, licensing, `.gitattributes` requirement

### Upstream / external
- `github.com/cowsay-org/cowsay` (latest tagged release) — source of vendored `.cow` files; the exact tag/SHA is to be pinned in `internal/cowsay/cows/SOURCE.md` when Phase 1 runs
- `pkg.go.dev/embed` — `//go:embed` syntax and `embed.FS` API
- `pkg.go.dev/github.com/sebdah/goldie/v2` — golden-file testing API and `-update` flag behavior

</canonical_refs>

<code_context>
## Existing Code Insights

The repo is greenfield — only `LICENSE` exists at the root. No Go files, no `go.mod`, no `.gitattributes` yet. Phase 1 establishes the entire module structure from scratch.

### Reusable Assets
- None yet — Phase 1 creates the foundation. PROJECT.md sets the module path; everything else is new.

### Established Patterns
- None in code. The patterns Phase 1 must establish (and that later phases will follow): unexported-only `internal/cowsay` API surface, golden-test fixture layout under `testdata/`, `.gitattributes`-enforced LF for cow files, `cows/SOURCE.md` + `cows/NOTICE` as the licensing/provenance pair.

### Integration Points
- `cmd/gosay/main.go` will be the only consumer of the `internal/cowsay` package in Phase 1. The exported API surface of `internal/cowsay` should be intentionally tiny — likely just a `Render(animal, message string, opts RenderOpts) (string, error)` and `ListCows() []string` to keep the seams clean for Phase 2 and Phase 3.

</code_context>

<specifics>
## Specific Ideas

- **Gopher likeness:** The hand-authored gopher.cow is the project's signature feature. Treat the user as the visual reviewer — present 2-3 ASCII variants during Phase 1 execution before committing.
- **Pitfall test inputs:** Even if golden coverage is scoped tight, the dynamic-terminator and backslash-unescape pitfalls each need an explicit, named test fixture in `testdata/golden/` so the regression coverage is unmistakable to future readers.
- **`daemon.cow` inclusion is a non-default choice:** Phase 1 must ship `cows/NOTICE` with an explicit paragraph noting Fedora's removal of `daemon.cow` and the project's deliberate decision to include it anyway. This is a paper trail item.
- **No-`flag` Phase 1 main.go is intentional:** Phase 1 main.go reads positional args via `os.Args[1:]`, joins them with spaces, and dispatches. `flag.Parse()` enters in Phase 2 when `-f`/`-l` arrive.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within Phase 1 scope. Items already deferred by the roadmap (Phase 2: stdin/-f/-l/--random; Phase 3: wrap/think/eyes/tongue/runewidth/-h; Phase 4: release pipeline) are documented in REQUIREMENTS.md traceability and ROADMAP.md.

</deferred>

---

*Phase: 1-first-runnable*
*Context gathered: 2026-05-20*
