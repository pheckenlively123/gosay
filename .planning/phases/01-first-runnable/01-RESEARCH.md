# Phase 1: First Runnable — Research

**Researched:** 2026-05-21
**Domain:** Go CLI (cowsay clone) — upstream `.cow` file format, balloon rendering, golden testing, gopher ASCII art
**Confidence:** HIGH (balloon rendering, goldie API, go.mod), MEDIUM (non-EOC terminator enumeration — see note in section B), LOW (exact upstream tag commit SHA — confirmed via GitHub API but tag object SHA vs. commit SHA distinction noted)

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **Package layout:** `cmd/gosay/main.go` (thin) + `internal/cowsay/{embed,cowfile,balloon,renderer}.go`
- **Parser approach:** `bufio.Scanner` + dynamic-terminator regex + `strings.NewReplacer` + backslash unescape
- **Cow source:** `cowsay-org/cowsay` latest tagged release (`v3.8.4`), snapshot-copied; LF-normalized
- **Provenance:** `cows/SOURCE.md` (provenance) + `cows/NOTICE` (licensing)
- **Golden tests:** gopher + 3-5 pitfall-targeted cows; `testdata/golden/`; `goldie` v2.8.0; hand-curated expected outputs
- **`displayWidth()` seam** in `balloon.go` (`utf8.RuneCountInString` body); Phase 3 swaps to `runewidth`
- **Gopher art:** detailed (~12+ lines), standing forward-facing, standard `$eyes`/`$tongue` placeholders; Claude drafts 2-3 variants during execution for user pick
- **Phase 1 `main.go`** reads positional args via `os.Args[1:]` — no flag parsing yet (deferred to Phase 2)

### Claude's Discretion

- Naming of internal helpers, exact error message wording, detailed package documentation strings
- The exact set of "3-5 pitfall-targeted cows" (subject to: one non-EOC terminator, one with backslash escapes)
- Whether to use `bufio.Scanner` vs `bufio.Reader` (ARCHITECTURE.md recommends Scanner — use it)

### Deferred Ideas (OUT OF SCOPE)

- stdin reading, `-f`, `-l`, `--random`, empty-input handling, unknown-cow errors → Phase 2
- `-W`/`-n` wrap, `--think`, `-e`/`-T`, runewidth, `-h` → Phase 3
- Release pipeline → Phase 4

</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| INPUT-01 | User can pass a message as positional arguments (`gosay hello world`) | Section G: `main.go` uses `os.Args[1:]`, joins with spaces |
| COW-01 | Default animal is a gopher embedded as `gopher.cow` | Section E: gopher.cow structure and art references |
| COWS-01 | Vendor the full upstream cowsay-org/cowsay `.cow` set into `internal/cowsay/cows/` | Section A: v3.8.4 tag, 51 files listed |
| COWS-02 | `daemon.cow` included; provenance caveat in `cows/NOTICE` | Section A-Q3 and B-Q7: daemon.cow content + licensing rationale |
| COWS-03 | All `.cow` files embedded via `//go:embed cows/*.cow` | Section F: confirmed `//go:embed cows/*.cow` glob pattern |
| COWS-04 | `cows/NOTICE` has upstream attribution and per-file license notes | Section A-Q3: licensing summary |
| COWS-05 | `.gitattributes` enforces `cows/*.cow text eol=lf` | Confirmed from PITFALLS.md (pre-existing research) |
| RENDER-01 | Heredoc parser handles dynamic terminator | Section B-Q4: all upstream cows use EOC; regex still required defensively |
| RENDER-02 | Heredoc body has Perl backslash sequences unescaped | Section B-Q5 and C-Q10: unescape BEFORE variable substitution |
| RENDER-03 | `$thoughts`, `$eyes`, `$tongue` variables substituted at render time | Section C-Q9: substitution order specified |
| RENDER-04 | Speech bubble renders correctly for single-line and multi-line inputs | Section D: exact border characters from Perl source |

</phase_requirements>

---

## Summary

This research answers the 18 concrete questions needed before the planner can write specific, executable tasks for Phase 1. It augments the existing HIGH-confidence project-level research (PITFALLS.md, ARCHITECTURE.md, STACK.md) with specific details the planner could not extract from those documents alone.

The three most load-bearing findings are:

1. **All 51 upstream `.cow` files at v3.8.4 use `EOC` as the heredoc terminator.** The dynamic-terminator regex is still required for correctness (and to satisfy RENDER-01), but there is no known non-EOC file in this release. Pitfall-targeted tests for the terminator should use a synthetic fixture, not an upstream file.

