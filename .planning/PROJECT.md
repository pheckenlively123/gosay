# gosay

## What This Is

`gosay` is a Go reimplementation of the classic `cowsay` CLI — a small toy that pipes a message through an ASCII-art animal so it appears to be "saying" it. Distributed as a single static binary, it ships with the full upstream cowsay menagerie embedded in the binary and defaults to a gopher instead of a cow as a nod to the language it's written in.

## Core Value

A single, fast, dependency-free Go binary that reproduces the fun of `cowsay` — message in, ASCII animal out — with no Perl, no external `.cow` files, and a gopher on by default.

## Requirements

### Validated

<!-- Shipped and confirmed valuable. -->

(None yet — ship to validate)

### Active

<!-- Current scope. Building toward these. -->

- [ ] CLI accepts a message as argument *and* via stdin (`echo hello | gosay`)
- [ ] Default animal is a gopher (not a cow)
- [ ] `-f <name>` selects a specific animal from the embedded set
- [ ] `-l` lists every embedded animal name
- [ ] `cowthink` behavior available (thought-bubble variant — likely a `-t/--think` flag or sibling command)
- [ ] Full upstream cowsay `.cow` files vendored into the repo and embedded with `//go:embed`
- [ ] Cross-platform binaries (linux / macOS / windows, amd64 + arm64) built and published by a GitHub Actions release workflow

### Out of Scope

<!-- Explicit boundaries. Includes reasoning to prevent re-adding. -->

- Importable library API (`gosay.Say(...)`) — explicitly declined; this is a CLI toy, not a reusable package
- Loading `.cow` files from disk at runtime — defeats the single-binary distribution; all animals are embedded
- Defining a gosay-native cow format — vendoring upstream `.cow` preserves the existing ecosystem and credit
- Animations / ANSI color / TUI — keep the toy small; nothing fancy beyond a gopher default
- Strict drop-in compatibility with Perl `cowsay`'s every flag — match the common ones; don't chase corner-case parity

## Context

- Author already knows Go reasonably well — this project is primarily an exploration of the GSD planning workflow on a small, finishable scope.
- Secondary goal: produce a toy that others might enjoy playing with (and forking).
- Upstream cowsay lives at [github.com/cowsay-org/cowsay](https://github.com/cowsay-org/cowsay) (the maintained successor to Tony Monroe's original); its `.cow` files use a small Perl heredoc-style format with variables like `$eyes`, `$tongue`, `$thoughts` — a parser will be needed to interpret them at render time.
- Go 1.26 is installed locally. `//go:embed` is the canonical way to bundle the vendored `.cow` files into the binary.
- A `LICENSE` already exists at the repo root (carried in before initialization).

## Constraints

- **Tech stack**: Pure Go, standard library where reasonable — keep the dependency tree near-empty so a single static binary remains effortless.
- **Compatibility**: Must accept the upstream `.cow` file format as-is (variables, heredocs, `binmode` quirks) so we can vendor without modification.
- **Distribution**: One self-contained binary per platform — no runtime data files, no install scripts.
- **Release**: All release artifacts must be produced by GitHub Actions (no manual local builds for releases).
- **Default**: The default animal *must* be a gopher; the cow is just one of many.

## Key Decisions

<!-- Decisions that constrain future work. Add throughout project lifecycle. -->

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Embed vendored upstream `.cow` files via `//go:embed` | Preserves the full menagerie and credits original work, while keeping a single-binary distribution | — Pending |
| Default animal is a gopher, not a cow | Signature differentiator — makes gosay feel like a Go-native toy, not just a port | — Pending |
| No library API; CLI only | Keep scope tight; the toy *is* the toy | — Pending |
| Cross-platform binaries via GitHub Actions | Reproducible releases, no local toolchain dependence for distribution | — Pending |
| Module path: `github.com/pheckenlively/gosay` | Standard `go install` path; matches author's GitHub identity | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd:complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-05-17 after initialization*
