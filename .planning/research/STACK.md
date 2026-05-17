# Stack Research

**Domain:** Go single-binary CLI tool (cowsay clone)
**Researched:** 2026-05-17
**Confidence:** HIGH (all major choices verified against current docs/packages)

---

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.22 (go.mod minimum) | Language | Go 1.26 is installed locally; set `go 1.22` in go.mod to maximise installer reach while keeping `//go:embed`, generics, and all needed stdlib. Go 1.22+ is actively supported (two-release support window covers 1.25 and 1.26 as of May 2026). |
| `//go:embed` + `embed.FS` | stdlib (Go 1.16+) | Bundle .cow files into binary | Zero-dependency, compile-time embedding. `embed.FS` gives `fs.ReadDir`, `fs.ReadFile`, and glob listing — everything gosay needs. No third-party asset bundler required. |
| `flag` (stdlib) | stdlib | CLI flag parsing | gosay has ~5 flags, one positional argument, and no subcommands. `flag` is sufficient, zero-dep, and idiomatic for single-command tools. Adding Cobra or Kong for this scale is over-engineering. |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `sebdah/goldie` | v2.8.0 (Oct 2025) | Golden-file testing for ASCII output | Use in all render tests. `g.Assert(t, "name", []byte(output))` pattern; regenerate with `go test -update ./...`. Saves copy-pasting multiline ASCII strings into test files. |
| `goreleaser/goreleaser` | v2 (current) | Cross-platform binary release pipeline | Use via `goreleaser/goreleaser-action@v6` in GitHub Actions. Produces mac/linux/windows × amd64/arm64 release archives and GitHub Release assets from a ~20-line `.goreleaser.yaml`. |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| `go test ./...` + stdlib `testing` | Unit and integration tests | No testify needed — `t.Errorf` + goldie covers all cases cleanly. Testify adds ~3 transitive deps for no meaningful gain on a toy CLI. |
| `go vet` + `staticcheck` | Static analysis | Run in CI; catches common mistakes. `staticcheck` is the de-facto Go linter, installable as `go install honnef.co/go/tools/cmd/staticcheck@latest`. |
| `goreleaser init` | Bootstrap `.goreleaser.yaml` | Run once locally; check the generated file into the repo. |

---

## Installation

```bash
# No external runtime deps — this is the whole point.
# Dev / test tooling only:
go get github.com/sebdah/goldie/v2

# GoReleaser (local dry-run only; CI uses the Action)
go install github.com/goreleaser/goreleaser/v2@latest
```

---

## Topic-by-Topic Decisions

### 1. CLI Flag Parsing — Use `flag` (stdlib)

**Winner: `flag` stdlib**

gosay's surface area: `-f <name>`, `-l`, `-t/--think`, `-W <cols>` (wrap width), message as positional arg or stdin. That is exactly the use case `flag` was designed for.

**Why not Cobra:** Cobra is the right choice when you need subcommands, persistent flags, shell completion generation, or Viper integration. gosay has none of those. Cobra also pulls in `spf13/pflag` as a mandatory dependency. Reputation is LOW on Context7 (flagged for reduced source quality).

**Why not Kong:** Kong (v1.15.0, Apr 2026) uses struct-tag-driven parsing, which is elegant but is genuine over-engineering for 5 flags. It adds a dependency and requires importing a non-stdlib module. Kong shines for multi-command CLIs with positional argument structs; gosay is flat.

**Why not urfave/cli:** Similar story to Cobra — command-centric framework, adds deps, not worth it for a toy with a flat command structure.

**Implementation pattern for gosay:**
```go
var (
    cowFile  = flag.String("f", "gopher", "Select a specific animal (use -l to list)")
    listCows = flag.Bool("l", false, "List all available animals")
    think    = flag.Bool("t", false, "Cowthink mode (thought bubbles)")
    wrapAt   = flag.Int("W", 40, "Wrap message at this many columns (0 = no wrap)")
)
flag.Parse()
```

**Confidence: HIGH** — verified against stdlib docs and community consensus.

---

### 2. Embedding .cow Files — `//go:embed` with `embed.FS`

**Winner: `embed.FS` with a `cows/` subdirectory**

