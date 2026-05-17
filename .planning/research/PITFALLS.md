# Pitfalls Research

**Domain:** Go CLI reimplementation of a Perl-originated tool (cowsay)
**Researched:** 2026-05-17
**Confidence:** HIGH (parsing/licensing), MEDIUM (build pipeline, Unicode edge cases)

---

## Critical Pitfalls

### Pitfall 1: Treating .cow Files as Simple Text Templates

**What goes wrong:**
A naive parser scans for `$eyes`, `$tongue`, and `$thoughts` and does string replacement, but silently misrenders cows that use backslash escape sequences. The result is ASCII art with doubled backslashes (`\\` appearing as literal `\\` instead of `\`), or at-sign corruption where `\@` renders as `\@` instead of `@`.

**Why it happens:**
`.cow` files are Perl source; inside a Perl heredoc (`<<"EOC"` or `<<EOC`), backslashes are Perl metacharacters:
- `\\` → renders as a single `\`
- `\@` → renders as `@` (escaping the array sigil)
- `\$` → renders as `$` (literal dollar sign, not a variable)

A Go parser that never interprets these sequences will double every backslash in the output. Many cows (including `default.cow`) use `\\` for parts of the body.

**How to avoid:**
After extracting the heredoc body, run a single unescape pass over the raw content before substituting variables:
1. `\\` → `\`
2. `\@` → `@`
3. `\$` → `$` (only for literal dollars that are NOT preceded by a known variable name)

Do NOT attempt general Perl string interpolation (`${expr}`, `@array` expansion, etc.) — those features appear in exotic cows only and are explicitly out of scope. Implement the three rules above and nothing more.

**Warning signs:**
- `default.cow` renders with `\\` visible in the tail/body
- Any cow with `/` characters looks wrong (they don't need escaping, so this is a sign you're over-escaping)
- Run `diff` between your output and upstream Perl cowsay on the first five cows

**Phase to address:** Core parsing phase (first implementation milestone).

---

### Pitfall 2: Unicode-Unaware Bubble Width Calculation

**What goes wrong:**
The speech bubble is drawn as a rectangle. Calculating `max(len(line))` using `len(line)` counts bytes; using `len([]rune(line))` counts Unicode code points. Neither is correct for CJK characters, emoji, or combining characters, which have display widths that differ from their byte or rune counts. The bubble ends up misaligned: too narrow for wide characters (CJK kanji each occupy 2 terminal columns) or too wide for combining characters (which occupy 0 additional columns).

**Why it happens:**
ASCII tooling authors use `len(s)` and it works for English. Go's rune model handles multibyte correctly but still doesn't account for terminal display width — a kanji `漢` is one rune but prints in two columns. This is a known, longstanding bug in the original Perl cowsay (Ubuntu bug #393212, Arch bug FS#48347).

**How to avoid:**
Use `github.com/mattn/go-runewidth` (the standard Go library for this problem, used by virtually every terminal UI library). Call `runewidth.StringWidth(line)` instead of `len([]rune(line))` for bubble sizing and word-wrap column counting. This handles CJK full-width, half-width, combining, and emoji width correctly.

Note: PROJECT.md says "pure Go, standard library where reasonable." `go-runewidth` is a single, small, well-maintained dependency — this is the one place where pulling it in is clearly justified. The alternative (implementing Unicode East Asian Width tables yourself) is far more error-prone.

**Warning signs:**
- Test with `echo "漢字テスト" | gosay` — bubble right edge misaligns with content
- Test with `echo "🐮🐄🐮" | gosay` — same symptom

**Phase to address:** Core rendering phase (text wrapping and bubble drawing).

---

### Pitfall 3: Licensing Omission — Vendoring .cow Files Without Attribution

**What goes wrong:**
The `.cow` files from upstream cowsay are not public domain. Vendoring them without attribution or license documentation may create legal ambiguity and will certainly be noticed by users and package maintainers who check licenses.

**Why it happens:**
Developers focus on functionality and overlook license compliance when vendoring assets from other projects. This is an easy oversight on a "toy" project.

**Details:**
Upstream cowsay (cowsay-org/cowsay) is licensed under **GPL v3 or later**. Most individual `.cow` files carry a dual-license: **GPL 1.0 or later, OR Artistic License 1.0**. Some files have distinct licenses:
- `apt.cow` — GPL only
- `gnu.cow`, `suse.cow` — WTFPL-2
- `kangaroo.cow` — GPL 2.0+
- `daemon.cow` — historically flagged with unclear provenance (removed from Fedora packaging)

Copyright holders: Tony Monroe (1999–2002), Andrew Janke (2016–2024), and GitHub contributors.

**How to avoid:**
1. Create `cows/NOTICE` (or `THIRD_PARTY_LICENSES.md`) listing: upstream project URL, copyright holders, and license(s).
2. Review and document any per-file license variations (especially `daemon.cow`).
3. `gosay`'s own `LICENSE` is already present in the repo root — ensure it is compatible or clearly scoped to the Go code, not the vendored cows.
4. Consider whether to vendor `daemon.cow` at all given its provenance concerns.
5. Reference the attribution in the `README`.

**Warning signs:**
- No `NOTICE` or `THIRD_PARTY_LICENSES` file in the repo
- `LICENSE` file only covers gosay's Go code with no mention of vendored cow files
- No comment in the `cows/` directory README pointing to upstream source

**Phase to address:** Asset vendoring phase (must be resolved before any public release).

---

### Pitfall 4: Heredoc Extraction Misidentifying the Terminator

**What goes wrong:**
The parser finds `$the_cow = <<"EOC";` (or `<<EOC;`, `<<'EOC'`) and extracts everything until a line that is exactly `EOC`. If the terminator varies (e.g., `END`, `EOT`, `COWS`) or is not anchored to the start of the line, the extraction silently includes too much or too little content, producing a garbled or blank animal.

**Why it happens:**
Some cow files use heredoc terminators other than `EOC`. The Perl heredoc spec requires the terminator to be on its own line with no leading whitespace. Implementations that hardcode `EOC` as the terminator string break on these files.

**How to avoid:**
- Capture the terminator string dynamically from the `$the_cow = <<"TERMINATOR"` or `$the_cow = <<TERMINATOR` line using a regex.
- Match the terminator against lines anchored at `^TERMINATOR$` (no leading/trailing whitespace).
- Handle all three heredoc quoting styles: `<<"EOC"` (interpolating), `<<'EOC'` (non-interpolating — same behavior for our purposes), `<<EOC` (interpolating, same as double-quoted).

**Warning signs:**
- A specific cow file renders as empty or as extra Perl code visible in the art
- Look for `.cow` files in the vendored set that do not use `EOC` as the terminator and test them first

**Phase to address:** Core parsing phase.

---

### Pitfall 5: `$thoughts` Variable — cowthink Mode Not Defaulting Correctly

**What goes wrong:**
In "say" mode, `$thoughts` is `\` (a backslash) and the bubble connector is `\`. In "think" mode, `$thoughts` is `o` and the bubble connectors are `(` / `)` for thought bubbles. Forgetting to vary `$thoughts` means `cowthink` draws a speech bubble with a `\` tail instead of a thought-bubble `o`.

Similarly: `$eyes` defaults to `oo`, `$tongue` defaults to `  ` (two spaces). If the user hasn't passed `-e` or `-T`, the parser must supply these defaults before substitution, not substitute empty strings.

**How to avoid:**
Define a `CowVars` struct with explicit defaults:
```go
type CowVars struct {
    Eyes     string // default "oo"
    Tongue   string // default "  " (two spaces)
    Thoughts string // default "\\" for say, "o" for think
}
```
Apply defaults before substitution. If the user passes `-e xx`, the struct field is overridden; otherwise the default is used.

**Warning signs:**
- `cowthink` variant uses `\` as the thought connector instead of `o`
- Eyes appear blank or tongue produces garbage output when no flag is passed
- A cow file that references `$tongue` but user never passed `-T` results in an empty pair of characters

**Phase to address:** Core rendering phase (CLI flag handling and variable default resolution).

---

## Moderate Pitfalls

### Pitfall 6: Terminal Width Detection Overcomplicated for This Use Case

**What goes wrong:**
Implementors try to detect the real terminal width via `TIOCGWINSZ` or `golang.org/x/term.GetSize()` to dynamically wrap at the current terminal width. This causes unexpected behavior when piping: `echo foo | gosay | cat` detects a non-terminal fd, falls back to 0 or errors, and either wraps at 0 (every word on its own line) or panics.

**The right behavior (matching upstream cowsay):**
Original cowsay hardcodes `-W 40` as the default. There is no automatic terminal width detection in the reference implementation. The user explicitly passes `-W 80` if they want wider output.

**How to avoid:**
Hard-code the default wrap width to 40 (matching upstream). Expose `-W <n>` to override. Do NOT attempt TIOCGWINSZ detection — it adds a dependency, complicates Windows cross-compilation, and deviates from the documented default. Checking `$COLUMNS` as a fallback is optional and low-value for this project scope.

**Warning signs:**
- `echo hello | gosay | wc -l` produces different output depending on terminal vs. pipe context
- Windows build fails due to syscall import for terminal size

**Phase to address:** Core rendering phase (wrap width implementation).

---

### Pitfall 7: go:embed Pattern Accidentally Excluding or Including Wrong Files

**What goes wrong:**
Two opposite failure modes:
1. `//go:embed cows` — excludes any `.cow` file whose name starts with `.` or `_`. Unlikely for standard cow files but possible for future additions.
2. `//go:embed all:cows` — includes `.DS_Store`, `Thumbs.db`, git-related files, and any other hidden files that crept into the `cows/` directory during development. These inflate binary size and may expose unintended content.

Additionally, `go:embed` is case-sensitive for pattern matching but behaves against the actual filesystem. On macOS (case-insensitive), you can reference `Cow.cow` and match `cow.cow` — but the same code breaks on Linux (case-sensitive). This can silently pass CI on a macOS developer machine and fail on Linux.

**How to avoid:**
- Use `//go:embed cows/*.cow` — this is explicit, includes only `.cow` files, no hidden files, no edge cases. It also documents intent clearly.
- Alternatively use `//go:embed cows` if you are confident no dot/underscore files exist and want subdirectory support.
- All cow file lookups must use lowercase names or normalize to lowercase at lookup time. Name all vendored files in lowercase.
- Never develop or test embed patterns only on macOS — verify on Linux (the GitHub Actions runner will surface this).

**Warning signs:**
- Binary size unexpectedly large
- Missing cow names that exist on the filesystem
- `-l` flag lists cows on one OS but not another

**Phase to address:** Asset embedding phase.

---

### Pitfall 8: Line Ending Corruption of Cow Art on Windows

**What goes wrong:**
`go:embed` reads files as raw bytes at compile time — line endings are preserved exactly as stored. If the vendored `.cow` files have LF endings (standard for files from a Unix/Linux repository) but the cow-rendering code on Windows decides to normalize or add `\r`, the ASCII art gains an extra carriage return before each newline. This shows as a rightward shift of every line in some terminals or corrupts the art entirely in others.

**Why it happens:**
Go source and `go:embed` embedded bytes are binary-clean — no automatic CRLF translation occurs in the `embed` package itself. However, if your rendering code writes `fmt.Println(line)` on Windows, `Println` only adds `\n`. The Windows terminal handles this correctly for most cases. The real risk is if anyone runs `git clone` with `core.autocrlf=true`, which would convert the vendored `.cow` files to CRLF on checkout. Then the embedded file contains `\r\n`, and the cow art renders with trailing `\r` on every line.

**How to avoid:**
- Add a `.gitattributes` file at the repo root specifying: `cows/*.cow text eol=lf` — this forces LF regardless of `core.autocrlf` settings.
- If defensive, strip `\r` when reading cow file bytes from the embedded FS before rendering.

**Warning signs:**
- Cow art has trailing `^M` visible in some terminals
- Art is misaligned only on Windows builds

**Phase to address:** Asset vendoring phase (set up `.gitattributes` when adding the `cows/` directory).

---

### Pitfall 9: GoReleaser Misconfiguration Causing Missing Binaries or Broken Releases

**What goes wrong:**
Common release pipeline failures:
1. Workflow triggers on every push instead of only on version tags — every commit attempts a full release.
2. `fetch-depth: 0` omitted from the checkout step — GoReleaser cannot determine the version from git tags and produces `v0.0.0-SNAPSHOT` or fails entirely.
3. `contents: write` permission not set on the workflow's `GITHUB_TOKEN` — GoReleaser cannot create the GitHub Release artifact and silently exits.
4. Version stamping in `ldflags` doesn't match the variable names in `main.go` — the binary reports `dev` as version at runtime.
5. The workflow runs the release action on pull requests or pushes to main instead of tag pushes — fails because no tag exists to version.

**How to avoid:**
- Trigger the release workflow only on `push: tags: ['v*']`.
- Always include `fetch-depth: 0` in the checkout action.
- Set `permissions: contents: write` in the workflow job.
- Use GoReleaser's standard ldflag template to stamp the version:
  ```yaml
  ldflags:
    - -s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}
  ```
  And in `main.go`: declare `var version = "dev"`, `var commit = "none"`, `var date = "unknown"`.
- Test locally first: `goreleaser release --snapshot --clean`.
- Never skip the `--snapshot` step before your first real release.

**Warning signs:**
- Release job succeeds but no artifacts appear on the GitHub Release page
- Binary `--version` flag prints `dev` after a tagged release
- Release action runs on every PR

**Phase to address:** Release pipeline phase.

---

### Pitfall 10: Scope Creep — Bikeshedding Traps

**What goes wrong:**
The project's small scope makes it an attractive canvas for "just one more feature." The following are known time-sinks that are explicitly out of scope per PROJECT.md and should be resisted:

| Trap | Why It Feels Tempting | Why It Destroys Scope |
|------|-----------------------|-----------------------|
| ANSI color output | "Make the gopher colorful!" | Adds terminal detection, color library, defeats piping, requires stripping logic |
| Library API (`gosay.Say(...)`) | "Others could import this" | PROJECT.md explicitly excludes it; forces API stability concerns |
| Runtime `.cow` file loading from disk | "Power users want custom cows" | Defeats single-binary distribution — the whole point of the project |
| `COWPATH` environment variable support | "Upstream supports it" | Same as above; partial compatibility is worse than none |
| Strict Perl cowsay flag parity | "Should match -bdgpstwy" | Chases edge cases; PROJECT.md says "common flags only" |
| Custom gosay-native cow format | "The Perl format is ugly" | Breaks ecosystem compatibility and discards all upstream artwork |
| Plugin system for new animals | "Extensibility!" | Over-engineering a toy; explicitly noted in PROJECT.md |
| Windows `.exe` icon / resource embedding | "Feels more professional" | Requires CGO or external tools; no material user benefit |

**How to avoid:**
When a new feature idea arises, check PROJECT.md's "Out of Scope" section first. If it's there, the answer is no. If it's not there but feels like it belongs there, add it during the next milestone review via `/gsd-transition`.

**Warning signs:**
- A new dependency appears in `go.mod` that isn't `go-runewidth`
- Any conversation starting with "wouldn't it be cool if..."
- A PR that adds more than one new flag to the CLI

**Phase to address:** All phases — this is a discipline issue, not a single-phase concern. The SUMMARY.md and PROJECT.md "Out of Scope" section are the enforcement mechanism.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| `strings.Replace` for all cow variable substitution | Fast to write | Breaks if a cow body contains literal `$eyes` text (rare but real) | Use regex anchored to whole-word match instead |
| Hardcoded `EOC` as heredoc terminator | Simpler parser | Silently misrenders cows with other terminators (`END`, `EOT`) | Never — capture terminator dynamically with one extra regex group |
| `len(s)` for bubble width | Works for ASCII test cases | Breaks for any non-ASCII input | Never — use `runewidth.StringWidth` from the start |
| Skip `.gitattributes` | Nothing to configure | CRLF corruption for Windows contributors | Acceptable to defer until first Windows bug report, but low cost to add upfront |
| No `NOTICE` file for vendored cows | Faster to ship | Legal/attribution gap; package maintainers will flag it | Never for a public release |

---

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| `go:embed` + cow files | Using `//go:embed cows` and finding cows excluded on some OS | Use `//go:embed cows/*.cow` for explicit, OS-consistent inclusion |
| GitHub Actions + GoReleaser | Missing `fetch-depth: 0`, missing `contents: write` | Follow GoReleaser docs explicitly; test with `--snapshot` first |
| `golang.org/x/term` for terminal width | Calling `term.GetSize(os.Stdin.Fd())` fails on Windows, piped stdin | Don't detect terminal width; use hardcoded 40 with `-W` override |
| Embedded cow FS lookup | Case-sensitive path on Linux vs case-insensitive on macOS | Normalize all cow names to lowercase; verify on Linux CI |

---

## "Looks Done But Isn't" Checklist

- [ ] **Bubble rendering:** Verify with CJK input — `echo "漢字テスト" | gosay` — bubble width correct
- [ ] **cowthink mode:** Verify `$thoughts` is `o` and bubble corners are `(` `)` not `\` `/`
- [ ] **Backslash unescaping:** Check `default.cow` — the tail `\` must appear as a single backslash, not `\\`
- [ ] **Escape at-signs:** Find a cow that uses `\@` and confirm it renders as `@`
- [ ] **License file:** Confirm `NOTICE` or `THIRD_PARTY_LICENSES` exists with cowsay-org attribution
- [ ] **Embedded cow count:** `-l` flag should list the same number of cows as files in `cows/`
- [ ] **Windows binary:** Cross-compiled Windows binary must run without CGO or external DLLs
- [ ] **Version stamping:** Tagged release binary — `gosay --version` must print the git tag, not `dev`
- [ ] **Stdin piping:** `echo hello | gosay` must work; `gosay hello` must also work
- [ ] **Long lines:** A 200-character message must wrap at 40 columns by default

---

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| .cow escape sequence misrendering | Core parsing (Phase 1) | Diff output vs. Perl cowsay on 5 cows |
| Unicode bubble width | Core rendering (Phase 1) | CJK and emoji input tests |
| Licensing / NOTICE file | Asset vendoring (Phase 1) | Repo contains NOTICE before first public push |
| Heredoc terminator varies | Core parsing (Phase 1) | Test cows with non-EOC terminators |
| cowthink `$thoughts` default | CLI flags (Phase 1) | `echo test | gosay -t` renders thought bubble |
| Terminal width detection complexity | Core rendering (Phase 1) | Hard-code 40; no TIOCGWINSZ calls |
| go:embed pattern gotchas | Asset embedding (Phase 1) | `//go:embed cows/*.cow`; verify on Linux CI |
| Line ending corruption | Asset vendoring (Phase 1) | `.gitattributes` set; test CRLF scenario |
| GoReleaser misconfiguration | Release pipeline (Phase 2) | Snapshot release works; tagged release produces artifacts |
| Scope creep | All phases | Check PROJECT.md Out of Scope before adding any feature |

---

## Sources

- [cowsay-org/cowsay — upstream repository](https://github.com/cowsay-org/cowsay)
- [cowsay-org Licensing.md — per-file license details](https://github.com/cowsay-org/cowsay/blob/master/doc-project/Licensing.md)
- [Ubuntu bug #393212 — cowsay UTF-8 multibyte length miscalculation](https://bugs.launchpad.net/bugs/393212)
- [Arch Linux bug FS#48347 — cowsay Unicode string length incorrect](https://bugs.archlinux.org/task/48347)
- [Go embed package docs — default exclusions, all: prefix](https://pkg.go.dev/embed)
- [golang/go issue #42328 — surprising inclusion of hidden files in embed.FS](https://github.com/golang/go/issues/42328)
- [golang/go issue #48888 — go:embed /* ignores subfolders with dot files](https://github.com/golang/go/issues/48888)
- [GoReleaser GitHub Actions docs — permissions, secrets, common mistakes](https://goreleaser.com/ci/actions/)
- [python-cowsay — how non-Perl ports handle cow file parsing](https://pypi.org/project/python-cowsay/)
- [James-Ansley/cowsay Python implementation notes](https://github.com/James-Ansley/cowsay)
- [quantum5/cowsay C++ implementation — escape sequence handling](https://github.com/quantum5/cowsay/blob/master/cowsay.cpp)
- [rivo/uniseg — Unicode text segmentation, display width in Go](https://github.com/rivo/uniseg)
- [golang.org/x/term — GetSize Windows limitations](https://pkg.go.dev/golang.org/x/term)
- [golang.org/x/text/width — East Asian Width tables](https://pkg.go.dev/golang.org/x/text/width)

---
*Pitfalls research for: Go cowsay reimplementation (gosay)*
*Researched: 2026-05-17*
