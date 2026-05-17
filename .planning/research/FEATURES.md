# Feature Landscape

**Domain:** cowsay-family CLI tool (Go reimplementation)
**Researched:** 2026-05-17
**Project scope:** "toy, not framework" — minimal, single binary, no library API

---

## Upstream Flag Surface Analysis

The canonical Perl `cowsay`/`cowthink` (cowsay-org/cowsay) defines this flag set:

| Flag | Purpose | Universally expected? |
|------|---------|----------------------|
| `-f <cowfile>` | Select animal | YES — core feature |
| `-l` | List animals | YES — core feature |
| `-W <col>` | Wrap column (default 40) | YES — users notice missing wrap |
| `-n` | Disable word wrap | YES — pipeline users need raw passthrough |
| `-e <xx>` | Custom eyes (2 chars) | MEDIUM — fun but niche |
| `-T <xx>` | Custom tongue (2 chars) | LOW — very niche |
| `-b` | Borg eyes (`==`) | LOW — mood preset |
| `-d` | Dead eyes (`xx`) + sticking-out tongue | LOW — mood preset |
| `-g` | Greedy eyes (`$$`) | LOW — mood preset |
| `-p` | Paranoid eyes (`@@`) | LOW — mood preset |
| `-s` | Stoned eyes (`**`) + tongue | LOW — mood preset |
| `-t` | Tired eyes (`--`) | LOW — mood preset |
| `-w` | Wired eyes (`OO`) | LOW — mood preset |
| `-y` | Youthful eyes (`..`) | LOW — mood preset |
| `-r` | Random cow selection | MEDIUM — common in clones |
| `-C` | True-color mode (only with `-r`) | LOW — fancy terminal only |
| `-h` | Help | YES — any CLI |

**Key finding:** The mood presets (`-b/-d/-g/-p/-s/-t/-w/-y`) are simply macros that set specific `-e` and `-T` values. They're a single switch statement on top of `-e`/`-T`. Without `-e`/`-T`, the presets cannot be implemented — so these features are tightly coupled.

---

## Table Stakes

Features users expect. Missing = product feels broken or unfinished.

| Feature | Why Expected | Complexity | Dependencies |
|---------|--------------|------------|--------------|
| Positional arg as message | `cowsay hello world` — the primary use case | Low | None |
| Stdin piping (`echo hello \| gosay`) | Equally expected by pipeline users | Low | None |
| Default animal is visible and works | First run must produce output | Low | Gopher .cow file |
| `-f <name>` animal selection | Selecting animals is the core game | Low | `.cow` file set embedded |
| `-l` list animals | Users need to discover what's available | Low | Embedded file list |
| Word wrap at 40 cols (default) | Without this, long messages break layout badly | Medium | Balloon renderer |
| `-W <col>` configurable wrap width | Pipeline users often want wider or narrower | Low | Balloon renderer |
| `-n` no-wrap passthrough | `fortune \| cowsay -n` is a classic use case | Low | Balloon renderer |
| Balloon drawing (speech bubble) | The defining visual element | Medium | Wrap logic |
| `cowthink` thought bubble variant | Universally expected alongside `cowsay` | Low | Balloon renderer |
| Help output (`-h`) | Any CLI without `-h` feels unfinished | Low | None |
| Clean error on bad `-f` name | Users typo animal names; a panic is unacceptable | Low | None |

### Stdin Behavior Details

Upstream behavior (HIGH confidence from man pages):
- If non-option arguments remain after flag parsing, they become the message; stdin is ignored.
- If no non-option arguments remain, cowsay reads from stdin.
- Both-at-once (piped AND arg) is not a supported combined mode — args win.
- Empty stdin: produces a balloon containing an empty string (or one empty line). This is technically valid output. Do not special-case it with an error.
- Multi-line stdin: each line is a paragraph; wrap applies per-line or globally depending on implementation. Upstream treats all stdin as one message, wrapping the whole thing.
- Very long single lines without spaces: do not wrap (no word boundary to break at); let them overflow or force-break at width. Most implementations overflow gracefully.

### Balloon Drawing Format

The speech bubble is the most fiddly rendering detail to get right:

```
 _________
< message >
 ---------
```

For multi-line:
```
 ___________
/ message   \
| line 2    |
\ line 3   /
 -----------
```