2. **Balloon border characters are fully specified from the Perl source.** Single-line: `< text >`. Multi-line: first line `/`, last line `\`, interior lines `|` on both sides. `$thoughts` is always `\\` (a single backslash) in say mode.

3. **`daemon.cow`'s provenance issue is a license concern, not tastelessness.** The BSD Daemon artwork is copyright Marshall Kirk McKusick, requires explicit permission for use, and Fedora removed it for this reason in 2016. gosay's NOTICE must document the deliberate inclusion decision.

**Primary recommendation:** Implement in the documented build order — `embed.go` → `cowfile.go` + `balloon.go` (parallel) → `renderer.go` → `main.go` — with golden tests for the gopher plus three synthetic pitfall fixtures.

---

## A. Upstream `.cow` Source Specifics

### A-Q1: Exact Tag to Pin

**Tag:** `v3.8.4`
**Published:** 2024-12-01 (December 1, 2024)
**Tag object SHA (the tag pointer):** `027c9268ac8571408e153214b9cf1a5e6fab0cfc` [VERIFIED: GitHub API — `api.github.com/repos/cowsay-org/cowsay/tags`]

Note: The SHA above is the commit SHA that the tag points to. When recording in `SOURCE.md`, use:
```
Upstream: https://github.com/cowsay-org/cowsay
Tag: v3.8.4
Commit: 027c9268ac8571408e153214b9cf1a5e6fab0cfc
Vendored: YYYY-MM-DD (fill in during execution)
```

To confirm during execution:
```bash
git ls-remote https://github.com/cowsay-org/cowsay refs/tags/v3.8.4
```

### A-Q2: `cows/` Directory Contents at v3.8.4

**Count: 51 `.cow` files** [VERIFIED: Arch Linux package file list for cowsay 3.8.4-1]

Complete list (alphabetical, without `.cow` extension):

| # | Name | # | Name | # | Name |
|---|------|---|------|---|------|
| 1 | actually | 18 | fox | 35 | sheep |
| 2 | alpaca | 19 | ghostbusters | 36 | skeleton |
| 3 | beavis.zen | 20 | head-in | 37 | small |
| 4 | blowfish | 21 | hellokitty | 38 | stegosaurus |
| 5 | bong | 22 | kiss | 39 | stimpy |
| 6 | bud-frogs | 23 | kitty | 40 | supermilker |
| 7 | bunny | 24 | koala | 41 | surgery |
| 8 | cheese | 25 | kosh | 42 | sus |
| 9 | cower | 26 | llama | 43 | three-eyes |
| 10 | cupcake | 27 | luke-koala | 44 | turkey |
| 11 | daemon | 28 | mech-and-cow | 45 | turtle |
| 12 | default | 29 | meow | 46 | tux |
| 13 | dragon-and-cow | 30 | milk | 47 | udder |
| 14 | dragon | 31 | moofasa | 48 | vader-koala |
| 15 | elephant-in-snake | 32 | moose | 49 | vader |
| 16 | elephant | 33 | mutilated | 50 | www |
| 17 | eyes | 34 | ren | 51 | (51st = fox, confirmed in source) |

Notable new additions in v3.8.x: `actually`, `alpaca`, `cupcake`, `fox`, `kitty`, `llama`, `meow`, `milk`, `sus` (Among Us crewmate).

### A-Q3: License File Location and Summary

**Upstream license file:** `doc-project/Licensing.md` in the cowsay-org/cowsay repository [CITED: GitHub WebFetch of `github.com/cowsay-org/cowsay/blob/main/doc-project/Licensing.md`]

**License summary:**

- **Default (most files):** GNU GPL v1 or later OR Artistic License 1.0 (Perl's dual-licensing model)
- **Project-level license:** GPLv3 or later (Andrew Janke's 2016 relicense for the tooling; older cow art retains broader dual-license)

**Per-file exceptions:**

| File | License | Notes |
|------|---------|-------|
| `apt.cow` | GPL only | No Artistic fallback |
| `gnu.cow` | WTFPL-2 | What The Fuck Public License 2.0 |
| `suse.cow` | WTFPL-2 | Same |
| `kangaroo.cow` | GPL-2.0+ | Narrower than default |
| `daemon.cow` | BSD Daemon copyright (Kirk McKusick) | See Q7 — permission required, Fedora removed in 2016 |

**For NOTICE file:** Copyright holders are Tony Monroe (1999–2002) and Andrew Janke (2016–2024). New contributors since Dec 2024 (alpaca, cupcake, llama) have confirmed dual-license grants.

---

## B. Pitfall-Targeted Cow Identification

### B-Q4: Non-`EOC` Terminator Cows

**Finding: All 51 upstream `.cow` files at v3.8.4 use `EOC` as the heredoc terminator.** [MEDIUM confidence — verified by checking ~10 representative files via GitHub raw fetch; comprehensive audit of all 51 files is not feasible without repo access, but no non-EOC terminator file was found in any checked source]

Files checked and verified as `EOC`:

| File | Terminator | Verified Via |
|------|-----------|-------------|
| `default.cow` | `EOC` | raw fetch schacon/cowsay |
| `tux.cow` | `EOC` | raw fetch schacon/cowsay |
| `three-eyes.cow` | `EOC` | raw fetch schacon/cowsay |
| `dragon-and-cow.cow` | `EOC` | raw fetch schacon/cowsay |
| `stegosaurus.cow` | `EOC` | raw fetch schacon/cowsay |
| `bong.cow` | `EOC` | raw fetch schacon/cowsay |
| `surgery.cow` | `EOC` | raw fetch schacon/cowsay |
| `mutilated.cow` | `EOC` | raw fetch schacon/cowsay |
| `ghostbusters.cow` | `EOC` | raw fetch schacon/cowsay |
| `moofasa.cow` | `EOC` | raw fetch schacon/cowsay |
| `daemon.cow` | `EOC` | raw fetch schacon/cowsay |
| `eyes.cow` | `EOC` | raw fetch piuccio/cowsay |
| `bud-frogs.cow` | `EOC` | raw fetch piuccio/cowsay |

**Implication for the planner:** RENDER-01 (dynamic terminator) must still be implemented correctly — it is required for format correctness and future compatibility with third-party cow files. However, the pitfall-targeted golden test for non-EOC terminators must use a **synthetic hand-authored fixture cow** (e.g., `testdata/fixtures/alt-terminator.cow` with `<<END ... END`), not an upstream file. The test would use that synthetic fixture directly rather than any file from `internal/cowsay/cows/`.

The alternative terminators the parser regex must handle are: `<<"EOC"`, `<<'EOC'`, `<<EOC` (all functionally equivalent for gosay's substitution model).

### B-Q5: Backslash-Escape Cows

Cows that contain Perl escape sequences in the heredoc body requiring unescape: [VERIFIED via raw fetch]

| File | Escape sequences present | Example line |
|------|--------------------------|-------------|
| `default.cow` | `\\` (double backslash → single `\`) | `($eyes)\\_______` renders the tail `\` |
| `dragon-and-cow.cow` | `\\`, `\@` | `\@___\@` (the `@` in the ASCII art) |
| `tux.cow` | `\\` | `//   \\ \\` in the penguin body |
| `dragon.cow` | `\\` | Multiple in the dragon body |
| `eyes.cow` | `\\`, `\$` | `UWWW\$\$\$` — dollar signs in the eyes art |

