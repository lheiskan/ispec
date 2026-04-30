---
type: adr
stable_id: ADR-0003
slug: install-via-symlinks
status: accepted
---

## Context

ispec needs to be reachable from arbitrary projects so a Claude Code session in any repo can invoke ispec's workflow skills and run the lint binary. Three viable install models:

- **Symlinks** from a fixed user-global location into ispec's source checkout.
- **Copy** of skills and binary into the user-global location at install time, with re-copy required on every ispec change.
- **Real distribution** via `go install`, Homebrew, or a package manager.

## Decision

Symlinks for v0.1. `make install` creates symlinks from `~/.claude/skills/` and `~/.local/bin/` into ispec's checkout. The `ispec-lint` binary is built locally with `make build`.

## Consequences

- Changes to ispec's skill files propagate immediately — no install step needed for skill edits. Critical while ispec's own design is still being iterated.
- The lint binary still requires `make build` after Go-source changes; the symlink doesn't rebuild on its own.
- Adopting projects rely on a single shared install rather than copies that must be kept in sync.
- The user must keep ispec's checkout in place; moving or deleting it breaks every install.
- No per-project version pinning. Every project sees the same live ispec.
- This will probably be superseded by a real distribution path (likely a copy-with-version-pin, or a Go-installable binary plus a separate skill-bundle mechanism) once ispec's surface stabilizes.

## Alternatives considered

- **Copy on install.** Rejected for v0.1: ispec is being actively developed and skill changes need to land instantly across consuming projects. Once ispec stabilizes, copy-with-version-pin will be more appropriate, captured in a future ADR that supersedes this one.
- **`go install` / package manager.** Rejected for v0.1: too heavy for a tool still finding its shape. The skills aren't binary content — they're markdown — so a Go-only distribution doesn't cover everything; a packaged install would still need a separate mechanism for skills.
