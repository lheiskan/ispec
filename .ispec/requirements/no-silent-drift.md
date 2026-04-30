---
type: requirement
stable_id: REQ-2
slug: no-silent-drift
stakeholders: [STK-1, STK-2]
satisfaction: partially-satisfied
---

## Statement

The system shall keep canon accurate as the codebase evolves, with no silent drift between specification and implementation.

## Rationale

The biggest threat to any AI-assisted spec-driven workflow is silent canon drift: the agent quietly updates the spec to match drifted code, and the spec stops being a contract — it becomes a paraphrase of whatever the implementation currently does. Once that happens, the spec loses its value as a regeneration source, an onboarding aid, and an alignment tool.

Defense is mechanical, not aspirational: pre-completion reconciliation, explicit user notification when canon changes, lint that catches dangling references and stale state.

## Currently

Partially satisfied. The pre-completion reconciliation step is specified in the spec-session skill ([[F-2]]) and the user-notification rule is encoded there. The lint binary ([[F-1]]) catches structural drift (dangling references, missing fields). Status invariants across linked cards (e.g. a feature `implemented` whose linked ADR is still `proposed`) and `paths:` glob validation are not yet implemented.
