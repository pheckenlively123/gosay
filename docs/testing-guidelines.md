# Testing Guidelines

Conventions for testing `gosay`. Stack: stdlib `testing` + `sebdah/goldie/v2` for golden files. No testify, no other test deps. Go 1.22 (`go.mod`), toolchain 1.26.

## Test commands

- Run everything: `go test ./...`
- Regenerate golden fixtures after an intentional render change: `go test -update ./internal/cowsay/` (the `-update` flag is provided by goldie; it rewrites every `.golden` file the package asserts against). Always inspect the diff of regenerated `.golden` files before committing — they are the spec.
- There is no CI config, Makefile, or `.goreleaser.yaml` in the repo yet. Do not reference CI in tests; assume `go test ./...` is the gate.

## Package layout

- `internal/cowsay/` — library: renderer, balloon, wrap, cowfile parser, embed. All golden tests live here.
- `cmd/gosay/` — CLI: `run()` and flag/exit-code behavior only. No golden tests here.
- Tests are `package cowsay` / `package main` (white-box, same package) — they call unexported helpers (`substituteVars`, `buildBalloon`, `parseCowBody`, `wrapMessage`, `hardBreak`, `wrapWords`, `displayWidth`) directly. Keep new tests white-box so they can drive production internals.

## Golden-file testing (the core pattern)

Use goldie for any test asserting **full multi-line rendered ASCII output**. Plain `t.Errorf`/`strings.Contains` is for everything else (substrings, exit codes, errors, counts, sort order, single-line balloon shapes).

Standard form — every golden test repeats this exactly:

```go
g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
out, err := Render("gopher", "hello", RenderOpts{})
if err != nil {
    t.Fatalf("Render: %v", err)
}
g.Assert(t, "gopher_say_hello", []byte(out))
```

Rules:
- Always pass `goldie.WithFixtureDir("testdata/golden")` — fixtures never live in the default `testdata/` root.
- Fixture name (2nd arg to `Assert`) is `snake_case`, no extension; goldie appends `.golden`. Names describe the scenario (`cjk_aligned_gopher`, `custom_eyes_tongue`, `think_say_hello`).
- Convert output to `[]byte(out)`.
- Test functions that use goldie are named `TestGolden_<Scenario>` (e.g. `TestGolden_GopherThink`). Non-golden tests in `golden_test.go` (like `TestListCows_IncludesGopher`) drop the prefix.
- Each golden test gets its own `goldie.New(t, …)` — do not share one `g` across tests.

What is golden-tested: the full `Render()` pipeline per cow/mode — gopher say, multiline, empty message, wrap, CJK alignment, think mode, custom eyes/tongue, and per-cow escape regressions (`default` backslash-unescape, `dragon-and-cow` `\@`, `three-eyes` two-eye quirk). When adding a render feature or a new cow with notable art, add a matching golden.

## Fixtures: golden vs. input

Two distinct kinds under `internal/cowsay/testdata/`:
- `testdata/golden/*.golden` — generated expected outputs (managed by `-update`, never hand-edit).
- `testdata/fixtures/*.cow` — hand-written **input** fixtures NOT in the embedded cow set. `non-eoc.cow` exists to exercise the dynamic heredoc-terminator path (`END` instead of `EOC`); it is read via `os.ReadFile("testdata/fixtures/non-eoc.cow")` and fed to `parseCowBody`. Put synthetic/edge-case `.cow` inputs here, not in `cows/`.

## Drive assertions through production code paths

A repeated, deliberate convention: tests that build expected output must reuse the same helpers `Render` uses, so the test can't drift from production.
- `TestRender_BraceForms` calls `substituteVars(...)` rather than re-implementing substitution.
- `TestGolden_NonEOCSayHello` rebuilds output via `buildBalloon` + `substituteVars` + the same `strings.TrimRight(..., "\n") + "\n"` trailing-newline normalization `Render` performs.