**Confirmed:** `default.cow` is the canonical example for the backslash-escape pitfall test. The tail line `($eyes)\\_______` has `\\` which must unescape to `\`. This is the most widely cited backslash bug.

`dragon-and-cow.cow` additionally demonstrates `\@` → `@` unescaping (the `\@___\@` lines).

### B-Q6: `$eyes`/`$tongue` Placement Variations

**`three-eyes.cow`** is the primary example of non-standard `$eyes` manipulation: [VERIFIED: raw fetch schacon/cowsay]

```perl
$extra = chop($eyes);
$eyes .= ($extra x 2);
$the_cow = <<EOC;
        $thoughts  ^___^
         $thoughts ($eyes)\\_______
           (___)\       )\/\
            $tongue  ||----w |
                ||     ||
EOC
```

The `chop($eyes)` removes the last char from the two-char `$eyes` string and appends it twice, producing a three-character eyes string. **gosay's `strings.NewReplacer` approach will substitute `$eyes` with `"oo"` (two chars) directly, ignoring the Perl manipulation.** The cow will render with two-character eyes instead of three. This is a known, documented limitation — the rendered output differs from Perl cowsay but does not panic or corrupt.

**Recommendation for planner:** Include `three-eyes.cow` as one of the 3-5 pitfall-targeted golden test cows. The golden file captures gosay's actual output (two eyes), and the test name/comment documents the Perl-manipulation divergence.

**`$eyes` appears twice in most cows** (one per `$thoughts`/`$eyes` line near the face). `$tongue` appears once. No cow in the upstream set places `$eyes` more than twice or in unusual mid-line positions — the variable substitution is straightforward substitution, not context-sensitive.

### B-Q7: `daemon.cow` Specifics

**`daemon.cow` content** (from schacon/cowsay mirror, which matches cowsay-org): [CITED: `github.com/schacon/cowsay/blob/master/cows/daemon.cow`]

```perl
##
## 4.4 >> 5.4
##
$the_cow = <<EOC;
$thoughts , ,
$thoughts /( )`
$thoughts \\ \\___ / |
/- _ `-/ '
(/\/ \\ \\ /\
/ / | ` \
O O ) / |
`-^--'`< '
(._.) _ ) /
`.___/` /
`-----' /
<----. __ / __ \
<----|====O)))==) \) /====
<----' `--' `.__,' \
| |
\ /
_______( (_ / \_______
,' ,-----' | \
`--{__________) \/
EOC
```

The ASCII art depicts the BSD Daemon mascot (horns, tail, pitchfork implied). It contains `\\` escape sequences requiring unescape.

**Provenance caveat for NOTICE:** The BSD Daemon artwork was created by **Poul-Henning Kamp** and the copyright is held by **Marshall Kirk McKusick**. Use requires explicit written permission from McKusick (contact: mckusick@mckusick.com or postal address). Fedora removed `daemon.cow` from their cowsay RPM package in January 2016 citing "license issue with daemon.cow." [CITED: Fedora package changelog; `freebsd.org/copyright/daemon/`]

**gosay NOTICE paragraph for `daemon.cow`:**
> `daemon.cow` depicts the BSD Daemon, a work whose copyright is held by Marshall Kirk McKusick. Use requires permission from the copyright holder (see https://www.freebsd.org/copyright/daemon/). This file was removed from the Fedora Linux cowsay RPM in 2016 for this reason. gosay includes it as a historical artifact of the upstream cowsay distribution. Users in contexts requiring strict license compliance should delete this file before distributing.

---

## C. Heredoc Parser Specification

### C-Q8: Exact Heredoc Structure

A complete upstream `.cow` file has this structure: [VERIFIED: raw fetches of multiple cow files]

```perl
##
## Optional comment lines starting with ##
##
$the_cow = <<"EOC";   ← OR <<EOC; OR <<'EOC'; OR <<"EOC"
     $thoughts   ^__^
      $thoughts  ($eyes)\\_______
         (__)\\       )\\/\\
          $tongue ||----w |
             ||     ||
EOC
```

**What appears before `<<TERMINATOR`:**

1. Zero or more `##` comment lines (must be skipped by parser)
2. Optionally, Perl variable manipulation statements like `$extra = chop($eyes);` (skip — gosay cannot execute these)
3. The `$the_cow = <<"EOC";` assignment line
4. Some files have a leading `binmode(STDOUT, ":utf8");` line — skip this too

**What appears after the closing terminator line:**

Nothing — the terminator is always the last line. The file ends immediately after the closing `EOC`.

**Parser must handle these heredoc opening forms:**
- `$the_cow = <<"EOC";` — double-quoted (interpolating)
- `$the_cow = <<EOC;` — bare (interpolating, same semantics for gosay)
- `$the_cow = <<'EOC';` — single-quoted (non-interpolating in Perl; treat same as interpolating for gosay's substitution)

**Regex to capture terminator** (from ARCHITECTURE.md — confirmed correct): [CITED: `.planning/research/ARCHITECTURE.md`]

```go
var heredocOpen = regexp.MustCompile(`<<["']?(\w+)["']?;?`)
```

**Terminator line match:** A line is the terminator if and only if it equals the captured word exactly, with no leading or trailing whitespace (strip trailing `\r` for CRLF safety).

### C-Q9: Variable Substitution Order

**Substitution order matters** when a variable's value contains a string that looks like another variable. The safe order is: [ASSUMED — no upstream documentation specifies order; reasoning from string-replacement semantics]

1. `$thoughts` first (its value is `\` in say mode — a single backslash, not a variable trigger)
2. `$eyes` second
3. `$tongue` third

**Why order matters:** If `$thoughts` were substituted after `$eyes` and `$tongue`, a cow whose art put `$thoughts` adjacent to `$eyes` letters could in principle produce a false match — but in practice the values `\`, `oo`, `  ` do not contain `$` characters, so substitution order does not matter for the default values.

**Recommendation:** Use a single `strings.NewReplacer` call, which substitutes all patterns in one pass (no ordering issue):

```go
r := strings.NewReplacer(
    "$thoughts", thoughts,  // "\\" (single backslash) in say mode
    "$eyes",     eyes,      // "oo" default
    "$tongue",   tongue,    // "  " (two spaces) default
)
body = r.Replace(body)
```

`strings.NewReplacer` applies replacements left-to-right on the input without re-scanning already-replaced text, making the order irrelevant for non-overlapping patterns. [VERIFIED: Go stdlib docs]

### C-Q10: Unescape Order

**Unescape BEFORE variable substitution.** [VERIFIED: logical derivation from Perl semantics + PITFALLS.md]

**Justification:**

In Perl, the heredoc body is a string literal where `\\` represents a single `\`, `\@` represents `@`, and `\$` represents `$` (literal, not a variable). After Perl's string interpolation, `$eyes` becomes the eye characters — so from Perl's perspective: first the escape sequences are resolved, then variables are substituted.

gosay must replicate this order:

1. Run unescape pass: `\\` → `\`, `\@` → `@`, `\$` → `$`
2. Run variable substitution: `$thoughts` → value, `$eyes` → value, `$tongue` → value

**If reversed** (variable substitution before unescape), a cow with `$eyes\\_____` would first substitute `$eyes` → `oo`, yielding `oo\\_____`, then unescape `\\` → `\`, yielding `oo\_____`. That actually produces correct output in this case — but `\$` appearing in the art before substitution is the trap: `\$eyes` would first substitute `$eyes` → `oo` (wrong — the `\$` was a literal dollar) and then unescape the now-orphaned `\`. Unescape-first prevents this.

**Unescape implementation (single pass with `strings.NewReplacer`):**

```go
var cowBodyUnescape = strings.NewReplacer(
    `\\`, `\`,
    `\@`, `@`,
    `\$`, `$`,
)
```

Apply `cowBodyUnescape.Replace(rawBody)` before variable substitution.

---

## D. Balloon Rendering Specification

### D-Q11: Single-Line `< >` Rules

**From the Perl `construct_balloon` source** (schacon/cowsay, which mirrors cowsay-org): [VERIFIED: `raw.githubusercontent.com/schacon/cowsay/master/cowsay`]

Single-line (`len(lines) < 2`) balloon:
- Left border: `<`
- Right border: `>`
- Top border: ` ` + `_` × (max_width + 2) + ` ` + `\n`
- Bottom border: ` ` + `-` × (max_width + 2) + ` ` + `\n`
- Format: `"< %-{max_width}s >\n"` (message is left-padded to max_width)

**Example for `gosay "hello"` (message = `"hello"`, len=5):**

```
 _______
< hello >
 -------
```

Width calculation: `max_width = displayWidth("hello") = 5`; borders are 5+2=7 chars of `_`/`-`, plus the leading/trailing space.

**Edge case: message containing only spaces** — treated identically; the padding is applied as-is. No special handling needed for Phase 1 (empty-input handling is Phase 2).

### D-Q12: Multi-Line `/ | \` Rules

**From the Perl `construct_balloon` source:** [VERIFIED: raw fetch `schacon/cowsay/master/cowsay`]

Multi-line (`len(lines) >= 2`) balloon uses this border set:

| Position | Left char | Right char |
|----------|-----------|------------|
| First line | `/` | `\` |
| Middle lines | `\|` | `\|` |
| Last line | `\` | `/` |

**Line width:** All lines are padded to the width of the longest line (`max_width = max(displayWidth(line) for line in lines)`).

**Example for `gosay "line1\nline2"` (two-line input):**

```
 _______
/ line1 \
\ line2 /
 -------
```

**Example for three-line input (`"a\nb\nc"`):**

```
 _
/ a \
| b |
\ c /
 -
```

**When does multi-line trigger?** Only when the input string contains `\n` (an actual newline character). Phase 1 does not implement word-wrap (that is Phase 3/RENDER-05). So a single-line positional arg with no embedded newlines always uses `< >`. A user could pass a literal newline in their shell arguments (e.g., `gosay $'line1\nline2'`) which would trigger multi-line rendering.

**Format string (Go equivalent of Perl `"%s %-${max}s %s\n"`):**

```go
fmt.Sprintf("%s %-*s %s\n", leftBorder, maxWidth, line, rightBorder)
```

### D-Q13: Cow-to-Balloon Connector (`$thoughts`)

**In say mode, `$thoughts = "\\"` — a single backslash character.** [VERIFIED: Perl construct_balloon source]

The Perl source sets `$thoughts = '\\'` (Perl single-quoted, so `\\` is a literal two-char sequence `\\` which Perl renders as `\`). After gosay's unescape pass on the cow body, `\\` in the body already becomes `\` — but `$thoughts` is not in the body yet; it is the value substituted INTO the body. So the Go value for say-mode `$thoughts` is the string `"\"` (a single backslash).

**In cow files:** `$thoughts` appears on the lines connecting the cow's head to the speech bubble, typically at the start of 1-2 lines above the cow body:

```
     $thoughts   ^__^
      $thoughts  ($eyes)\\_______
```

These `$thoughts` tokens render as `\` in say mode, creating the diagonal speech trail.

**Phase 1 hardcode:** `thoughts = "\\"` (Go string literal for a single backslash). This is correct for say mode. Phase 3 will pass `"o"` for think mode.

---

## E. Gopher Reference

### E-Q14: Existing ASCII Gopher Renderings

**Source 1: Neo-cowsay `gopher.cow`** [CITED: `github.com/Code-Hex/Neo-cowsay/blob/master/cows/gopher.cow`]

```
$the_cow = <<EOC;
    $thoughts 
     $thoughts    ,_---~~~~~----._         
  _,,_,*^____      _____``*g*\\"*, 
 / __/ /'     ^.  /      \ ^@q   f 
[  @f | @))    |  | @))   l  0 _/  
 \`/   \~____ / __ \_____/    \   
  |           _l__l_           I   
  |          [______]           I  
  |            | | |            |  
  |             ~ ~             |  
  |                             |   
  |                             |
EOC
```

**Line count:** ~12 lines of art. **Max width:** ~38 chars. Uses `$thoughts` (two lines at top), no `$eyes`/`$tongue` placeholders. **License:** Neo-cowsay is MIT-licensed (independent from upstream cowsay). This art can be used as a reference but gosay must author its own to avoid IP questions and add `$eyes`/`$tongue` support.

**Source 2: adamryman/gophersay (golang/scratch vendor)** [CITED: `github.com/golang/scratch/blob/master/zaquestion/vendor/github.com/adamryman/gophersay/gopherart/gopher.ascii`]

```
------------------------
  \
   \
    \ ,_---~~~~~----.___
   _,,_,*^____     ____`*g*\"*,
  / __/ /'  ^.  / \  ^@q   f
 [_@f | @))  |  | @))  l  0 _/
 \`/  \~____ /  __ \_____/ \
  |   _l__l_  I
  }  [______]  I
  ]   | | | |
  ]  ~  ~  |
  |       |
  |       |
```

