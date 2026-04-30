---
type: feature
stable_id: F-3
slug: install-pathway
status: implemented
targets: [REQ-3]
paths:
  - Makefile
  - INSTALL.md
---

## Description

A Makefile-driven install path that creates symlinks under the user's global Claude skill directory and on the user's `$PATH`, making ispec's workflow skills and the `ispec-lint` binary available to any Claude Code session in any project.

## Behavior

- `make build` produces the `ispec-lint` binary at the project root.
- `make install` creates symlinks:
  - `~/.claude/skills/<skill-name>` → `<ispec-checkout>/.claude/skills/<skill-name>` for each of the five workflow skills.
  - `~/.local/bin/ispec-lint` → `<ispec-checkout>/ispec-lint`.
- Existing entries at install destinations (real files or other symlinks) are left in place; the install reports `skipping` for those. The user can `make uninstall && make install` to refresh.
- `make uninstall` removes only ispec's own symlinks; does not delete the ispec source or build artifacts.
- `make clean` removes the built binary.
- Install destinations are configurable via `make install SKILLS_DST=... BIN_DST=...`.

INSTALL.md documents prerequisites (Go toolchain, `~/.local/bin` on `$PATH`), the install steps, verification, and update / uninstall flow.

## Limitations (deferred)

- Symlinks point at the live ispec checkout, so any change to ispec immediately propagates to all adopting projects. No per-project version pinning, no drift detection.
- The lint binary requires manual `make build` when ispec source changes; the symlink doesn't trigger a rebuild.
- Distribution via `go install`, Homebrew, or system package managers is out of scope; intended for now is local-checkout-based use only. See [[ADR-0003]] for the choice and its supersession plan.
- Multi-user / system-wide installation is out of scope.