Recommended layout:
```
internal/
  cowfiles/
    embed.go          // var FS embed.FS
    cows/
      gopher.cow      // custom default
      cow.cow
      tux.cow
      ... (full upstream set)
```

`embed.go`:
```go
package cowfiles

import "embed"

//go:embed cows/*.cow
var FS embed.FS
```

Key operations:
- **List all animals:** `fs.ReadDir(FS, "cows")` — returns `[]fs.DirEntry`, strip `.cow` suffix for display.
- **Look up by name:** `FS.ReadFile("cows/" + name + ".cow")` — returns `[]byte`.
- **Custom gopher default:** Just name the file `gopher.cow` and default `-f` to `"gopher"`.

**Why `embed.FS` over `[]byte` or `string`:** `embed.FS` is the right type for a directory of files. It implements `fs.FS` so you get `ReadDir`, `ReadFile`, and can use `fs.Sub` if needed. Single `[]byte` embedding is only appropriate for embedding one known file at compile time.

**Confirmed patterns:** The `cows/*.cow` glob pattern embeds all .cow files. By convention, embed directives go in a dedicated package so the large binary blob doesn't pollute your main package's compile units.

**Confidence: HIGH** — verified against official embed package docs (pkg.go.dev/embed).

---

### 3. .cow File Parsing — Hand-rolled string replacer (Neo-cowsay pattern)

**Winner: Hand-rolled parser modeled on Code-Hex/Neo-cowsay**

The upstream `.cow` format is a Perl heredoc:
```perl
$the_cow = <<"EOC";
     $thoughts   ^__^
     $thoughts   ($eyes)\\_______
...
EOC
```

The parse algorithm (confirmed from Neo-cowsay source):
1. Scan lines, find the `$the_cow = <<"EOC"` (or `<<EOC`) opener.
2. Collect lines until a bare `EOC` terminator.
3. Strip lines starting with `##` (Perl comments embedded in some cow files).
4. Apply variable substitution with `strings.NewReplacer`:
   - `\\\\` → `\\`
   - `\\@` → `@`
   - `\\$` → `$`
   - `$eyes` / `${eyes}` → eye chars (default `oo`)
   - `$tongue` / `${tongue}` → tongue chars (default `  `)
   - `$thoughts` / `${thoughts}` → bubble connector char (`\` for say, `o` for think)

This is the complete parsing problem. It is ~50 lines of Go. **Do not import a Perl interpreter.**

**Why not import Neo-cowsay directly:** Neo-cowsay v2.0.4 (Feb 2022) is four years old, requires its own embedded cow set (competing with gosay's), and is structured as a library API — the opposite of gosay's CLI-only requirement. Use it as a reference implementation only.

**Why not regex:** The heredoc terminator is always a bare `EOC` on its own line. A simple line-by-line state machine is clearer and faster than a regex.

**Confidence: HIGH** — parsing logic confirmed by reading Neo-cowsay's `cowsay.go` `strings.NewReplacer` block directly.

---

### 4. Text Wrapping / Balloon Rendering — Hand-roll it (~30 lines)

**Winner: stdlib `strings.Fields` + manual column counter**

`mitchellh/go-wordwrap` (v1.0.1, last published Sep 2020) is archived — Mitchell Hashimoto has stepped back from Go and the repo has no activity since 2020. It also has a known edge case: words longer than the wrap limit are not split, they overflow. For cowsay's use case (short terminal messages) this is usually fine, but taking an archived dependency for 30 lines of logic is not justified.

**Hand-rolled approach:**
```go
func wrapText(s string, limit int) string {
    if limit <= 0 {
        return s
    }
    words := strings.Fields(s)
    var b strings.Builder
    lineLen := 0
    for i, w := range words {
        if lineLen > 0 && lineLen+1+len(w) > limit {
            b.WriteByte('\n')
            lineLen = 0
        } else if i > 0 {
            b.WriteByte(' ')
            lineLen++
        }
        b.WriteString(w)
        lineLen += len(w)
    }
    return b.String()
}
```

The balloon renderer then finds the longest wrapped line, draws a box of that width, and left-pads each line. This is the core of the original cowsay algorithm and takes ~60 lines in Go.

**Why not muesli/reflow:** Handles ANSI escape sequences — a feature gosay explicitly does not need (out of scope: no ANSI color). Adds a dependency for zero benefit here.

**Confidence: HIGH** — mitchellh archive confirmed by search; hand-rolling confirmed as the right call for a ~30-line problem.

---

### 5. Testing — stdlib `testing` + `sebdah/goldie` v2

**Winner: stdlib `testing` for assertions, goldie for snapshot/golden-file comparison**

For a CLI that renders ASCII art, the primary test need is: "given this input, does the full rendered output match exactly?" That is the golden file pattern.

**goldie v2.8.0** (Oct 2025, actively maintained):
```go
func TestSay(t *testing.T) {
    out := render("hello world", "gopher", false, 40)
    g := goldie.New(t)
    g.Assert(t, "gopher-hello-world", []byte(out))
}
// First run: go test -update ./...  (creates testdata/gopher-hello-world.golden)
// Subsequent runs: go test ./...    (compares against stored golden)
```

Golden files live in `testdata/` (goldie's default), are committed to the repo, and make diffs in CI immediately visible.

**Why not testify:** `require.Equal` and `assert.Equal` work fine for simple string comparison, but produce noisy diffs for multi-line ASCII output. goldie uses `ClassicDiff` or `ColoredDiff` which shows line-level differences, exactly what you want. Also, testify adds `stretchr/objx`, `davecgh/go-spew`, `pmezard/go-difflib` — three extra transitive dependencies the project doesn't need.

**Why not stdlib only:** `t.Errorf("got:\n%s\nwant:\n%s", got, want)` with a hard-coded `want` string in source works, but updating 50 `.cow` renderings by hand after a balloon formatting change is painful. goldie's `-update` flag makes mass-regeneration a one-liner.

**Confidence: HIGH** — goldie v2.8.0 confirmed on pkg.go.dev; test pattern is well-established Go idiom.

---

### 6. Release Pipeline — GoReleaser v2 + GitHub Actions

**Winner: GoReleaser v2 via `goreleaser/goreleaser-action@v6`**

Minimal `.goreleaser.yaml` for gosay:
```yaml
version: 2

builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}}

