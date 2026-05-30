# Phase 2: Input and Cow Selection - Pattern Map

**Mapped:** 2026-05-30
**Files analyzed:** 6 (4 modified, 2 new test artifacts)
**Analogs found:** 6 / 6

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `cmd/gosay/main.go` | CLI entrypoint | request-response | `cmd/gosay/main.go` (current) | exact — same file, extending |
| `cmd/gosay/main_test.go` | CLI test | request-response | `cmd/gosay/main_test.go` (current) | exact — same file, extending |
| `internal/cowsay/cowfile.go` | package/model | CRUD | `internal/cowsay/cowfile.go` (current) | exact — same file, small addition |
| `internal/cowsay/embed.go` | utility | batch/list | `internal/cowsay/embed.go` (current) | read-only reference |
| `internal/cowsay/renderer.go` | service | request-response | `internal/cowsay/renderer.go` (current) | read-only reference |
| `internal/cowsay/*_test.go` + `testdata/golden/` | test + golden fixture | batch | `internal/cowsay/golden_test.go` (current) | exact role match |

---

## Pattern Assignments

### `cmd/gosay/main.go` (CLI entrypoint, request-response)

**Analog:** `cmd/gosay/main.go` (current Phase 1 version — extend, do not replace architecture)

**Current imports block** (lines 1–10):
```go
package main

import (
    "fmt"
    "io"
    "os"
    "strings"

    "github.com/pheckenlively/gosay/internal/cowsay"
)
```

