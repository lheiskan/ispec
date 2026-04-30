---
type: stakeholder
stable_id: STK-1
slug: project-authors
status: active
---

## Identity

Humans designing, building, and maintaining AI-assisted software projects who choose to adopt ispec to manage their project's documentation. They are the primary editors of an ispec project's canon and the operators of the workflows ispec defines.

## Goals

- Produce canonical documentation that survives iterative, prompt-driven development.
- Have the documentation be accurate at any point in the project's lifetime without re-reading the entire codebase.
- Enable future regeneration, training, and user-doc production from the same source material.
- Use AI assistance throughout — for both authoring and maintaining cards — without the AI silently corrupting the canon.

## Constraints

- Limited time to maintain documentation; the system must work in parallel with implementation rather than gating it.
- Expect the workflow to align with existing AI-assisted tooling (Claude Code, MCP, skills).
- Prefer markdown-and-text artifacts over proprietary formats so the canon is portable and PR-reviewable.

## Communication

Direct interaction with the AI agent in the project's working tree. Configuration via the project repo (skills, lint config, schemas).