archives:
  - format: tar.gz
    format_overrides:
      - goos: windows
        format: zip

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
```

GitHub Actions workflow (`.github/workflows/release.yml`):
```yaml
on:
  push:
    tags:
      - "v*"

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: stable
      - uses: goreleaser/goreleaser-action@v6
        with:
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

This produces: `gosay_linux_amd64.tar.gz`, `gosay_linux_arm64.tar.gz`, `gosay_darwin_amd64.tar.gz`, `gosay_darwin_arm64.tar.gz`, `gosay_windows_amd64.zip`, `gosay_windows_arm64.zip` — all attached to the GitHub Release automatically.

**Why not hand-rolled matrix:** A matrix strategy spins up 6 separate runner VMs (one per OS×arch combination) costing ~6× the compute time. Go cross-compilation runs on a single linux/amd64 runner in under 30 seconds. GoReleaser knows this and exploits it. The config is also ~20 lines vs ~130 lines for equivalent matrix YAML.

**Why not `wangyoucao577/go-release-action`:** Less maintained, no changelog generation, no archive format control. GoReleaser is the canonical choice used by the Go ecosystem (kubectl, Hugo, Cobra, etc.).

**Confidence: HIGH** — GoReleaser v2 config confirmed via Context7 docs; action tag confirmed as goreleaser-action@v6.

---

### 7. Go Version Target — `go 1.22` in go.mod

**Winner: `go 1.22` as minimum in go.mod**

Current Go versions as of May 2026:
- **go1.26.3** (2026-05-07) — installed locally, latest stable
- **go1.25.10** (2026-05-07) — still supported
- go1.24 — end of support (three-release window: 1.24 is now superseded by 1.25 and 1.26)

**Recommendation:** Set `go 1.22` in go.mod. Rationale:
- `//go:embed` arrived in Go 1.16 — that's the hard floor.
- Go 1.22 introduced loop variable scoping fix (commonly cited regression fix, makes the module feel current without being bleeding edge).
- Anyone with Go 1.22+ installed can build and `go install` gosay without a toolchain upgrade.
- Setting `go 1.26` would be honest about local dev, but would reject `go install` from users who haven't upgraded yet (Go's forward-compat rule: toolchain older than the go directive refuses to build).
- Use a `toolchain go1.26.3` line to tell `go mod tidy` to use the latest patch but keep the minimum at 1.22.

