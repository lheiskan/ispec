---
type: feature
stable_id: F-2
slug: start-session-skill
status: implemented
targets: [REQ-1, REQ-2]
paths:
  - skills/ispec-start-session/SKILL.md
---

## Description

The main session-loop skill, invoked at the start of any unit of work that may produce code or canon changes. Per [[ADR-0001]], a single skill covers all trigger types via classification, with confirmation points only at moments where user input matters.

## Behavior

The skill walks the agent through ten phases:

1. Classify the trigger (bad implementation / missed corner case / new feature / refactor / ops change).
2. Search existing cards for related context.
3. Open a session card.
4. Detect cross-repo work and surface it.
5. Draft the contract (features, requirements, spec amendments, ADRs, out-of-scope) and wait for user approval.
6. Apply canon updates to the working tree (uncommitted).
7. Implement; surface mid-coding spec amendments rather than diverging silently.
8. Pre-completion reconciliation; explicitly notify user when canon was updated to match implementation.
9. Commit canon and code together with stable IDs in the message.
10. Suggest closing the session.

## Limitations (deferred)

- The skill is hand-written and relies on the agent following its instructions; no machine-enforced gating between phases.
- Cross-repo session protocol depends on sibling repos using ispec; the non-ispec-upstream fallback (capture as ADR locally) is described but not yet exercised.
- Session card schema is defined but no automated tooling for opening, updating, or closing sessions exists yet — the agent edits files directly.