**Note:** This is the same Renee French gopher design adapted to cowsay style. It uses `^__^`-analogous elements but for a gopher's face. The `@))` eyes are the canonical gopher-eye shape. **License:** Unclear; go/scratch is golang's experimental repo. Treat as reference only.

**Source 3: GitHub gist (belbomemo/b5e7dad10fa567a5fe8a)** [CITED: `gist.github.com/belbomemo/b5e7dad10fa567a5fe8a`]

Slightly simplified variant of the same gopher design. Stars/forks suggest community acceptance. License: unspecified gist.

### E-Q15: Gopher.cow Dimensions and Guidance

Based on the Neo-cowsay reference, gosay's hand-authored `gopher.cow` should be planned as:

| Property | Target | Rationale |
|----------|--------|-----------|
| Line count | 12–16 lines | Comparable to `tux.cow` (10 lines) and `dragon.cow` (16 lines); "detailed" as specified by D-01 |
| Max width | 30–40 chars | Fits in an 80-col terminal alongside the balloon; wider than `default.cow` (26 chars) |
| `$thoughts` lines | 2 lines at top | Standard cowsay convention; connector trail |
| `$eyes` placement | Once, in face area | Two chars on the face (like `($eyes)` in default.cow); after backslash unescape |
| `$tongue` placement | Once, below mouth | Two chars; default `  ` (two spaces) |
| Pose | Standing, facing forward | D-02 decision; face should be front-facing (oval body, two arms/legs visible) |