The border characters differ for single-line (`< >`) vs multi-line (`/ \` on first/last, `| |` on middle). Getting this right is required — it's what cowsay IS.

Thought bubbles use `( )` on every line and `o o o` as the thought trail.

---

## Differentiators

Features that would make gosay stand out. This project explicitly wants minimal scope — these are candidates only if they cost almost nothing.

| Feature | Value Proposition | Complexity | Tradeoff | Verdict |
|---------|-------------------|------------|---------|---------|
| **Gopher default** | Signature Go-native identity; no other cowsay starts with gopher | Low — just a different default cow name | None, already decided | BUILD — already in scope |
| **`--think`/`-t` flag on `gosay`** | Users get both modes from one binary; no `cowthink` symlink needed | Low — one boolean flag changes bubble chars | The flag `-t` conflicts with upstream's "tired" eye preset — use `--think` or a separate binary to avoid ambiguity | BUILD with `--think` long flag |
| **`cowthink` as second binary** | Matches upstream muscle memory; `cowthink` is well-known | Low — second `main.go` or `argv[0]` detection | Doubles install footprint slightly; two binaries to document | OPTIONAL — add as symlink target or second binary if `--think` is already there |
| **Random animal mode** | Fun for `fortune \| gosay -r`; common in clones | Low — `rand.Intn(len(cows))` | Not requested; adds one flag; risk: users expect it so might as well | CANDIDATE — low cost, see analysis below |
| **`-e`/`-T` eye and tongue customization** | Allows `gosay -e ^^ hello` fun; table stakes in Neo-cowsay | Low — just substitute into cow template vars | Core `.cow` parser already needs `$eyes`/`$tongue` vars; these flags cost near nothing once parser exists | BUILD — near-zero marginal cost |
| **Mood presets (`-d` dead, `-s` stoned etc.)** | Complete upstream compat; users who know cowsay expect them | Low — macros over `-e`/`-T` once those exist | Adds 8 flags to help output (visual clutter); presets are niche in practice | SKIP — not requested; adds flag noise |
| **Small curated set of gopher variants** | `gopher-beret.cow`, `gopher-witch.cow` etc. in the embedded set | Medium — requires making/sourcing ASCII art | Fun but introduces a curation and maintenance burden | SKIP for v1 — revisit if upstream gopher.cow exists |

### Random Animal Mode: Tradeoff Analysis

Random mode (`-r` in upstream, `--random` in Neo-cowsay) is the single most common feature found in every Go clone surveyed. Implementation cost is trivial — pick a random index from the embedded names list. The use case `fortune | gosay -r` is genuinely fun and is how the tool tends to get demoed.

**Recommendation:** Add `--random` flag. It costs ~5 lines of code once the animal list exists for `-l`. The project's "toy" framing actually makes this a good fit — it's playful, not enterprisey.

### `cowthink` Integration Recommendation

Upstream approach: `cowthink` is the same binary as `cowsay`, detected via `argv[0]` (symlink convention). Neo-cowsay ships two separate binaries (`cowsay` and `cowthink` as distinct `cmd/` entries, each a thin wrapper that passes a `Thinking()` option).

**Recommended approach for gosay:** Add `--think` as a long flag to `gosay`. Do not use short `-t` (conflicts with upstream "tired" preset, even if presets are skipped — user confusion risk). Optionally install a `cowthink` binary that is either:
- A second tiny `main.go` that calls `gosay` logic with think mode forced, or
- A symlink to `gosay` with `argv[0]` detection.

The symlink approach has no additional binary to build/release. The second-binary approach is cleaner for cross-platform (Windows has no symlinks in PATH). Given cross-platform is a stated goal, prefer either the flag-only approach or two separate `main.go` wrappers under `cmd/gosay` and `cmd/cowthink`.

---

## Anti-Features

Features to explicitly NOT build. This is the most important section for this project.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| **Library API** (`gosay.Say(...)`) | PROJECT.md explicitly declined it; adds surface area and semver obligations | CLI only; internal packages unexported |
| **COWPATH / runtime `.cow` file loading** | Defeats single-binary distribution; PROJECT.md explicitly declined | All animals embedded via `//go:embed`; no filesystem lookup at runtime |
| **ANSI color / rainbow / aurora output** | PROJECT.md says nothing fancy beyond a gopher default; Neo-cowsay has this if users want it | Keep output plain ASCII; let users pipe to `lolcat` if they want color |
| **Animations / TUI** | Scope creep; this is a toy, not a dashboard | Static output only |
| **Mood presets (`-b/-d/-g/-p/-s/-t/-w/-y`)** | 8 flags for minor eye-glyph substitutions; niche; bloats help output; `-t` ambiguous with `--think` | If `-e`/`-T` are built, users can replicate any preset manually |
| **`-r`/`--random` with true-color `-C` flag** | Color output is explicitly out of scope | `--random` without color is fine; do not add `-C` |
| **gosay-native `.cow` format** | PROJECT.md explicitly declined defining a new format; breaks the existing ecosystem | Vendor upstream `.cow` files as-is; write a small interpreter for `$eyes`, `$tongue`, `$thoughts` |
| **Interactive fuzzy cow selector** (`-f -`) | Neo-cowsay's party trick; requires a terminal UI dependency; conflicts with "near-empty dependency tree" | `-l` to list, `-f name` to pick; that's enough |
| **Plugin system / third-party cow collections** | cowsay-org has pluggable collections; this is a toy, not a platform | Embedded set only; users can fork |
| **Super cows mode** (`--super`) | Easter egg in Neo-cowsay; no defined behavior; pure noise | Not building undocumented Easter eggs |
| **Windows `.cow` symlink support** | `.cow` files themselves just need to be embedded; no symlink issues arise with embed | Non-issue with `//go:embed` approach |
| **Strict Perl `.cow` execution** | Full Perl eval is a security hole and requires Perl at runtime | Parse only the `$eyes`, `$tongue`, `$thoughts` variable substitution that upstream cow files actually use; ignore arbitrary Perl code |
| **Web/HTTP API wrapper** | Not a server; nobody asked | It's a CLI |

---

## Feature Dependencies

```
Positional arg input
Stdin input
  └── both → message normalization (trim trailing newlines, join lines)
        └── Balloon renderer
              ├── Word wrap (default 40 cols)
              ├── -W wrap width
              ├── -n no-wrap
              └── think mode (changes border chars: < > vs ( ))
                    └── --think flag / cowthink binary

.cow file parser (variable substitution: $eyes, $tongue, $thoughts)
  ├── -e eyes flag
  ├── -T tongue flag
  └── -f cowfile flag
        └── -l list (same embedded file set, just list names)
              └── --random (same list, pick one)
```

**Critical dependency:** The `.cow` parser is the keystone. Everything visual depends on it — balloon rendering, animal output, eye/tongue substitution, think-mode `$thoughts` character. Build the parser before the CLI flag surface.

---

## MVP Recommendation

Prioritize (in build order):

1. `.cow` parser — variable substitution for `$eyes`, `$tongue`, `$thoughts`; returns the `$the_cow` body as a string
2. Balloon renderer — single-line and multi-line speech bubbles; correct border chars; word wrap at 40 default; `-W`; `-n`
3. Embedded animal set — `//go:embed` all upstream `.cow` files; gopher as default
4. Core CLI — message from args or stdin; `-f`; `-l`; `-e`; `-T`; `--think`; `--random`; `-W`; `-n`; `-h`
5. Release workflow — cross-platform binaries via GitHub Actions
6. `cowthink` binary (thin wrapper or symlink) — optional stretch goal; `--think` flag covers the use case

Defer/never:
- Color, animation, TUI
- Library API
- Mood presets
- COWPATH / runtime file loading
- Interactive fuzzy finder

---

## Sources

- Upstream cowsay man page: https://man.archlinux.org/man/cowsay.1.en (HIGH confidence)
- cowsay-org/cowsay: https://github.com/cowsay-org/cowsay (HIGH confidence)
- Neo-cowsay v2 API: https://pkg.go.dev/github.com/Code-Hex/Neo-cowsay/v2 (HIGH confidence)
- Cowthink symlink mechanics: https://www.mankier.com/1/cowsay (HIGH confidence)
- cowsay -l output format: https://manpages.opensuse.org/Tumbleweed/cowsay/cowsay.1.en.html (HIGH confidence)
- cowsay Wikipedia: https://en.wikipedia.org/wiki/Cowsay (MEDIUM confidence)
- Neo-cowsay gopher.cow: https://github.com/Code-Hex/Neo-cowsay/blob/master/cows/gopher.cow (MEDIUM confidence)
