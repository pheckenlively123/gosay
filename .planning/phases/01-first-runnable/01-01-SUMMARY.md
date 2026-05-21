---
phase: 01-first-runnable
plan: "01"
subsystem: scaffold
tags: [go-module, gitattributes, walking-skeleton, cli-entry-point]
dependency_graph:
  requires: []
  provides:
    - go.mod (module declaration, Go version target, toolchain pin)
    - .gitattributes (COWS-05 line-ending policy)
    - cmd/gosay/main.go (CLI entry point stub, os.Args reading)
  affects: []
tech_stack:
  added:
    - Go 1.22 minimum (go.mod)
    - toolchain go1.26.3
  patterns:
    - os.Args positional arg reading (no flag.Parse)
    - stderr usage message + os.Exit(1) on missing args
key_files:
  created:
    - go.mod
    - .gitattributes
    - cmd/gosay/main.go
    - .gitignore
  modified: []
decisions:
  - "go 1.22 minimum in go.mod for broadest go install reach while keeping //go:embed and generics"
  - "toolchain go1.26.3 pins local builds to latest patch without raising minimum"
  - "No require block in go.mod for Plan 01-01 — goldie/v2 enters in Plan 01-04 when first golden test exists"
  - ".gitattributes lands before any cow file to enforce LF endings (COWS-05, PITFALLS #8)"
  - "Phase 1 main.go uses os.Args[1:] only — flag.Parse() deferred to Phase 2"
metrics:
  duration: "1m 34s"
  completed: "2026-05-21"
  tasks_completed: 2
  files_created: 4
---

# Phase 1 Plan 01: Walking Skeleton Summary

Go module scaffold and minimal CLI entry point — `go build ./...` succeeds and `go run ./cmd/gosay hello` prints a deterministic stub line to stdout.

## What Was Built

### Task 1: Initialize Go module and line-ending policy

**go.mod** (exact contents):
```
module github.com/pheckenlively/gosay

go 1.22

toolchain go1.26.3
```

**`.gitattributes`** (exact contents):
```
cows/*.cow text eol=lf
```

The `.gitattributes` file lands before any `.cow` file is committed (Plan 01-02 vendors the cow files), ensuring the LF-enforcement policy is active when the first cow file is checked in. This is the load-bearing CRLF defense from PITFALLS.md Pitfall 8.

### Task 2: Add walking-skeleton main.go

**`cmd/gosay/main.go`** — 16 lines including package clause and imports.

Key behavior:
- Reads `os.Args[1:]` (no `flag.Parse()`)
- Joins args with spaces: `message := strings.Join(os.Args[1:], " ")`
- If no args: prints `usage: gosay <message>` to stderr and exits 1
- Otherwise: prints `gosay (walking skeleton): would render gopher saying: <message>` to stdout

**Verified `go run ./cmd/gosay hello` stdout:**
```
gosay (walking skeleton): would render gopher saying: hello
```

**Verified `go run ./cmd/gosay hello world` stdout:**
```
gosay (walking skeleton): would render gopher saying: hello world
```

**Verified `go run ./cmd/gosay` (no args) stderr:**
```
usage: gosay <message>
```

### Additional: .gitignore

Added `.gitignore` for the compiled `gosay` binary (generated artifact) and `*.test` files.

## Verification Results

```
go build ./...          # exit 0 (no packages before Task 2; exit 0 after)
go build ./cmd/gosay    # exit 0
go run ./cmd/gosay hello world | grep "hello world"  # found
grep "^module github.com/pheckenlively/gosay$" go.mod  # found
grep "^cows/*.cow text eol=lf$" .gitattributes         # found
```

All success criteria met:
- [x] `go.mod` exists with exact module path, go directive, toolchain directive
- [x] `.gitattributes` exists with `cows/*.cow text eol=lf`
- [x] `cmd/gosay/main.go` builds; reads positional args via `os.Args[1:]`; prints deterministic stub; no `flag.Parse()`; no stdin
- [x] `go build ./...` exits 0
- [x] `go run ./cmd/gosay hello` prints a line containing `hello` to stdout
- [x] `go run ./cmd/gosay` (no args) writes usage to stderr and exits non-zero

## Commits

| Task | Commit | Message |
|------|--------|---------|
| Task 1 | `490ba6b` | `chore(01-01): initialize Go module and line-ending policy` |
| Task 2 | `918ba2b` | `feat(01-01): add walking-skeleton cmd/gosay/main.go` |
| Post-task | `70db6b3` | `chore(01-01): add .gitignore for compiled binary` |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical Functionality] Added .gitignore for compiled binary**
- **Found during:** Task 2 verification — `go build ./cmd/gosay` left an untracked `gosay` binary
- **Issue:** The compiled binary was untracked. While not blocking, leaving generated artifacts untracked in git creates noise and risks accidental commits.
- **Fix:** Created `.gitignore` with `gosay` and `*.test` entries; committed alongside the task.
- **Files modified:** `.gitignore`
- **Commit:** `70db6b3`

None of the plan-specified tasks deviated — both executed exactly as written.

## Threat Surface Scan

No new security-relevant surface introduced. Plan 01-01 creates:
- `go.mod` — read by Go toolchain at build time (T-01-01-01: accepted per threat register)
- `.gitattributes` — enforces LF endings (T-01-01-04: mitigation applied as planned)
- `cmd/gosay/main.go` — echoes user input verbatim to stdout (T-01-01-02: accepted per threat register)

No new network endpoints, auth paths, file access patterns, or schema changes. Threat register entries T-01-01-01 through T-01-01-04 are all addressed or accepted as documented in the plan.

## Self-Check: PASSED

Files created:
- FOUND: /work/gosay/go.mod
- FOUND: /work/gosay/.gitattributes
- FOUND: /work/gosay/cmd/gosay/main.go
- FOUND: /work/gosay/.gitignore

Commits verified:
- FOUND: 490ba6b (chore(01-01): initialize Go module and line-ending policy)
- FOUND: 918ba2b (feat(01-01): add walking-skeleton cmd/gosay/main.go)
- FOUND: 70db6b3 (chore(01-01): add .gitignore for compiled binary)