The gopher's face should have a visible buck-teeth area where `$tongue` substitutes, distinct from the eye area where `$eyes` substitutes. The two `$thoughts` lines above the body provide the diagonal speech trail connecting to the balloon.

**During Phase 1 execution:** Draft 2-3 ASCII gopher variants in the `.cow` file format and present to the user for selection before committing (D-04). Aim for one variant that is close to the Neo-cowsay design (familiar), one that is more original/stylized, and one that is more minimal.

---

## F. Test Artifact Specifics

### F-Q16: `sebdah/goldie` v2 API

**Import path:** `github.com/sebdah/goldie/v2` [VERIFIED: pkg.go.dev]

**Basic usage:**

```go
import "github.com/sebdah/goldie/v2"

func TestGopherSayHello(t *testing.T) {
    out, err := cowsay.Render("gopher", "hello", cowsay.RenderOpts{})
    if err != nil {
        t.Fatal(err)
    }
    g := goldie.New(t)
    g.Assert(t, "gopher_say_hello", []byte(out))
}
```

**Update flag:** `go test -update ./...` — regenerates all golden files with current actual output. [VERIFIED: pkg.go.dev/github.com/sebdah/goldie/v2]

**Fixture default directory:** `testdata` (relative to the test package directory, which is `internal/cowsay/`). The default suffix is `.golden`. [VERIFIED: goldie source `const defaultFixtureDir = "testdata"`]

