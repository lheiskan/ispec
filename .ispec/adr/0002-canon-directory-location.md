---
type: adr
stable_id: ADR-0002
slug: canon-directory-location
status: accepted
---

## Context

The canon needs a default home in each project. The original sketch proposed `docs/`. Two issues surfaced when dogfooding ispec on itself:

1. Many real-world projects (including ispec) already have hand-authored content in `docs/`. Putting cards there mixes tool-managed source-of-truth with human-authored long-form content.
2. The convention in adjacent tooling (`.git/`, `.claude/`) is hidden tool-managed directories at the project root.

## Decision

Cards live in `.ispec/` by default. The `docs/` directory is left to the project: human-authored long-form content (README, design notes, worked examples), and any artifacts the project chooses to derive from the canon with other tools. ispec itself writes nothing there — see [[ADR-0006]].

The location is configurable via `ispec lint -canon-dir <subdir>` for projects that want a different layout.

## Consequences

- Mirrors the convention of `.git/` / `.claude/`: tool-managed, hidden, source-of-truth.
- Existing projects with `docs/` content can adopt ispec without conflict.
- The separation is by *authority*, not by source-versus-output: `.ispec/` is tool-managed and lint-verified, `docs/` is the project's own. Anything derived from the canon by other tooling can land in `docs/` without colliding with cards.

## Alternatives considered

- **`docs/` as default.** Rejected: conflicts with existing project content; conflates source with output.
- **`spec/` as default.** Rejected: too narrow — the canon includes stakeholders, sessions, and ADRs, not just spec.
- **Configurable with no default.** Rejected: increases adoption friction; a sensible default is more useful than perfect flexibility.

Implemented in [[F-1]] as the default `-canon-dir` flag value.
