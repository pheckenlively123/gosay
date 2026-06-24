# Security Guidelines

Security conventions for `gosay`, a zero-dependency cowsay reimplementation. The threat surface is small: all input is local (CLI args, stdin) and all cow art ships embedded in the binary. These rules capture the non-obvious defenses already in place — preserve them when modifying code.

## Trust boundary: the embedded filesystem is the allowlist

All cow art is bundled via `//go:embed cows/*.cow` (`internal/cowsay/embed.go`). There is no runtime disk access, no user-supplied `.cow` files, no `COWPATH`. Treat `cowFS` as the sole source of cow data.

- `readCowFile(name)` builds the path as `"cows/" + name + ".cow"` and reads from `cowFS`. **This concatenation is intentionally NOT a vulnerability**: `embed.FS` confines reads to the embedded tree. Traversal attempts (`-f ../../etc/passwd`) resolve to a non-existent embedded path and return `fs.ErrNotExist` — never an escape. Verified: `cows/../../etc/passwd.cow` → file does not exist.
- **Do not** replace `embed.FS` reads with `os.ReadFile`, `os.Open`, or `filepath.Join` against a real directory. That would turn the `-f` value into a genuine path-traversal sink. If external cow loading is ever added, it MUST validate `name` (reject `/`, `\`, `..`) before any OS filesystem call.
- `ListCows` filters `cowFS.ReadDir("cows")` to `.cow` suffixes only; keep this filter so non-cow embedded files can't leak into listings or `--random` selection.

## Error design: sentinel errors hide internals from users

The package distinguishes "cow not found" from real I/O/parse failures using a sentinel, so user-facing messages never expose embed paths or Go internals.

- `ErrUnknownCow` (`cowfile.go`) is the only error callers branch on. `LoadCow` maps `fs.ErrNotExist` → `fmt.Errorf("%w: %s", ErrUnknownCow, name)`; all other errors are wrapped with context but remain opaque.
- `main.go` checks `errors.Is(err, cowsay.ErrUnknownCow)` and prints exactly `gosay: unknown cowfile "<name>"` — the raw internal error (e.g. `open cows/...: file does not exist`) is never shown for the unknown-cow case.
- When adding new failure modes, follow this split: define/reuse a sentinel for conditions the user can act on; wrap everything else with `%w` and let `main.go`'s fallback (`fmt.Fprintln(stderr, err)`) handle it. Do not echo attacker-controlled `name` into messages beyond the already-quoted form.
- On any error path, write to `stderr` and produce **no** stdout (tested: unknown cow yields empty stdout). Keep stdout clean so output is never half-rendered.

## Cow-file parsing: ordering of unescape vs. substitution is security-load-bearing

`.cow` files use Perl heredoc syntax with `$thoughts`/`$eyes`/`$tongue` placeholders. The parser (`cowfile.go`) and renderer (`renderer.go`) deliberately separate escape resolution from variable substitution.

- **Unescape happens at parse time, substitution at render time — never merge them.** `cowBodyUnescape` (resolves `\\`→`\`, `\@`→`@`, `\$`→`$`) runs inside `parseCowBody`, *before* `substituteVars` ever sees the body. This means a literal `\$eyes` in art becomes `$eyes` text and is NOT substituted — only genuine `$eyes` placeholders are. Reversing this order would let escaped sequences be reinterpreted as live variables. Golden tests (`golden_test.go`) lock this ordering in; do not "optimize" them away.
- Substitution is a single `strings.NewReplacer` pass (`substituteVars`) covering both `$var` and `${var}` forms. A single pass is intentional: replacement values are never re-scanned, so a user-supplied `-e`/`-T` value containing another placeholder string (e.g. `-e '$tongue'`) is inserted literally and cannot trigger a second substitution round. Keep all substitutions in this one replacer; never substitute iteratively or recursively.
- `parseCowBody` only ever recognizes the heredoc opener via the anchored regex `\$the_cow\s*=\s*<<["']?(\w+)["']?;?`. It does **not** execute, `eval`, or interpret any other Perl in the file — the rest of the `.cow` Perl preamble is ignored. Do not add general Perl interpretation; the format is parsed as inert text.
- The scanner buffer is capped (`scanner.Buffer(64KiB, 1MiB max)`). This bounds memory per line for embedded art. If non-embedded input is ever parsed, this cap becomes a real DoS control — keep it.
- A missing terminator or missing opener returns an error rather than silently consuming the rest of the file. Preserve these failure returns.

## User input is rendered as inert text — no shell, no eval

CLI args and stdin become the balloon message; `-e`/`-T`/`-W` tune rendering. None of these reach a shell, `exec`, template engine, or format string.

- The message is only ever passed to `strings`/`fmt.Fprintf` for layout (`balloon.go`, `wrap.go`). There is **no** `os/exec`, no `text/template`, no `format`-string-from-user-input anywhere. Keep it that way — never pass user message/flag content as the format argument to `Printf`-family calls (always use `%s`).
- `gosay` does **not** sanitize ANSI/terminal control sequences in the message — by design, matching upstream cowsay (a user echoing escape codes into their own terminal is their choice). Do not add silent stripping that would break legitimate art; if escape-sequence neutralization is ever wanted, make it an explicit opt-in flag, not a hidden default.

## Robust handling of malformed / hostile text input

The wrapper hardens against degenerate Unicode rather than trusting input is well-formed.

- `hardBreak` (`wrap.go`) advances by `utf8.DecodeRuneInString` **byte size**, never by display width. This guarantees every emitted chunk is valid UTF-8 even when wrapping mid-glyph, and guarantees termination: a single rune wider than the wrap width (e.g. a 2-column CJK glyph under `-W 1`) is emitted whole and advanced past, so there is **no infinite loop or data loss**. Do not change the advance to be width-based.
- Width math uses `runewidth.StringWidth` consistently (`balloon.go`/`wrap.go`) so border alignment can't be desynced by multi-width characters.

## Input validation in the CLI layer (`main.go`)

Flag combinations are validated up front and rejected with exit code 1 + a clean stderr message:

- `-f` + `--random` together → rejected (ambiguous selection).
- `-l` + (message | `--random` | explicit `-f`) → rejected.
- Explicit non-positive `-W` (without `-n`) → rejected; an *unset* `-W` (0) is a sentinel meaning "default 40", resolved downstream. Keep explicit-vs-default detection via `fs.Visit`, not by inspecting the zero value.
- `fs.Usage` is set to a no-op before `Parse` so `-h`/`--help` does not bleed usage onto stderr; help goes to stdout with exit 0. Preserve this.

## Dependencies & supply chain

- Keep the dependency tree minimal (currently `mattn/go-runewidth`, plus `sebdah/goldie` for tests). Every new dependency widens the supply-chain surface of a single-binary tool — prefer stdlib (the project explicitly rejects cobra, urfave/cli, testify, etc.).
- Releases must be produced by GitHub Actions / GoReleaser, never hand-built locally — this keeps a reproducible, auditable build provenance for the distributed binary.
- Run `go vet` + `staticcheck` and (recommended) `govulncheck` in CI before release to catch known-vulnerable dependency versions.
