<!-- GSD:project-start source:PROJECT.md -->

## Project

**gosay**

`gosay` is a Go reimplementation of the classic `cowsay` CLI — a small toy that pipes a message through an ASCII-art animal so it appears to be "saying" it. Distributed as a single static binary, it ships with the full upstream cowsay menagerie embedded in the binary and defaults to a gopher instead of a cow as a nod to the language it's written in.

**Core Value:** A single, fast, dependency-free Go binary that reproduces the fun of `cowsay` — message in, ASCII animal out — with no Perl, no external `.cow` files, and a gopher on by default.

### Constraints

- **Tech stack**: Pure Go, standard library where reasonable — keep the dependency tree near-empty so a single static binary remains effortless.
- **Compatibility**: Must accept the upstream `.cow` file format as-is (variables, heredocs, `binmode` quirks) so we can vendor without modification.
- **Distribution**: One self-contained binary per platform — no runtime data files, no install scripts.
- **Release**: All release artifacts must be produced by GitHub Actions (no manual local builds for releases).
- **Default**: The default animal *must* be a gopher; the cow is just one of many.

<!-- GSD:project-end -->

<!-- GSD:stack-start source:research/STACK.md -->

## Technology Stack

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

## Installation

# No external runtime deps — this is the whole point.

# Dev / test tooling only:

# GoReleaser (local dry-run only; CI uses the Action)

## Topic-by-Topic Decisions

### 1. CLI Flag Parsing — Use `flag` (stdlib)

### 2. Embedding .cow Files — `//go:embed` with `embed.FS`

- **List all animals:** `fs.ReadDir(FS, "cows")` — returns `[]fs.DirEntry`, strip `.cow` suffix for display.
- **Look up by name:** `FS.ReadFile("cows/" + name + ".cow")` — returns `[]byte`.
- **Custom gopher default:** Just name the file `gopher.cow` and default `-f` to `"gopher"`.

### 3. .cow File Parsing — Hand-rolled string replacer (Neo-cowsay pattern)

### 4. Text Wrapping / Balloon Rendering — Hand-roll it (~30 lines)

### 5. Testing — stdlib `testing` + `sebdah/goldie` v2

### 6. Release Pipeline — GoReleaser v2 + GitHub Actions

### 7. Go Version Target — `go 1.22` in go.mod

- **go1.26.3** (2026-05-07) — installed locally, latest stable
- **go1.25.10** (2026-05-07) — still supported
- go1.24 — end of support (three-release window: 1.24 is now superseded by 1.25 and 1.26)
- `//go:embed` arrived in Go 1.16 — that's the hard floor.
- Go 1.22 introduced loop variable scoping fix (commonly cited regression fix, makes the module feel current without being bleeding edge).
- Anyone with Go 1.22+ installed can build and `go install` gosay without a toolchain upgrade.
- Setting `go 1.26` would be honest about local dev, but would reject `go install` from users who haven't upgraded yet (Go's forward-compat rule: toolchain older than the go directive refuses to build).
- Use a `toolchain go1.26.3` line to tell `go mod tidy` to use the latest patch but keep the minimum at 1.22.

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| `flag` stdlib | `alecthomas/kong` v1.15.0 | If gosay grows subcommands (e.g. `gosay say` / `gosay think` as separate subcommands rather than a flag) |
| `flag` stdlib | `spf13/cobra` | If gosay ever needs shell completion generation or man-page output |
| Hand-rolled word wrap | `mitchellh/go-wordwrap` v1.0.1 | If you need it fast and accept an archived dep — it works fine for short strings |
| `sebdah/goldie` v2 | Plain `t.Errorf` with string literals | For very small test suites (< 5 tests) where golden file machinery feels heavy |
| GoReleaser v2 | Hand-rolled matrix GHA | If you need Windows ARM64 MSI installers or Homebrew tap automation specifically — then GoReleaser Pro adds more value |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| `spf13/cobra` | Over-engineered for 5 flat flags; adds pflag dep; no subcommands needed | `flag` stdlib |
| `urfave/cli` v2/v3 | Command-centric framework; adds deps; not worth it for a single-command toy | `flag` stdlib |
| `mitchellh/go-wordwrap` | Archived since 2020; Mitchell Hashimoto stepped away from Go; would be the project's only dependency for code you can write in 30 lines | Hand-rolled `strings.Fields` loop |
| `github.com/Code-Hex/Neo-cowsay/v2` | Last updated Feb 2022; structured as a library (contradicts gosay's CLI-only scope); brings its own embedded cow set; use as reference only | Hand-rolled parser |
| `stretchr/testify` | Adds 3 transitive deps for `assert.Equal`; goldie covers the multi-line ASCII diff case better | `sebdah/goldie` v2 + stdlib `testing` |
| Hand-rolled GitHub Actions matrix | 6 VMs × 6 combinations = slow; 130+ lines of YAML; no changelog | GoReleaser v2 |

## Version Compatibility

| Package | Go Minimum | Notes |
|---------|-----------|-------|
| `sebdah/goldie` v2.8.0 | Go 1.18+ (uses generics internally) | Compatible with go.mod `go 1.22` target |
| `goreleaser/goreleaser-action` v6 | N/A (GitHub Action, not a Go dep) | Wraps GoReleaser v2; only in CI |
| `alecthomas/kong` v1.15.0 | Go 1.21+ | Only if subcommands ever added |
| `embed.FS` (stdlib) | Go 1.16 | gosay's hard floor |

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

<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->

## Conventions

Conventions not yet established. Will populate as patterns emerge during development.
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->

## Architecture

Architecture not yet mapped. Follow existing patterns found in the codebase.
<!-- GSD:architecture-end -->

<!-- GSD:skills-start source:skills/ -->

## Project Skills

No project skills found. Add skills to any of: `.claude/skills/`, `.agents/skills/`, `.cursor/skills/`, `.github/skills/`, or `.codex/skills/` with a `SKILL.md` index file.
<!-- GSD:skills-end -->

<!-- GSD:workflow-start source:GSD defaults -->

## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:

- `/gsd:quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd:debug` for investigation and bug fixing
- `/gsd:execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->

<!-- GSD:profile-start -->

## Developer Profile

> Profile not yet configured. Run `/gsd:profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
