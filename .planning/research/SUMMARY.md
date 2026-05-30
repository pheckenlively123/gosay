# Project Research Summary

**Project:** gosay
**Domain:** Go single-binary CLI tool (cowsay clone)
**Researched:** 2026-05-17
**Confidence:** HIGH

## Executive Summary

gosay is a Go reimplementation of the classic `cowsay` CLI — a small toy that wraps a message in an ASCII speech bubble next to an ASCII animal. The recommended build approach is pure Go with zero runtime dependencies: use `//go:embed` to bundle the full upstream cowsay-org/cowsay `.cow` file set into the binary, hand-roll a ~50-line heredoc parser to handle the Perl-variable format, and hand-roll a ~90-line balloon renderer. The only external dependency is `github.com/mattn/go-runewidth` for correct Unicode display-width calculation — an exception to the "stdlib only" rule that is non-negotiable if non-ASCII input is to be supported. Testing uses `sebdah/goldie` for golden-file snapshot comparisons, and releases use GoReleaser v2 via GitHub Actions.

The technical risk surface is small but specific. Three parsing bugs trip up every Go cowsay implementation: (1) failing to unescape Perl backslash sequences (`\\` → `\`, `\@` → `@`) from the heredoc body, (2) hardcoding `EOC` as the heredoc terminator instead of capturing it dynamically, and (3) using byte-length instead of display-width for bubble sizing. All three are well-understood and can be addressed from day one with a good test suite. The licensing obligation for vendored `.cow` files (GPL/Artistic dual-license with some per-file variations) requires a `NOTICE` file before any public release.

The scope is tight and intentionally so. The `.cow` file parser is the keystone — every visual feature depends on it. Build order follows the dependency graph: scaffold → embed + cow listing → parser → balloon renderer → renderer → CLI wiring → golden tests → release pipeline. Phases 1 and 2 of that graph (parser and balloon) are independent and can proceed in parallel.

## Key Findings

### Recommended Stack

Use the Go standard library for everything except display-width calculation and golden-file testing. The CLI flag surface (`-f`, `-l`, `-W`, `-n`, `--think`, `-e`, `-T`, `--random`) fits comfortably in `flag` stdlib — adding Cobra or Kong for five flat flags is over-engineering. Embedding the `cows/` directory is handled natively by `//go:embed cows/*.cow` on an `embed.FS` variable. Release cross-compilation (linux/darwin/windows × amd64/arm64) is handled by GoReleaser v2 in a single ~20-line config file run on one CI runner, not a six-VM matrix.

**Core technologies:**
- Go 1.22 (go.mod minimum) / toolchain go1.26.3 — `go 1.22` maximizes installer reach while keeping `//go:embed`, generics, and all needed stdlib; `toolchain` line pins the local version
- `embed.FS` (stdlib) — bundle the entire `cows/` directory into the binary at compile time; zero-cost listing via `fs.ReadDir`
- `flag` (stdlib) — sufficient for five flat flags; no subcommands needed
- `github.com/mattn/go-runewidth` — the one justified external dep; required for correct CJK/emoji bubble width
- `sebdah/goldie` v2.8.0 — golden-file test snapshots; `go test -update ./...` regenerates all fixtures after a rendering change
- GoReleaser v2 + `goreleaser/goreleaser-action@v6` — cross-platform release artifacts from a tag push

**What not to use:** Cobra (overkill), `mitchellh/go-wordwrap` (archived 2020), `Code-Hex/Neo-cowsay` (library API, competing embedded cow set — reference only), `golang.org/x/term` for terminal width detection (breaks in pipelines, adds platform complexity, hardcoded 40 is correct).

### Expected Features

**Must have (table stakes):**
- Message from positional args (`gosay hello world`)
- Message from stdin (`echo hello | gosay`)
- Default animal is the gopher, not the cow
- `-f <name>` to select any embedded animal
- `-l` to list all embedded animals
- Speech bubble with correct single-line (`< >`) and multi-line (`/ | \`) borders
- Word wrap at 40 columns default; `-W <n>` to override; `-n` to disable
- `--think` flag for thought-bubble variant (bubble uses `( )` borders, `o` thoughts trail)
- `-e <xx>` and `-T <xx>` for eye and tongue customization (near-zero cost once parser handles `$eyes`/`$tongue`)
- `--random` flag to pick a random animal (5 lines of code once `-l` exists)
- Clean error on unknown `-f` name; `-h` help output

**Should have:**
- `cowthink` as a second binary (thin `cmd/cowthink/main.go` wrapper with think mode forced) — adds Windows-friendly parity with the symlink convention; low cost after `--think` exists

**Defer to v2+ / never:**
- Mood presets (`-b/-d/-g/-p/-s/-t/-w/-y`) — add flag noise for niche eye glyphs; user can do the same with `-e`
- ANSI color output, animations, TUI — explicitly out of scope
- Library API — explicitly out of scope
- Runtime `.cow` file loading from disk — defeats single-binary goal
- Interactive fuzzy cow finder — requires TUI dependency

### Architecture Approach

The codebase is four components plus a data directory, totaling roughly 400 lines of Go. `cmd/gosay/main.go` stays thin (flag parsing, stdin/arg reading, dispatch, error exit — under 60 lines). All logic lives in `internal/cowsay/` to keep it testable in isolation and unexported to prevent accidental library use. The `cows/` directory sits under `internal/cowsay/cows/` so the `//go:embed` declaration and the data it embeds are in the same package.

**Major components:**
1. `internal/cowsay/embed.go` + `cows/` — `embed.FS` declaration; vendored upstream `.cow` files; `gopher.cow` hand-authored as the default animal
2. `internal/cowsay/cowfile.go` — heredoc parser (`bufio.Scanner` line scan, dynamic terminator capture via regex, backslash unescape pass); `ListCows()` and `LoadCow(name)` functions
3. `internal/cowsay/balloon.go` — word-wrap with `runewidth.StringWidth` for display-width; single-line vs multi-line border logic; think vs say bubble characters
4. `internal/cowsay/renderer.go` — `strings.NewReplacer` substitution of `$thoughts`/`$eyes`/`$tongue`; concatenates balloon + substituted cow body
5. `cmd/gosay/main.go` — wires the above; reads flags and input; prints result; exits non-zero on error

Build order within the internal package: `embed.go` (data layer) → `cowfile.go` and `balloon.go` in parallel (independent) → `renderer.go` (requires both) → `main.go` (wires everything).

### Critical Pitfalls

1. **Backslash unescape omission** — After heredoc extraction, run a single unescape pass (`\\`→`\`, `\@`→`@`, `\$`→`$`) before variable substitution. Without this, `default.cow`'s tail renders as `\\`. Verify by diffing output against Perl cowsay on the first five cows.

2. **Hardcoded `EOC` terminator** — Some cow files use `END`, `EOT`, or other terminators. Capture the terminator string dynamically from the `<<TERMINATOR` line with a regex group; never hardcode `EOC`. A file with a non-EOC terminator silently renders blank or garbled otherwise.

3. **Byte-length bubble width** — Using `len(s)` or `len([]rune(s))` for bubble sizing breaks on CJK (2 columns per char) and emoji. Use `runewidth.StringWidth(line)` throughout. This is the one non-stdlib dependency; its absence is a bug, not a tradeoff.

4. **Licensing / NOTICE omission** — Vendored `.cow` files are GPL/Artistic dual-licensed; some files have individual licenses (`apt.cow` is GPL-only; `daemon.cow` has unclear provenance and was removed from Fedora packaging). A `cows/NOTICE` file with upstream attribution is required before any public release. Decide explicitly whether to include `daemon.cow`.

5. **GoReleaser misconfiguration** — Three lines break releases: omitting `fetch-depth: 0` from checkout (produces `v0.0.0-SNAPSHOT`), omitting `permissions: contents: write` (silently produces no artifacts), and triggering on push-to-main instead of `push: tags: ['v*']`. Test with `goreleaser release --snapshot --clean` before the first real tag.

## Implications for Roadmap

### Phase 1: Core Engine

**Rationale:** The parser and balloon renderer are pure functions with no external dependencies — they can be fully built, unit-tested, and validated before any CLI or release concern is touched. This phase produces the most value per line of code and front-loads the trickiest correctness work (backslash unescaping, dynamic terminator capture, display-width balloon sizing).

**Delivers:** A working `internal/cowsay` package with embedded cow files, a correct heredoc parser, a correct balloon renderer, and a correct renderer combining them. All three critical parsing pitfalls are eliminated here.

**Addresses:** Cow file embedding, `.cow` parser, balloon rendering, `$thoughts`/`$eyes`/`$tongue` variable substitution, word wrap, think vs say mode.

**Avoids:** Backslash misrendering, hardcoded terminator, byte-length bubble width, `$thoughts` default misconfiguration.

**Must resolve:** Add `cows/NOTICE` with upstream attribution in this phase, before any public commit of the vendored files. Add `.gitattributes` enforcing `cows/*.cow text eol=lf`.

**Research flag:** No additional research needed — patterns are fully documented.

---

### Phase 2: CLI and Full Feature Surface

**Rationale:** With the engine in place, wiring the CLI is straightforward plumbing. All flags can be implemented against already-tested internals. This is also when golden tests are added (the engine must be stable first for golden files to be meaningful).

**Delivers:** A working `gosay` binary with the full flag surface: `-f`, `-l`, `-W`, `-n`, `--think`, `-e`, `-T`, `--random`, `-h`. Stdin and positional arg input both work. Golden test fixtures are committed. Error handling is clean.

**Addresses:** All table-stakes features. Gopher as default. `--random`. Clean error on bad `-f` name.

**Avoids:** Terminal width detection complexity (hardcode 40, expose `-W`; no `TIOCGWINSZ`). Scope creep (no mood presets, no color, no library API).

**Research flag:** No additional research needed — flag surface and stdin behavior are documented from upstream man page.

---

### Phase 3: Release Pipeline

**Rationale:** Release infrastructure comes last because it needs a working binary and a real git tag to be meaningful. Build it once and it runs forever.

**Delivers:** `.goreleaser.yaml`, `.github/workflows/release.yml`, and a validated first tagged release producing six platform archives (linux/darwin/windows × amd64/arm64) attached to the GitHub Release. `gosay --version` prints the git tag.

**Addresses:** Cross-platform distribution requirement. `go install github.com/pheckenlively/gosay/cmd/gosay@latest` works.

**Avoids:** GoReleaser misconfiguration — validate with `--snapshot` before tagging.

**Research flag:** No additional research needed — GoReleaser v2 config is fully documented in STACK.md.

---

### Phase 4: cowthink Binary (Optional Stretch)

**Rationale:** `--think` flag on `gosay` fully covers the use case. A separate `cowthink` binary is purely for muscle-memory compatibility with users who know the upstream tool. Low cost after Phase 2 (thin `cmd/cowthink/main.go` that forces think mode). Include if the author wants it; skip if not.

**Delivers:** A second binary in the release archives. `cowthink hello` works as expected.

**Research flag:** Standard patterns; no research needed.

---

### Phase Ordering Rationale

- Parser and balloon are dependency-free — building them first means every subsequent phase builds on tested, correct foundations rather than debugging rendering bugs through the CLI layer.
- CLI wiring is deferred until the engine is stable so that golden test fixtures reflect final rendering behavior and don't need to be regenerated due to engine bugs.
- Release pipeline is last because it depends on a working binary, a real tag, and GitHub repository configuration that cannot be validated until both exist.
- The `daemon.cow` licensing question must be resolved in Phase 1 (before any public commit of vendored files) — it is not a Phase 3 concern.

### Research Flags

Phases with well-documented patterns (skip additional research):
- **Phase 1:** All patterns documented. `strings.NewReplacer` substitution, `bufio.Scanner` heredoc extraction, `embed.FS` glob, `runewidth.StringWidth` usage — all verified.
- **Phase 2:** Flag surface and stdin behavior from upstream man page. `flag` stdlib usage is trivial.
- **Phase 3:** GoReleaser v2 config is in STACK.md verbatim.
- **Phase 4:** Thin wrapper; no research needed.

No phases require a `--research-phase` pass before planning can proceed.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All packages verified on pkg.go.dev; versions confirmed May 2026; GoReleaser v2 config tested against docs |
| Features | HIGH | Upstream man page and cowsay-org/cowsay source reviewed; Neo-cowsay behavior compared |
| Architecture | HIGH | Package layout follows Go official module layout docs; component boundaries validated against Neo-cowsay reference |
| Pitfalls | HIGH (parsing/licensing), MEDIUM (Unicode edge cases) | Parsing pitfalls confirmed by reading upstream cow files directly; Unicode width caveats confirmed from open bug reports |

**Overall confidence:** HIGH

### Gaps to Address

- **`daemon.cow` inclusion:** The cow file has unclear provenance and was removed from Fedora packaging. Decide before vendoring whether to include it. If excluded, document the exclusion in `cows/NOTICE`. This is a judgment call for the author, not a research gap.

- **gopher.cow source:** PROJECT.md specifies a gopher as the default but no canonical `gopher.cow` exists in cowsay-org/cowsay upstream. Neo-cowsay has a gopher.cow but it is under that project's license. The author needs to either hand-author a gopher ASCII art or obtain a clearly-licensed one. This is the only feature with an unresolved asset dependency.

- **`--think` short flag alias:** Research confirms `-t` is deliberately avoided (conflicts with upstream "tired" preset) in favor of `--think` long flag only. Confirm author preference — long-only is the safe default.

## Open Questions for Requirements

These are human decisions that must be resolved before requirements can be finalized:

1. **`daemon.cow` inclusion** — Include despite Fedora's removal for provenance concerns, or exclude with a note in NOTICE? No correct answer — author call. The conservative choice is exclusion.

2. **Full upstream set vs curated subset** — Vendor all ~60 `.cow` files from cowsay-org/cowsay, or curate a smaller set? Recommendation: vendor the full set. Binary size impact is modest; it preserves ecosystem compatibility.

3. **Minimum Go version in go.mod** — Research recommends `go 1.22` for broad installer reach. Does the author care about users on 1.22–1.24, or is 1.25+ sufficient? Recommendation: `go 1.22` is the right call.

4. **`cowthink` binary** — Build a second `cmd/cowthink/main.go` binary (Phase 4), or is `gosay --think` sufficient? Low cost either way; author preference.

5. **`gopher.cow` authorship** — Hand-author original ASCII gopher art, or adapt from a clearly-licensed source? Must be resolved before Phase 1 can complete. Neo-cowsay's gopher.cow is a candidate if licensing is compatible.

## Sources

### Primary (HIGH confidence)
- `pkg.go.dev/embed` — embed.FS API, glob patterns
- `pkg.go.dev/github.com/sebdah/goldie/v2` v2.8.0 — confirmed Oct 2025
- `go.dev/doc/devel/release` — Go 1.26.3 latest stable May 2026
- `go.dev/blog/toolchain` — go vs toolchain directive semantics
- `goreleaser.com/customization` + Context7 — GoReleaser v2 config
- cowsay-org/cowsay upstream + Licensing.md — feature surface, per-file licenses
- `github.com/Code-Hex/Neo-cowsay` cowsay.go — parsing reference
- Upstream cowsay man page (man.archlinux.org) — flag surface, stdin behavior

### Secondary (MEDIUM confidence)
- Ubuntu bug #393212 / Arch bug FS#48347 — Unicode width miscalculation in cowsay
- golang/go issue #42328, #48888 — go:embed hidden file inclusion edge cases
- nmyk.io/cowsay — text/template approach (contrast reference)

### Tertiary (informational)
- python-cowsay, quantum5/cowsay (C++) — non-Go implementations consulted for escape-sequence handling patterns

---
*Research completed: 2026-05-17*
*Ready for roadmap: yes*
