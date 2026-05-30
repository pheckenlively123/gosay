# Walking Skeleton — gosay

**Phase:** 1
**Generated:** 2026-05-21

## Capability Proven End-to-End

A user can run `go run ./cmd/gosay "hello"` (or `gosay "hello"` after `go build`) and see "hello" wrapped in a `< hello >` speech bubble next to a gopher rendered from an embedded `.cow` file — every layer from `os.Args` through the heredoc parser, balloon renderer, variable substitution, and `embed.FS` is exercised on every invocation.

The walking-skeleton subset (Plan 01-01) proves the smallest version of that pipeline: `go build` succeeds, the binary runs, and stdout contains a deterministic string. Subsequent plans replace the hardcoded stub with the real renderer without altering the architectural decisions below.

## Architectural Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Language / version | Go; `go 1.22` in `go.mod`, `toolchain go1.26.3` | Maximises `go install` reach (1.22 is the broadest still-supported floor) while pinning local builds to the latest patch. `//go:embed` arrived in Go 1.16 — well below the floor. |
| Module path | `github.com/pheckenlively/gosay` | Locked in PROJECT.md / CLAUDE.md; matches the upstream GitHub home. |
| CLI framework | None — `os.Args[1:]` only in Phase 1 | Phase 1 has one positional arg and no flags; `flag.Parse()` enters in Phase 2 when `-f`/`-l` arrive. |
| Package layout | `cmd/gosay/main.go` (thin) + `internal/cowsay/{embed,cowfile,balloon,renderer}.go` + `internal/cowsay/cows/` | Per D-01..D-19 in CONTEXT.md and ARCHITECTURE.md. `internal/` prevents library import (project is CLI-only). Cow files live next to `embed.go` so the `//go:embed cows/*.cow` glob resolves cleanly. |
| Exported API surface (internal pkg) | `Render(animal, message string, opts RenderOpts) (string, error)` and `ListCows() ([]string, error)` | Minimal seam — `ListCows` is unused in Phase 1 main.go but is exported because Phase 2 needs it without refactor. `RenderOpts` carries `Eyes`/`Tongue`/`Thoughts` so Phase 3 can pass non-defaults. |
| Cow file vendoring | Snapshot copy from `cowsay-org/cowsay` v3.8.4 (commit `027c9268ac8571408e153214b9cf1a5e6fab0cfc`); LF-normalized; recorded in `internal/cowsay/cows/SOURCE.md` | Per D-05..D-09. One-time copy avoids submodule complexity; `SOURCE.md` documents refresh procedure. |
| Cow embedding | `//go:embed cows/*.cow` on a single `embed.FS` in `internal/cowsay/embed.go` | Canonical Go 1.16+ idiom. Explicit `*.cow` glob excludes `NOTICE`, `SOURCE.md`, and any stray hidden files. |
| Line-ending discipline | `.gitattributes` at repo root with `cows/*.cow text eol=lf` | Prevents CRLF corruption on Windows checkouts (PITFALLS.md #8). Lands in Plan 01-01 before any cow files are committed. |
| Heredoc parser | `bufio.Scanner` + `regexp.MustCompile(`<<["']?(\w+)["']?;?`)` for dynamic terminator capture; `strings.NewReplacer` for backslash unescape; `strings.NewReplacer` for variable substitution; **unescape BEFORE variable substitution** | Per CONTEXT D-locked. Matches Perl semantics (RESEARCH §C.10) and avoids the three classic parser pitfalls. |
| Variable substitution | Single `strings.NewReplacer` covering both bare (`$eyes`) and brace (`${eyes}`) forms for `$thoughts`/`$eyes`/`$tongue` | `strings.NewReplacer` applies left-to-right without re-scanning replacements (RESEARCH §C.9), so order does not matter for the default values. Brace forms protect against third-party cows down the line. |
| Display-width abstraction | Unexported `displayWidth(s string) int` in `internal/cowsay/balloon.go`; Phase 1 body is `utf8.RuneCountInString(s)` | Per D-16..D-19. Single seam swaps to `runewidth.StringWidth` in Phase 3 with zero call-site changes. A `t.Skip(...)` CJK golden test documents the gap. |
| Golden testing | `github.com/sebdah/goldie/v2` v2.8.0 + hand-curated `.golden` fixtures in `internal/cowsay/testdata/golden/` | Per D-11..D-15. No Perl-cowsay dependency in CI; the user reviews each golden once during Phase 1 execution. |
| Default animal | `gopher` — hand-authored `internal/cowsay/cows/gopher.cow` | Project identity (PROJECT.md). User picks from 2-3 ASCII variants drafted during Plan 01-04 execution (D-04). |
| Build / test commands | `go build ./...`, `go vet ./...`, `go test ./...`, `go test -update ./internal/cowsay` (regenerate goldens) | Standard Go toolchain. No `make`, no shell scripts, no external task runner. |
| Deployment target | None in Phase 1 — local `go build` / `go install`. Release pipeline (GoReleaser + GitHub Actions) lands in Phase 4. | Walking-skeleton scope is "binary builds and runs locally"; distribution is a later vertical slice. |

## Stack Touched in Phase 1

- [x] Project scaffold — `go.mod`, `go.sum`, repo-root `.gitattributes`, `cmd/gosay/main.go`
- [x] Data layer — `internal/cowsay/cows/*.cow` (51 vendored + `gopher.cow`), `internal/cowsay/cows/{NOTICE,SOURCE.md}`, `internal/cowsay/embed.go` with `embed.FS`
- [x] Engine — `internal/cowsay/cowfile.go` (heredoc parser), `internal/cowsay/balloon.go` (single-line `< >` + multi-line `/ | \`), `internal/cowsay/renderer.go` (variable substitution + concatenation)
- [x] CLI entry — `cmd/gosay/main.go` reads `os.Args[1:]`, joins with spaces, dispatches to `cowsay.Render("gopher", message, cowsay.RenderOpts{})`, prints to stdout, exits non-zero on error
- [x] Test runner — stdlib `testing` + `sebdah/goldie/v2`; goldens in `internal/cowsay/testdata/golden/`
- [x] Local full-stack run — `go run ./cmd/gosay "hello"` produces the gopher saying hello on stdout; `go test ./...` exits 0

## Out of Scope (Deferred to Later Slices)

Phase 1 produces a runnable binary, **not** the full feature set. The following are explicitly out of scope and must not be added by Phase 1 plans:

- **stdin reading** (`echo hello | gosay`) — Phase 2 / INPUT-02
- **`-f <name>` cow selection flag** — Phase 2 / COW-02
- **`-l` list animals flag** — Phase 2 / COW-03
- **`--random` flag** — Phase 2 / COW-04
- **Unknown-cow error handling** — Phase 2 / COW-05
- **Empty-input handling** (empty bubble) — Phase 2 / INPUT-04
- **Any use of the `flag` stdlib package** — Phase 2 (main.go uses `os.Args[1:]` only)
- **Word wrap (`-W <n>` / `-n`)** — Phase 3 / RENDER-05
- **Thought-bubble mode (`--think`)** — Phase 3 / RENDER-07
- **Custom eyes / tongue flags (`-e`, `-T`)** — Phase 3 / RENDER-08
- **`runewidth.StringWidth` (replace `displayWidth` body)** — Phase 3 / RENDER-06; Phase 1 leaves a `t.Skip` CJK golden test as the documented gap
- **`-h` / `--help`** — Phase 3 / HELP-01
- **GoReleaser config, GitHub Actions release workflow, `--version`, `go install` validation** — Phase 4 / DIST-01..05

## Subsequent Slice Plan

Each later phase adds one vertical slice on top of this skeleton without altering its architectural decisions:

- **Phase 2: Input and Cow Selection** — wire `flag.Parse()` into `main.go`; add stdin reading; expose `-f`/`-l`/`--random`; expose `ListCows`; add empty-input handling and unknown-cow error path
- **Phase 3: Full Flag Surface** — add `-W`/`-n` word wrap to `balloon.go`; add `--think` (swap `$thoughts`, balloon corners); add `-e`/`-T`; pull in `mattn/go-runewidth`; swap `displayWidth` body to `runewidth.StringWidth`; remove CJK `t.Skip`; add `-h`/`--help`
- **Phase 4: Release Pipeline** — add `.goreleaser.yaml`; add `.github/workflows/release.yml`; add `--version` flag with `-ldflags` stamping; validate `go install github.com/pheckenlively/gosay/cmd/gosay@latest`