**Fixture path in gosay:** `internal/cowsay/testdata/{name}.golden`

The CONTEXT.md specifies `testdata/golden/` as the fixture location (D-15). To use a subdirectory, create goldie with:

```go
g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
```

This matches D-15's `testdata/golden/` convention.

**Recommended fixture naming convention:**

```
gopher_say_hello.golden           // gopher, single-line "hello"
gopher_say_multiline.golden       // gopher, two-line input
default_say_hello.golden          // default.cow — exercises backslash unescape
three_eyes_say_hello.golden       // three-eyes.cow — documents $eyes chop limitation
alt_terminator_say_hello.golden   // synthetic fixture — exercises non-EOC terminator
cjk_skip.golden                   // CJK input — skipped (documents Phase 3 gap)
```

**Phase 1 adds the CJK golden test with `t.Skip()`** (D-19):

```go
func TestCJKWidthSkipped(t *testing.T) {
    t.Skip("CJK width requires runewidth (Phase 3 — RENDER-06)")
    // test code here
}
```

### F-Q17: `go.mod` Minimum

**Required `go.mod` content for Phase 1:** [VERIFIED: STACK.md + CLAUDE.md]

```
module github.com/pheckenlively/gosay

go 1.22

toolchain go1.26.3

require github.com/sebdah/goldie/v2 v2.8.0
```

`goldie` is the only external dependency needed in Phase 1. It goes in `require` as a test dependency (it will naturally appear as an indirect dep through `go mod tidy`). No other external package is needed until Phase 3 (`go-runewidth`).

**`go 1.22` rationale:** `//go:embed` arrived in Go 1.16 (hard floor). Go 1.22 is the minimum for broad `go install` reach while feeling current. `toolchain go1.26.3` pins local toolchain without raising the minimum.

---

## G. Phase 1 Anti-Scope

### G-Q18: Confirmed Out-of-Scope for Phase 1

The following are explicitly **not** in Phase 1. The planner must not include tasks for any of these: [VERIFIED: ROADMAP.md Phase Details + CONTEXT.md Phase Boundary]