```
// go.mod
go 1.22

toolchain go1.26.3
```

**Confidence: HIGH** — Go release timeline confirmed at go.dev/doc/devel/release; toolchain directive behavior confirmed from go.dev/blog/toolchain.

---

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| `flag` stdlib | `alecthomas/kong` v1.15.0 | If gosay grows subcommands (e.g. `gosay say` / `gosay think` as separate subcommands rather than a flag) |
| `flag` stdlib | `spf13/cobra` | If gosay ever needs shell completion generation or man-page output |
| Hand-rolled word wrap | `mitchellh/go-wordwrap` v1.0.1 | If you need it fast and accept an archived dep — it works fine for short strings |
| `sebdah/goldie` v2 | Plain `t.Errorf` with string literals | For very small test suites (< 5 tests) where golden file machinery feels heavy |
| GoReleaser v2 | Hand-rolled matrix GHA | If you need Windows ARM64 MSI installers or Homebrew tap automation specifically — then GoReleaser Pro adds more value |

---

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| `spf13/cobra` | Over-engineered for 5 flat flags; adds pflag dep; no subcommands needed | `flag` stdlib |
| `urfave/cli` v2/v3 | Command-centric framework; adds deps; not worth it for a single-command toy | `flag` stdlib |
| `mitchellh/go-wordwrap` | Archived since 2020; Mitchell Hashimoto stepped away from Go; would be the project's only dependency for code you can write in 30 lines | Hand-rolled `strings.Fields` loop |
| `github.com/Code-Hex/Neo-cowsay/v2` | Last updated Feb 2022; structured as a library (contradicts gosay's CLI-only scope); brings its own embedded cow set; use as reference only | Hand-rolled parser |
| `stretchr/testify` | Adds 3 transitive deps for `assert.Equal`; goldie covers the multi-line ASCII diff case better | `sebdah/goldie` v2 + stdlib `testing` |
| Hand-rolled GitHub Actions matrix | 6 VMs × 6 combinations = slow; 130+ lines of YAML; no changelog | GoReleaser v2 |

---

## Version Compatibility

| Package | Go Minimum | Notes |
|---------|-----------|-------|
| `sebdah/goldie` v2.8.0 | Go 1.18+ (uses generics internally) | Compatible with go.mod `go 1.22` target |
| `goreleaser/goreleaser-action` v6 | N/A (GitHub Action, not a Go dep) | Wraps GoReleaser v2; only in CI |
| `alecthomas/kong` v1.15.0 | Go 1.21+ | Only if subcommands ever added |
| `embed.FS` (stdlib) | Go 1.16 | gosay's hard floor |

---

## Sources

- `pkg.go.dev/embed` — embed.FS API, glob patterns, `fs.Sub` usage (HIGH confidence)
- `pkg.go.dev/github.com/alecthomas/kong` v1.15.0 — current version confirmed Apr 2026 (HIGH)
- `pkg.go.dev/github.com/sebdah/goldie/v2` v2.8.0 — current version confirmed Oct 2025 (HIGH)
- `pkg.go.dev/github.com/mitchellh/go-wordwrap` v1.0.1 — archived, last published Sep 2020 (HIGH)
- `github.com/Code-Hex/Neo-cowsay` — `cowsay.go` `strings.NewReplacer` block reviewed directly (HIGH)
- `go.dev/doc/devel/release` — Go 1.26.3 latest stable as of May 2026 (HIGH)
- `go.dev/blog/toolchain` — go vs toolchain directive semantics (HIGH)
- Context7 `/goreleaser/goreleaser` — GoReleaser v2 GitHub Actions config (HIGH)
- Context7 `/alecthomas/kong` — kong struct tag API and `kong.Parse` example (HIGH)
- `goreleaser.com/customization/builds/builders/go/` — goos/goarch minimal config (HIGH)

---

*Stack research for: Go single-binary CLI (gosay / cowsay clone)*
*Researched: 2026-05-17*
