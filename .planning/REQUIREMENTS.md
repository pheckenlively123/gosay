# Requirements: gosay

**Defined:** 2026-05-18
**Core Value:** A single, fast, dependency-free Go binary that reproduces the fun of `cowsay` — message in, ASCII animal out — with no Perl, no external `.cow` files, and a gopher on by default.

## v1 Requirements

Requirements for initial release. Each maps to a roadmap phase.

### Input

- [ ] **INPUT-01**: User can pass a message as positional arguments (`gosay hello world`)
- [x] **INPUT-02**: User can pipe a message via stdin (`echo hello | gosay`)
- [x] **INPUT-03**: When both stdin and positional args are present, positional args win (matches upstream cowsay behavior)
- [x] **INPUT-04**: Empty input produces a valid empty bubble (no panic, no garbled output)

### Cow Selection

- [ ] **COW-01**: Default animal is a gopher (hand-authored from the Go mascot, embedded as `gopher.cow`)
- [x] **COW-02**: `-f <name>` selects a specific animal from the embedded set
- [ ] **COW-03**: `-l` lists every embedded animal name (upstream cowsay columnar format — `Cow files:` header + names in wrapped columns)
- [ ] **COW-04**: `--random` picks a random animal each invocation
- [x] **COW-05**: Unknown `-f <name>` exits non-zero with a clear error message

### Cow Files

- [ ] **COWS-01**: Vendor the full upstream cowsay-org/cowsay `.cow` set into `internal/cowsay/cows/`
- [ ] **COWS-02**: `daemon.cow` is included; its provenance caveat is documented in `cows/NOTICE`
- [ ] **COWS-03**: All `.cow` files are embedded into the binary via `//go:embed cows/*.cow`
- [ ] **COWS-04**: `cows/NOTICE` provides upstream attribution and notes per-file license variances (`apt.cow` GPL-only, `gnu.cow`/`suse.cow` WTFPL-2, `kangaroo.cow` GPL-2.0+, `daemon.cow` provenance caveat, etc.)
- [ ] **COWS-05**: `.gitattributes` enforces `cows/*.cow text eol=lf` (Windows-safe vendoring)

### Rendering

- [ ] **RENDER-01**: Heredoc parser handles dynamic terminator (`<<TERMINATOR ... TERMINATOR`, not just `EOC`)
- [ ] **RENDER-02**: Heredoc body has Perl backslash sequences unescaped (`\\` → `\`, `\@` → `@`, `\$` → `$`)
- [ ] **RENDER-03**: `$thoughts`, `$eyes`, `$tongue` variables are substituted at render time
- [ ] **RENDER-04**: Speech bubble renders correctly for both single-line (`< >`) and multi-line (`/ | \`) inputs
- [ ] **RENDER-05**: Word-wrap defaults to 40 columns; `-W <n>` overrides; `-n` disables wrapping
- [ ] **RENDER-06**: Bubble sizing uses display-width (`runewidth.StringWidth`), not byte/rune count — CJK / emoji / combining characters render with correct bubble width
- [ ] **RENDER-07**: `--think` flag swaps to thought-bubble form (`( )` borders, `o` thoughts trail, `$thoughts = "o"`)
- [ ] **RENDER-08**: `-e <xx>` customises eyes (default `oo`); `-T <xx>` customises tongue (default `  `)

### Distribution

- [ ] **DIST-01**: Single static Go binary per platform (no runtime data files)
- [ ] **DIST-02**: GitHub Actions release workflow builds release artifacts on tag push (`v*`)
- [ ] **DIST-03**: GoReleaser v2 produces archives for linux / darwin / windows × amd64 / arm64
- [ ] **DIST-04**: `go install github.com/pheckenlively/gosay/cmd/gosay@latest` installs a working binary
- [ ] **DIST-05**: `gosay --version` prints the git tag (set via `-ldflags`)

### Help & Usability

- [ ] **HELP-01**: `-h` / `--help` prints usage with every flag documented and example invocations

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

(None — gosay v1 covers everything in scope. Future ideas land in `.planning/todos/` or `--backlog`.)

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Mood preset flags (`-b/-d/-g/-p/-s/-t/-w/-y`) | Flag noise for niche eye glyphs; `-e` gives full control. `-t` would also collide with `--think` semantics |
| ANSI color output | Explicitly excluded in PROJECT.md — keep the toy small |
| Animations / TUI | Explicitly excluded in PROJECT.md — keep the toy small |
| Library API (`gosay.Say(...)`) | Explicitly excluded in PROJECT.md — `internal/` package enforces this structurally |
| Runtime `.cow` file loading from disk | Defeats single-binary distribution; all cows are embedded |
| Separate `cowthink` binary | `gosay --think` covers the use case; user-selected single-binary scope |
| Terminal-width auto-detection (`TIOCGWINSZ` / `$COLUMNS`) | Original cowsay hard-codes 40; auto-detection breaks piped output and complicates Windows builds |
| Native gosay `.cow` format | Vendoring upstream preserves the ecosystem and credit; defining a new format adds work for no gain |
| Interactive fuzzy cow finder | Requires TUI dependency; out of project shape |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| INPUT-01 | Phase 1 | Pending |
| INPUT-02 | Phase 2 | Complete |
| INPUT-03 | Phase 2 | Complete |
| INPUT-04 | Phase 2 | Complete |
| COW-01 | Phase 1 | Pending |
| COW-02 | Phase 2 | Complete |
| COW-03 | Phase 2 | Pending |
| COW-04 | Phase 2 | Pending |
| COW-05 | Phase 2 | Complete |
| COWS-01 | Phase 1 | Pending |
| COWS-02 | Phase 1 | Pending |
| COWS-03 | Phase 1 | Pending |
| COWS-04 | Phase 1 | Pending |
| COWS-05 | Phase 1 | Pending |
| RENDER-01 | Phase 1 | Pending |
| RENDER-02 | Phase 1 | Pending |
| RENDER-03 | Phase 1 | Pending |
| RENDER-04 | Phase 1 | Pending |
| RENDER-05 | Phase 3 | Pending |
| RENDER-06 | Phase 3 | Pending |
| RENDER-07 | Phase 3 | Pending |
| RENDER-08 | Phase 3 | Pending |
| DIST-01 | Phase 4 | Pending |
| DIST-02 | Phase 4 | Pending |
| DIST-03 | Phase 4 | Pending |
| DIST-04 | Phase 4 | Pending |
| DIST-05 | Phase 4 | Pending |
| HELP-01 | Phase 3 | Pending |

**Coverage:**

- v1 requirements: 28 total (4 INPUT + 5 COW + 5 COWS + 8 RENDER + 5 DIST + 1 HELP)
- Mapped to phases: 28
- Unmapped: 0

Note: The requirements file header stated 27 but the defined requirement list contains 28 entries. All 28 are mapped.

---
*Requirements defined: 2026-05-18*
*Last updated: 2026-05-18 after roadmap creation — traceability table populated*
