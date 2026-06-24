# gosay

> A zero-dependency Go reimplementation of the classic `cowsay` — with a gopher on by default.

`gosay` pipes a message through an ASCII-art animal so it appears to be "saying" it. You give it text (as arguments or on stdin), and it prints a speech (or thought) balloon attached to the animal of your choice.

It's a faithful take on the original `cowsay`, with a few deliberate differences:

- **A gopher by default**, not a cow — a nod to the language it's written in. The cow is just one of many animals.
- **A single static binary.** No Perl, no runtime data files, no install scripts. The entire upstream cowsay menagerie (50 animals) plus a hand-authored gopher is embedded directly in the binary via `//go:embed`.
- **Near-zero dependencies.** Pure Go, standard-library-first.

## Quick start

Install with the Go toolchain:

```bash
go install github.com/pheckenlively/gosay/cmd/gosay@latest
```

This drops a `gosay` binary in `$(go env GOPATH)/bin` (make sure that's on your `PATH`). Then:

```bash
gosay hello
```

You can also pipe input in:

```bash
echo "moo" | gosay -f default
```

## Usage

```
Usage: gosay [flags] [message...]
```

If you pass a message as arguments, those win. Otherwise `gosay` reads from stdin (when piped). Running it interactively with no message prints usage.

| Flag | Description |
| --- | --- |
| `-e <eyes>` | Set the eye characters (default `"oo"`). |
| `-f <name>` | Select an animal from the embedded set (default `gopher`). |
| `-l` | List all available animals. |
| `-n` | Disable word wrapping; preserve all input whitespace. Overrides `-W`. |
| `-T <tongue>` | Set the tongue characters (default two spaces). |
| `-W <cols>` | Wrap the message at this many display columns (default `40`). |
| `--random` | Pick a random animal. |
| `--think` | Use a thought bubble `( )` instead of a speech bubble `< >`. |
| `-h`, `--help` | Show help. |

A few examples:

```bash
gosay --think -e "^^" "thinking..."     # thought bubble, custom eyes
echo hi | gosay -f tux                  # Tux the penguin says hi
gosay -l                                # list every animal
gosay --random "surprise!"              # roll the dice
```

Notes:
- `-f` and `--random` cannot be combined.
- `-l` cannot be combined with a message or an animal selection.
- An explicit non-positive `-W` (e.g. `-W 0`) is a usage error — unless `-n` is also set.

## Tech stack & key design decisions

- **Pure Go, stdlib-first.** The only runtime dependency is [`mattn/go-runewidth`](https://github.com/mattn/go-runewidth) for correct display-width math. Test-only: [`sebdah/goldie`](https://github.com/sebdah/goldie) for golden-file tests.
- **Two packages, clean seam.** `cmd/gosay` is a thin CLI layer; `internal/cowsay` is a pure rendering library with no CLI concerns.
- **Embedded, inert cow files.** `.cow` files are vendored from upstream cowsay and embedded in the binary. The Perl preamble is never executed, and the embedded filesystem acts as a strict allowlist.
- **Defaults resolved downstream.** Eyes/tongue/width defaults live in the renderer, not in `main`, with the zero value used as the "not set" sentinel.

For the full architecture, rendering pipeline ordering, and embedded-asset provenance rules, see **[AGENTS.md](AGENTS.md)**.

## Project structure

```
gosay/
├── cmd/gosay/
│   └── main.go            # CLI layer: flags, stdin/arg resolution, exit codes
├── internal/cowsay/       # Pure rendering library
│   ├── embed.go           # embedded FS + animal listing
│   ├── cowfile.go         # .cow heredoc parsing
│   ├── renderer.go        # Render entry point + variable substitution
│   ├── balloon.go         # speech / thought bubble layout
│   ├── wrap.go            # word wrap + hard break
│   └── cows/              # embedded .cow art (gopher + 50 vendored animals)
├── docs/                  # contributor guidelines (see below)
├── AGENTS.md              # architecture & conventions for contributors / AI agents
└── go.mod
```

## Building & testing

```bash
# Build the binary
go build ./cmd/gosay

# Run it without installing
go run ./cmd/gosay hello

# Run all tests
go test ./...

# Regenerate golden files after an intentional render change, then review the diff
go test -update ./internal/cowsay/

# Static analysis (expected before a release)
go vet ./...
staticcheck ./...
```

## For contributors and AI agents

Before changing code in a given area, read the relevant guideline:

- **[AGENTS.md](AGENTS.md)** — overall architecture, package layout, rendering pipeline, code style, and common pitfalls.
- **[docs/security-guidelines.md](docs/security-guidelines.md)** — embedded-FS trust boundary, unescape/substitution ordering, input validation.
- **[docs/error-handling-guidelines.md](docs/error-handling-guidelines.md)** — `main`/`run` split, exit codes, sentinel errors, user-facing messages.
- **[docs/testing-guidelines.md](docs/testing-guidelines.md)** — golden-file pattern, fixture layout, and table-driven conventions.

Cow-file provenance and the re-vendoring procedure are in [`internal/cowsay/cows/SOURCE.md`](internal/cowsay/cows/SOURCE.md); per-file licensing is in `internal/cowsay/cows/NOTICE`.

## License

See [LICENSE](LICENSE). Vendored `.cow` art retains its upstream licensing — see `internal/cowsay/cows/NOTICE`.
