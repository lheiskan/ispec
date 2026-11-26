---
type: adr
stable_id: ADR-0001
slug: skill-driven-workflow
status: accepted
---

## Context

The session loop has multiple phases — intake, classification, design, implementation, reconciliation, close. Two designs are available: explicit phase-transition slash commands the user invokes, or skill-driven workflow where the agent classifies the trigger and walks the loop with confirmation points only at moments where user input is load-bearing.

## Decision

Skill-driven. The user describes what they want in natural language; the [[F-2]] (`ispec-start-session`) skill classifies the trigger and walks the phases. Confirmation points are surfaced only where they genuinely matter:

- The contract step (before coding starts), where scope and proposed canon changes are agreed.
- Mid-coding amendment surfacing, when implementation forces a spec change.
- Pre-completion reconciliation, when the implementation diverges from the original feature definition.

A small number of explicit user-facing commands (`ispec-defer`, `ispec-close-session`) cover cases where user intent is the primary input.

## Consequences

- Less ceremony. The user works normally; skills steer.
- Phase boundaries remain real — they shape what artifacts are produced — but stay internal to the agent's process.
- One skill can cover multiple trigger types (bug, feature, refactor, ops change) via classification, rather than splitting per trigger type. See the trigger classification list in [[F-2]].
- The risk of silent scope expansion or canon drift is concentrated at the few explicit confirmation points, where it can be surfaced.

## Alternatives considered

- **Slash command per phase** (`/intake`, `/design`, `/implement`, etc.). Rejected: forces the user to remember the workflow and drive transitions; high friction; breaks if a step is forgotten.
- **Single skill with no confirmation points.** Rejected: silent scope expansion and silent canon drift become inevitable. The whole point of the system is to keep canon honest; that requires moments where the user ratifies what the agent is about to do.
