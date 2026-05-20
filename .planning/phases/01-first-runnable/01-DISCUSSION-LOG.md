# Phase 1: First Runnable - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-20
**Phase:** 1-first-runnable
**Areas discussed:** Gopher ASCII art, Cow-vendoring mechanism, Golden-test coverage scope, Display-width abstraction now vs later

---

## Gopher ASCII art

### Q1 — How detailed should the gopher.cow art be?

| Option | Description | Selected |
|--------|-------------|----------|
| Minimalist (~4-6 lines) | Iconic, blocky, instantly readable in a tiny terminal | |
| Medium (~8-10 lines) | Recognisably the Go gopher — buck teeth, hands, classic round-bodied silhouette | |
| Detailed (~12+ lines) | Full Renee-French-style mascot — arms out, feet visible, more textured shading | ✓ |

**User's choice:** Detailed (~12+ lines)
**Notes:** User selected the preview alongside (preview was the classic koala silhouette, used as a density/proportions reference; actual art will be redrawn as the Go mascot)

### Q2 — What pose should the gopher strike?

| Option | Description | Selected |
|--------|-------------|----------|
| Standing, facing forward | Classic Renee French stance — round body, arms at sides, buck teeth visible | ✓ |
| Sitting / squatting | Lower silhouette, paws on the ground in front | |
| Profile / side view | Gopher in profile facing the speech bubble | |
| You decide | Claude picks whatever reads best | |

**User's choice:** Standing, facing forward
**Notes:** Matches the most recognisable mascot stance

### Q3 — Where in the gopher should `$eyes` and `$tongue` placeholders sit?

| Option | Description | Selected |
|--------|-------------|----------|
| Standard (eyes on face, tongue between buck teeth) | Matches upstream cow files; `-e XX -T --` will work as expected in Phase 3 | ✓ |
| Eyes only, tongue absent | Cleaner art, breaks `-T` parity | |
| Skip placeholders entirely for v1 | Hardcode `oo` eyes and no tongue — fastest, but breaks Phase 3 `-e`/`-T` on the gopher | |

**User's choice:** Standard placeholder layout
**Notes:** Preserves cowsay flag parity from day one

### Q4 — How should I deliver the gopher art during execution?

| Option | Description | Selected |
|--------|-------------|----------|
| I'll draft 2-3 variants; you pick during execution | Best balance of speed and visual review | ✓ |
| I'll pick one based on this discussion | Faster; revisions still possible | |
| Author it now during this discussion | Push creative work into discuss-phase | |

**User's choice:** Draft 2-3 variants; pick during execution
**Notes:** Visual review at execution time is the natural moment for art decisions

---

## Cow-vendoring mechanism

### Q1 — How should we get the upstream cow files into the repo?

| Option | Description | Selected |
|--------|-------------|----------|
| One-time snapshot copy | Manually copy files from a specific upstream commit; record provenance | ✓ |
| `scripts/vendor-cows.sh` | Repeatable shell-script snapshot from a pinned tag | |
| Git submodule of cowsay-org/cowsay | Most reproducible; adds submodule friction | |
| Go generate + download at build | Build-time download; adds network dep to dev/CI | |

**User's choice:** One-time snapshot copy
**Notes:** Simplest path; upstream rarely changes; manual re-sync acceptable

### Q2 — Which upstream source should we snapshot from?

| Option | Description | Selected |
|--------|-------------|----------|
| `cowsay-org/cowsay` latest tagged release | Maintained successor to Tony Monroe's original; pin to exact tag | ✓ |
| `cowsay-org/cowsay` tip-of-main | May include in-progress changes; pin commit SHA | |
| Tony Monroe's original | Historical authenticity; fewer cows | |

**User's choice:** cowsay-org/cowsay latest tagged release
**Notes:** Stable, maintained, full menagerie

### Q3 — Anything to transform/filter during the snapshot copy?

| Option | Description | Selected |
|--------|-------------|----------|
| Copy as-is + normalize line endings to LF | Verbatim copy with CRLF → LF normalization; `.gitattributes` enforces eol=lf going forward | ✓ |
| Strip Perl comments / dead lines + consolidate metadata into NOTICE | Smaller files; loses per-file context | |
| You decide | Claude chooses based on parser simplicity / licensing clarity | |

**User's choice:** Copy as-is + LF normalization
**Notes:** Preserves upstream fidelity; eol=lf handles Windows safely

### Q4 — Where should the SOURCE attribution metadata live?

| Option | Description | Selected |
|--------|-------------|----------|
| `internal/cowsay/cows/SOURCE.md` | Provenance file next to cow files; NOTICE handles licensing separately | ✓ |
| Inline in `cows/NOTICE` | One file for both concerns | |
| In the top-level README only | No file in cows/ itself | |

**User's choice:** Separate `SOURCE.md` next to the cow files
**Notes:** Clean separation — NOTICE for licensing, SOURCE for provenance

---

