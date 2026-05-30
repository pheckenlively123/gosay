---
phase: 02-input-and-cow-selection
reviewed: 2026-05-30T00:00:00Z
depth: standard
files_reviewed: 5
files_reviewed_list:
  - cmd/gosay/main.go
  - cmd/gosay/main_test.go
  - internal/cowsay/cowfile.go
  - internal/cowsay/cowfile_test.go
  - internal/cowsay/golden_test.go
findings:
  critical: 1
  warning: 4
  info: 3
  total: 8
status: issues_found
---

# Phase 2: Code Review Report

**Reviewed:** 2026-05-30T00:00:00Z
**Depth:** standard
**Files Reviewed:** 5
**Status:** issues_found

## Summary

Reviewed the Phase 2 flag-based CLI surface (`cmd/gosay/main.go`) and the
`ErrUnknownCow` sentinel plumbing in `internal/cowsay/cowfile.go`, plus their
tests. The cross-module error chain is sound: `LoadCow` wraps with
`fmt.Errorf("%w: %s", ErrUnknownCow, name)`, `Render` re-wraps with `%w`, and
`main.go`'s `errors.Is(err, cowsay.ErrUnknownCow)` correctly unwraps the full
chain. Path-traversal via `-f` is not exploitable because `embed.FS.ReadFile`
rejects any path failing `fs.ValidPath` (`..`, leading `/`) with
`fs.ErrNotExist`, which maps cleanly to the unknown-cow message. Terminal-escape
injection through the `-f` value into stderr is neutralized by the `%q` verb,
which Go-quotes control bytes.

The dominant defect is a reachable panic in the `--random` path when the cow
pool is empty (`rand.Intn(0)`). There are also several robustness and
test-coverage gaps worth addressing.

## Critical Issues

### CR-01: `--random` panics on an empty cow pool (`rand.Intn(0)`)

**File:** `cmd/gosay/main.go:105-112`
**Issue:** When `randomFlag` is set, the code does
`animal = names[rand.Intn(len(names))]`. `ListCows()` can legitimately return a
non-error empty slice — `cowFS.ReadDir("cows")` succeeds even if the directory
contains only non-`.cow` files (the loop in `embed.go` skips every entry that
fails the `.cow` suffix check, see `internal/cowsay/embed.go:21-27`). With an
empty `names`, `rand.Intn(0)` panics ("invalid argument to Intn"), crashing the
process with a stack trace instead of a clean exit code. The `ListCows()` error
return is checked, but the empty-result case is not.
**Fix:**
```go
names, err := cowsay.ListCows()
if err != nil {
    fmt.Fprintln(stderr, err)
    return 1
}
if len(names) == 0 {
    fmt.Fprintln(stderr, "gosay: no cows available")
    return 1
}
animal = names[rand.Intn(len(names))]
```

## Warnings

### WR-01: `io.ReadAll(os.Stdin)` is unbounded

**File:** `cmd/gosay/main.go:127`
**Issue:** Piped stdin is read with `io.ReadAll`, which has no upper bound. A
process feeding an unbounded or very large stream (e.g. `yes | gosay`) forces
gosay to buffer the entire input in memory before rendering, and the rendered
balloon would be enormous. For a toy CLI this is not a security hole, but a
bounded read is cheap insurance and matches the spirit of "message in, animal
out."
**Fix:** Cap the read with `io.LimitReader`, e.g.
`data, err := io.ReadAll(io.LimitReader(os.Stdin, 1<<20))` (1 MiB), and
optionally warn when the limit is hit.

### WR-02: `run()` reads `os.Stdin` / `isTTY(os.Stdin)` directly — stdin path is untestable

**File:** `cmd/gosay/main.go:121,127`
**Issue:** `run()` accepts injected `stdout`/`stderr` writers but reaches out to
the package-global `os.Stdin` for both the TTY check and the read. As the
comment in `main_test.go:9-16` admits, this leaves the entire piped-stdin branch
(the `io.ReadAll` + `TrimSuffix` logic, INPUT path) with no unit coverage —
`TestRun_NoArgs` is forced to accept *either* exit code 0 or 1 because it cannot
control stdin. A bug in the stdin trimming or read-error handling would not be
caught by any test in this phase.
**Fix:** Add a `stdin io.Reader` parameter to `run()` (and a way to signal
TTY-ness, e.g. pass `isTTY` result or an interface), so the piped, empty, and
read-error stdin cases can be exercised deterministically:
```go
func run(args []string, stdin io.Reader, stdout, stderr io.Writer, stdinIsTTY bool) int
```

