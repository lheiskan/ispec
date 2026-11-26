---
name: ispec-defer
description: Use this skill when the user wants to explicitly defer work surfaced during an `ispec-start-session` — typically to force smaller scope than the agent proposed, or to capture discovered work for later without implementing it now. Captures deferred items as cards with cascading status updates across linked requirements and ADRs.
---

# ispec-defer

Captures work for later. Invoked when the user wants to force smaller scope than the agent proposed, when investigation surfaces follow-on work that is separable from the current session, or when a feature is being explicitly removed from the active set.

## When to invoke

- User says "defer X" or "let's leave X for later" during an `ispec-start-session` contract negotiation.
- Agent proposed a multi-feature scope and user wants to narrow to one.
- During implementation, a follow-on feature surfaces that the user does not want to do now.
- Standalone — user wants to defer an existing active feature that has not yet been implemented.

## Process

### 1. Identify items to defer

Confirm with the user which items to defer. Each item is one of:

- A **feature** card (status `proposed` or `active`).
- A **draft requirement** that has not yet been ratified (no satisfying implemented feature).
- A **draft ADR** that has not yet been ratified (status `proposed`).

If the user names an item ambiguously, search and confirm the specific stable ID before proceeding.

### 2. Cascade status updates

For each deferred item, apply the cascade:

- **Feature** → set `status: deferred`. Add `discovered_during: <session-id>` if not already set (this is the session that surfaced or de-scoped it).
- **Linked requirement** → check whether any other feature with status `active` or `implemented` still satisfies it. If not, the requirement's *satisfaction* becomes `unsatisfied`. The requirement card itself stays — it is still a real need.
- **Linked ADR** → if the ADR was specifically about how to implement this feature and is not yet realized in code, demote `accepted` → `proposed`. If the ADR captures a broader decision that survives independently, leave it `accepted`.
- **Draft requirement / ADR** that the user is also choosing to defer along with the feature → stay `proposed`, retain the link to the deferred feature, retain `discovered_during`.

The reverse cascade runs when the deferred work is later picked up: feature `deferred` → `active`, ADR `proposed` → `accepted`, requirement satisfaction recomputes.

### 3. Confirm with the user

Show the deferred set and the resulting status changes:

> "Deferring:
> - F-019 (key rotation procedure) → `deferred`
> - ADR-0017 (rotation cadence) → demoted `accepted` → `proposed`
> - REQ-23 (operator can rotate secrets) → satisfaction now `unsatisfied`
>
> Confirm?"

Wait for explicit user confirmation before writing the changes.

### 4. Update the session card

Add the deferred items to `deferred_cards` in the active session card, each with a one-line rationale:

```yaml
deferred_cards:
  - id: F-019
    rationale: separable ops procedure, not blocking initial deploy
  - id: ADR-0017
    rationale: tied to F-019, demoted to proposed
```

If there is no active session (the user is invoking `ispec-defer` standalone), open a one-off session card to record the deferral and close it immediately.

### 5. Confirm the deferred work is visible

Deferred cards *are* the backlog — there is no document to regenerate
(ADR-0006). Run `ispec status` and confirm the newly deferred items appear
under its deferred list. That, plus `discovered_during:` on each card, is what
lets the user find this work again when starting a new session.

If a deferred item does not show up, the card's `status:` is wrong — fix the
card rather than reporting the deferral as complete.

## Notes

- **Deferred is not committed-to.** A deferred card is a discovery, not a promise. The user can later archive it as `abandoned` if the discovery turns out not to be needed. Status transitions should be cheap.
- **Research is preserved.** A deferred feature with associated draft requirement and ADR keeps the work that has already been done visible. Tentative decisions stay in `proposed` state without pretending to be canon.
- **Defer cascades, undefer cascades back.** When a deferred feature is later activated, the agent should run the reverse cascade — promote the linked ADR back to `accepted` (after implementation), recompute requirement satisfaction.

## Anti-patterns to resist

- **Deferring without surfacing the cascade.** The user should see what other cards are affected, not just the feature. Hidden cascades produce inconsistent status across linked cards.
- **Pushing back on user-chosen scope.** If the user explicitly asks to defer, the agent should defer cleanly and not relitigate. Surface concerns once if relevant ("note: F-019 covers a security-relevant gap"), then defer.
- **Losing the discovery context.** Always set `discovered_during` and add a rationale. Without context, a deferred card is just a name on a backlog and rots into "what was this again?"