## Golden-test coverage scope

### Q1 — Which cows should get golden-test coverage in Phase 1?

| Option | Description | Selected |
|--------|-------------|----------|
| Gopher + 3-5 pitfall-targeted cows | Snapshot the default plus diverse cows that exercise dynamic-terminator and backslash-unescape pitfalls | ✓ |
| Gopher only | Minimal test surface; relies on unit tests for parser correctness | |
| All vendored cows (~60) | Maximum safety net; biggest commitment; may force decisions on imperfect-render cows | |

**User's choice:** Gopher + 3-5 pitfall-targeted cows
**Notes:** Locks correctness on the hard cases without test-bloat

### Q2 — What's the source of truth for golden files?

| Option | Description | Selected |
|--------|-------------|----------|
| Hand-curated expected outputs | Author expected ASCII per case; review during Phase 1; no Perl dep | ✓ |
| Diff against Perl cowsay output | Verifies parity with upstream; adds Perl dep to CI | |
| Generate-on-first-run | First `-update` produces goldens; risk of locking-in early bugs | |

**User's choice:** Hand-curated expected outputs
**Notes:** Reviewed during Phase 1 execution; no external runtime dependency

### Q3 — What input messages should golden tests cover?

| Option | Description | Selected |
|--------|-------------|----------|
| Short single-line (`hello`) | Smallest case — exercises `< >` borders | ✓ |
| Multi-line (`line1\nline2`) | Exercises `/ | \` multi-line borders | ✓ |
| Empty input | Edge case from INPUT-04 — Phase 1 only does positional args but should still produce a valid empty bubble | |
| ASCII-only long string (no wrapping in Phase 1) | Locks "long input is one long bubble line" as documented Phase 1 behaviour | |

**User's choice:** Short single-line + Multi-line
**Notes:** Empty-input and long-input goldens deferred to Phase 2 / Phase 3 when those code paths land

### Q4 — Where should the golden test fixtures live?

| Option | Description | Selected |
|--------|-------------|----------|
| `internal/cowsay/testdata/golden/` | Standard Go convention; `goldie` default; tests next to code | ✓ |
| `internal/cowsay/cowfile_test_golden/` | Custom directory; loses some goldie defaults | |
| Top-level `testdata/golden/` | Centralized; less idiomatic for `internal/` packages | |

**User's choice:** `internal/cowsay/testdata/golden/`
**Notes:** Standard idiom

---

## Display-width abstraction now vs later

### Q1 — Build a `displayWidth(s string) int` seam in Phase 1?

| Option | Description | Selected |
|--------|-------------|----------|
| Yes — build the seam now | Phase 1 defines `displayWidth(s) = utf8.RuneCountInString(s)`; Phase 3 just swaps the body | ✓ |
| No — hardcode `len([]rune(s))` and refactor in Phase 3 | Slightly more refactor surface; fine for a small codebase | |
| You decide | Claude picks based on Phase 1 simplicity vs Phase 3 ergonomics | |

**User's choice:** Build the seam now
**Notes:** Small upfront cost; clean diff in Phase 3

### Q2 — Where should `displayWidth` live?

| Option | Description | Selected |
|--------|-------------|----------|
| In `internal/cowsay/balloon.go` as an unexported function | Next to its only caller in Phase 1 | ✓ |
| Its own `internal/cowsay/width.go` file from the start | Clear separation; signals "this is the seam" | |
| Inline as a `var displayWidth = func(...)` closure variable | Runtime swap capability we don't actually need | |

**User's choice:** Unexported function in `balloon.go`
**Notes:** Promote to a `width.go` file only if a second caller emerges in Phase 3

### Q3 — Should Phase 1 add a test that documents 'non-ASCII width is broken until Phase 3'?

| Option | Description | Selected |
|--------|-------------|----------|
| Yes — add a `t.Skip(...)` test with a comment | Pending test that Phase 3 just removes the skip on | ✓ |
| No — just a comment in balloon.go | Function-level comment only | |
| Nothing — REQUIREMENTS.md already documents it | Don't litter code with future-pointers | |

**User's choice:** Add a skipped/pending test
**Notes:** The Phase 3 change becomes a one-line removal; test artefact documents the gap

---

## Claude's Discretion

- Internal helper naming, error-message wording, and detailed package doc strings within the layout constraints above
- The exact set of "3-5 pitfall-targeted cows" subject to the requirement that at least one uses a non-`EOC` terminator and at least one has backslash escapes
- `bufio.Scanner` vs `bufio.Reader` for the heredoc parser (default to `bufio.Scanner` per ARCHITECTURE.md)

## Deferred Ideas

None — discussion stayed within Phase 1 scope. Items already deferred by the roadmap (Phase 2: stdin / `-f` / `-l` / `--random`; Phase 3: wrap / think / eyes / tongue / runewidth / `-h`; Phase 4: release pipeline) are documented in REQUIREMENTS.md traceability and ROADMAP.md.