### WR-03: `TestRun_NoArgs` asserts a tautology and cannot fail on the no-args contract

**File:** `cmd/gosay/main_test.go:17-33`
**Issue:** Because `run()` consults the real `os.Stdin`, the test accepts both
exit code 0 and exit code 1 (`if code != 0 && code != 1`), and only checks
stderr/stdout when `code == 1`. Under `go test` stdin is typically a pipe
(non-TTY), so in CI this test takes the `code == 0` branch and asserts nothing
about the usage message — the named contract ("no args + no pipe -> usage, exit
1") is never actually verified. This is a flaky/vacuous test masquerading as
coverage.
**Fix:** After applying WR-02's injection seam, split into two deterministic
tests: one with a non-TTY empty `stdin` asserting the empty-bubble render
(exit 0), and one with `stdinIsTTY=true` asserting usage on stderr + exit 1.

### WR-04: Unknown-cow detection only covers `fs.ErrNotExist`, not invalid embed paths

**File:** `internal/cowsay/cowfile.go:92-99`
**Issue:** `LoadCow` maps to `ErrUnknownCow` only when
`errors.Is(err, fs.ErrNotExist)`. `embed.FS.ReadFile` returns a `*fs.PathError`
wrapping `fs.ErrInvalid` (not `ErrNotExist`) for names that fail `fs.ValidPath`
— e.g. `-f "../foo"` -> `cows/../foo.cow`, or `-f ""` -> `cows/.cow`. Those
fall through to the generic `fmt.Errorf("load cow %q: %w", name, err)` branch,
which in `main.go` is printed verbatim via `fmt.Fprintln(stderr, err)`
(line 142). That generic message leaks the internal `cows/...` embed path that
`TestRun_CowFlag_UnknownAnimal` (main_test.go:97-100) explicitly asserts must
not leak — the test only happens to pass because `"nosuchcow"` yields
`ErrNotExist`, not because the `..`/empty cases are handled.
**Fix:** Treat invalid paths as unknown cows too:
```go
if errors.Is(err, fs.ErrNotExist) || errors.Is(err, fs.ErrInvalid) {
    return ParsedCow{}, fmt.Errorf("%w: %s", ErrUnknownCow, name)
}
```
and add a test for `LoadCow("../etc/passwd")` / `LoadCow("")`.

## Info

### IN-01: `formatCowList` wrap width is a bare magic number

**File:** `cmd/gosay/main.go:28`
**Issue:** `const wrapWidth = 76` is local and undocumented beyond a comment.
Acceptable, but the "approximately 76" relationship to upstream cowsay's column
shape is fragile; if upstream is ever matched precisely this should be derived,
not guessed.
**Fix:** Leave as-is, or add a brief reference to the upstream constant it
mirrors.

### IN-02: Usage string duplicated between `fs.Usage` and the no-args branch

**File:** `cmd/gosay/main.go:65,123`
**Issue:** The literal `"usage: gosay [-f name] [-l] [--random] [message...]"`
appears twice. If flags change, the two can drift out of sync.
**Fix:** Call `fs.Usage()` in the no-args branch instead of re-printing the
literal, or hoist the string to a package const.

### IN-03: `--random` ignores an explicit `-f` only after parse, but default `-f` value is silently discarded

**File:** `cmd/gosay/main.go:60,104-112`
**Issue:** `cowName` defaults to `"gopher"`; on the `--random` path `animal` is
overwritten and the default is discarded, which is correct. Minor: the
`fs.Visit` "was -f explicitly set" detection (lines 73-78) is the right idiom,
but a short comment noting that the default `"gopher"` must never count as
"explicit" would prevent a future refactor from breaking the D-06 conflict
check.
**Fix:** Add a one-line comment; no code change required.

---

_Reviewed: 2026-05-30T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
