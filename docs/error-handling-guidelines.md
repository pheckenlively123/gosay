# Error Handling Guidelines

These rules capture the error-handling conventions actually used in `gosay`. Follow them when adding or modifying code. They are derived from `cmd/gosay/main.go`, `internal/cowsay/*.go`, and the corresponding tests.

## Architecture: the `main`/`run` split

- `main()` does exactly one thing: `os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))`. Never put logic, printing, or error handling in `main()`.
- `run(args []string, stdout, stderr io.Writer) int` owns all CLI logic and **returns an exit code** instead of calling `os.Exit`. This keeps it unit-testable (see `main_test.go`, which drives `run` directly with `bytes.Buffer`s).
- `os.Exit` appears in exactly one place — inside `main()`. Do not call `os.Exit`, `log.Fatal`, or `panic` anywhere else.
- All output goes through the injected `stdout`/`stderr` writers, **not** `os.Stdout`/`os.Stderr` or `fmt.Print*` to the default streams. (Exception: `run` reads `os.Stdin` directly — there is no injection seam for stdin, and tests acknowledge this gap explicitly.)

## Exit code conventions

- `0` = success, including non-render successes: `-h`/`--help`, `-l` listing, and an empty-but-valid message that renders an empty bubble.
- `1` = any error: usage/flag errors, flag conflicts, unknown cow, stdin read failure, render failure.
- There are only two exit codes. Do not introduce additional codes without a strong reason.
- On any exit-1 path, **nothing is written to stdout** — error text goes to stderr only. Tests assert `stdout.Len() == 0` on error paths; preserve this.

## Sentinel errors

- The one sentinel is `cowsay.ErrUnknownCow` (`errors.New("unknown cowfile")`), defined in `cowfile.go`.
- Use a sentinel only when a caller needs to **discriminate** a specific failure to produce a distinct user-facing message. `ErrUnknownCow` exists so `main` can print `gosay: unknown cowfile "<name>"` instead of leaking the embed path.
- Create the sentinel at the boundary where the condition is first known. In `LoadCow`, a `fs.ErrNotExist` from the embedded FS is translated into `ErrUnknownCow`:
  ```go
  if errors.Is(err, fs.ErrNotExist) {
      return ParsedCow{}, fmt.Errorf("%w: %s", ErrUnknownCow, name)
  }
  ```
  Note the sentinel is wrapped with `%w` (so `errors.Is` keeps working) and the name is appended as plain context, not via a second `%w`.

## Wrapping and propagation

- Wrap with `fmt.Errorf("<context>: %w", err)` when adding context while preserving the chain. Context strings are lowercase, no trailing punctuation, and describe the operation: `"listing embedded cows: %w"`, `"load cow %q: %w"`, `"render: %w"`, `"scanner error reading cow file: %w"`.
- Each layer adds one short context prefix. `readCowFile` → `LoadCow` (`load cow %q`) → `Render` (`render`) builds a readable chain without restating lower layers.
- Use `%w` for **at most one** wrapped error per `Errorf`. When you want a sentinel plus extra data, wrap the sentinel with `%w` and add the data with a non-`%w` verb (`%s`/`%q`), as in the `ErrUnknownCow` example above.
- Create fresh errors with `errors.New` (no wrapping) only for leaf conditions that have no underlying cause — e.g. `errors.New("no heredoc opener found in cow file")`. Use `fmt.Errorf` (no `%w`) when you need to interpolate but there is no underlying error to chain: `fmt.Errorf("heredoc terminator %q not found in cow file", marker)`.

## Discrimination: `errors.Is`, never type assertions

- Always discriminate errors with `errors.Is(err, Target)`. There are **no** type assertions or `errors.As` anywhere in this codebase; do not add them unless a typed error carrying fields is genuinely required.
- `main` branches on `errors.Is(err, flag.ErrHelp)` and `errors.Is(err, cowsay.ErrUnknownCow)`. `LoadCow` branches on `errors.Is(err, fs.ErrNotExist)`.
- When you add a sentinel, add a `errors.Is`-based test alongside it (see `TestLoadCow_Nonexistent_SentinelError`). Test the sentinel relationship, not the error string.

## User-facing error messages

- Internal/library errors print verbatim via `fmt.Fprintln(stderr, err)` — acceptable for I/O and parse failures that don't have a dedicated user message.
- Errors that warrant a polished message get an explicit, hand-written string prefixed with `gosay: ` and printed instead of the raw error. Examples:
  - `gosay: unknown cowfile %q` (uses `%q` to quote the name)
  - `gosay: cannot combine -f and --random`
  - `gosay: -l cannot be combined with a message or animal selection`
  - `gosay: -W must be a positive number of columns`
- Plain usage errors (bad flags, interactive with no input) print `usage: gosay [flags] [message...]` to stderr.
- **Never leak internal details** in user-facing messages: no embed paths (`cows/<name>.cow`), no Go type names, no wrapped chain. The unknown-cow path deliberately drops the `cows/...:` detail by translating to `ErrUnknownCow`. There is a regression test asserting stderr does **not** contain `cows/` (`TestRun_CowFlag_UnknownAnimal`); keep messages clean enough to pass it.

## Flag parsing and validation

- Use `flag.NewFlagSet("gosay", flag.ContinueOnError)` (never `flag.ExitOnError`) so parse errors return through `run` as exit code 1 rather than exiting the process.
- Call `fs.SetOutput(stderr)` so the flag package's own diagnostics route to the injected stderr.
- Set `fs.Usage = func() {}` (a no-op) **before** `Parse` to suppress the flag package's automatic usage dump on `-h`/`--help`; otherwise help text bleeds onto stderr. Help is then printed manually to **stdout** with exit 0 on `errors.Is(err, flag.ErrHelp)`.
- Validate flag combinations after parsing and return exit 1 with a `gosay: ...` message for conflicts (`-f`+`--random`, `-l`+message/selection, explicit non-positive `-W`). Use `fs.Visit` to distinguish "explicitly set" from "default/zero value" when a zero value is itself a meaningful sentinel (e.g. `-W 0` vs. unset `-W`).

## Embedded FS error handling

- Embedded FS reads (`cowFS.ReadDir`, `cowFS.ReadFile`) can fail; always check and wrap their errors. `ListCows` wraps `ReadDir` failures (`listing embedded cows: %w`).
- `readCowFile` returns the raw FS error unwrapped; the **translation to `ErrUnknownCow`** happens one level up in `LoadCow` via the `fs.ErrNotExist` check. Keep this separation — low-level readers stay generic, the semantic mapping lives at the load boundary.
- After `ListCows()` succeeds in `--random`, the code indexes `names[rand.Intn(len(names))]` without a length guard, relying on the embed always being non-empty. Don't replicate that assumption for new pools that could legitimately be empty.

## Parser (heredoc) errors

- `parseCowBody` distinguishes three failure modes, each with a distinct message: scanner error (`%w`-wrapped), no opener found, and missing terminator (includes the captured terminator token via `%q`). Tests assert on substrings of these messages (`"no heredoc opener"`, the terminator token), so keep the distinguishing words stable when editing.

## Testing error paths

- Drive `run` directly with `bytes.Buffer` stdout/stderr; assert on exit code, stdout emptiness, and stderr substring — not on full error strings.
- For sentinels, assert `errors.Is`. For user-facing messages, assert the stable distinguishing substring (`"cannot combine"`, `unknown cowfile "name"`) and assert the absence of leaked internals (`cows/`).
