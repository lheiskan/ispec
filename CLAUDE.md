# ispec

A system for producing and maintaining requirements, specifications, and decisions in AI-assisted software development projects.

## Where to start

- `docs/design.md` — overall design: artifact taxonomy, data model, session loop, status lifecycles, defer mechanism, cross-repo sessions, bootstrap path, lint rules, open questions.
- `docs/schemas.md` — concrete frontmatter schemas for each card type. The lint binary checks against these.
- `docs/worked-examples/cross-repo-session.md` — synthetic worked example walking a cross-repo session end to end, used to validate the design.
- `.claude/skills/bootstrap/SKILL.md` — initialize a brand-new project (Phase 0 spike → Phase 1 retrospective canon → Phase 2 steady state).
- `.claude/skills/retrofit/SKILL.md` — adopt ispec on an existing project with code, history, and possibly existing docs. Always scope deliberately; do not retroactively document everything.
- `.claude/skills/spec-session/SKILL.md` — main workflow skill (invoke at the start of any unit of work in a bootstrapped project).
- `.claude/skills/session-close/SKILL.md` — close-out skill (invoke when work is complete and committed).
- `.claude/skills/defer/SKILL.md` — explicit deferral of work surfaced during a session.

## Tooling

- `cmd/ispec-lint/main.go` — Go binary that walks `docs/`, parses card frontmatter, validates per-type schema rules, and resolves cross-card references (frontmatter ID lists and body `[[wikilinks]]`). Cross-repo references warn but are not yet resolved.
  - Build: `go build ./cmd/ispec-lint` (deposits `./ispec-lint`).
  - Run: `./ispec-lint -root <project-root>`.
  - Exits 0 on clean, 1 on any error, 2 on missing `docs/`.

## Status

Early design. Skills and the design doc exist; the card store layout (`docs/stakeholders/`, `docs/requirements/`, etc.) and the lint / view-generation tooling are not yet implemented.

The next concrete steps are listed in `docs/design.md` under "Open design questions."
