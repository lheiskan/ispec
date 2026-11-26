---
type: feature
stable_id: F-3
slug: install-pathway
status: implemented
targets: [REQ-3]
paths:
  - cmd/ispec/**
  - embed.go
  - Makefile
  - INSTALL.md
---

## Description

Installation of ispec as a single Go binary carrying its own agent-facing
assets, so that adopting a project requires no ispec checkout — at install
time or afterwards. Per [[ADR-0004]].

## Behavior

- `go install github.com/lheiskan/ispec/cmd/ispec@latest` fetches, builds and
  installs the binary. No clone at any point.
- The binary embeds `skills/*/SKILL.md` at build time (`embed.go`), so it can
  install the workflow skills with no source tree present.
- `ispec harness install` writes the skills and the session-start hook into
  the target harness. See [[F-5]] for the integration itself.
- From a checkout, `make install` runs `go install` and then has the
  *installed* copy install the harness integration, so the recorded hook path
  does not depend on the checkout surviving.
- `make build` / `make test` / `make clean` cover local development;
  `make check` and `make uninstall` delegate to the binary.

## Limitations (deferred)

- No per-project version pinning. The installed binary is whatever was last
  installed, shared by every project on the machine. Now tractable —
  `go install ...@v1.2.3` selects a version where a symlink could not — but
  not implemented.
- No prebuilt release binaries and no package-manager distribution (Homebrew,
  apt). Installing requires a Go toolchain.
- Editing a skill requires rebuilding the binary for the change to take
  effect; `go run ./cmd/ispec harness install` is the development loop.
  Inherent to [[ADR-0004]], not an oversight.
- `make install` resolves `GOBIN` via `go env`; an unusual Go environment may
  need the binary installed by hand.
