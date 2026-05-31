# Phase 3: Full Flag Surface - Research

**Researched:** 2026-05-31
**Domain:** Go CLI flag surface, display-width text layout, Unicode balloon rendering
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Word Wrap (RENDER-05)**
- D-01: Default wrap width is 40 columns; `-W <n>` overrides it; `-n` disables word-wrapping entirely.
- D-02: Long words that exceed the wrap width are hard-broken mid-word (Perl-cowsay behavior) so the bubble is always a guaranteed rectangle no wider than the requested width. No overflow mode.
- D-03: Wrapping is display-width aware — both the wrap boundary and the mid-word hard-break count display columns (`runewidth.StringWidth`), not bytes or runes. This shares the same width primitive as D-06 so a CJK string wraps and sizes consistently.
- D-04: Existing newlines are preserved, then each resulting line is wrapped to the width. Input is split on `\n` first (building on the current `strings.Split(message, "\n")` foundation in `balloon.go`), then word-wrap applies within each line. `-n` keeps the explicit line breaks but skips the word-wrapping step.

**Eyes / Tongue (RENDER-08)**
- D-05: `-e <eyes>` and `-T <tongue>` pass through verbatim — any length, no validation, no truncation. Matches upstream cowsay exactly and requires zero validation code.
- D-06 (eyes/tongue edge): `RenderOpts` currently treats an empty `Eyes`/`Tongue`/`Thoughts` string as "use the default" (`renderer.go` default-fills empties). An explicit `-e ""` is indistinguishable from "flag not passed." Acceptable to leave as-is, but the planner should consciously decide (e.g., a sentinel/`*string` or "explicitly set" tracking) rather than discover it by accident.

**Display Width / runewidth (RENDER-06)**
- D-07: Swap the body of the existing `displayWidth(s string) int` seam in `internal/cowsay/balloon.go` from `utf8.RuneCountInString(s)` to `runewidth.StringWidth(s)`. `github.com/mattn/go-runewidth` is the one approved external dependency for this phase.
- D-08: Fix the byte-padding bug alongside the width swap: `balloon.go` currently pads with `fmt.Fprintf("%-*s", maxWidth, line)`, and Go's `%-*s` pads by byte width, not display width. Padding must be computed in display columns (`maxWidth - displayWidth(line)` trailing spaces).
- D-09: Remove the Phase 1 `t.Skip(...)`-marked CJK golden test and replace it with a real golden that asserts the correctly-aligned bubble for `漢字テスト`.