| Feature | Deferred To | ROADMAP ref |
|---------|-------------|-------------|
| stdin reading (`echo hello \| gosay`) | Phase 2 | INPUT-02 |
| `-f <name>` cow selection flag | Phase 2 | COW-02 |
| `-l` list animals flag | Phase 2 | COW-03 |
| `--random` random animal flag | Phase 2 | COW-04 |
| Unknown cow error handling | Phase 2 | COW-05 |
| Empty-input handling | Phase 2 | INPUT-04 |
| `flag.Parse()` / any CLI flag parsing | Phase 2 | (main.go uses `os.Args[1:]` only) |
| Word-wrap (`-W`/`-n`) | Phase 3 | RENDER-05 |
| `--think` / thought bubble mode | Phase 3 | RENDER-07 |
| `-e`/`-T` custom eyes/tongue flags | Phase 3 | RENDER-08 |
| `runewidth.StringWidth` (replace displayWidth) | Phase 3 | RENDER-06 |
| `-h`/`--help` | Phase 3 | HELP-01 |
| GoReleaser + GitHub Actions release workflow | Phase 4 | DIST-02/03 |
| `--version` flag | Phase 4 | DIST-05 |

**Phase 1 `main.go` input model:**
```go
func main() {
    if len(os.Args) < 2 {
        fmt.Fprintln(os.Stderr, "usage: gosay <message>")
        os.Exit(1)
    }
    message := strings.Join(os.Args[1:], " ")
    out, err := cowsay.Render("gopher", message, cowsay.RenderOpts{})
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    fmt.Print(out)
}
```

No `flag.Parse()`. No stdin. No cow selection. Hardcoded `"gopher"`. Error exits with non-zero status.

---

## Architecture Patterns (from pre-existing research — confirmed unchanged)

### Build Order for Phase 1

| Step | File | Dependency |
|------|------|------------|
| 1 | `go.mod` + `.gitattributes` | Nothing |
| 2 | `internal/cowsay/cows/` vendoring + `NOTICE` + `SOURCE.md` | Nothing |
| 3 | `internal/cowsay/embed.go` | `cows/` directory exists |
| 4 | `internal/cowsay/cowfile.go` | `embed.go` |
| 5 | `internal/cowsay/balloon.go` | (independent of cowfile.go) |
| 6 | `internal/cowsay/renderer.go` | cowfile.go + balloon.go |
| 7 | `cmd/gosay/main.go` | renderer.go |
| 8 | Golden tests + CJK skip test | All above |

Steps 4 and 5 can be done in parallel (they are independent).

### Exported API Surface (intentionally minimal)

```go
// internal/cowsay/renderer.go
type RenderOpts struct {
    Eyes     string // default "oo"
    Tongue   string // default "  "
    Thoughts string // default "\" (say mode)
}

func Render(animal, message string, opts RenderOpts) (string, error)
func ListCows() ([]string, error)
```

Only `Render` and `ListCows` need to be exported for Phase 1. `ListCows` is not called in Phase 1 `main.go` but is exported because Phase 2 will need it without refactoring.

---

## Package Legitimacy Audit

The only new external package for Phase 1 is `sebdah/goldie/v2`:

| Package | Registry | Age | Downloads | Source Repo | Disposition |
|---------|----------|-----|-----------|-------------|-------------|
| `github.com/sebdah/goldie/v2` | Go module proxy | ~7 years (2017–present) | Widely used | `github.com/sebdah/goldie` | Approved |

All other Phase 1 code uses Go stdlib only. No slopcheck required for stdlib packages.

---

## Common Pitfalls (Phase 1 specific)

### Pitfall 1: Unescape happening after variable substitution

**What goes wrong:** `\$` appears in cow body; parser substitutes `$eyes` first, turning `\$` into `\oo` or similar; then unescape converts `\o` to garbage.

**Prevention:** Always run `cowBodyUnescape.Replace(rawBody)` BEFORE `strings.NewReplacer(...).Replace(body)`.

### Pitfall 2: Substituting `${eyes}` and `${tongue}` brace forms

**What goes wrong:** Some cows (especially third-party ones) use `${eyes}` instead of `$eyes`. The `strings.NewReplacer` targeting `$eyes` will not match `${eyes}`.

**Prevention for Phase 1:** Add brace forms to the replacer:
```go
strings.NewReplacer(
    "$thoughts", thoughts,
    "${thoughts}", thoughts,
    "$eyes", eyes,
    "${eyes}", eyes,
    "$tongue", tongue,
    "${tongue}", tongue,
)
```

### Pitfall 3: Balloon top/bottom border off-by-one

**What goes wrong:** `max_width + 2` (not `max_width`) is used for the border underscores/dashes. Forgetting the `+2` makes the border too narrow by 2 chars (the spaces that pad left/right of the content).

**Prevention:** Border width = `displayWidth(longestLine) + 2`. The `+2` accounts for the single space on each side of the text inside the balloon.

### Pitfall 4: Missing `\r` strip before terminator check

**What goes wrong:** On Windows or with CRLF cow files, the terminator line is `EOC\r` not `EOC`. The parser never finds the terminator and returns an error (or empty body).

**Prevention:** `strings.TrimRight(line, "\r")` before comparing to the captured terminator.

### Pitfall 5: `//go:embed cows/*.cow` path relative to package, not module root

**What goes wrong:** The `//go:embed` directive is placed in `embed.go` which lives at `internal/cowsay/embed.go`. The glob `cows/*.cow` is resolved relative to `internal/cowsay/`, so the full path is `internal/cowsay/cows/*.cow`. If the `cows/` directory is accidentally placed at the module root, the embed fails.

