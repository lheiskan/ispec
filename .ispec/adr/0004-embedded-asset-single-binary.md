---
type: adr
stable_id: ADR-0004
slug: embedded-asset-single-binary
status: accepted
supersedes: [ADR-0003]
---

## Context

[[ADR-0003]] chose symlinks for v0.1 and predicted its own supersession. Its
cost had become concrete:

- The install points into a live checkout the user must keep forever, in the
  right place, on the right branch. `INSTALL.md` had to warn against moving or
  deleting it.
- `make install` skipped any destination that already existed, so a stale
  hand-copied skill won permanently and nothing reported the conflict.
- The install was a shell loop in a Makefile: no Windows, and `make` required
  to install a Go program.
- The binary needed a manual rebuild after any source change; the symlink
  could not notice.

ADR-0003 rejected real distribution on one specific ground: "the skills aren't
binary content — they're markdown — so a Go-only distribution doesn't cover
everything; a packaged install would still need a separate mechanism for
skills."

That premise is false under `go:embed`. A `//go:embed skills/*/SKILL.md`
directive makes the markdown *part of* the binary, read at build time. One
artifact then carries both halves, and there is no second mechanism to build.

## Decision

Distribute ispec as a single Go binary with its agent-facing assets embedded,
installed with `go install github.com/lheiskan/ispec/cmd/ispec@latest`. The
binary installs its own skills via `ispec harness install`, reading them from
itself rather than from a checkout.

Consequently:

- `cmd/ispec-lint` becomes `cmd/ispec`, with `lint` as a subcommand, because
  the binary now does more than lint.
- Skills move from `.claude/skills/` to `skills/` at the repository root.
  `go:embed` skips dot-prefixed directories, and the assets are
  harness-neutral — which harness they are installed into is
  [[ADR-0005]]'s concern, not a fact baked into where they are stored.
- The validation engine moves to `internal/lint`, so `status` and the
  session-start directive can read a canon without shelling out to a lint.

## Consequences

- No checkout is retained, or ever required. `go install ...@latest` fetches,
  builds and installs in one step, and the module path already supports it.
- Pure Go: installs on Windows, and needs no `make`.
- Skills cannot drift from the binary, because they are the binary. A stale
  install is now detectable by comparing embedded bytes to installed bytes,
  which is what `harness check` does.
- **Editing a skill now requires a rebuild to take effect.** This is the real
  cost, and it lands hardest on ispec's own development, where skills are
  iterated constantly. `go run ./cmd/ispec harness install` is the one-command
  answer; `harness check` reports when the installed copy has fallen behind.
- ispec's own repository no longer carries project-scoped skills under
  `.claude/`, so ispec dogfoods the same global install path its adopters use.
- Versions are still not pinned per project: the installed binary is whatever
  was last installed. Per-project pinning stays deferred, but is now a
  tractable problem — `go install ...@v1.2.3` is a version selector, where a
  symlink was not.

## Alternatives considered

- **Keep symlinks, add a `check` verb.** Rejected: drift detection is the
  smaller half of the problem. The checkout dependency and the
  shell-and-`make` install are what make the first-run experience bad.
- **Copy skills from a checkout at install time.** Rejected: still requires a
  checkout at install time, and reintroduces the "separate mechanism for
  skills" that embedding removes.
- **Embed skills but keep them in `.claude/skills/`.** Possible — `go:embed
  all:.claude/skills` would reach them. Rejected because it asserts that
  ispec's assets belong to Claude Code, which contradicts [[ADR-0005]].
