---
type: requirement
stable_id: REQ-1
slug: durable-canonical-documentation
stakeholders: [STK-1]
satisfaction: partially-satisfied
---

## Statement

The system shall enable production of canonical project documentation — requirements, decisions, specifications, stakeholders — that is *atemporal* (describes what the project *is*), independent of the order in which work was done.

## Rationale

Prompt-driven development produces a chronological narrative by default: prompts, code edits, commit messages. Downstream uses — regeneration, agent training, user-doc production — need an atemporal reference. Without an explicit canon, the project drifts into "code is the spec" territory, and consumers must reverse-engineer intent from implementation.

## Currently

Partially satisfied. The artifact taxonomy and skills exist; the lint binary checks structural integrity; the canon directory is in place. View generation (which would aggregate the canon into traditional documents like a full SPEC.md) is not yet implemented.

See [[F-1]] for the lint binary and [[F-2]] for the spec-session skill that drives canon authoring.