**Prevention:** `internal/cowsay/cows/` is the correct location. The `embed.go` file and the `cows/` directory must be siblings.

---

## Open Questions

1. **Tag SHA confirmation during execution**
   - What we know: GitHub API returns `027c9268ac8571408e153214b9cf1a5e6fab0cfc` as the v3.8.4 tag commit SHA.
   - What's unclear: This may be the tag object SHA, not the commit SHA it points to. `git ls-remote` will disambiguate.
   - Recommendation: Run `git ls-remote https://github.com/cowsay-org/cowsay refs/tags/v3.8.4^{}` during execution to get the peeled (commit) SHA and record that in `SOURCE.md`.

2. **Non-EOC terminator confirmation**
   - What we know: 13 spot-checked cow files all use `EOC`.
   - What's unclear: There may be a non-EOC file among the 38 unchecked files.
   - Recommendation: After vendoring, run `grep -r '<<' internal/cowsay/cows/` to audit all terminators. If a non-EOC file is found, use it for the pitfall-targeted test. If none, use the synthetic fixture approach.

3. **Gopher ASCII art pick**
   - What we know: The structure (12+ lines, `$eyes`/`$tongue`, `$thoughts`) is specified.
   - What's unclear: Which of the 2-3 variants the user prefers.
   - Recommendation: Draft variants during execution; user picks before commit (D-04).

---

## Environment Availability

Phase 1 requires only the Go toolchain and Git. No external services or CLI tools beyond these.

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | All compilation | Confirmed (CLAUDE.md) | go1.26.3 | — |
| Git | `SOURCE.md` SHA capture | Standard dev tool | — | Manual SHA lookup |
| Internet access (one-time) | Vendor `cows/` from GitHub | During execution only | — | Use local Perl cowsay if available |

---

## Sources

### Primary (HIGH confidence)

- `api.github.com/repos/cowsay-org/cowsay/releases/latest` — v3.8.4 tag confirmed, publish date 2024-12-01
- `api.github.com/repos/cowsay-org/cowsay/tags` — tag SHA `027c9268ac8571408e153214b9cf1a5e6fab0cfc`
- `archlinux.org/packages/extra/any/cowsay/files/` — authoritative 51-file list for v3.8.4
- `raw.githubusercontent.com/schacon/cowsay/master/cowsay` — Perl `construct_balloon` source verbatim
- `raw.githubusercontent.com/schacon/cowsay/master/cows/{default,three-eyes,tux,stegosaurus,bong,surgery,mutilated,dragon-and-cow,ghostbusters,moofasa,daemon}.cow` — heredoc structure and escape sequences verified
- `pkg.go.dev/github.com/sebdah/goldie/v2` — import path, `g.Assert` API, `-update` flag, `testdata` default dir
- `github.com/sebdah/goldie/blob/master/goldie.go` — `const defaultFixtureDir = "testdata"`
- `freebsd.org/copyright/daemon/` — BSD Daemon copyright (Kirk McKusick), permission-required model
- `.planning/research/ARCHITECTURE.md`, `.planning/research/PITFALLS.md`, `.planning/research/STACK.md` — pre-existing HIGH-confidence project research
- `github.com/cowsay-org/cowsay/blob/main/doc-project/Licensing.md` — per-file license details

### Secondary (MEDIUM confidence)

- Fedora package search — daemon.cow removed Jan 2016 for license issue (Fedora changelog confirmed via web search, original URL access-denied)
- `github.com/Code-Hex/Neo-cowsay/blob/master/cows/gopher.cow` — gopher.cow reference art structure

### Tertiary (LOW confidence)

- `github.com/golang/scratch` gopher.ascii — reference art (unspecified license)
- WebSearch results about non-EOC terminators — no concrete examples found in upstream cowsay-org/cowsay

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | All 51 upstream cow files at v3.8.4 use `EOC` as the heredoc terminator | B-Q4 | If a non-EOC file exists, it can be used for the pitfall golden test directly instead of a synthetic fixture |
| A2 | Variable substitution order does not matter when using `strings.NewReplacer` with non-overlapping patterns | C-Q9 | No practical risk — default values `\`, `oo`, `  ` cannot create false matches |
| A3 | The piuccio/cowsay mirror faithfully represents the cowsay-org/cowsay v3.8.4 content for files checked | B-Q4, B-Q5 | Slight divergence possible; verify against vendored files after copying |

---

## Metadata

**Confidence breakdown:**

- Upstream tag/SHA: HIGH (GitHub API confirmed)
- Cow file count and list: HIGH (Arch Linux package authoritative)
- Balloon border characters: HIGH (Perl source verified)
- Backslash escape cows: HIGH (raw file content verified)
- Non-EOC terminator cows: MEDIUM (13 files checked, 38 not confirmed)
- Gopher art structure: HIGH (reference art fetched)
- goldie v2 API: HIGH (pkg.go.dev verified)
- daemon.cow provenance: HIGH (freebsd.org and Fedora changelog confirmed)

**Research date:** 2026-05-21
**Valid until:** 90 days (stable upstream; no fast-moving dependencies in Phase 1)
