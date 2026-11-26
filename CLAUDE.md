# ispec

A system for producing and maintaining requirements, specifications, and decisions in AI-assisted software development projects.

## Where to start

- `docs/design.md` — overall design: artifact taxonomy, data model, session loop, status lifecycles, defer mechanism, cross-repo sessions, bootstrap path, lint rules, open questions.
- `docs/schemas.md` — concrete frontmatter schemas for each card type. The lint binary checks against these.
- `docs/worked-examples/cross-repo-session.md` — synthetic worked example walking a cross-repo session end to end, used to validate the design.
- `skills/ispec-bootstrap/SKILL.md` — initialize a brand-new project (Phase 0 spike → Phase 1 retrospective canon → Phase 2 steady state).
- `skills/ispec-retrofit/SKILL.md` — adopt ispec on an existing project with code, history, and possibly existing docs. Always scope deliberately; do not retroactively document everything.
- `skills/ispec-start-session/SKILL.md` — main workflow skill (invoke at the start of any unit of work in a bootstrapped project).
- `skills/ispec-close-session/SKILL.md` — close-out skill (invoke when work is complete and committed).
- `skills/ispec-defer/SKILL.md` — explicit deferral of work surfaced during a session.

## Tooling

- `cmd/ispec/` — the `ispec` binary. Subcommands:
  - `ispec lint [-root DIR] [-canon-dir NAME]` — validate card schemas and resolve cross-card references (frontmatter ID lists and body `[[wikilinks]]`). Cross-repo references warn but are not yet resolved.
  - `ispec status` — card counts, open sessions, deferred features, lint state.
  - `ispec harness install|check|uninstall|session-start` — install ispec into an agent harness. See `INSTALL.md`.
  - Exit codes: 0 clean, 1 problems found, 2 no canon directory.
- `internal/lint/` — the card schema and reference resolution. `internal/status/` — facts about a canon. `internal/harness/` — per-harness placement plus the session-start directive.
- `embed.go` — embeds `skills/*/SKILL.md` into the binary, so an install needs no checkout (ADR-0004).
- Build: `make build`. Test: `make test`. Install: `make install`.
- **Editing a skill requires a rebuild to take effect.** Dev loop: `make install`. (`go run ./cmd/ispec harness install` is refused — `go run`'s binary lives in the build cache, so the hook would record a path that vanishes.)

## Status

Early. The data model, schemas, workflow skills, and lint binary exist and are
used on real projects; ispec's own canon lives in `.ispec/` and its lint runs
clean against itself.

Not built yet: view generation (deriving `SPEC.md` / manuals / runbooks from
audience-tagged cards), cross-repo reference resolution (repo-prefixed
references parse and warn but are not resolved), the full set of cross-card
status invariants, and agent-agnostic packaging.

Remaining design questions are listed in `docs/design.md` under "Open design
questions."
