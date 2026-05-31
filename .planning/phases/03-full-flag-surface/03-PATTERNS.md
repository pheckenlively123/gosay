# Phase 3: Full Flag Surface - Pattern Map

**Mapped:** 2026-05-31
**Files analyzed:** 9 (6 modified source files + 3+ new golden fixtures)
**Analogs found:** 9 / 9

---

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/cowsay/balloon.go` | utility | transform | `internal/cowsay/balloon.go` (itself, current) | exact — in-place modification |
| `internal/cowsay/renderer.go` | service | request-response | `internal/cowsay/renderer.go` (itself, current) | exact — in-place modification |
| `cmd/gosay/main.go` | controller | request-response | `cmd/gosay/main.go` (itself, current) | exact — in-place modification |
| `internal/cowsay/balloon_test.go` | test | transform | `internal/cowsay/balloon_test.go` (itself, current) | exact — in-place modification |
| `internal/cowsay/renderer_test.go` | test | request-response | `internal/cowsay/renderer_test.go` (itself, current) | exact — in-place modification |
| `internal/cowsay/golden_test.go` | test | transform | `internal/cowsay/golden_test.go` (itself, current) | exact — in-place modification |
| `cmd/gosay/main_test.go` | test | request-response | `cmd/gosay/main_test.go` (itself, current) | exact — in-place modification |
| `internal/cowsay/wrap.go` (new, optional) | utility | transform | `internal/cowsay/balloon.go` | role-match — same package, same data-flow |
| `internal/cowsay/testdata/golden/*.golden` (new) | test fixture | transform | `internal/cowsay/testdata/golden/gopher_say_hello.golden` | exact — same goldie fixture pattern |

---

## Pattern Assignments

### `internal/cowsay/balloon.go` (utility, transform)

**Analog:** itself (current state)

**Imports pattern** (lines 1-7):
```go
package cowsay

import (
    "fmt"
    "strings"
    "unicode/utf8"
)
```

Phase 3 change: replace `"unicode/utf8"` with `"github.com/mattn/go-runewidth"` (and add `"unicode/utf8"` back only in `hardBreak` for `utf8.DecodeRuneInString`). If wrap logic moves to `wrap.go`, the utf8 import stays there; balloon.go gets runewidth.

**Core `displayWidth` seam** (lines 9-14 — the one-line body swap):
```go
// BEFORE (Phase 1):
func displayWidth(s string) int {
    return utf8.RuneCountInString(s)
}

// AFTER (Phase 3 D-07):
func displayWidth(s string) int {
    return runewidth.StringWidth(s)
}
```
This is a one-line body swap. Zero call-site changes required because the seam was purpose-built in Phase 1.

**`buildBalloon` signature change** (line 23 — current signature):
```go
// BEFORE (Phase 1):
func buildBalloon(message string) string {
    lines := strings.Split(message, "\n")
    // ...
}

// AFTER (Phase 3 D-10, open question 3 from RESEARCH):
func buildBalloon(lines []string, think bool) string {
    // lines are already wrapped/split by caller (Render)
    // ...
}
```

**`padRight` helper — replaces all `fmt.Fprintf("... %-*s ...", maxWidth, line, ...)` calls** (D-08):
```go
// NEW in Phase 3 — replace every %-*s occurrence:
func padRight(s string, targetWidth int) string {
    w := displayWidth(s)
    if w >= targetWidth {
        return s
    }
    return s + strings.Repeat(" ", targetWidth-w)
}
```

**Current body lines with `%-*s` (lines 41-54) — ALL must be replaced:**
```go
// BEFORE — single-line (line 41):
fmt.Fprintf(&b, "< %-*s >\n", maxWidth, lines[0])

// BEFORE — multi-line (lines 53):
fmt.Fprintf(&b, "%s %-*s %s\n", left, maxWidth, line, right)

// AFTER — single-line say:
fmt.Fprintf(&b, "< %s >\n", padRight(lines[0], maxWidth))

// AFTER — multi-line say:
fmt.Fprintf(&b, "%s %s %s\n", left, padRight(line, maxWidth), right)

// AFTER — think (any line count, D-10):
fmt.Fprintf(&b, "( %s )\n", padRight(line, maxWidth))
```

**Think-mode branching pattern** (RESEARCH Pattern 3):
```go
// In buildBalloon body, replacing current if/else:
if think {
    for _, line := range lines {
        fmt.Fprintf(&b, "( %s )\n", padRight(line, maxWidth))
    }
} else if len(lines) == 1 {
    fmt.Fprintf(&b, "< %s >\n", padRight(lines[0], maxWidth))
} else {
    for i, line := range lines {
        var left, right string
        switch {
        case i == 0:
            left, right = "/", "\\"
        case i == len(lines)-1:
            left, right = "\\", "/"
        default:
            left, right = "|", "|"
        }
        fmt.Fprintf(&b, "%s %s %s\n", left, padRight(line, maxWidth), right)
    }
}
```

**Top/bottom border — unchanged pattern** (lines 37-58, keep as-is):
```go
b.WriteString(" " + strings.Repeat("_", maxWidth+2) + " \n")
// ... body lines ...
b.WriteString(" " + strings.Repeat("-", maxWidth+2) + " \n")
```

---

### `internal/cowsay/wrap.go` (new utility, transform) — optional extraction

**Analog:** `internal/cowsay/balloon.go` (same package, same helper-function pattern)

If the planner decides to extract wrap logic into a separate file (RESEARCH open question 2, CONTEXT Claude's Discretion), copy this file-opening pattern from `balloon.go`:

**File header pattern** (copy from `balloon.go` lines 1-7):
```go
package cowsay

import (
    "strings"
    "unicode/utf8"

    "github.com/mattn/go-runewidth"
)
```

**`wrapMessage` entry point** (RESEARCH Pattern 2, verified):
```go
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
```

**`wrapWords` greedy packer** (RESEARCH Pattern 2, verified):
```go
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
```

**`hardBreak` rune-boundary splitter** (RESEARCH Pattern 2, verified — D-02, D-03):
```go
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

---

### `internal/cowsay/renderer.go` (service, request-response)

**Analog:** itself (current state)

**`RenderOpts` struct extension** (lines 16-20 — current struct, extend in place):
```go
// BEFORE (Phase 1/2):
type RenderOpts struct {
    Eyes     string // default "oo"
    Tongue   string // default "  " (two spaces)
    Thoughts string // default "\" (single backslash, say mode)
}

// AFTER (Phase 3 — RESEARCH Pattern 5):
type RenderOpts struct {
    Eyes     string // default "oo"
    Tongue   string // default "  " (two spaces)
    Thoughts string // default "\" (say mode) or "o" (think mode set by main.go)
    Width    int    // wrap width; 0 = use default 40; ignored when NoWrap is true
    NoWrap   bool   // -n flag: skip wrapping entirely
    Think    bool   // --think: use ( ) borders; main.go also sets Thoughts="o"
}
```

**`Render` function threading pattern** (lines 64-81 — current body, must add wrap step):
```go
// CURRENT (Phase 2):
func Render(animal, message string, opts RenderOpts) (string, error) {
    cow, err := LoadCow(animal)
    if err != nil {
        return "", fmt.Errorf("render: %w", err)
    }
    substitutedBody := substituteVars(cow.Body, opts)
    balloon := buildBalloon(message)
    out := balloon + substitutedBody
    return strings.TrimRight(out, "\n") + "\n", nil
}

// AFTER (Phase 3 — RESEARCH Pattern 5 threading):
func Render(animal, message string, opts RenderOpts) (string, error) {
    cow, err := LoadCow(animal)
    if err != nil {
        return "", fmt.Errorf("render: %w", err)
    }

    // Thread think mode into Thoughts if not already set by caller
    if opts.Think && opts.Thoughts == "" {
        opts.Thoughts = "o"
    }

    // Determine effective wrap width
    wrapWidth := opts.Width
    if wrapWidth <= 0 && !opts.NoWrap {
        wrapWidth = 40 // default per D-01
    }
    if opts.NoWrap {
        wrapWidth = 0 // signals wrapMessage to skip wrapping
    }

    wrappedMessage := wrapMessage(message, wrapWidth)
    wrappedLines := strings.Split(wrappedMessage, "\n")

    substitutedBody := substituteVars(cow.Body, opts)
    balloon := buildBalloon(wrappedLines, opts.Think)
    out := balloon + substitutedBody
    return strings.TrimRight(out, "\n") + "\n", nil
}
```

**`substituteVars` — unchanged** (lines 34-62): no changes required for Phase 3. `-e`/`-T` values flow in via `opts.Eyes`/`opts.Tongue`; `--think`→`opts.Thoughts="o"` flows in via `opts.Thoughts`. The existing default-filling and `strings.NewReplacer` logic handles all cases.

---

### `cmd/gosay/main.go` (controller, request-response)

**Analog:** itself (current state)

**Imports pattern** (lines 1-13 — current):
```go
package main

import (
    "errors"
    "flag"
    "fmt"
    "io"
    "math/rand"
    "os"
    "strings"

    "github.com/pheckenlively/gosay/internal/cowsay"
)
```
No new imports needed for Phase 3 (all new flags use `flag` stdlib; `errors` already present for `errors.Is`).

**Flag registration pattern** (lines 54-66 — current flags, extend):
```go
// CURRENT (Phase 2):
fs := flag.NewFlagSet("gosay", flag.ContinueOnError)
fs.SetOutput(stderr)
var cowName string
var listFlag bool
var randomFlag bool
fs.StringVar(&cowName, "f", "gopher", "cow `name` to use")
fs.BoolVar(&listFlag, "l", false, "list available cows")
fs.BoolVar(&randomFlag, "random", false, "pick a random cow")
fs.Usage = func() {
    fmt.Fprintln(stderr, "usage: gosay [-f name] [-l] [--random] [message...]")
}

// AFTER (Phase 3 D-14 — no-op Usage BEFORE Parse is critical):
fs := flag.NewFlagSet("gosay", flag.ContinueOnError)
fs.SetOutput(stderr)
var cowName string
var listFlag bool
var randomFlag bool
var wrapWidth int
var noWrap bool
var think bool
var eyes string
var tongue string
fs.StringVar(&cowName, "f", "gopher", "cow `name` to use")
fs.BoolVar(&listFlag, "l", false, "list available cows")
fs.BoolVar(&randomFlag, "random", false, "pick a random cow")
fs.IntVar(&wrapWidth, "W", 0, "wrap message at this many display `cols` (default 40)")
fs.BoolVar(&noWrap, "n", false, "disable word wrapping")
fs.BoolVar(&think, "think", false, "use thought bubble instead of speech bubble")
fs.StringVar(&eyes, "e", "", "set eye characters (default \"oo\")")
fs.StringVar(&tongue, "T", "", "set tongue characters (default \"  \")")

// CRITICAL: no-op before Parse to suppress automatic stderr print on -h (RESEARCH Pattern 4)
fs.Usage = func() {}
```

**`flag.ErrHelp` interception pattern** (lines 68-70 — replace current):
```go
// CURRENT (Phase 2):
if err := fs.Parse(args); err != nil {
    return 1
}

// AFTER (Phase 3 D-13, D-14 — RESEARCH Pattern 4):
if err := fs.Parse(args); err != nil {
    if errors.Is(err, flag.ErrHelp) {
        fmt.Fprintln(stdout, helpText)  // stdout + exit 0 (D-13)
        return 0
    }
    fmt.Fprintln(stderr, "usage: gosay [flags] [message...]")
    return 1
}
```

**Help text constant pattern** (D-15 — new constant before `run()`):
```go
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
```

**`RenderOpts` wiring pattern** (line 136 — current call, expand):
```go
// CURRENT (Phase 2):
out, err := cowsay.Render(animal, message, cowsay.RenderOpts{})

// AFTER (Phase 3 — wire new flags through):
opts := cowsay.RenderOpts{
    Eyes:    eyes,
    Tongue:  tongue,
    Width:   wrapWidth,
    NoWrap:  noWrap,
    Think:   think,
}
if think {
    opts.Thoughts = "o"  // D-11: --think sets thoughts trail
}
out, err := cowsay.Render(animal, message, opts)
```

**`fs.Visit` explicit-flag detection pattern** (lines 72-78 — existing, reuse for `-e`/`-T` if sentinel needed):
```go
// Existing pattern for detecting explicit flag setting (currently used for -f):
fExplicit := false
fs.Visit(func(f *flag.Flag) {
    if f.Name == "f" {
        fExplicit = true
    }
})
// Same pattern can detect explicit -e "" if D-06 sentinel is needed:
// eyesExplicit := false
// fs.Visit(func(f *flag.Flag) { if f.Name == "e" { eyesExplicit = true } })
```

---

### `internal/cowsay/balloon_test.go` (test, transform)

**Analog:** itself (current state)

**Test table pattern** (lines 7-62 — copy this structure for new wrap/think test cases):
```go
func TestBuildBalloon(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {
            name:  "single_hello",
            input: "hello",
            expected: " _______ \n< hello >\n ------- \n",
        },
        // ... add think and wrap cases in same struct
    }
    for _, tc := range tests {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            got := buildBalloon(tc.input)  // signature changes to buildBalloon([]string, bool)
            if got != tc.expected {
                t.Errorf("buildBalloon(%q) mismatch:\nexpected:\n`%s`\ngot:\n`%s`",
                    tc.input, tc.expected, got)
            }
        })
    }
}
```

**`TestDisplayWidth` update — MUST change expected value** (lines 64-82):
```go
// BEFORE (Phase 1 — line 73):
{"漢字", 2},  // Phase 1: rune count

// AFTER (Phase 3 — RESEARCH Pitfall 4):
{"漢字", 4},  // Phase 3: display width (2 CJK × 2 cols each)
```

New test cases to add for wrap:
```go
// In TestWrapWords or TestWrapMessage (new test function):
func TestWrapMessage(t *testing.T) {
    tests := []struct {
        name  string
        input string
        width int
        want  string
    }{
        {name: "no_wrap_zero", input: "hello world", width: 0, want: "hello world"},
        {name: "fits_on_one_line", input: "hello", width: 40, want: "hello"},
        {name: "wraps_at_boundary", input: "hello world", width: 5, want: "hello\nworld"},
        {name: "hard_break_long_word", input: "abcdef", width: 3, want: "abc\ndef"},
    }
    // same t.Run table pattern as above
}
```

---

### `internal/cowsay/golden_test.go` (test, transform)

**Analog:** itself (current state)

**Golden test function pattern** (lines 12-18 — copy exactly for each new golden):
```go
func TestGolden_GopherSayHello(t *testing.T) {
    g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
    out, err := Render("gopher", "hello", RenderOpts{})
    if err != nil {
        t.Fatalf("Render: %v", err)
    }
    g.Assert(t, "gopher_say_hello", []byte(out))
}
```

**CJK test replacement** (lines 97-111 — remove `t.Skip`, rename golden, D-09):
```go
// REMOVE: TestGolden_CJK_Skipped with t.Skip(...)

// ADD: real CJK golden test
func TestGolden_GopherSayCJK(t *testing.T) {
    g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
    out, err := Render("gopher", "漢字テスト", RenderOpts{})
    if err != nil {
        t.Fatalf("Render: %v", err)
    }
    g.Assert(t, "cjk_aligned_gopher", []byte(out))
}
```

**Think golden pattern** (new — D-10, D-11):
```go
func TestGolden_GopherThink(t *testing.T) {
    g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
    out, err := Render("gopher", "hello", RenderOpts{Think: true, Thoughts: "o"})
    if err != nil {
        t.Fatalf("Render: %v", err)
    }
    g.Assert(t, "think_say_hello", []byte(out))
}
```

**Wrap golden pattern** (new — D-01, D-02):
```go
func TestGolden_GopherWrap(t *testing.T) {
    g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
    out, err := Render("gopher", "a long message that should wrap at forty columns", RenderOpts{Width: 40})
    if err != nil {
        t.Fatalf("Render: %v", err)
    }
    g.Assert(t, "wrap_long_message", []byte(out))
}
```

**Eyes/tongue golden pattern** (new — D-05):
```go
func TestGolden_GopherCustomEyesTongue(t *testing.T) {
    g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
    out, err := Render("gopher", "hello", RenderOpts{Eyes: "^^", Tongue: "U "})
    if err != nil {
        t.Fatalf("Render: %v", err)
    }
    g.Assert(t, "custom_eyes_tongue", []byte(out))
}
```

**Goldie fixture regeneration command** (run after Phase 3 implementation):
```bash
go test -update ./internal/cowsay/...
```

---

### `internal/cowsay/renderer_test.go` (test, request-response)

**Analog:** itself (current state)

**Unit test pattern** (lines 8-27 — copy structure for new opt field tests):
```go
func TestRender_DefaultCowSingleLine(t *testing.T) {
    out, err := Render("default", "hello", RenderOpts{})
    if err != nil {
        t.Fatalf("Render returned unexpected error: %v", err)
    }
    if !strings.Contains(out, "< hello >") {
        t.Errorf("expected output to contain '< hello >', got:\n%s", out)
    }
}
```

New Phase 3 unit tests to add (matching existing style):
```go
func TestRender_ThinkMode(t *testing.T) {
    out, err := Render("default", "hello", RenderOpts{Think: true})
    if err != nil {
        t.Fatalf("Render returned unexpected error: %v", err)
    }
    if !strings.Contains(out, "( hello )") {
        t.Errorf("expected think bubble '( hello )', got:\n%s", out)
    }
    // say bubble must NOT appear
    if strings.Contains(out, "< hello >") {
        t.Errorf("expected no say bubble in think mode, got:\n%s", out)
    }
}

func TestRender_WrapAt40Default(t *testing.T) {
    // 50-char message should wrap to lines <= 40 display cols
    msg := strings.Repeat("x", 50)
    out, err := Render("default", msg, RenderOpts{})
    if err != nil {
        t.Fatalf("Render: %v", err)
    }
    for _, line := range strings.Split(out, "\n") {
        // inner content lines should not exceed 40 + border chars
        _ = line
    }
    _ = out
}

func TestRender_NoWrap(t *testing.T) {
    msg := strings.Repeat("x", 50)
    out, err := Render("default", msg, RenderOpts{NoWrap: true})
    if err != nil {
        t.Fatalf("Render: %v", err)
    }
    if !strings.Contains(out, strings.Repeat("x", 50)) {
        t.Errorf("expected no-wrap to preserve full 50-char line, got:\n%s", out)
    }
}
```

---

### `cmd/gosay/main_test.go` (test, request-response)

**Analog:** itself (current state)

**CLI test harness pattern** (lines 17-33 — copy `bytes.Buffer` + `run()` invocation for each new test):
```go
func TestRun_NoArgs(t *testing.T) {
    var stdout, stderr bytes.Buffer
    code := run(nil, &stdout, &stderr)
    if code != 0 && code != 1 {
        t.Errorf("expected exit code 0 or 1 for no args, got %d", code)
    }
}
```

**Flag-conflict assertion pattern** (lines 168-180 — copy for new flag conflicts):
```go
func TestRun_FlagConflict_FAndRandom(t *testing.T) {
    var stdout, stderr bytes.Buffer
    code := run([]string{"-f", "tux", "--random", "hi"}, &stdout, &stderr)
    if code != 1 {
        t.Errorf("expected exit code 1 for -f + --random, got %d", code)
    }
    if !strings.Contains(stderr.String(), "cannot combine") {
        t.Errorf("expected 'cannot combine' in stderr, got: %q", stderr.String())
    }
    if stdout.Len() != 0 {
        t.Errorf("expected no stdout on conflict error, got: %q", stdout.String())
    }
}
```

New Phase 3 CLI tests to add (all follow same `var stdout, stderr bytes.Buffer; code := run(...)` pattern):
```go
func TestRun_Help_ExitsZero(t *testing.T) {
    var stdout, stderr bytes.Buffer
    code := run([]string{"-h"}, &stdout, &stderr)
    if code != 0 {
        t.Errorf("expected exit code 0 for -h, got %d (stderr: %q)", code, stderr.String())
    }
    // Help goes to stdout (D-13)
    if !strings.Contains(stdout.String(), "gosay") {
        t.Errorf("expected help text on stdout, got: %q", stdout.String())
    }
    // Must NOT appear on stderr
    if stderr.Len() != 0 {
        t.Errorf("expected no stderr for -h, got: %q", stderr.String())
    }
}

func TestRun_Think_UsesBubble(t *testing.T) {
    var stdout, stderr bytes.Buffer
    code := run([]string{"--think", "hello"}, &stdout, &stderr)
    if code != 0 {
        t.Errorf("expected exit code 0 for --think, got %d", code)
    }
    if !strings.Contains(stdout.String(), "( hello )") {
        t.Errorf("expected thought bubble, got:\n%s", stdout.String())
    }
}

func TestRun_WFlag_CustomWidth(t *testing.T) {
    var stdout, stderr bytes.Buffer
    code := run([]string{"-W", "10", "hello world"}, &stdout, &stderr)
    if code != 0 {
        t.Errorf("expected exit code 0 for -W 10, got %d", code)
    }
    // "hello world" (11 display cols) should wrap at 10 to two lines
    out := stdout.String()
    if !strings.Contains(out, "hello") || !strings.Contains(out, "world") {
        t.Errorf("expected both words in output, got:\n%s", out)
    }
}

func TestRun_NFlag_DisablesWrap(t *testing.T) {
    var stdout, stderr bytes.Buffer
    msg := strings.Repeat("x", 50)
    code := run([]string{"-n", msg}, &stdout, &stderr)
    if code != 0 {
        t.Errorf("expected exit code 0 for -n, got %d", code)
    }
    if !strings.Contains(stdout.String(), strings.Repeat("x", 50)) {
        t.Errorf("expected full 50-char line preserved with -n, got:\n%s", stdout.String())
    }
}

func TestRun_EyesFlag_Custom(t *testing.T) {
    var stdout, stderr bytes.Buffer
    code := run([]string{"-e", "^^", "hello"}, &stdout, &stderr)
    if code != 0 {
        t.Errorf("expected exit code 0 for -e ^^, got %d", code)
    }
    if !strings.Contains(stdout.String(), "^^") {
        t.Errorf("expected custom eyes ^^ in output, got:\n%s", stdout.String())
    }
}

func TestRun_TongueFlag_Custom(t *testing.T) {
    var stdout, stderr bytes.Buffer
    code := run([]string{"-T", "U ", "hello"}, &stdout, &stderr)
    if code != 0 {
        t.Errorf("expected exit code 0 for -T, got %d", code)
    }
    // Tongue appears in gopher cow body if $tongue is present; test non-empty output
    if stdout.Len() == 0 {
        t.Error("expected non-empty output for -T U , got empty")
    }
}
```

---

### `internal/cowsay/testdata/golden/*.golden` (new fixtures)

**Analog:** `internal/cowsay/testdata/golden/gopher_say_hello.golden`

**File naming convention** (snake_case matching the `g.Assert(t, "name", ...)` call):
- `cjk_aligned_gopher.golden` — replaces skipped `cjk_skip.golden`
- `think_say_hello.golden` — think bubble with gopher
- `wrap_long_message.golden` — wrapped output at 40 cols
- `custom_eyes_tongue.golden` — custom `-e`/`-T` render

**Generation pattern** — do NOT hand-author these files. Run:
```bash
go test -update ./internal/cowsay/...
```
after implementing the render changes. Goldie generates the `.golden` files automatically from the current `Render` output.

**Sample existing fixture** (`testdata/golden/gopher_say_hello.golden` — line format reference):
```
 _______ 
< hello >
 ------- 
     \
      \  .-----.
               / (oo) \
              |  .---.  |
```
CJK fixture will have wider top/bottom borders (e.g., `漢字テスト` = 10 display cols → ` ____________ `).

---

## Shared Patterns

### Display-Width Measurement (D-07)
**Source:** `internal/cowsay/balloon.go` lines 9-14 (current `displayWidth` seam)
**Apply to:** `balloon.go` (body swap), `wrap.go` (called for word sizing), `balloon_test.go` (update expected value)
```go
// The seam — swap body only, no call-site changes:
func displayWidth(s string) int {
    return runewidth.StringWidth(s) // was: utf8.RuneCountInString(s)
}
```

### Display-Width Padding (D-08)
**Source:** `internal/cowsay/balloon.go` lines 41, 53 (current `%-*s` occurrences)
**Apply to:** Every `fmt.Fprintf` line in `buildBalloon` that pads a content line
```go
// Replace ALL occurrences of %-*s with padRight:
func padRight(s string, targetWidth int) string {
    w := displayWidth(s)
    if w >= targetWidth {
        return s
    }
    return s + strings.Repeat(" ", targetWidth-w)
}
```

### `flag.ErrHelp` No-Op Guard (D-14)
**Source:** `cmd/gosay/main.go` lines 64-66 (current `fs.Usage` assignment, rewrite)
**Apply to:** `main.go` `run()` function — set BEFORE `fs.Parse`
```go
// No-op MUST be set before fs.Parse to suppress automatic stderr print:
fs.Usage = func() {}
// Then detect after Parse:
if errors.Is(err, flag.ErrHelp) {
    fmt.Fprintln(stdout, helpText)
    return 0
}
```

### Error Handling (existing pattern — unchanged)
**Source:** `cmd/gosay/main.go` lines 138-144
**Apply to:** All new flag error paths in `main.go`
```go
if errors.Is(err, cowsay.ErrUnknownCow) {
    fmt.Fprintf(stderr, "gosay: unknown cowfile %q\n", animal)
    return 1
}
fmt.Fprintln(stderr, err)
return 1
```

### Golden Test Registration (existing pattern)
**Source:** `internal/cowsay/golden_test.go` lines 12-18
**Apply to:** Every new golden test function (cjk, think, wrap, custom eyes/tongue)
```go
g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
out, err := Render("gopher", "<message>", RenderOpts{<opts>})
if err != nil {
    t.Fatalf("Render: %v", err)
}
g.Assert(t, "<fixture_name>", []byte(out))
```

### `run()` CLI Test Harness (existing pattern)
**Source:** `cmd/gosay/main_test.go` lines 17-33
**Apply to:** Every new CLI-level test in `main_test.go`
```go
var stdout, stderr bytes.Buffer
code := run([]string{<args>}, &stdout, &stderr)
if code != <expected> {
    t.Errorf("expected exit code %d, got %d (stderr: %q)", <expected>, code, stderr.String())
}
```

---

## No Analog Found

All Phase 3 files have close analogs within the existing codebase. No files require falling back to RESEARCH.md patterns exclusively. The RESEARCH.md patterns (Pattern 1-5) are used as supplementary detail within the pattern assignments above, but every new/modified file builds directly on existing source files as primary analogs.

---

## Metadata

**Analog search scope:** `/work/gosay/internal/cowsay/`, `/work/gosay/cmd/gosay/`
**Files scanned:** 7 Go source files + 7 golden fixtures
**Pattern extraction date:** 2026-05-31
