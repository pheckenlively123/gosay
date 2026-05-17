# Architecture Research

**Domain:** Go CLI text-processing toy (cowsay clone)
**Researched:** 2026-05-17
**Confidence:** HIGH

## Standard Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     CLI Entry Point                          │
│  cmd/gosay/main.go — flag parse, stdin/arg read, dispatch   │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                  internal/cowsay/                            │
│  ┌────────────┐  ┌─────────────┐  ┌──────────────────────┐  │
│  │  balloon   │  │   cowfile   │  │   renderer           │  │
│  │  (wrap +   │  │   (parse +  │  │   (substitute vars + │  │
│  │   format)  │  │   lookup)   │  │    emit final art)   │  │
│  └─────┬──────┘  └──────┬──────┘  └──────────────────────┘  │
│        │                │                 ▲                  │
│        └────────────────┴─────────────────┘                  │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                  cows/ (embed.FS)                            │
│  gopher.cow  default.cow  dragon.cow  …  (vendored)         │
└─────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Notes |
|-----------|----------------|-------|
| `cmd/gosay/main.go` | Flag parsing, stdin/arg read, mode dispatch (say vs think), error exit | Thin: ≤60 lines. No logic here. |
| `internal/cowsay/balloon.go` | Word-wrap message to width; build speech/thought balloon string | Pure function: `(message, width, think bool) -> string` |
| `internal/cowsay/cowfile.go` | Parse upstream `.cow` heredoc; resolve name→content from embed.FS; list available names | Regex extract heredoc body, expose `ParsedCow` struct |
| `internal/cowsay/renderer.go` | Substitute `$thoughts/$eyes/$tongue` into parsed cow body; concatenate balloon + cow | Pure function: `(balloon, cow, eyes, tongue, thoughts) -> string` |
| `cows/` (embedded) | Static `.cow` file tree vendored from upstream cowsay-org/cowsay | Never modified; `//go:embed cows/*` in `internal/cowsay/embed.go` |

## Recommended Project Structure

```
gosay/
├── go.mod                        # module github.com/pheckenlively/gosay
├── go.sum
├── LICENSE
├── .github/
│   └── workflows/
│       └── release.yml           # goreleaser or matrix build
├── cmd/
│   └── gosay/
│       └── main.go               # package main — flag parse + dispatch only
├── internal/
│   └── cowsay/
│       ├── embed.go              # //go:embed cows/* + var CowFS embed.FS
│       ├── cowfile.go            # ParseCow(), ListCows(), heredoc regex
│       ├── balloon.go            # BuildBalloon()
│       ├── renderer.go           # Render()
│       ├── cowfile_test.go
│       ├── balloon_test.go
│       └── renderer_test.go
├── cows/                         # vendored upstream .cow files (read-only)
│   ├── gopher.cow                # DEFAULT — hand-authored gopher ASCII art
│   ├── default.cow
│   ├── dragon.cow
│   └── …                        # full cowsay-org/cowsay menagerie
└── testdata/
    └── golden/                   # golden output files for end-to-end tests
        ├── gopher_say_hello.golden
        └── …
```

### Structure Rationale

- **`cmd/gosay/`** — Follows the Go standard for a repo that might eventually expose multiple binaries (even if it never does). `go install github.com/pheckenlively/gosay/cmd/gosay@latest` is the canonical install path. Keeps `main.go` trivially thin.
- **`internal/cowsay/`** — Using `internal/` explicitly prevents importers (the project is CLI-only; no library API is a stated constraint). Groups all logic under one package without needing sub-packages — this is appropriate for a ~400-line codebase.
- **`cows/`** at repo root — The `//go:embed` directive requires the path to be relative to the package directory. Placing `cows/` under `internal/cowsay/cows/` works cleanly and keeps it next to the embed declaration. Alternative: root-level `cows/` with embed declared in a `cmd/` file — messier. Prefer `internal/cowsay/cows/` so the declaration and the data live together.
- **`testdata/golden/`** at repo root — Go's test tooling treats `testdata/` as off-limits to the build tool; golden files live there conventionally.