**Think Mode (RENDER-07)**
- D-10: `--think` renders a thought bubble: the bubble border uses `(` on the left and `)` on the right of every content line — both single-line and every line of multi-line input (replacing `< >` and the angled `/ | \`). Top/bottom borders remain underscores/dashes.
- D-11: `--think` sets `RenderOpts.Thoughts = "o"` (vs the say-mode default `\`), driving the `$thoughts` substitution already wired in `renderer.go`.
- D-12: `--think` is long-only; no `-t` alias.

**Help (HELP-01)**
- D-13: Explicit `-h`/`--help` prints the full usage to stdout and exits 0. Error-triggered usage continues to go to stderr with a non-zero exit.
- D-14: Implementation: `flag` package (in `ContinueOnError` mode) returns `flag.ErrHelp` when `-h`/`-help`/`--help` is requested. `run()` must `errors.Is(err, flag.ErrHelp)` after `fs.Parse`, print the full help to stdout, and return 0.
- D-15: Help text is a clean Go-native usage block: a synopsis line, each flag with a one-line description, then 2–3 example invocations. Does NOT reproduce Perl cowsay's help layout. Examples are mandatory.

### Claude's Discretion
- Exact `RenderOpts` field shape for width/no-wrap/think (e.g. `Width int`, `NoWrap bool`, `Think bool`) and whether wrapping lives in the renderer, the balloon builder, or a small `wrap.go` helper.
- Exact wrap algorithm (greedy word packing via `strings.Fields` + per-word display-width accounting) and how the mid-word hard-break splits a too-long word.
- `-W` edge cases (`-W 0`, negative, or absurdly large) — pick sane behavior.
- East-Asian ambiguous-width characters: default to `go-runewidth`'s standard (ambiguous = width 1; do not enable `EastAsianWidth`) unless a golden test reveals a concrete problem.
- Exact help-text wording, flag-description phrasing, and which 2–3 examples to show.
- Whether to promote `displayWidth` to its own `width.go` file now that a second consideration (wrap) shares it.

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within Phase 3 scope. Release pipeline (GoReleaser, GitHub Actions, `--version`, `go install`) remains Phase 4. Mood-preset flags and a `-t` alias remain explicitly Out-of-Scope.
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| RENDER-05 | Word-wrap defaults to 40 columns; `-W <n>` overrides; `-n` disables wrapping | Display-width-aware greedy wrap + hard-break algorithm; `RenderOpts` field additions; balloon.go integration |
| RENDER-06 | Bubble sizing uses display-width (`runewidth.StringWidth`), not byte/rune count | `go-runewidth` v0.0.24 API verified; `displayWidth` seam swap; `%-*s` padding fix; CJK golden replacement |
| RENDER-07 | `--think` flag swaps to thought-bubble form (`( )` borders, `o` thoughts trail) | Upstream Perl source verified: `( )` on every line in think mode; `substituteVars` already handles `Thoughts="o"` |
| RENDER-08 | `-e <xx>` customises eyes; `-T <xx>` customises tongue | Pass-through verbatim; `substituteVars` already handles substitution; D-06 sentinel edge case noted |
| HELP-01 | `-h`/`--help` prints usage with every flag documented and example invocations | `flag.ErrHelp` behavior verified by experiment; no-op Usage pattern to suppress double-print; stdout + exit 0 path |
</phase_requirements>

---

## Summary

Phase 3 adds five user-visible capabilities on top of the working Phase 2 shell: word wrap, display-width correction, think mode, custom eyes/tongue, and help output. The integration surface is narrow — all changes land in `balloon.go`, `renderer.go`, `cmd/gosay/main.go`, and the test files. No new packages are created; `RenderOpts` gains three new fields; `buildBalloon` gains a `think bool` parameter and a wrap step upstream of balloon construction.

The most technically nuanced work is the two-part display-width fix (D-07 + D-08): swapping `displayWidth` from `utf8.RuneCountInString` to `runewidth.StringWidth` is necessary but not sufficient — `fmt.Fprintf("%-*s", maxWidth, line)` pads by bytes, so multi-byte lines get under-padded. The correct pattern is to manually append `strings.Repeat(" ", maxWidth-displayWidth(line))` trailing spaces. Both changes must land in the same commit for the CJK golden test to pass.

The `flag.ErrHelp` interception requires one non-obvious setup step: `fs.Usage` must be set to a no-op function before calling `fs.Parse`, because the flag package calls `fs.Usage()` automatically before returning `ErrHelp`. Without the no-op, setting `fs.SetOutput(stderr)` means the flag package will print to stderr before the caller gets to redirect help to stdout. Verified by live experiment.

**Primary recommendation:** Implement in dependency order — (1) add `go-runewidth` + fix `displayWidth` + fix padding bug, (2) add wrap step to `buildBalloon`, (3) add `--think` border variant, (4) wire `-W`/`-n`/`--think`/`-e`/`-T` flags in `main.go`, (5) wire `flag.ErrHelp` help path, (6) update/add golden and unit tests.

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Display-width measurement | `internal/cowsay/balloon.go` (`displayWidth` seam) | — | Seam was purpose-built in Phase 1 for this exact swap; all callers go through one function |
| Word wrap | `internal/cowsay/balloon.go` (pre-balloon step) | Possibly `internal/cowsay/wrap.go` helper | Wrapping must happen before `buildBalloon` so max line width is computed on already-wrapped lines |
| Balloon border selection (say vs think) | `internal/cowsay/balloon.go` (`buildBalloon`) | — | `buildBalloon` already owns all border logic |
| `Thoughts` variable substitution | `internal/cowsay/renderer.go` (`substituteVars`) | — | Already wired; `--think` just sets `RenderOpts.Thoughts="o"` before calling `Render` |
| Flag registration and parsing | `cmd/gosay/main.go` (`run()`) | — | Thin-main pattern; all flags live here |
| Help text | `cmd/gosay/main.go` (`run()`) | — | Help intercepts `flag.ErrHelp` inside `run()`; printed to `stdout` writer injected into `run()` |
| Option threading | `internal/cowsay/RenderOpts` struct | — | Phase 1 established the struct; Phase 3 adds `Width int`, `NoWrap bool`, `Think bool` fields |

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/mattn/go-runewidth` | v0.0.24 [VERIFIED: pkg.go.dev] | Display-width measurement for CJK/emoji/combining chars | 3,302 packages import it; the universal Go terminal-width library; MIT license; 13-year history |
| `flag` (stdlib) | stdlib | CLI flag parsing | Already in use; Phase 3 adds 5 new flags to the existing `FlagSet` |
| `github.com/sebdah/goldie/v2` | v2.8.0 [VERIFIED: pkg.go.dev] | Golden-file testing for new render paths | Already in use; `go test -update ./...` regenerates fixtures |

### Supporting (transitive — not called directly)
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/clipperhouse/uax29/v2` | v2.2.0 [VERIFIED: pkg.go.dev] | Unicode grapheme segmentation used internally by `go-runewidth` v0.0.24 | Indirect dep — `go-runewidth` uses it for grapheme-cluster-aware `StringWidth` on non-ASCII strings; do not import directly |

### Installation
```bash
go get github.com/mattn/go-runewidth@v0.0.24
go mod tidy
```

Running `go get` on `go-runewidth` v0.0.24 also pulls `github.com/clipperhouse/uax29/v2 v2.2.0` as an indirect dependency. This is expected — `go-runewidth` v0.0.24 introduced grapheme-cluster-aware width measurement. Both entries will appear in `go.mod`; `go mod tidy` will mark them appropriately.

**Version verification:**
```
github.com/mattn/go-runewidth latest = v0.0.24 (published 2026-05-29) [VERIFIED: pkg.go.dev]
github.com/clipperhouse/uax29/v2 = v2.2.0 (published 2025-09-14) [VERIFIED: go mod download]
```

---

## Package Legitimacy Audit

> slopcheck was available but returned a non-JSON error when invoked against Go module paths (it targets npm packages). Manual verification performed instead — see notes below.

| Package | Registry | Age | Imported By | Source Repo | slopcheck | Disposition |
|---------|----------|-----|-------------|-------------|-----------|-------------|
| `github.com/mattn/go-runewidth` | pkg.go.dev (Go) | 13 years (created 2013-06-21) | 3,302 modules | github.com/mattn/go-runewidth | N/A (Go, not npm) | Approved [VERIFIED: pkg.go.dev + GitHub API] |
| `github.com/clipperhouse/uax29/v2` | pkg.go.dev (Go) | 6 years (created 2020-04-15) | Legitimate — Unicode spec implementation | github.com/clipperhouse/uax29 | N/A (Go, not npm) | Approved [VERIFIED: pkg.go.dev + GitHub API] |

**Notes:**
- `go-runewidth`: 700 GitHub stars, 97 forks, actively maintained (last updated 2026-05-30). MIT license. Used by virtually every Go terminal UI library (bubbletea, tcell, lipgloss, etc.).
- `clipperhouse/uax29`: Implements UAX #29 Unicode text segmentation. 116 stars, actively maintained. This is a legitimate Unicode standard implementation library.
- slopcheck is an npm-focused tool; it does not evaluate Go modules. Verification was done via `pkg.go.dev`, GitHub API stats, and `go mod download` (cryptographic hash verification via go.sum).

**Packages removed due to slopcheck [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

---

## Architecture Patterns

### System Architecture Diagram

```
gosay "long message" | gosay --think -W 20 -e "^^" -T "~~"
        │
        ▼
cmd/gosay/main.go — run(args, stdout, stderr) int
  fs.Parse → ErrHelp? → print full help to stdout → return 0
  fs.Parse → other err → print usage to stderr → return 1
  Parse -W (int,40), -n (bool), --think (bool), -e (string), -T (string)
  Build RenderOpts{Eyes, Tongue, Thoughts("o" if think), Width, NoWrap, Think}
        │
        ▼
internal/cowsay/Render(animal, message, opts)
        │
        ├── balloon.go: wrapMessage(message, opts.Width, opts.NoWrap)
        │       split on "\n" → for each line: wrapWords(line, width)
        │       wrapWords: greedy word packing with displayWidth accounting
        │       hardBreak: split word at display-column boundary (never mid-rune)
        │
        ├── balloon.go: buildBalloon(wrappedLines, opts.Think)
        │       displayWidth() → runewidth.StringWidth() [D-07]
        │       maxWidth = max display width across all lines
        │       top border = "_" × (maxWidth+2)
        │       think=true  → every line: "( " + padRight(line,maxWidth) + " )"
        │       think=false, 1 line → "< " + padRight(line,maxWidth) + " >"
        │       think=false, N lines → "/ | \" border set, padRight on each
        │       padRight = line + spaces×(maxWidth-displayWidth(line)) [D-08]
        │       bottom border = "-" × (maxWidth+2)
        │
        └── renderer.go: substituteVars(cowBody, opts)
                $thoughts → opts.Thoughts ("o" think / "\" say)
                $eyes → opts.Eyes (default "oo")
                $tongue → opts.Tongue (default "  ")
                → balloon + substituted cow body
```

### Recommended Project Structure Changes

```
internal/cowsay/
├── balloon.go          ← MODIFIED: displayWidth body swap + padding fix + wrap step + think borders
├── balloon_test.go     ← MODIFIED: update TestDisplayWidth for CJK (expect 4 not 2); add wrap + think tests
├── renderer.go         ← MODIFIED: add Width int, NoWrap bool, Think bool to RenderOpts
├── golden_test.go      ← MODIFIED: replace t.Skip CJK test; add think/wrap/eyes golden tests
├── testdata/golden/
│   ├── cjk_aligned_gopher.golden   ← NEW (replaces cjk_skip.golden placeholder)
│   ├── think_say_hello.golden      ← NEW
│   ├── wrap_long_message.golden    ← NEW
│   └── custom_eyes_tongue.golden   ← NEW
cmd/gosay/
├── main.go             ← MODIFIED: register -W/-n/--think/-e/-T; ErrHelp intercept; full help block
└── main_test.go        ← MODIFIED: add tests for -W/-n/--think/-e/-T/-h/--help
```

### Pattern 1: Display-Width-Aware Padding (D-07 + D-08 fix)

**What:** Replace the Phase 1 `displayWidth` body AND the `fmt.Fprintf("%-*s", ...)` padding call. Both must change together.
**When to use:** Every place a line is padded to `maxWidth` inside `buildBalloon`.
**Example:**
```go
// Source: verified by running /tmp/padtest.go in research session

// Step 1: swap displayWidth body
import "github.com/mattn/go-runewidth"

func displayWidth(s string) int {
    return runewidth.StringWidth(s)
}

// Step 2: replace fmt.Fprintf("%-*s", maxWidth, line) everywhere in buildBalloon
func padRight(s string, targetWidth int) string {
    w := displayWidth(s)
    if w >= targetWidth {
        return s
    }
    return s + strings.Repeat(" ", targetWidth-w)
}

// Usage in buildBalloon (single-line say):
fmt.Fprintf(&b, "< %s >\n", padRight(line, maxWidth))
// Usage in buildBalloon (multi-line say):
fmt.Fprintf(&b, "%s %s %s\n", left, padRight(line, maxWidth), right)
// Usage in buildBalloon (think, any line count):
fmt.Fprintf(&b, "( %s )\n", padRight(line, maxWidth))
```

### Pattern 2: Display-Width-Aware Word Wrap (D-02, D-03, D-04)

**What:** Greedy word-packing with display-width accounting + hard-break for overlong words at display-column boundary. Never splits inside a rune.
**When to use:** `wrapMessage()` called before `buildBalloon`, once per `Render` call.
**Example:**
```go
// Source: verified by running /tmp/wraptest.go in research session

// wrapMessage preserves existing newlines then word-wraps each resulting line.
// width <= 0 means no-wrap (pass through with newlines preserved).
func wrapMessage(message string, width int) string {
    if width <= 0 {
        return message
    }
    inputLines := strings.Split(message, "\n")
    var out []string
    for _, line := range inputLines {
        out = append(out, wrapWords(line, width)...)
    }
    return strings.Join(out, "\n")
}

// wrapWords wraps a single input line (no existing newlines) to width display columns.
// Long words are hard-broken at display-column boundaries, never inside a rune.
func wrapWords(line string, width int) []string {
    words := strings.Fields(line)
    if len(words) == 0 {
        return []string{""} // preserve blank lines
    }
    var result []string
    var current strings.Builder
    currentW := 0
    for _, word := range words {
        wordW := displayWidth(word)
        if currentW == 0 {
            if wordW <= width {
                current.WriteString(word)
                currentW = wordW
            } else {
                chunks := hardBreak(word, width)
                for i, chunk := range chunks {
                    if i < len(chunks)-1 {
                        result = append(result, chunk)
                    } else {
                        current.WriteString(chunk)
                        currentW = displayWidth(chunk)
                    }
                }
            }
        } else if currentW+1+wordW <= width {
            current.WriteByte(' ')
            current.WriteString(word)
            currentW += 1 + wordW
        } else {
            result = append(result, current.String())
            current.Reset()
            currentW = 0
            if wordW <= width {
                current.WriteString(word)
                currentW = wordW
            } else {
                chunks := hardBreak(word, width)
                for i, chunk := range chunks {
                    if i < len(chunks)-1 {
                        result = append(result, chunk)
                    } else {
                        current.WriteString(chunk)
                        currentW = displayWidth(chunk)
                    }
                }
            }
        }
    }
    if current.Len() > 0 {
        result = append(result, current.String())
    }
    return result
}

// hardBreak splits s into chunks of at most width display columns.
// Rune boundaries are respected — never splits inside a multi-byte rune.
func hardBreak(s string, width int) []string {
    var result []string
    for len(s) > 0 {
        var chunk strings.Builder
        chunkW := 0
        remaining := s
        for len(remaining) > 0 {
            r, size := utf8.DecodeRuneInString(remaining)
            rw := displayWidth(string(r))
            if chunkW+rw > width {
                break
            }
            chunk.WriteRune(r)
            chunkW += rw
            remaining = remaining[size:]
        }
        if chunk.Len() == 0 {
            break // safety: single rune wider than width — skip it
        }
        result = append(result, chunk.String())
        s = s[len(s)-len(remaining):]
    }
    return result
}
```

### Pattern 3: Think-Mode Balloon Border (D-10, confirmed against upstream Perl source)

**What:** Upstream cowsay's `construct_balloon` uses `( )` on ALL 6 border positions (up-left, up-right, down-left, down-right, left, right) in think mode. This means every content line, whether single-line or multi-line, gets `( )` borders. The existing say-mode branching (`< >` for single-line, `/ | \` for multi-line) does NOT apply in think mode.

**Upstream Perl (authoritative):**
```perl
if ($0 =~ /think/i) {
    $thoughts = 'o';
    @border = qw[ ( ) ( ) ( ) ];  # all 6 positions are ( or )
}
```
[CITED: github.com/cowsay-org/cowsay bin/cowsay — `construct_balloon` subroutine]

**Go implementation:**
```go
if think {
    for _, line := range lines {
        fmt.Fprintf(&b, "( %s )\n", padRight(line, maxWidth))
    }
} else if len(lines) == 1 {
    fmt.Fprintf(&b, "< %s >\n", padRight(lines[0], maxWidth))
} else {
    for i, line := range lines {
        // existing / | \ branching
    }
}
```

**Verified output** (from research session):
```
 ____________ 
( 漢字テスト )
( hello      )
( abc        )
 ------------ 
```

### Pattern 4: `flag.ErrHelp` Interception (D-13, D-14)

**What:** The `flag` package calls `fs.Usage()` automatically before returning `ErrHelp`. If `fs.SetOutput(stderr)` is active, this causes help to be printed to stderr before the caller can redirect it to stdout. The fix: set `fs.Usage` to a no-op before calling `fs.Parse`, then print full help manually to `stdout` if `ErrHelp` is detected.

**Verified by live experiment:**
```go
// Source: verified by running /tmp/flagtest3.go in research session

fs := flag.NewFlagSet("gosay", flag.ContinueOnError)
fs.SetOutput(stderr)  // parse errors go to stderr

// CRITICAL: set no-op before Parse to suppress automatic Usage call on -h
fs.Usage = func() {}

if err := fs.Parse(args); err != nil {
    if errors.Is(err, flag.ErrHelp) {
        // -h or --help was requested — print full help to stdout, exit 0
        printFullHelp(stdout, fs)
        return 0
    }
    // Other parse errors: print usage to stderr, exit 1
    printUsage(stderr, fs)
    return 1
}
```

**Key facts verified:**
- `flag.ErrHelp` is returned by `fs.Parse` when `-h`, `-help`, or `--help` is encountered and no such flag is explicitly defined [VERIFIED: `go doc flag ErrHelp` + `go doc flag.FlagSet.Parse`]
- The flag package calls `fs.Usage()` BEFORE returning `ErrHelp` in `ContinueOnError` mode [VERIFIED: live experiment + `go doc -src flag.FlagSet.Parse`]
- `errors.Is(err, flag.ErrHelp)` is the correct detection idiom [VERIFIED: same]
- If Phase 3 registers `--think`, `-W`, `-n`, `-e`, `-T` but NOT `-h`, then passing `-h` correctly triggers `ErrHelp` [VERIFIED: live experiment]

### Pattern 5: `RenderOpts` Extension

**What:** Three new fields added to `RenderOpts` to carry wrap/think options from `main.go` to `Render`/`buildBalloon`.
**Example:**
```go
type RenderOpts struct {
    Eyes     string // default "oo"
    Tongue   string // default "  " (two spaces)
    Thoughts string // default "\" (say mode) or "o" (think mode)
    Width    int    // default 40; 0 means "use default 40" unless NoWrap is set
    NoWrap   bool   // -n flag: skip wrapping entirely
    Think    bool   // --think: use ( ) borders and set Thoughts="o"
}
```

**Threading in `Render`:**
```go
func Render(animal, message string, opts RenderOpts) (string, error) {
    // Thread think mode into Thoughts before substituteVars
    if opts.Think && opts.Thoughts == "" {
        opts.Thoughts = "o"
    }
    
    // Determine wrap width
    wrapWidth := opts.Width
    if wrapWidth <= 0 && !opts.NoWrap {
        wrapWidth = 40 // default
    }
    if opts.NoWrap {
        wrapWidth = 0 // signals wrapMessage to skip wrapping
    }
    
    wrappedMessage := wrapMessage(message, wrapWidth)
    wrappedLines := strings.Split(wrappedMessage, "\n")
    
    balloon := buildBalloon(wrappedLines, opts.Think)
    // ...
}
```

### Anti-Patterns to Avoid

- **Swapping only `displayWidth` body without fixing `%-*s` padding:** The bubble will measure correctly but pad incorrectly. Both changes must land together. [VERIFIED: live experiment shows `fmt.Sprintf("%-*s", 10, "漢字テスト")` produces 20-char string, not 15-char]
- **Wrapping after balloon build:** Wrap must happen before `buildBalloon` so that `maxWidth` reflects already-wrapped line widths, not the original long line.
- **Splitting a rune mid-byte in `hardBreak`:** Always use `utf8.DecodeRuneInString` to find rune boundaries; never slice `s[i:]` using a byte index that lands inside a multi-byte rune.
- **Setting `fs.Usage` AFTER calling `fs.Parse`:** The Usage interception for ErrHelp only works if the no-op is set before Parse.
- **Using `-t` as a think alias:** `-t` is reserved (see Out-of-Scope mood flags). D-12 explicitly prohibits it.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Terminal display width of Unicode strings | Custom East Asian Width table | `runewidth.StringWidth(s)` | UAX #11 tables are 3,000+ entries; `go-runewidth` handles CJK, emoji, combining, surrogates correctly and is actively maintained |
| Word wrap with display-width accounting | Nothing — hand-roll this | `wrapWords()` helper (30 lines per CLAUDE.md) | `mitchellh/go-wordwrap` is archived 2020; no display-width awareness; `go-runewidth` has a `Wrap()` method but it doesn't match gosay's hard-break behavior |
| Unicode grapheme segmentation | Custom grapheme cluster logic | `go-runewidth` v0.0.24 (uses `clipperhouse/uax29` internally) | The indirect dep is purposely used by `runewidth`; do not import `uax29` directly |

**Key insight:** `go-runewidth` has a `Wrap(s string, w int)` convenience function, but it wraps on word boundaries without hard-breaking long words. Since D-02 requires hard-breaking long words (Perl-cowsay behavior), the wrap step must be hand-rolled using `runewidth.StringWidth` as the measurement primitive.

---

## Common Pitfalls

### Pitfall 1: The Two-Part Display-Width Fix
**What goes wrong:** Developer swaps `displayWidth` from rune count to `runewidth.StringWidth` but leaves `fmt.Fprintf("%-*s", maxWidth, line)` unchanged. CJK bubble right borders still misalign.
**Why it happens:** `%-*s` pads by **byte** width in Go. For `漢字テスト` (15 bytes, 5 runes, 10 display cols) with `maxWidth=10`, `fmt.Sprintf("%-*s", 10, "漢字テスト")` produces a 20-byte string (no-op padding) instead of the correct 15-byte string. The right border appears 5 columns too far to the right.
**How to avoid:** Replace every `fmt.Fprintf(&b, "... %-*s ...", maxWidth, line, ...)` with `fmt.Fprintf(&b, "... %s ...", padRight(line, maxWidth), ...)`.
**Warning signs:** `echo "漢字テスト" | gosay` shows a right border that doesn't align with the top/bottom borders.

### Pitfall 2: `flag.ErrHelp` Prints to stderr Before Return
**What goes wrong:** Developer sets `fs.SetOutput(stderr)` and `fs.Usage = fullHelpFunc`, expecting to control where help goes after `ErrHelp`. But the flag package calls `fs.Usage()` BEFORE returning `ErrHelp`. Help goes to stderr even though the caller tries to redirect it to stdout.
**Why it happens:** `flag.FlagSet.Parse` calls `f.usage()` (which calls `fs.Usage`) as part of the `ErrHelp` handling path in `ContinueOnError` mode — before returning the error. Verified in Go stdlib source via `go doc -src flag.FlagSet.Parse`.
**How to avoid:** Set `fs.Usage = func() {}` (no-op) before `fs.Parse`. Detect `ErrHelp` afterward and print the full help block manually to `stdout`.

### Pitfall 3: Wrap Step After Balloon Build
**What goes wrong:** Wrapping is applied to the final output string (after balloon build) rather than to the message before balloon build. Result: wrap breaks appear inside the balloon borders, mangling the box.
**Why it happens:** Seems like "just format the output"; misses that `buildBalloon` uses `strings.Split(message, "\n")` to determine line count and `maxWidth`.
**How to avoid:** Wrap must run first: `wrappedLines := strings.Split(wrapMessage(message, width), "\n")`, then `buildBalloon(wrappedLines, think)`.
**Warning signs:** `gosay -W 5 "hello world"` produces a balloon with the text `hello\nworld` inside single-line `< >` brackets.

### Pitfall 4: `TestDisplayWidth` Still Expects Rune-Count Behavior
**What goes wrong:** Swapping `displayWidth` body breaks the existing `balloon_test.go` test case that asserts `displayWidth("漢字") == 2` (the rune-count value). Go tests fail red.
**Why it happens:** The Phase 1 test explicitly documents the rune-count limitation — after Phase 3, the expectation must flip to `4` (the correct display width).
**How to avoid:** When swapping `displayWidth`, update `TestDisplayWidth` to expect `displayWidth("漢字") == 4`. Also remove the `t.Skip` from `TestGolden_CJK_Skipped` and replace it with a passing golden assertion.
**Warning signs:** `go test ./internal/cowsay/... ` fails with `displayWidth("漢字") = 4, want 2`.

### Pitfall 5: rune Splitting in `hardBreak`
**What goes wrong:** `hardBreak` slices the string at a byte offset that falls inside a multi-byte rune, producing invalid UTF-8 output.
**Why it happens:** Temptation to use `s[i:]` where `i` was incremented by display width rather than by rune size.
**How to avoid:** Always use `utf8.DecodeRuneInString(remaining)` to get both the rune and its byte `size`, then advance by `remaining = remaining[size:]`.

### Pitfall 6: `-W 0` or Negative Width
**What goes wrong:** User passes `-W 0` or `-W -5`; undefined behavior in the wrap algorithm.
**How to avoid:** In `Render` (or `wrapMessage`), treat `Width <= 0` as "use default 40" unless `NoWrap` is explicitly set. Document this in the help text.

---

## Code Examples

### Correct Balloon Width Computation (post-Phase 3)
```go
// Source: verified by running /tmp/balloontest.go in research session
// 漢字テスト = 5 CJK chars, display width 10, but only 5 runes / 15 bytes

// After Phase 3: displayWidth("漢字テスト") == 10 (correct)
// Before Phase 3: displayWidth("漢字テスト") == 5 (wrong — rune count)

// CJK say bubble (multi-line):
//  ____________ 
// / 漢字テスト \    ← right border at column 14 = 1 + 1 + 10 + 1 + 1
// | hello      |    ← "hello" (5 cols) + 5 spaces = 10, right border at col 14
// \ abc        /    ← "abc" (3 cols) + 7 spaces = 10
//  ------------ 

// CJK think bubble:
//  ____________ 
// ( 漢字テスト )    ← every line uses ( )
// ( hello      )
// ( abc        )
//  ------------ 
```

### Updated `TestDisplayWidth` Expected Values
```go
// Source: existing balloon_test.go; MUST be updated in Phase 3
{"漢字", 2},  // Phase 1 expected value — WRONG after runewidth swap
// becomes:
{"漢字", 4},  // Phase 3 correct value: 2 CJK chars × 2 display cols each
```

### Help Text Structure (D-15)
```go
// Source: CONTEXT.md D-15 + flag package idioms
const helpText = `gosay — make a gopher say something

Usage: gosay [flags] [message...]

Flags:
  -e <eyes>     Set eye characters (default "oo")
  -f <name>     Select animal from embedded set (default "gopher")
  -l            List available animals
  -n            Disable word wrapping (preserve all input whitespace)
  -T <tongue>   Set tongue characters (default "  ")
  -W <cols>     Wrap message at this many display columns (default 40)
  --random      Pick a random animal
  --think       Use thought bubble ( ) instead of speech bubble < >

Examples:
  gosay hello
  echo hi | gosay -f tux
  gosay --think -e "^^" "thinking..."`

// In run():
fs.Usage = func() {} // no-op to suppress automatic stderr print on -h
if err := fs.Parse(args); err != nil {
    if errors.Is(err, flag.ErrHelp) {
        fmt.Fprintln(stdout, helpText)  // stdout + exit 0
        return 0
    }
    fmt.Fprintln(stderr, "usage: gosay [flags] [message...]")
    return 1
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `utf8.RuneCountInString` for display width | `runewidth.StringWidth` | Phase 3 / RENDER-06 | CJK/emoji bubbles now align correctly |
| `fmt.Fprintf("%-*s", maxWidth, line)` byte padding | Manual `strings.Repeat(" ", maxWidth-displayWidth(line))` | Phase 3 / D-08 | Right border aligns for CJK input |
| No word wrap | Greedy display-width-aware wrap at 40 cols | Phase 3 / RENDER-05 | Long messages fit a terminal column |
| No think mode | `--think` sets `( )` borders + `Thoughts="o"` | Phase 3 / RENDER-07 | Thought bubble mode available |
| Hard-coded `RenderOpts{}` in main.go | `RenderOpts{Width, NoWrap, Think, Eyes, Tongue}` | Phase 3 | All flags wire through |

**go-runewidth version note:** v0.0.24 (published 2026-05-29) introduced a transitive dependency on `github.com/clipperhouse/uax29/v2` for grapheme-cluster-aware width measurement. Earlier versions (e.g. v0.0.16) used `rivo/uniseg`. The extra transitive dep is expected and legitimate. [VERIFIED: go.mod inspection of cached module]

**Deprecated/outdated:**
- `utf8.RuneCountInString` for display width: still valid for pure code-point counting but incorrect for terminal display width. Retain only in `TestDisplayWidth` update notes.
- `%-*s` format verb for display-aligned padding: still valid for byte-aligned padding of ASCII strings; avoid for Unicode terminal output.

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `go-runewidth` v0.0.24 `EastAsianWidth` default is `false` (ambiguous chars = width 1) | Standard Stack | Ambiguous-width CJK chars might render at width 2, making the CJK golden fail. Mitigation: verify with the golden test before committing. | [ASSUMED — verified `DefaultCondition.EastAsianWidth=false` in source, but runtime locale detection via `handleEnv()` could override on a CJK system] |
| A2 | `runewidth.StringWidth` correctly counts `漢字テスト` as 10 (5 × 2-wide chars) | Code Examples | Golden test would fail with wrong width. Low risk — tested with standard Unicode CJK double-width codepoints. [ASSUMED — not verified with a live binary against the golden file] |

**If this table is empty:** N/A — two assumed claims are listed above.

---

## Open Questions (RESOLVED)

All three were Claude's-Discretion items; resolved during planning (commit `2a88fd3`).

1. **D-06 sentinel for explicit `-e ""`** — **RESOLVED:** declined. Plan 03-04 Task 1 consciously accepts that `-e ""` maps to the default and does NOT add a `*string`/`EyesSet` sentinel — keeping `main.go` thin per D-06. The verbatim-passthrough decision (D-05) makes this edge pathological and not worth the added surface.
   - What we know: `renderer.go` treats `opts.Eyes == ""` as "use default oo". An explicit `-e ""` is indistinguishable from "flag not passed."
   - Original recommendation: `*string` or `EyesSet bool`. Rejected in favor of simplicity.

2. **`displayWidth` promotion to `width.go`** — **RESOLVED:** promoted. Plan 03-02 Task 1 creates `internal/cowsay/wrap.go` housing `wrapMessage`/`wrapWords`/`hardBreak` (and the wrap path's use of `displayWidth`), keeping `balloon.go` focused on border construction.
   - What we know: Both wrap and balloon sizing now call `displayWidth`. Phase 1 D-17 left open whether to move it to its own file.

3. **`buildBalloon` signature change** — **RESOLVED:** changed to `buildBalloon(lines []string, think bool) string`. Plan 03-02 Task 2 makes the signature final (with `think` stubbed to fall through), Plan 03-03 Task 1 fills the `( )` think branch. `Render` calls `wrapMessage` first, then passes `[]string` to `buildBalloon` — both functions stay pure and independently testable.
   - What we know: Currently `buildBalloon(message string) string` with `strings.Split` inside. Phase 3 needs it to accept already-wrapped `[]string` lines and a `think bool`.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | All compilation | ✓ | go1.26.3 | — |
| `go get` / module download | Adding `go-runewidth` | ✓ | (network access) | — |
| `go test ./...` | Running existing tests | ✓ | stdlib | — |
| `go test -update ./...` | Regenerating goldie fixtures | ✓ | goldie v2.8.0 already in go.mod | — |

**Missing dependencies with no fallback:** None.

---

## Security Domain

> `security_enforcement: true`, `security_asvs_level: 1`. Phase 3 is a pure CLI text-processing tool with no network, no auth, no database, and no file system writes (all output goes to stdout/stderr).

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | CLI tool — no users, no sessions |
| V3 Session Management | No | Stateless process |
| V4 Access Control | No | No resources to protect |
| V5 Input Validation | Partial | `-e`/`-T` are intentionally pass-through verbatim (D-05); `-W` int parsing handled by `flag` package; no injection surface since output is terminal ASCII art |
| V6 Cryptography | No | No cryptographic operations |
| V7 Error Handling | Yes | Help → stdout + exit 0; parse errors → stderr + exit 1; unknown cow → clean error without internal path leak (already enforced in Phase 2) |

### Known Threat Patterns for CLI text tools

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Malicious input causing OOM (huge `-W` value or huge message) | DoS | `flag` int parsing prevents non-numeric; very large `-W` just wraps at a large width (bounded by input length); no allocation loops proportional to `-W` value |
| Internal path disclosure via error messages | Information Disclosure | Already mitigated in Phase 2: `gosay: unknown cowfile "name"` does not include `cows/name.cow` path |
| Terminal escape injection via `-e`/`-T` | Tampering | Out of scope — ANSI color is explicitly excluded. The verbatim pass-through (D-05) is intentional and documented. Users who pass escape sequences own the consequences. No sanitization required. |

**Overall security posture:** ASVS Level 1 is trivially satisfied for a stateless ASCII-art CLI. No security-specific tasks are required in Phase 3.

---

## Sources

### Primary (HIGH confidence)
- `go doc flag ErrHelp`, `go doc flag.FlagSet.Parse`, `go doc -src flag.FlagSet.Parse` — ErrHelp behavior, ContinueOnError mode [VERIFIED]
- `/home/dev/go/pkg/mod/github.com/mattn/go-runewidth@v0.0.24/runewidth.go` — `StringWidth`, `RuneWidth`, `EastAsianWidth`, `DefaultCondition` [VERIFIED: source code read]
- Live experiment (`/tmp/flagtest*.go`, `/tmp/padtest.go`, `/tmp/wraptest.go`, `/tmp/balloontest.go`) — all core behaviors verified by running Go programs [VERIFIED]
- `pkg.go.dev/github.com/mattn/go-runewidth` v0.0.24 — version, import count, publish date [VERIFIED]
- `github.com/cowsay-org/cowsay bin/cowsay` — `construct_balloon` Perl source for think mode border specification [CITED: fetched via WebFetch]

### Secondary (MEDIUM confidence)
- `go mod download` output — confirms `go-runewidth` v0.0.24 pulls `clipperhouse/uax29/v2 v2.2.0` as indirect dep [VERIFIED]
- GitHub API (`api.github.com/repos/mattn/go-runewidth`, `api.github.com/repos/clipperhouse/uax29`) — package age, star count, maintenance status [VERIFIED]
- Existing `balloon_test.go`, `golden_test.go`, `renderer.go`, `balloon.go`, `main.go` — current code read directly; Phase 3 changes isolated [VERIFIED]

### Tertiary (LOW confidence)
- None in this research.

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — `go-runewidth` API read from source; goldie already in use; `flag` behavior verified by experiment
- Architecture: HIGH — existing code read; integration surface is narrow and well-defined by Phase 1/2 seams
- Pitfalls: HIGH — two-part display-width fix and ErrHelp interception both verified by live Go programs
- Think-mode bubble shape: HIGH — read directly from upstream Perl `construct_balloon` source

**Research date:** 2026-05-31
**Valid until:** 2026-08-31 (go-runewidth is stable; flag stdlib is stable; bubble shape matches upstream Perl)