Phase 2 adds to the import block: `"errors"`, `"flag"`, `"io"`, `"math/rand"`, `"unicode/utf8"` (or just use `"io"` already present). The full new import set will be roughly:
```go
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

**Testable harness pattern** (lines 15–28 in current `main.go`):
```go
// run assembles the message from args, renders it, and writes the result.
// It returns the process exit code so the logic is testable without os.Exit.
// args is os.Args[1:] (the program arguments, excluding the binary name).
func run(args []string, stdout, stderr io.Writer) int {
    if len(args) < 1 {
        fmt.Fprintln(stderr, "usage: gosay <message>")
        return 1
    }
    message := strings.Join(args, " ")
    out, err := cowsay.Render("gopher", message, cowsay.RenderOpts{})
    if err != nil {
        fmt.Fprintln(stderr, err)
        return 1
    }
    fmt.Fprint(stdout, out)
    return 0
}
```

**Key constraint:** The `run(args []string, stdout, stderr io.Writer) int` signature MUST NOT change — `main_test.go` calls it directly. Phase 2 replaces the body of `run` with `flag`-parsing logic while keeping this exact signature.

**`main()` pattern** (lines 30–32):
```go
func main() {
    os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
```
This two-liner MUST remain unchanged.

**Phase 2 body structure for `run` (new logic to implement):**

```
1. Create a flag.FlagSet (not flag.CommandLine) so tests can call run() multiple times cleanly.
2. Define flags: cowName (string, default "gopher"), listCows (bool), randomCow (bool).
3. Override FlagSet.Usage to print: "usage: gosay [-f name] [-l] [--random] [message...]" to stderr, return 1.
4. Parse flags from args using fs.Parse(args).
5. Conflict checks:
   a. -f + --random together → fmt.Fprintln(stderr, "gosay: cannot combine -f and --random") → return 1
   b. -l with any message args or -f/--random → fmt.Fprintln(stderr, "gosay: -l cannot be combined with message or animal selection") → return 1
6. If -l: call cowsay.ListCows(), format columnar, write to stdout, return 0.
7. Determine animal: if --random, pick rand.Intn(len(names)) from cowsay.ListCows() result.
8. Resolve message:
   a. If positional args remain (fs.Args()), join with " ".
   b. Else if stdin is NOT a TTY (os.ModeCharDevice check): io.ReadAll(os.Stdin), trim exactly one trailing "\n".
   c. Else (interactive terminal, no args): print usage, return 1.
9. Call cowsay.Render(animal, message, cowsay.RenderOpts{}).
10. Error check: errors.Is(err, cowsay.ErrUnknownCow) → fmt.Fprintf(stderr, "gosay: unknown cowfile %q\n", animal) → return 1.
    Any other error → fmt.Fprintln(stderr, err) → return 1.
11. fmt.Fprint(stdout, out), return 0.
```

**TTY detection pattern (D-02):**
```go
// isTTY reports whether f is connected to a character device (interactive terminal).
func isTTY(f *os.File) bool {
    fi, err := f.Stat()
    if err != nil {
        return false
    }
    return fi.Mode()&os.ModeCharDevice != 0
}
```

**flag.FlagSet pattern (enables multiple run() calls in tests):**
```go
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
if err := fs.Parse(args); err != nil {
    return 1
}
```

---

### `cmd/gosay/main_test.go` (CLI test, request-response)

**Analog:** `cmd/gosay/main_test.go` (current Phase 1 version — extend, do not remove existing tests)

**Test file header pattern** (lines 1–7):
```go
package main

import (
    "bytes"
    "strings"
    "testing"
)
```

**Existing test structure pattern** (lines 9–21):
```go
func TestRun_NoArgs(t *testing.T) {
    var stdout, stderr bytes.Buffer
    code := run(nil, &stdout, &stderr)
    if code != 1 {
        t.Errorf("expected exit code 1 for no args, got %d", code)
    }
    if stdout.Len() != 0 {
        t.Errorf("expected no stdout output, got: %q", stdout.String())
    }
    if !strings.Contains(stderr.String(), "usage:") {
        t.Errorf("expected usage message on stderr, got: %q", stderr.String())
    }
}
```

**Phase 2 new tests to add (copy this pattern for each):**

```go
// Pattern: always declare var stdout, stderr bytes.Buffer; call run(); check code + output.

func TestRun_FlagConflict_FAndRandom(t *testing.T) {
    var stdout, stderr bytes.Buffer
    code := run([]string{"-f", "tux", "--random"}, &stdout, &stderr)
    if code != 1 {
        t.Errorf("expected exit code 1 for -f + --random, got %d", code)
    }
    if !strings.Contains(stderr.String(), "cannot combine") {
        t.Errorf("expected conflict message on stderr, got: %q", stderr.String())
    }
    if stdout.Len() != 0 {
        t.Errorf("expected no stdout on conflict, got: %q", stdout.String())
    }
}

func TestRun_ListFlag_AloneExits0(t *testing.T) {
    var stdout, stderr bytes.Buffer
    code := run([]string{"-l"}, &stdout, &stderr)
    if code != 0 {
        t.Errorf("expected exit code 0 for -l alone, got %d (stderr: %q)", code, stderr.String())
    }
    if !strings.Contains(stdout.String(), "Cow files:") {
        t.Errorf("expected 'Cow files:' header in -l output, got:\n%s", stdout.String())
    }
}

func TestRun_ListFlag_WithMessage_IsError(t *testing.T) {
    var stdout, stderr bytes.Buffer
    code := run([]string{"-l", "hello"}, &stdout, &stderr)
    if code != 1 {
        t.Errorf("expected exit code 1 for -l + message, got %d", code)
    }
}

func TestRun_CowFlag_KnownAnimal(t *testing.T) {
    var stdout, stderr bytes.Buffer
    code := run([]string{"-f", "tux", "hello"}, &stdout, &stderr)
    if code != 0 {
        t.Errorf("expected exit code 0 for -f tux, got %d (stderr: %q)", code, stderr.String())
    }
}

func TestRun_CowFlag_UnknownAnimal(t *testing.T) {
    var stdout, stderr bytes.Buffer
    code := run([]string{"-f", "nosuchcow", "hello"}, &stdout, &stderr)
    if code != 1 {
        t.Errorf("expected exit code 1 for unknown cow, got %d", code)
    }
    if !strings.Contains(stderr.String(), `unknown cowfile "nosuchcow"`) {
        t.Errorf("expected clean unknown-cow message on stderr, got: %q", stderr.String())
    }
    // Must NOT leak internal path like "cows/nosuchcow.cow"
    if strings.Contains(stderr.String(), "cows/") {
        t.Errorf("stderr must not leak internal cow path, got: %q", stderr.String())
    }
}

func TestRun_Random_ReturnsMember(t *testing.T) {
    // --random must pick an animal that exists in ListCows(); not which one.
    var stdout, stderr bytes.Buffer
    code := run([]string{"--random", "hello"}, &stdout, &stderr)
    if code != 0 {
        t.Errorf("expected exit code 0 for --random, got %d (stderr: %q)", code, stderr.String())
    }
    if stdout.Len() == 0 {
        t.Error("expected non-empty stdout for --random")
    }
}

func TestRun_EmptyMessage_Renders(t *testing.T) {
    // gosay "" must produce a valid empty bubble (exit 0, non-empty output).
    var stdout, stderr bytes.Buffer
    code := run([]string{""}, &stdout, &stderr)
    if code != 0 {
        t.Errorf("expected exit code 0 for empty-string arg, got %d (stderr: %q)", code, stderr.String())
    }
    if stdout.Len() == 0 {
        t.Error("expected non-empty stdout for empty message arg")
    }
}
```

**Note on stdin testing:** Tests that exercise the stdin path (piped input) require injecting a reader into `run`. The current `run` signature uses `os.Stdin` directly for stdin. To keep the signature stable, the TTY check and stdin read can be factored as an unexported helper that tests can skip by always providing positional args. If a stdin injection seam is needed, add it as an internal helper rather than changing the public `run` signature.

---

### `internal/cowsay/cowfile.go` (model, CRUD)

**Analog:** `internal/cowsay/cowfile.go` (current — small addition only)

**Current imports block** (lines 1–10):
```go
package cowsay

import (
    "bufio"
    "bytes"
    "errors"
    "fmt"
    "regexp"
    "strings"
)
```

Phase 2 adds `"io/fs"` to the import block:
```go
import (
    "bufio"
    "bytes"
    "errors"
    "fmt"
    "io/fs"
    "regexp"
    "strings"
)
```

**Sentinel error to add (new, near top of file after package declaration):**
```go
// ErrUnknownCow is returned by LoadCow when the named cow does not exist in
// the embedded set. Callers use errors.Is(err, cowsay.ErrUnknownCow) to
// distinguish a missing cow from other (I/O or parse) errors.
var ErrUnknownCow = errors.New("unknown cowfile")
```

**Current `LoadCow` function** (lines 86–96):
```go
func LoadCow(name string) (ParsedCow, error) {
    data, err := readCowFile(name)
    if err != nil {
        return ParsedCow{}, fmt.Errorf("load cow %q: %w", name, err)
    }
    body, err := parseCowBody(data)
    if err != nil {
        return ParsedCow{}, fmt.Errorf("load cow %q: %w", name, err)
    }
    return ParsedCow{Name: name, Body: body}, nil
}
```

**Phase 2 modified `LoadCow` — wrap `fs.ErrNotExist` with `ErrUnknownCow`:**
```go
func LoadCow(name string) (ParsedCow, error) {
    data, err := readCowFile(name)
    if err != nil {
        if errors.Is(err, fs.ErrNotExist) {
            return ParsedCow{}, fmt.Errorf("%w: %s", ErrUnknownCow, name)
        }
        return ParsedCow{}, fmt.Errorf("load cow %q: %w", name, err)
    }
    body, err := parseCowBody(data)
    if err != nil {
        return ParsedCow{}, fmt.Errorf("load cow %q: %w", name, err)
    }
    return ParsedCow{Name: name, Body: body}, nil
}
```

**Error chain:** `cowsay.ErrUnknownCow` is the sentinel; the cow name is appended to the error message. `renderer.go`'s `Render` wraps this with `fmt.Errorf("render: %w", err)`, so `errors.Is(err, cowsay.ErrUnknownCow)` still returns `true` through that wrapping chain. No changes to `renderer.go` are required.

**Existing test to update in `cowfile_test.go`** (lines 160–168, `TestLoadCow_Nonexistent`):
```go
func TestLoadCow_Nonexistent(t *testing.T) {
    _, err := LoadCow("does-not-exist")
    if err == nil {
        t.Fatal("expected error for non-existent cow, got nil")
    }
    if !strings.Contains(err.Error(), "does-not-exist") {
        t.Errorf("expected error to mention 'does-not-exist', got: %v", err)
    }
}
```
Phase 2 extends this test (or adds a sibling) to also assert `errors.Is(err, cowsay.ErrUnknownCow)`:
```go
func TestLoadCow_Nonexistent_SentinelError(t *testing.T) {
    _, err := LoadCow("does-not-exist")
    if err == nil {
        t.Fatal("expected error for non-existent cow, got nil")
    }
    if !errors.Is(err, ErrUnknownCow) {
        t.Errorf("expected errors.Is(err, ErrUnknownCow) = true, got err: %v", err)
    }
}
```

---

### `internal/cowsay/embed.go` (read-only reference)

**No changes in Phase 2.** Used as-is by `main.go` for `-l` and `--random`.

**`ListCows()` signature** (lines 15–30 of `embed.go`):
```go
func ListCows() ([]string, error) {
    entries, err := cowFS.ReadDir("cows")
    if err != nil {
        return nil, fmt.Errorf("listing embedded cows: %w", err)
    }
    names := make([]string, 0, len(entries))
    for _, e := range entries {
        n := e.Name()
        if !strings.HasSuffix(n, ".cow") {
            continue
        }
        names = append(names, strings.TrimSuffix(n, ".cow"))
    }
    sort.Strings(names)
    return names, nil
}
```

**Key properties for callers:**
- Returns sorted, `.cow`-stripped names — use directly as the `--random` pool and the `-l` source.
- Returns `([]string, error)` — `main.go` must handle the error case (if embed FS is broken, exit non-zero with the raw error; this path is practically unreachable in production but must not panic).

**Usage pattern in `main.go` for `--random`:**
```go
names, err := cowsay.ListCows()
if err != nil {
    fmt.Fprintln(stderr, err)
    return 1
}
animal = names[rand.Intn(len(names))]
```

**Usage pattern in `main.go` for `-l` (columnar output):**
```go
names, err := cowsay.ListCows()
if err != nil {
    fmt.Fprintln(stderr, err)
    return 1
}
// Write columnar listing to stdout — see Shared Patterns: -l Columnar Formatting.
```

---

### `internal/cowsay/renderer.go` (read-only reference)

**No changes in Phase 2.** The `Render` call site and error wrapping already chain `ErrUnknownCow` transparently via `%w`.

**`Render` signature** (line 64 of `renderer.go`):
```go
func Render(animal, message string, opts RenderOpts) (string, error)
```

**Error wrapping chain** (lines 64–68):
```go
func Render(animal, message string, opts RenderOpts) (string, error) {
    cow, err := LoadCow(animal)
    if err != nil {
        return "", fmt.Errorf("render: %w", err)
    }
    // ...
```

Because `Render` uses `%w`, an `ErrUnknownCow` from `LoadCow` is preserved through the chain. `main.go` checks `errors.Is(err, cowsay.ErrUnknownCow)` after calling `Render` — no need to call `LoadCow` directly.

---

### `internal/cowsay/*_test.go` + `testdata/golden/` (golden tests)

**Analog:** `internal/cowsay/golden_test.go` (current)

**goldie setup pattern** (lines 12–19 of `golden_test.go`):
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

**Import pattern for golden tests** (lines 1–9 of `golden_test.go`):
```go
package cowsay

import (
    "os"
    "strings"
    "testing"

    goldie "github.com/sebdah/goldie/v2"
)
```

**Phase 2 golden tests to add:**

```go
// TestGolden_GopherSayEmpty verifies D-03: empty message renders a valid empty bubble.
// buildBalloon("") already produces " __ \n<  >\n -- \n" (confirmed by TestBuildBalloon).
// This golden captures the full output (balloon + gopher body).
func TestGolden_GopherSayEmpty(t *testing.T) {
    g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
    out, err := Render("gopher", "", RenderOpts{})
    if err != nil {
        t.Fatalf("Render: %v", err)
    }
    g.Assert(t, "gopher_say_empty", []byte(out))
}
```

**Note:** The `-l` columnar output is produced by formatting logic in `main.go`, not by `internal/cowsay`. Its golden test therefore lives in `cmd/gosay/main_test.go` (assert `stdout.String()` contains expected column structure) rather than as a `goldie` golden file. Alternatively, if a `FormatCowList(names []string) string` helper is extracted to `internal/cowsay`, its golden test follows the same `goldie.New(t, goldie.WithFixtureDir("testdata/golden"))` setup pattern.

**Updating golden files:** Run `go test -update ./...` to regenerate `.golden` files after implementation. The `-update` flag is from goldie. Review each golden manually before committing.

---

## Shared Patterns

### Error wrapping via `%w`
**Source:** All existing `internal/cowsay/*.go` files
**Apply to:** `cowfile.go` `LoadCow` modification, any new helpers
```go
// Wrap errors with %w so callers can use errors.Is / errors.As.
// Sentinel errors (ErrUnknownCow) use fmt.Errorf("%w: <detail>", sentinel).
// Internal errors use fmt.Errorf("context: %w", err).
return ParsedCow{}, fmt.Errorf("%w: %s", ErrUnknownCow, name)
return ParsedCow{}, fmt.Errorf("load cow %q: %w", name, err)
```

### Stderr/exit-code discipline
**Source:** `cmd/gosay/main.go` (lines 16–27)
**Apply to:** All new error paths in `run()`
```go
// All errors print to stderr (never stdout); all error paths return non-zero.
// Success output goes to stdout; function returns 0.
fmt.Fprintln(stderr, "gosay: <message>")
return 1
```

### Test harness: bytes.Buffer + run()
**Source:** `cmd/gosay/main_test.go` (lines 9–21)
**Apply to:** All new CLI-level test functions
```go
var stdout, stderr bytes.Buffer
code := run([]string{...args...}, &stdout, &stderr)
// Assert code, stdout.String(), stderr.String()
```

### goldie fixture convention
**Source:** `internal/cowsay/golden_test.go` (lines 12–19)
**Apply to:** Any new package-level render golden tests
```go
g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
g.Assert(t, "snake_case_test_name", []byte(out))
// Golden files live at: internal/cowsay/testdata/golden/<name>.golden
// Regenerate: go test -update ./internal/cowsay/...
```

### -l Columnar Formatting (Claude's discretion per D-07)
**Source:** No existing analog — new behavior. Follow upstream cowsay's general shape.
**Apply to:** The `-l` output path in `run()`

Upstream cowsay produces output like:
```
Cow files:
actually  alpaca  beavis.zen  blowfish  bong  bud-frogs  bunny  cheese
cower     cupcake daemon      default   dragon-and-cow  dragon ...
```

Implementation guidance:
- Print `"Cow files:\n"` as the header to `stdout`.
- Join names with a single space separator, wrapped at approximately 72–80 characters per line (exact width is Claude's discretion per D-07).
- A simple approach: accumulate names on the current line; if adding the next name would exceed the wrap width, emit a newline first.
- The output must be deterministic (same order every run) — `ListCows()` already returns sorted names.
- Write to `stdout io.Writer`, not `os.Stdout` directly, so tests can capture it.

---

## No Analog Found

All Phase 2 files have close analogs in the existing codebase. No files require falling back to research patterns.

The TTY detection helper (`isTTY` using `os.Stdin.Stat()` / `os.ModeCharDevice`) has no existing analog — it is new code. The pattern is standard Go:
```go
fi, err := os.Stdin.Stat()
if err != nil { /* treat as non-TTY */ }
isTerminal := fi.Mode()&os.ModeCharDevice != 0
```
This is called only when no positional args are present (D-01/D-02).

---

## Metadata

**Analog search scope:** `/work/gosay/cmd/gosay/`, `/work/gosay/internal/cowsay/`
**Files scanned:** 11 Go source files + 1 go.mod
**Pattern extraction date:** 2026-05-30