**Correction to layout above:** `cows/` should be nested under `internal/cowsay/` so the embed glob works cleanly from that package:

```
internal/
└── cowsay/
    ├── cows/          ← lives here, not at repo root
    │   ├── gopher.cow
    │   └── …
    └── embed.go       ← //go:embed cows/*
```

## Architectural Patterns

### Pattern 1: Thin `main`, fat `internal`

**What:** `main.go` parses flags, reads stdin/args, calls `cowsay.Render(...)`, prints result, exits. No logic in `main`.

**When to use:** Always for CLI tools — makes the core testable without subprocess invocation.

**Trade-offs:** Slightly more files than a flat single-file program. Worth it because `internal/cowsay` is trivially unit-testable.

**Example:**
```go
// cmd/gosay/main.go
func main() {
    cfg := parseFlags()
    msg := readMessage(cfg)
    out, err := cowsay.Render(cowsay.Options{
        Message: msg,
        Animal:  cfg.animal,
        Width:   cfg.width,
        Think:   cfg.think,
        Eyes:    cfg.eyes,
        Tongue:  cfg.tongue,
    })
    if err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
    fmt.Print(out)
}
```

### Pattern 2: Runtime variable substitution with `strings.Replacer`

**What:** Parse the `.cow` heredoc body at load time (extract the ASCII art between `<<EOC` and `EOC`). At render time, call `strings.NewReplacer("$thoughts", t, "$eyes", e, "$tongue", tng).Replace(body)`.

**When to use:** For gosay's substitution needs — three known variables, no conditionals, no loops. This is the correct tool.

**Trade-offs vs `text/template`:**
- `strings.Replacer` is ~25% faster for this use case (benchmark evidence from Go community)
- Zero dependency on template syntax leaking into `.cow` files
- Upstream `.cow` files use Perl `$var` syntax, not `{{.Var}}` — a Replacer maps Perl names directly, while `text/template` would require converting every vendored file (nmyk.io/cowsay does this conversion; we explicitly do not want to modify vendored files)

**Trade-offs vs build-time codegen:** Build-time codegen (generating a `cows.go` map at `go generate` time) adds tooling complexity and a code-generation step in CI for zero runtime benefit. `embed.FS` + runtime parse is simpler and still gives a single binary.

**Example:**
```go
// internal/cowsay/renderer.go
func Render(balloon, cowBody, eyes, tongue, thoughts string) string {
    r := strings.NewReplacer(
        "$thoughts", thoughts,
        "$eyes",     eyes,
        "$tongue",   tongue,
    )
    return balloon + "\n" + r.Replace(cowBody)
}
```

### Pattern 3: Heredoc extraction via `bufio.Scanner` line scan

**What:** Read the `.cow` file line by line. Skip until a line matches `<<(\w+)` (heredoc open). Accumulate lines until a line is exactly the heredoc terminator. Return accumulated body.

**When to use:** For the `.cow` parser. The format is simple enough that a regex or line scanner beats importing a Perl parser.

**Trade-offs:** Handles 99% of upstream cow files. Edge cases: some advanced cow files define extra Perl variables before the heredoc (`$extra = chop($eyes)` in three-eyes.cow). These manipulations will silently be lost. For an MVP that handles the common menagerie, this is acceptable — document it and note which files may render oddly.