When you need a "expected" value, prefer calling the real helper over hardcoding a parallel implementation.

## Table-driven tests

Used for pure helpers with many small cases: `wrapMessage` (`TestWrapMessage`), `buildBalloon` (`TestBuildBalloon`), `displayWidth` (`TestDisplayWidth`). Convention:
- `tests := []struct{ name string; ... }{...}` with a `name` field, then `t.Run(tc.name, …)`.
- Include the `tc := tc` capture line before `t.Run` (pre-1.22 habit retained in-repo; keep it for consistency).
- Subtest names are `snake_case` (`wraps_at_word_boundary`, `think_two_line`, `cjk_splits_by_display_cols`).
- `t.Parallel()` and `t.Helper()` are NOT used anywhere — do not introduce them without reason.
- Cover CJK/wide-rune and empty-string cases explicitly; width tests use `strings.Repeat("x", N)` to construct exact-length inputs.

## Testing `run()` (CLI seam)

`main.go` exposes `func run(args []string, stdout, stderr io.Writer) int`; `main()` is just `os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))`. Test the CLI through `run`:

```go
var stdout, stderr bytes.Buffer
code := run([]string{"-f", "tux", "hello"}, &stdout, &stderr)
```

Assert on the returned exit `code`, `stdout.String()`, and `stderr.String()`. Conventions enforced by existing tests, follow them:
- Errors/usage go to **stderr** with exit `1`; help (`-h`/`--help`) goes to **stdout** with exit `0`; on error stdout must be empty, and on success stderr must be empty.
- Error messages must not leak internal paths (e.g. assert stderr does NOT contain `"cows/"`).
- `run` reads `os.Stdin` directly (no injection seam). The no-args path is therefore environment-dependent: `TestRun_NoArgs` accepts **either** exit 0 (piped/empty stdin → empty bubble) or exit 1 (TTY → usage). Do not assert a single code for stdin-dependent paths; the TTY behavior is covered at the process level, not in `run` unit tests.
- `run` tests assert substrings (`< hello >`, `( hello )`, `Cow files:`), exit codes, and stream routing — never golden files.
- Flag-combination contracts each get a named test: conflicts (`-f`+`--random` → "cannot combine", `-l`+message/`--random`), precedence (`-n` overrides `-W`), composition (`--random`+`--think` is allowed). Add one when introducing a flag interaction.
- `TestHelpText_MentionsRealDefaults` guards the hand-maintained `helpText` const against drift from real defaults (eyes `"oo"`, width `40`, tongue `"  "`). If you change a default in the renderer, update this test in lockstep.

## Naming & documentation conventions

- Test func names: `Test<Unit>_<Scenario>` (`TestRun_CowFlag_UnknownAnimal`, `TestLoadCow_Nonexistent_SentinelError`, `TestParseCowBody_CRLFTerminator`). Golden tests use the `TestGolden_` prefix.
- Many tests carry a doc comment citing the requirement/decision ID they lock in (e.g. `RENDER-05`, `D-09`, `INPUT-04`, review IDs like `L2`). When a test exists to pin a specific spec decision or known quirk, document that in the comment — especially for intentional deviations from upstream cowsay (see `TestGolden_ThreeEyesSayHello`).
- Error-path tests use `errors.Is(err, ErrUnknownCow)` for sentinel checks plus a substring check that the bad name appears in the message.

## Coverage expectations

Per-layer split: cowfile parsing (heredoc terminators EOC/dynamic/quoted/CRLF, backslash unescape, missing opener/terminator), embed/listing (count ≥51, sorted, no `.cow` suffix, key cows present), wrap (word boundary, hard break, rune-safety, CJK, no-data-loss), balloon (say + think borders, padding, empty), renderer (substitution, custom eyes/tongue, wrap defaults, think mode), golden (full pipeline per scenario), and CLI (`run` flags/exit codes). New functionality should land tests at the lowest layer it touches **plus** a golden if it changes rendered output.
