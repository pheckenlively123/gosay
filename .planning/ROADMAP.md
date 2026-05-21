# Roadmap: gosay

## Overview

gosay ships in four vertical slices. Phase 1 produces a real, runnable binary — a gopher says hello with a correctly rendered speech bubble. Each subsequent phase adds a user-visible capability on top of the working foundation: full input and cow selection in Phase 2, the complete flag surface in Phase 3, and the automated release pipeline in Phase 4. Every phase delivers something users can invoke and observe; no phase is a pure infrastructure layer.

## Phases

**Phase Numbering:**

- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: First Runnable** - `gosay "hello"` prints a gopher with a speech bubble; parser, balloon, renderer, and CLI wired end-to-end; cows vendored with NOTICE
- [ ] **Phase 2: Input and Cow Selection** - stdin, `-f`, `-l`, `--random`, empty input, error handling
- [ ] **Phase 3: Full Flag Surface** - `--think`, `-W`, `-n`, `-e`, `-T`, runewidth, `-h`/`--help`
- [ ] **Phase 4: Release Pipeline** - GoReleaser, GitHub Actions, `--version`, `go install`

## Phase Details

### Phase 1: First Runnable

**Goal**: User can run `gosay "hello"` and see a gopher saying hello — parser, balloon renderer, variable substitution, and CLI wired through `main.go`; all vendored cows present with attribution
**Mode:** mvp
**Depends on**: Nothing (first phase)
**Requirements**: INPUT-01, COW-01, COWS-01, COWS-02, COWS-03, COWS-04, COWS-05, RENDER-01, RENDER-02, RENDER-03, RENDER-04
**Success Criteria** (what must be TRUE):

  1. `gosay "hello"` prints a gopher ASCII figure with a `< hello >` speech bubble — no panics, no garbled output
  2. `gopher.cow` exists as a hand-authored file in `internal/cowsay/cows/`; it is the default animal
  3. The heredoc parser correctly handles a dynamic terminator (not hardcoded `EOC`) and unescapes Perl backslash sequences (`\\` → `\`, `\@` → `@`, `\$` → `$`); golden tests cover both pitfalls
  4. Multi-line input renders with `/ | \` borders and single-line input with `< >` borders
  5. `cows/NOTICE` exists with upstream cowsay-org attribution and per-file license notes; `.gitattributes` enforces `cows/*.cow text eol=lf`

**Plans**: 4 plans
Plans:
**Wave 1**

- [x] 01-01-PLAN.md — Walking skeleton: go.mod + .gitattributes + minimal main.go (positional-arg read, stub output)

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 01-02-PLAN.md — Cow data layer: vendor 50 upstream .cow files + NOTICE + SOURCE.md + embed.go (//go:embed + ListCows)

**Wave 3** *(blocked on Wave 2 completion)*

- [ ] 01-03-PLAN.md — Rendering engine: cowfile parser (dynamic terminator + backslash unescape) + balloon (single/multi-line) + renderer (variable substitution)

**Wave 4** *(blocked on Wave 3 completion)*

- [ ] 01-04-PLAN.md — Wire-up: gopher.cow (user-pick gate) + main.go calls cowsay.Render + golden tests (gopher + 4 pitfall cows + skipped CJK)

### Phase 2: Input and Cow Selection

**Goal**: User can pipe messages via stdin, select any embedded animal with `-f`, list all animals with `-l`, pick a random one with `--random`, and receive a clear error for unknown cows
**Mode:** mvp
**Depends on**: Phase 1
**Requirements**: INPUT-02, INPUT-03, INPUT-04, COW-02, COW-03, COW-04, COW-05
**Success Criteria** (what must be TRUE):

  1. `echo "hello" | gosay` prints a gopher saying hello (stdin path works)
  2. `gosay -f tux "hello"` prints tux saying hello (any embedded animal selectable by name)
  3. `gosay -l` lists every embedded animal name, one per line
  4. `gosay --random "hello"` prints some animal saying hello (animal varies across invocations)
  5. `gosay -f nosuchcow "hello"` exits non-zero with a human-readable error message; empty input produces a valid empty bubble with no panic

**Plans**: TBD

### Phase 3: Full Flag Surface

**Goal**: User has access to the complete flag set — word wrap control, thought-bubble mode, custom eyes and tongue, Unicode-correct bubble sizing, and documented help output
**Mode:** mvp
**Depends on**: Phase 2
**Requirements**: RENDER-05, RENDER-06, RENDER-07, RENDER-08, HELP-01
**Success Criteria** (what must be TRUE):

  1. A 200-character message wraps at 40 columns by default; `-W 80` wraps at 80; `-n` disables wrapping entirely
  2. `gosay --think "hello"` renders a thought bubble (`( )` borders, `o` thought trail) instead of a speech bubble
  3. `gosay -e XX -T -- "hello"` renders the gopher with `XX` eyes and `--` tongue
  4. `echo "漢字テスト" | gosay` produces a bubble whose right edge aligns correctly — `runewidth.StringWidth` used throughout, not `len()`
  5. `gosay -h` (and `gosay --help`) prints usage documentation covering every flag with example invocations

**Plans**: TBD

### Phase 4: Release Pipeline

**Goal**: A tagged push to GitHub triggers an automated release producing static binaries for all six platform/arch combinations; `gosay --version` prints the git tag; `go install` works
**Mode:** mvp
**Depends on**: Phase 3
**Requirements**: DIST-01, DIST-02, DIST-03, DIST-04, DIST-05
**Success Criteria** (what must be TRUE):

  1. Pushing a `v*` tag triggers the GitHub Actions release workflow and produces six archives (linux/darwin/windows × amd64/arm64) attached to the GitHub Release
  2. `gosay --version` prints the git tag (e.g., `v1.0.0`), not `dev`
  3. `go install github.com/pheckenlively/gosay/cmd/gosay@latest` installs a working binary with no CGO or runtime data files required
  4. A local `goreleaser release --snapshot --clean` dry-run completes without errors before the first real tag

**Plans**: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. First Runnable | 2/4 | In Progress|  |
| 2. Input and Cow Selection | 0/TBD | Not started | - |
| 3. Full Flag Surface | 0/TBD | Not started | - |
| 4. Release Pipeline | 0/TBD | Not started | - |
