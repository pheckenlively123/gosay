---
phase: 02-input-and-cow-selection
plan: "01"
subsystem: cowsay
tags: [error-handling, sentinel-error, tdd]
dependency_graph:
  requires: []
  provides: [cowsay.ErrUnknownCow sentinel, LoadCow fs.ErrNotExist wrapping]
  affects: [main.go COW-05 error message (plan 02-02)]
tech_stack:
  added: []
  patterns: [sentinel error via errors.New, errors.Is wrapping with fmt.Errorf %w]
key_files:
  created: []
  modified:
    - internal/cowsay/cowfile.go
    - internal/cowsay/cowfile_test.go
decisions:
  - "Use fmt.Errorf(\"%w: %s\", ErrUnknownCow, name) to wrap the sentinel with the cow name, preserving errors.Is compatibility without leaking the internal embed path (cows/<name>.cow)"
metrics:
  duration: ~5 minutes
  completed: "2026-05-30T23:23:46Z"
---

# Phase 02 Plan 01: Add ErrUnknownCow Sentinel Error Summary

**One-liner:** Exported `ErrUnknownCow` sentinel wraps `fs.ErrNotExist` in `LoadCow`, enabling `main.go` to detect a missing cow via `errors.Is` without leaking the `cows/<name>.cow` internal embed path.

## What Was Built

Added a package-level sentinel error `ErrUnknownCow` to `internal/cowsay/cowfile.go` and updated `LoadCow` to wrap `fs.ErrNotExist` with it. This establishes the error contract that plan 02-02 (COW-05 user-facing message) depends on.

### Key Changes

- `internal/cowsay/cowfile.go`: Added `"io/fs"` import, exported `var ErrUnknownCow = errors.New("unknown cowfile")` with doc comment, and added `errors.Is(err, fs.ErrNotExist)` guard in `LoadCow` that returns `fmt.Errorf("%w: %s", ErrUnknownCow, name)` instead of the old path-bearing error.
- `internal/cowsay/cowfile_test.go`: Added `"errors"` import and new `TestLoadCow_Nonexistent_SentinelError` test verifying `errors.Is(err, ErrUnknownCow)` is true for a missing cow.

## TDD Gate Compliance

| Gate | Commit | Status |
|------|--------|--------|
| RED (test) | `63c1ede` | PASS — test compiled with undefined `ErrUnknownCow`, build failed as expected |
| GREEN (impl) | `7eafed5` | PASS — all LoadCow tests pass including new sentinel test |
| REFACTOR | N/A | No refactor needed — implementation was already clean |

## Task Summary

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 (RED) | Add failing sentinel-error test | `63c1ede` | cowfile_test.go |
| 1 (GREEN) | Add ErrUnknownCow sentinel + LoadCow wrapping | `7eafed5` | cowfile.go |

## Verification Results

- `go test ./internal/cowsay/... -run 'TestLoadCow' -v` — PASS (all 3 variants)
- `go test ./internal/cowsay/...` — PASS (29 tests, 1 skipped)
- `go build ./...` — PASS
- `go vet ./internal/cowsay/...` — PASS (no findings)
- `TestRender_UnknownCow` also passes, confirming the sentinel propagates through `Render`'s `%w` wrapping

## Deviations from Plan

None — plan executed exactly as written.

## Threat Surface Scan

No new network endpoints, auth paths, file access patterns, or schema changes introduced. The `ErrUnknownCow` sentinel reduces information disclosure (T-02-01) by replacing the path-bearing error with a name-only message. T-02-02 (embed.FS traversal) remains accepted — `fs.ValidPath` rules cause traversal names to fail as `fs.ErrNotExist` and route to the sentinel.

## Self-Check: PASSED

- `internal/cowsay/cowfile.go` — file exists with `var ErrUnknownCow`, `"io/fs"` import, and `errors.Is(err, fs.ErrNotExist)` guard
- `internal/cowsay/cowfile_test.go` — file exists with `TestLoadCow_Nonexistent_SentinelError`
- Commits `63c1ede` and `7eafed5` exist in git log