**Example:**
```go
// internal/cowsay/cowfile.go
var heredocOpen = regexp.MustCompile(`<<["']?(\w+)["']?;?`)

func parseCowBody(data []byte) (string, error) {
    scanner := bufio.NewScanner(bytes.NewReader(data))
    var marker string
    var lines []string
    for scanner.Scan() {
        line := scanner.Text()
        if marker == "" {
            if m := heredocOpen.FindStringSubmatch(line); m != nil {
                marker = m[1]
            }
            continue
        }
        if strings.TrimRight(line, "\r") == marker {
            return strings.Join(lines, "\n"), nil
        }
        lines = append(lines, line)
    }
    return "", fmt.Errorf("heredoc terminator %q not found", marker)
}
```

### Pattern 4: `embed.FS` for the cow directory

**What:** Single `//go:embed cows/*` on a package-level `embed.FS` variable.

**When to use:** Always — this is the canonical Go 1.16+ idiom for embedding a directory of data files into a binary. HIGH confidence.

**Trade-offs vs `map[string]string` codegen:** `embed.FS` requires zero build-time tooling, is understood by `go build` natively, and supports `fs.ReadDir` for listing animals. A generated map would be ~4,000 lines of generated Go with no benefit. Do not do this.

```go
// internal/cowsay/embed.go
package cowsay

import "embed"

//go:embed cows/*
var cowFS embed.FS
```

Listing and loading:
```go
func ListCows() ([]string, error) {
    entries, err := cowFS.ReadDir("cows")
    // strip .cow suffix, sort, return
}

func loadCow(name string) ([]byte, error) {
    return cowFS.ReadFile("cows/" + name + ".cow")
}
```

## Data Flow

### End-to-End Pipeline

```
User invokes: gosay "Hello, gopher!"
     │
     ▼
cmd/gosay/main.go
  parseFlags()  →  animal="gopher", width=40, think=false, eyes="oo", tongue="  "
  readMessage() →  "Hello, gopher!" (from arg or stdin)
     │
     ▼
internal/cowsay/cowfile.go
  LoadCow("gopher")
    cowFS.ReadFile("cows/gopher.cow")   ← embed.FS (in-binary)
    parseCowBody(data)                  ← heredoc regex → raw ASCII body
    → ParsedCow{Body: "...ascii art with $thoughts/$eyes/$tongue placeholders..."}
     │
     ▼
internal/cowsay/balloon.go
  BuildBalloon("Hello, gopher!", width=40, think=false)
    wordwrap(message, width)            ← pure string transform
    buildBorder(lines)                  ← top/bottom border + side chars
    → " _______________\n< Hello, gopher! >\n ---------------"
     │
     ▼
internal/cowsay/renderer.go
  Render(balloon, cow.Body, eyes="oo", tongue="  ", thoughts="\\")
    strings.NewReplacer(...)Replace(cow.Body)
    balloon + "\n" + substituted_art
    → final string
     │
     ▼
cmd/gosay/main.go
  fmt.Print(result)  →  stdout
```

### `cowthink` variant

The `think` boolean flows through the same pipeline. `BuildBalloon` receives `think=true` and uses `( )` instead of `< >` border characters and `o` instead of `\` for the thoughts trail. The cow body is identical — only balloon changes. No separate code path needed.

### `-l` list mode

```
main.go -l flag detected
  →  cowsay.ListCows() → []string{"beavis.zen","bud-frogs","default",...}
  →  sort + print one per line
  →  exit 0
```

## Gopher-as-Default: Where It Lives

The gopher default is a **single constant** in `cmd/gosay/main.go`:

```go
const defaultAnimal = "gopher"
```

`flag.StringVar(&cfg.animal, "f", defaultAnimal, "cowfile name")` uses it. No special-casing elsewhere — the pipeline is animal-agnostic. The `gopher.cow` file itself lives in `internal/cowsay/cows/gopher.cow` as a hand-authored ASCII gopher, vendored alongside the upstream cow files.

## Build Order Implications

The pipeline has clear dependency layers. Build in this order for a working end-to-end skeleton as early as possible:

| Phase | What to Build | Why This Order |
|-------|--------------|----------------|
| 1 | Repo scaffold + `go.mod` + CI skeleton | Unblocks everything |
| 2 | `internal/cowsay/embed.go` + `cows/` vendoring + `ListCows()` | Establishes the data foundation; tests can immediately check cow count |
| 3 | `cowfile.go` — heredoc parser + `LoadCow()` | Required before render; unit-testable in isolation with sample `.cow` bytes |
| 4 | `balloon.go` — word-wrap + balloon builder | Independent of cow parsing; pure string logic; easiest to unit test |
| 5 | `renderer.go` — variable substitution + combine | Assembles phases 3+4; produces first end-to-end output |
| 6 | `cmd/gosay/main.go` — flags, stdin, dispatch, error handling | Wires everything into the binary |
| 7 | Golden tests + `-l` flag + `-t/--think` flag | Polish after skeleton works |
| 8 | GitHub Actions release workflow | Last because it needs a working binary |

Phases 3 and 4 are independent and can be built in parallel. Phase 5 requires both.

## Testing Surface

| Layer | Test Type | Rationale |
|-------|-----------|-----------|
| `parseCowBody()` | Unit tests with inline `.cow` byte literals | Verify heredoc extraction on normal + edge-case files |
| `BuildBalloon()` | Unit tests, table-driven | Pure function; test wrap at boundary widths, single-line, multi-line, think vs say |
| `Render()` | Unit tests | 3 replacements; trivial but worth testing $thoughts multi-char |
| `ListCows()` | Unit test checking count + presence of "gopher" and "default" | Catches embed misconfiguration early |
| Full pipeline (say hello with gopher) | Golden test in `testdata/golden/` | Catches regressions to rendered output without brittle string literals |
| CLI flags | Subprocess test or integration test via `os/exec` | Test `-f`, `-l`, `-W`, stdin reading |

**Golden test pattern:** store expected `.golden` files in `testdata/golden/`. Support `-update` flag in tests to regenerate. Use standard library `os.ReadFile` + `bytes.Equal` — no external golden-file library needed for a project this size.

## Anti-Patterns

### Anti-Pattern 1: Logic in `main.go`

**What people do:** Put word-wrap, cow parsing, or output formatting directly in `main.go` because it's quick.

**Why it's wrong:** `main.go` can't be tested without subprocess invocation. The logic becomes untestable in isolation.

**Do this instead:** Keep `main.go` to ≤60 lines of flag parsing and wiring. All transformations live in `internal/cowsay/`.

### Anti-Pattern 2: Converting vendored `.cow` files to Go templates

**What people do:** Transform Perl `$var` syntax to `{{.Var}}` and use `text/template`.

**Why it's wrong:** Requires modifying every vendored file, making future upstream syncs painful. Also loses compatibility with advanced cow files that do Perl string manipulation.

**Do this instead:** Parse the heredoc body as raw text. Use `strings.Replacer` at render time to substitute `$eyes`, `$tongue`, `$thoughts`. Accept that Perl-manipulating cow files (a small minority) may render imperfectly.

### Anti-Pattern 3: Embedding `cows/` with a generated `map[string]string`

**What people do:** Run `go generate` to produce a Go file with `var cows = map[string]string{"default": "..."}`.

**Why it's wrong:** Adds a code-generation step to the build, inflates the generated Go source, makes diffs unreadable, and provides zero benefit over `embed.FS`.

**Do this instead:** `//go:embed cows/*` on an `embed.FS` variable. `go build` handles it natively.

### Anti-Pattern 4: Over-engineering the `ParsedCow` type

**What people do:** Make `ParsedCow` an interface, add factory methods, builder pattern, etc.

**Why it's wrong:** This is a ~400-line codebase. Interfaces and builders are engineering overhead with no payoff.

**Do this instead:** A plain struct with exported fields is sufficient:
```go
type ParsedCow struct {
    Name string
    Body string // raw ASCII art with $-placeholders intact
}
```

## Sources

- Go module layout — official docs: https://go.dev/doc/modules/layout
- No-nonsense Go package layout (2024): https://laurentsv.com/blog/2024/10/19/no-nonsense-go-package-layout.html
- Neo-cowsay (reference implementation): https://github.com/Code-Hex/Neo-cowsay
- nmyk.io/cowsay API (text/template approach, for contrast): https://pkg.go.dev/nmyk.io/cowsay
- Go embed package docs: https://pkg.go.dev/embed
- Cowsay cow file format (Context7 / cowsay-org): https://github.com/cowsay-org/cowsay
- Golden file testing in Go: https://ieftimov.com/posts/testing-in-go-golden-files/

---
*Architecture research for: gosay (Go cowsay clone)*
*Researched: 2026-05-17*
