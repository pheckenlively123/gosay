# AGENTS.md

`gosay` is a zero-dependency Go reimplementation of `cowsay`: a message comes in (via CLI args or stdin), an ASCII-art animal "says" it, and the result goes to stdout. The full upstream cowsay menagerie is embedded in the binary via `//go:embed`, with a hand-authored gopher as the default animal. The codebase is deliberately small and stdlib-first — a thin CLI (`cmd/gosay`) over a pure rendering library (`internal/cowsay`).

## Docs index

Read the relevant guideline before touching code in that domain. These are authoritative and not duplicated below:

- `docs/security-guidelines.md` — security patterns, embed FS trust boundary, unescape/substitution ordering, input validation.
- `docs/error-handling-guidelines.md` — main/run split, exit codes, sentinel errors, wrapping/propagation, user-facing messages.
- `docs/testing-guidelines.md` — goldie golden-file pattern, fixture layout, table-driven conventions, CLI seam testing.

## Architecture & package layout

- Two packages only. `cmd/gosay/main.go` is the CLI layer (flag parsing, stdin/arg resolution, exit codes). `internal/cowsay/` is the pure rendering library. Keep this split: no rendering logic in `main`, no CLI/flag/exit concerns in `cowsay`.
- The library is one flat package split by responsibility, one concern per file: `embed.go` (embedded FS + listing), `cowfile.go` (heredoc parsing + `ErrUnknownCow`), `renderer.go` (the `Render` entry point + variable substitution), `balloon.go` (speech/thought bubble layout), `wrap.go` (word wrap + hard break). Add new functionality to the matching file rather than creating new ones; introduce a new file only for a genuinely new concern.
- The pipeline is fixed and ordering-sensitive: `LoadCow` (read + parse + unescape) → resolve effective `RenderOpts` defaults/width → `wrapMessage` → `buildBalloon` → `substituteVars` → trailing-newline normalization. Several steps are load-bearing in order (unescape before substitute; wrap before balloon so `maxWidth` is computed on wrapped lines). Do not reorder without reading the security guidelines and inline comments first.

## The `RenderOpts` pattern (defaults live downstream)

- `RenderOpts` uses the zero value as the "not set" sentinel for every field. Empty `Eyes`/`Tongue`/`Thoughts` and non-positive `Width` are resolved to real defaults (`"oo"`, two spaces, `\`/`o`, 40 cols) inside `Render`/`substituteVars`, **not** in `main`. `main` registers empty-string/zero flag defaults precisely so this single source of truth isn't bypassed.
- Consequence: the `helpText` const in `main.go` describes *effective* defaults and intentionally differs from the flag-registered defaults. `TestHelpText_MentionsRealDefaults` guards this — change a default in the renderer and update help text + that test in lockstep.
- To distinguish "explicitly set" from "zero value" for a flag whose zero is meaningful (e.g. `-W 0`), use `fs.Visit`, never a `== 0` check.

## Embedded assets & provenance

- `.cow` files live in `internal/cowsay/cows/` and are embedded via `//go:embed cows/*.cow`. They are vendored from upstream cowsay (see `cows/SOURCE.md` for the tag/commit and the documented refresh procedure; `cows/NOTICE` for per-file licensing). `gopher.cow` is hand-authored original art and must stay out of any re-vendor overwrite.
- All `.cow` files are LF-only, enforced by `.gitattributes` (`cows/*.cow text eol=lf`). Preserve LF endings when adding or refreshing cow files; the parser tolerates CRLF but the repo standard is LF.
- The `.cow` format is parsed as inert text — only the anchored heredoc opener regex is recognized; the surrounding Perl preamble is ignored, never executed. Do not add Perl interpretation.

## Code style

- Stdlib-first, minimal dependencies. Current deps are `mattn/go-runewidth` (display-width math) and `sebdah/goldie` (test-only). The project explicitly rejects cobra, urfave/cli, kong, testify, and go-wordwrap — do not add CLI frameworks, assertion libraries, or wrappers for code that fits in a few lines. New dependencies need a strong justification.
- All display-width and column math goes through `runewidth.StringWidth`/`RuneWidth` (via the `displayWidth` helper) so CJK/emoji borders stay aligned — never use `len()` or rune count for width.
- Exported identifiers and every non-trivial helper carry a doc comment, often citing the requirement/decision ID (e.g. `D-01`, `RENDER-05`, `Pitfall 3`) that the code locks in. When code encodes a deliberate spec decision or upstream-quirk deviation, document the *why* and the ID inline.
- Substitution is a single `strings.NewReplacer` pass covering both `$var` and `${var}` forms — never iterate or recurse substitutions.

## Common pitfalls

- Don't swap `embed.FS` reads for `os.ReadFile`/`filepath.Join` against a real directory — that turns `-f` into a path-traversal sink. The concatenation in `readCowFile` is safe only because `embed.FS` confines it.
- Don't merge unescape and substitution, or move unescape out of `parseCowBody`.
- Don't change `hardBreak` to advance by display width instead of UTF-8 byte size (guarantees termination + valid UTF-8).
- Keep stdout clean on error paths: every exit-1 path writes to stderr only and produces zero stdout bytes.

## Test & PR expectations

- Tests are white-box and drive production internals directly. Build "expected" values by calling the real helper (`substituteVars`, `buildBalloon`), not by re-implementing it.
- After any intentional render change, regenerate goldens with `go test -update ./internal/cowsay/` and review the `.golden` diff before committing — golden files are the spec, never hand-edited.
- Land tests at the lowest layer a change touches, plus a golden if rendered output changes, plus a `run`-level test for any new flag interaction.
- `go test ./...` is the gate. `go vet` + `staticcheck` (+ `govulncheck`) are expected before release.
