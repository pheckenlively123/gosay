# Cow File Provenance

## Source

- **Upstream:** https://github.com/cowsay-org/cowsay
- **Tag:** v3.8.4
- **Commit:** 027c9268ac8571408e153214b9cf1a5e6fab0cfc
- **Vendored:** 2026-05-21

## Files Vendored

50 of the 51 upstream `.cow` files from `share/cowsay/cows/` were vendored into this
directory. The file `gopher.cow` is intentionally absent — it is hand-authored in this
project (Plan 01-04) as an original work rather than derived from the upstream set.

The file `flaming-sheep.cow` was present in the upstream tag but is not included here
as it was not in the verified file list from the Phase 1 research (§A-Q2).

## Refresh Procedure

To re-vendor the cow files from a newer upstream release:

1. Clone the desired tag:
   ```
   git clone --depth 1 --branch <new-tag> https://github.com/cowsay-org/cowsay /tmp/cowsay-<new-tag>
   ```

2. Confirm the commit SHA:
   ```
   git -C /tmp/cowsay-<new-tag> rev-parse HEAD
   git ls-remote https://github.com/cowsay-org/cowsay refs/tags/<new-tag>
   ```

3. Copy the cow files:
   ```
   cp /tmp/cowsay-<new-tag>/share/cowsay/cows/*.cow internal/cowsay/cows/
   ```

4. Remove files that are not part of gosay's vendored set:
   ```
   rm -f internal/cowsay/cows/gopher.cow
   # Remove any other newly-added upstream files not in the research-vetted list
   ```

5. Verify LF line endings (should return 0):
   ```
   grep -rlP '\r' internal/cowsay/cows/*.cow | wc -l
   ```

6. Update the `Tag`, `Commit`, and `Vendored` fields in this file.

7. Run the embed tests:
   ```
   go test ./internal/cowsay/...
   ```

8. Review any new files against `internal/cowsay/cows/NOTICE` for license changes.
