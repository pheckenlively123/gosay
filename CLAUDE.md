@AGENTS.md

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

## Build & Test Commands

```bash
# Run all tests
go test ./...

# Regenerate golden files after an intentional render change
go test -update ./internal/cowsay/

# Static analysis (expected before release)
go vet ./...
staticcheck ./...
```

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
