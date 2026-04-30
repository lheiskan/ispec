---
type: session
slug: 2026-04-30-bootstrap
trigger: bootstrap ispec onto itself; relocate canon to .ispec/ and author the load-bearing initial cards
classification: bootstrap
status: resolved
affected_cards:
  - STK-1
  - STK-2
  - REQ-1
  - REQ-2
  - ADR-0001
  - ADR-0002
  - SPEC-data-model
  - F-1
  - F-2
deferred_cards: []
commits: []
outcome: |
  ispec now has an initial canon. STK-1 / STK-2 establish the two
  active stakeholders (project authors; AI agents). REQ-1 / REQ-2
  state the two load-bearing requirements (durable canonical
  documentation; no silent drift). ADR-0001 / ADR-0002 capture the
  two design decisions made during this bootstrap (skill-driven
  workflow over slash-command transitions; canon directory at
  .ispec/ rather than docs/). SPEC-data-model documents the card
  taxonomy and reference rules. F-1 / F-2 record the two
  implemented features (lint binary; spec-session skill).

  Lint passes clean. Coverage is deliberately scoped to the load-
  bearing minimum; additional ADRs (atemporal canon vs session log
  split, tooling stack rationale), additional requirements (scoped
  AI loading, multi-audience views), additional features (bootstrap
  skill, retrofit skill, defer skill, session-close skill), and
  additional spec sections (session loop, status lifecycles,
  cross-repo handling) are deferred to follow-up sessions.

  The canon directory location decision (.ispec/ rather than docs/)
  necessitated a small change to the lint binary (configurable
  -canon-dir flag, default .ispec) and updates to design.md,
  schemas.md, and the bootstrap / retrofit / spec-session /
  session-close skill files. These were applied as part of this
  session.
---

## Trigger

Dogfood ispec on itself: take the design corpus and tooling produced over the previous April 30, 2026 dialogue and capture canon for it. Decide where the canon directory should live, then author the load-bearing minimum.

## Pre-decisions

- Use `.ispec/` as the canon directory rather than `docs/` (formalized as [[ADR-0002]]).
- Stay scope-controlled — author the minimum viable canon, defer the rest to follow-up sessions.

## Discoveries during the session

- The schema field `satisfies_features` on spec cards reads slightly backwards (a spec doesn't "satisfy" a feature; rather, a feature *implements* what the spec describes). The current framing — "this spec satisfies the documentation need of these features" — works but is not ideal naming. Captured as a follow-up consideration; not changed in this session to avoid scope creep.
- The bootstrap skill assumes a brand-new project ships a working slice in commit 1 and has minimal canon to read in Phase 1. ispec is unusual — its "code" *is* its design (skills + lint + design corpus). Phase 1 worked because there was already substantial design material to draft cards from; the dialogue was less "what does this code mean?" and more "which design decisions deserve ADRs?"

## Outcome

See frontmatter `outcome` field.
