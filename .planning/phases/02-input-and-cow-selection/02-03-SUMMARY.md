---
phase: 02-input-and-cow-selection
plan: "03"
subsystem: cmd/gosay
tags: [flag-parsing, cow-listing, random-selection, conflict-guards, columnar-output]
dependency_graph:
  requires: [flag-based run() scaffold (plan 02-02), cowsay.ListCows() (plan 02-01)]
  provides: [-l columnar listing, --random pool selection, D-06/D-09 conflict guards]
  affects: [Phase 2 complete — all COW-03/COW-04 requirements met]
tech_stack:
  added: [math/rand]
  patterns: [fs.Visit explicit-flag detection, formatCowList() helper for deterministic columnar output, rand.Intn bounded to ListCows() pool]
key_files:
  created: []
  modified:
    - cmd/gosay/main.go
    - cmd/gosay/main_test.go
decisions:
  - "Use fs.Visit (not default-value comparison) to detect explicit -f: avoids false positive when gopher is the default and user only passes --random"
  - "formatCowList() unexported helper in main package: deterministic, unit-testable, keeps run() readable"
  - "Wrap width set to 76 characters per line: matches upstream cowsay general column shape per D-07 (Claude's discretion)"
  - "No seeded/injectable RNG for --random (D-05): package-level math/rand.Intn only"
  - "D-09 guard checks fExplicit (not cowName != gopher) for consistency with D-06 detection"
metrics:
  duration: ~15 minutes
  completed: "2026-05-30T00:00:00Z"
---

# Phase 02 Plan 03: -l listing and --random selection Summary

**One-liner:** Added `-l` columnar cow listing (Cow files: header, 76-char wrap, no default marker) and `--random` pool selection from full ListCows(), with D-06 (-f+--random) and D-09 (-l+input) mutual-exclusion guards; backed by five new CLI tests.

## What Was Built

Extended `run()` in `cmd/gosay/main.go` with two new flags and their conflict guards on top of the plan-02 FlagSet scaffold, completing the Phase 2 selection surface.

### Key Changes

**`cmd/gosay/main.go`:**
- Added `"math/rand"` import
- Registered `-l` (bool) and `--random` (bool) on the existing FlagSet; extended `fs.Usage` string to include them
- Added `fExplicit` detection via `fs.Visit` to distinguish explicit `-f` from the default `"gopher"`
- D-06 guard: `randomFlag && fExplicit` → `"gosay: cannot combine -f and --random"` → return 1
- D-09 guard: `listFlag && (len(fs.Args()) > 0 || randomFlag || fExplicit)` → `"gosay: -l cannot be combined with a message or animal selection"` → return 1
- `-l` branch: calls `cowsay.ListCows()`, formats via `formatCowList()`, writes to stdout, returns 0
- `--random` branch: calls `cowsay.ListCows()`, picks `names[rand.Intn(len(names))]`, assigns to `animal`
- Added `formatCowList(names []string) string`: "Cow files:\n" header + space-separated names wrapped at 76 chars per line
- All plan-02 behavior preserved: TTY gating, args-win, empty-bubble, `errors.Is(err, cowsay.ErrUnknownCow)`

**`cmd/gosay/main_test.go`:**
- `TestRun_ListFlag_AloneExits0`: code 0, `Cow files:` prefix, `gopher` present, no `default)` marker
- `TestRun_ListFlag_WithMessage_IsError`: code 1, no stdout
- `TestRun_ListFlag_WithRandom_IsError`: code 1
- `TestRun_FlagConflict_FAndRandom`: code 1, `cannot combine` on stderr, no stdout
- `TestRun_Random_ReturnsMember`: code 0, non-empty stdout

## Task Summary

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Add -l listing, --random selection, and conflict guards to run() | `82c40b3` | cmd/gosay/main.go |
| 2 | Add CLI tests for listing, random membership, and conflict errors | `264ba2f` | cmd/gosay/main_test.go |

## Verification Results

- `go build ./...` — PASS
- `go vet ./...` — PASS (no findings)
- `go test ./...` — PASS (all packages green, 12/12 cmd tests pass)
- `go run ./cmd/gosay -l | head -1` — `Cow files:`
- `go run ./cmd/gosay -l | tr ' ' '\n' | grep -c gopher` — 1 (present, plain)
- `go run ./cmd/gosay -l | grep -c 'default)'` — 0 (no marker)
- `go run ./cmd/gosay -l; echo "exit=$?"` — exit=0
- `go run ./cmd/gosay --random hello; echo "exit=$?"` — renders animal with hello, exit=0
- `go run ./cmd/gosay -f tux --random hello; echo "exit=$?"` — `cannot combine -f and --random`, exit=1
- `go run ./cmd/gosay -l hello; echo "exit=$?"` — exit=1
- `grep -n 'math/rand' cmd/gosay/main.go` — matched (line 8)
- `grep -n 'rand.Intn' cmd/gosay/main.go` — matched (line 111)
- `grep -n 'Cow files:' cmd/gosay/main.go` — matched (lines 25, 30)
- `grep -n 'cannot combine -f and --random' cmd/gosay/main.go` — matched

## Deviations from Plan

None — plan executed exactly as written.

The `fs.Visit`-based explicit-flag detection was specified directly in the plan task action (line 97) and was implemented as described.

## Threat Surface Scan

No new network endpoints, auth paths, or external file access introduced.

- T-02-05 (Information Disclosure — `-l` listing): ACCEPTED. `-l` prints only embedded cow base names (public, vendored). No filesystem paths or internal detail.
- T-02-06 (Tampering — `--random` index): ACCEPTED. `rand.Intn(len(names))` is bounded to the ListCows() slice; chosen name is always a known embedded cow, so downstream `LoadCow` cannot hit `ErrUnknownCow`.
- T-02-02 (Tampering — `-f` traversal): MITIGATED. D-06 guard additionally blocks `-f`+`--random` ambiguity; path traversal still resolves to `fs.ErrNotExist` → clean `ErrUnknownCow` message (unchanged from plan-02).

## Self-Check: PASSED

- `cmd/gosay/main.go` — exists with `math/rand`, `rand.Intn`, `Cow files:`, `cannot combine -f and --random`, `formatCowList`, `fExplicit`, `fs.Visit`
- `cmd/gosay/main_test.go` — exists with `TestRun_ListFlag_AloneExits0`, `TestRun_ListFlag_WithMessage_IsError`, `TestRun_ListFlag_WithRandom_IsError`, `TestRun_FlagConflict_FAndRandom`, `TestRun_Random_ReturnsMember`
- Commits `82c40b3` and `264ba2f` exist in git log
