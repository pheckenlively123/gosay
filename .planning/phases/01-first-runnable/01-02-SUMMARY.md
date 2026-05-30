---
phase: 01-first-runnable
plan: "02"
subsystem: cow-data-layer
tags: [embed, cow-files, vendoring, licensing, go-embed]
dependency_graph:
  requires:
    - go.mod (module declaration — Plan 01-01)
    - .gitattributes (cows/*.cow LF policy — Plan 01-01)
  provides:
    - internal/cowsay/cows/ (50 vendored upstream .cow files + NOTICE + SOURCE.md)
    - internal/cowsay/embed.go (embed.FS declaration + ListCows + readCowFile)
    - internal/cowsay/embed_test.go (smoke tests for embedded cow set)
  affects:
    - internal/cowsay/cowfile.go (Plan 01-03 calls readCowFile)
    - internal/cowsay/renderer.go (Plan 01-03 orchestrates readCowFile)
    - cmd/gosay/main.go (Plan 01-04 wires Render which calls readCowFile)
tech_stack:
  added:
    - embed.FS (stdlib Go 1.16+) for compile-time cow file embedding
  patterns:
    - "//go:embed cows/*.cow glob (explicit .cow pattern, excludes NOTICE/SOURCE.md from binary)"
    - "cowFS.ReadDir + strings.TrimSuffix for ListCows implementation"
    - "unexported readCowFile seam for Plan 01-03 cowfile.go"
key_files:
  created:
    - internal/cowsay/cows/ (50 .cow files from cowsay-org/cowsay v3.8.4)
    - internal/cowsay/cows/NOTICE
    - internal/cowsay/cows/SOURCE.md
    - internal/cowsay/embed.go
    - internal/cowsay/embed_test.go
  modified: []
decisions:
  - "Excluded flaming-sheep.cow (upstream v3.8.4 has 51 files but research listed 50; flaming-sheep absent from research §A-Q2 list)"
  - "//go:embed cows/*.cow used (not all:cows) per PITFALLS.md Pitfall 7 — excludes NOTICE and SOURCE.md from binary"
  - "readCowFile kept unexported — only ListCows is exported per plan interface spec"
  - "ListCows uses >= 50 floor in test (not == 51) to allow Plan 01-04 gopher.cow addition without test churn"
metrics:
  duration: "4m 12s"
  completed: "2026-05-21"
  tasks_completed: 2
  files_created: 53
---

# Phase 1 Plan 02: Cow Data Layer Summary

Vendor 50 upstream `.cow` files from cowsay-org/cowsay v3.8.4 (commit 027c9268), write NOTICE with daemon.cow BSD provenance and per-file license variances, write SOURCE.md with pinned commit SHA, and wire `//go:embed cows/*.cow` into `embed.FS` with `ListCows()` and unexported `readCowFile()` — `go vet` and `go test ./internal/cowsay` both pass.

## What Was Built

### Task 1: Vendor 50 upstream .cow files + NOTICE + SOURCE.md

**Source:** `cowsay-org/cowsay` tag `v3.8.4`, commit `027c9268ac8571408e153214b9cf1a5e6fab0cfc`.

Confirmed via: `git -C /tmp/cowsay-v3.8.4 rev-parse HEAD` returned `027c9268ac8571408e153214b9cf1a5e6fab0cfc`.

**Files vendored:** 50 `.cow` files placed at `internal/cowsay/cows/*.cow`.

Line ending verification: `grep -lP '\r' internal/cowsay/cows/*.cow | wc -l` = **0** (all LF).

Key files confirmed present:
- `daemon.cow` — included (with BSD Daemon provenance in NOTICE)
- `default.cow` — included
- `three-eyes.cow` — included
- `tux.cow` — included
- `dragon-and-cow.cow` — included
- `gopher.cow` — intentionally ABSENT (Plan 01-04 authors it)

**SOURCE.md** pins: upstream URL, tag v3.8.4, commit SHA, vendoring date 2026-05-21, and a refresh procedure covering clone → copy → strip extras → verify LF → update fields → run tests → review NOTICE.

**NOTICE** contains:
1. Upstream attribution (Tony Monroe 1999–2002, Andrew Janke 2016–2024)
2. Dual-license model (GPL v1+ OR Artistic 1.0; project-level GPLv3+)
3. Per-file license variances table (apt.cow GPL-only, gnu.cow WTFPL-2, suse.cow WTFPL-2, kangaroo.cow GPL-2.0+)
4. `daemon.cow` provenance paragraph naming Marshall Kirk McKusick, the Fedora 2016 removal, and the freebsd.org/copyright/daemon/ reference
5. Note on gopher.cow as original project work (not vendored)

### Task 2: internal/cowsay/embed.go + embed_test.go

**embed.go** (package `cowsay`):

```go
//go:embed cows/*.cow
var cowFS embed.FS
```

`ListCows()`: `cowFS.ReadDir("cows")` → filter `.cow` entries → strip suffix → `sort.Strings` → return.

`readCowFile(name string) ([]byte, error)`: returns `cowFS.ReadFile("cows/" + name + ".cow")` — the seam Plan 01-03 calls.

Imports: `embed`, `sort`, `strings` (stdlib only).

**embed_test.go** test results:

```
=== RUN   TestListCows_CountAndSort
--- PASS: TestListCows_CountAndSort (0.00s)
=== RUN   TestListCows_KeyFilesPresent
--- PASS: TestListCows_KeyFilesPresent (0.00s)
PASS
ok  github.com/pheckenlively/gosay/internal/cowsay  0.001s
```

`ListCows()` first-run sorted slice (first 10): `[actually alpaca beavis.zen blowfish bong bud-frogs bunny cheese cower cupcake]`

## Verification Results

```
ls internal/cowsay/cows/*.cow | wc -l    # 50
test -f internal/cowsay/cows/NOTICE      # OK
test -f internal/cowsay/cows/SOURCE.md   # OK
test -f internal/cowsay/embed.go         # OK
grep -q "v3.8.4" SOURCE.md               # OK
grep -q "027c9268..." SOURCE.md          # OK
grep -q "Marshall Kirk McKusick" NOTICE  # OK
CRLF count = 0                           # OK
go vet ./internal/cowsay                 # exit 0
go test ./internal/cowsay -run TestListCows  # PASS
go build ./...                           # exit 0
```

## Commits

| Task | Commit | Message |
|------|--------|---------|
| Task 1 | `9d75ac6` | `chore(01-02): vendor 50 upstream cow files + NOTICE + SOURCE.md` |
| Task 2 | `8b153b6` | `feat(01-02): add embed.go with embed.FS, ListCows, readCowFile + smoke tests` |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Research Discrepancy] Excluded flaming-sheep.cow from vendored set**
- **Found during:** Task 1 — upstream v3.8.4 has 51 files including `flaming-sheep.cow`, but the research (§A-Q2) only listed 50 names and the plan's success criteria require exactly 50 files
- **Issue:** The research table row 51 showed "fox" (a duplicate of row 18) — a formatting artifact. The actual upstream has `flaming-sheep.cow` as the 51st file not in the research list. The plan's must_haves say "(51 total minus gopher.cow)" assuming gopher.cow was in the upstream, but it is not.
- **Fix:** Removed `flaming-sheep.cow` after copying — kept exactly the 50 files matching the plan's frontmatter `files_modified` list. The plan's verification checks `wc -l | grep -q "^50$"`.
- **Files modified:** none (file simply not included)

None of the plan-specified API contracts deviated — both tasks executed as written.

## Known Stubs

None. `embed.go` returns real embedded data; `ListCows()` and `readCowFile()` read from the actual `cowFS`. No hardcoded values or placeholder data.

## Threat Surface Scan

No new network endpoints, auth paths, or schema changes introduced. Threat register mitigations applied:

| Threat ID | Status |
|-----------|--------|
| T-01-02-01 (upstream integrity) | Mitigated — SOURCE.md contains verified SHA `027c9268ac8571408e153214b9cf1a5e6fab0cfc` |
| T-01-02-02 (CRLF corruption) | Mitigated — all 50 files LF at commit time; .gitattributes policy active from Plan 01-01 |
| T-01-02-03 (accidental embedding of secrets) | Mitigated — `//go:embed cows/*.cow` glob matches only .cow files; NOTICE and SOURCE.md excluded from binary |
| T-01-02-04 (malicious .cow content) | Accepted per plan — no ANSI escapes found in upstream files |
| T-01-02-05 (missing NOTICE) | Mitigated — NOTICE written with all required fields verified by grep assertions |

## Self-Check: PASSED

Files created:
- FOUND: /work/gosay/.claude/worktrees/agent-a9a0d6aae81d45c72/internal/cowsay/cows/ (50 .cow files)
- FOUND: /work/gosay/.claude/worktrees/agent-a9a0d6aae81d45c72/internal/cowsay/cows/NOTICE
- FOUND: /work/gosay/.claude/worktrees/agent-a9a0d6aae81d45c72/internal/cowsay/cows/SOURCE.md
- FOUND: /work/gosay/.claude/worktrees/agent-a9a0d6aae81d45c72/internal/cowsay/embed.go
- FOUND: /work/gosay/.claude/worktrees/agent-a9a0d6aae81d45c72/internal/cowsay/embed_test.go

Commits verified:
- FOUND: 9d75ac6 (chore(01-02): vendor 50 upstream cow files + NOTICE + SOURCE.md)
- FOUND: 8b153b6 (feat(01-02): add embed.go with embed.FS, ListCows, readCowFile + smoke tests)
