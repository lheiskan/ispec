---
name: ispec-start-session
description: Use this skill when starting work on any substantive change to the project — a new feature, a bug fix, a refactor, an ops change, or any prompt that would result in code being written. Walks the agent through trigger classification, capture of requirements / features / ADRs, an explicit user-approved contract, implementation, and pre-completion reconciliation. Produces atemporal canon (requirement / feature / spec / ADR cards) alongside the code, and ensures both stay aligned.
---

# ispec-start-session

The main workflow for a unit of project work. The user describes what they want; this skill frames the agent's behavior through the entire loop until the work is committed and the session is ready to close.

## When to invoke

Invoke at the start of any unit of work that would result in code or canon changes. The skill itself classifies the work and adapts — there is no separate skill for bugs vs. features vs. refactors.

Skip only for trivial chores that touch nothing in canon (formatting, dependency bump that does not change behavior, comment fix). When in doubt, invoke — the classification step is cheap.

**For brand-new projects:** if the repo has no ispec canon yet (no `.ispec/stakeholders/`, `.ispec/requirements/`, etc.), do not run this skill. Hand off to the `ispec-bootstrap` skill instead — bootstrap explicitly defers canon authoring until the first working slice exists, which is what a brand-new project actually needs.

## Process

### 1. Classify the trigger

Read the user's prompt and search existing cards. Decide which classification best fits:

- **Bad implementation** — code diverges from existing spec. No canon change. The session links to the existing spec section; the commit references it.
- **Missed corner case** — spec was incomplete. Amend the relevant spec section, or add a corner-case card under the relevant feature, whichever is cleaner.
- **New feature, no existing requirement** — draft a requirement card and a feature card linked to it.
- **New feature, requirement exists** — draft a feature card linked to the existing requirement.
- **Refactor or non-functional change** — usually no requirement / spec change. If a new design choice is being made, draft an ADR.
- **Operational change** — operator-side requirement / spec / ADR amendment. Same shape, audience tagged `operator`.

If unsure between two classifications, ask the user briefly. Do not silently pick.

### 2. Search for related context

Before drafting anything new:

- Search existing requirement cards for ones this work might already satisfy or refine.
- Search existing feature cards for ones this work amends rather than creates.
- Search deferred cards (status `deferred`) — this work may be picking one up.
- Search related stakeholder cards to confirm whose need is driving this.

Reuse existing cards in preference to creating new ones. The card graph stays cleaner when names and IDs are stable.

### 3. Open a session card

Create `.ispec/sessions/<YYYY-MM-DD>-<short-slug>.md` with frontmatter:

```yaml
---
type: session
slug: <YYYY-MM-DD>-<short-slug>
trigger: <one-line description of what initiated this work>
classification: <one of the categories above>
status: open
affected_cards: []
deferred_cards: []
commits: []
outcome: ""
---

## Trigger

(one paragraph describing what initiated this work, in the user's words)

## Pre-decisions

(tooling / scope / approach decisions the user has already made or that fall out of the contract step before coding starts)

## Discoveries during the session

(filled in at reconciliation — observations about ispec or the project that surfaced during the work, anything surprising or that didn't fit. The ispec-related ones flow back as triggers for future ispec sessions.)

## Outcome

(filled at session close)
```

The body sections accumulate state through the loop. The "Discoveries during the session" section is the load-bearing dogfood-loop mechanism for ispec — capture observations as they surface during the work, do not try to fix ispec mid-session.

### 4. Detect cross-repo work

If investigation suggests the work crosses into a sibling repo (e.g. a deployment-tooling repo, a shared-library repo, an infrastructure project) — especially one that also uses ispec — surface this *before* drafting the contract:

- Identify the dependency repo and the capability that's missing or needs amendment.
- If the dependency repo uses ispec: cross-repo session is possible. Sibling repos are configured in `.ispec/repos.yml` (if present) and use repo-prefixed stable IDs (e.g. `deploy-tool:F-004`).
- If the dependency repo does not use ispec: cross-repo canon updates aren't possible; instead the session will capture an ADR locally documenting the external dependency, and any workaround as a deferred feature with a clear "remove once X" condition.

Tell the user: "this work requires changes in <other-repo> as well as <this-repo>; we can extend this session to cover both, or defer the <other-repo> part." User decides scope.

### 5. Draft the contract

Before writing any code, present an explicit contract to the user:

- **Feature(s)** to be implemented in this session. If multiple features group naturally, that is fine; if not, propose one and surface the rest as deferral candidates.
- **Requirement(s)** the feature(s) target. Existing IDs preferred; draft new requirements only if none fit. For cross-repo work, draft requirements in the appropriate repo — upstream repos express their requirements as general capabilities (not client-specific), with `originated_during: <client-repo>:<session-id>` for traceability.
- **Spec amendments** as a diff per repo (proposed sections / changes shown explicitly).
- **ADRs** to be drafted, with the question each addresses and a one-line proposed answer. Cross-repo ADRs land in the repo where the decision is made; downstream features carry `depends_on: [<upstream-repo>:<feature-id>]`.
- **Out of scope** — what was discovered but is not being done now, with brief rationale. These become deferred cards on user approval.

Ask the user explicitly: "Does this contract reflect what you want me to do?" Wait for approval. Do not start coding before approval.

### 6. Apply canon updates to the working tree

After approval, write the requirement / feature / spec / ADR changes to the working tree. **Do not commit them yet.** Canon updates and code updates commit together once the implementation is reconciled with them.

This intermediate uncommitted state is what makes mid-coding amendments cheap. Committing canon up-front and then amending mid-coding produces confusing history.

### 7. Implement

Write the code. Direction can change during coding; if the spec needs amending, amend the (still-uncommitted) spec rather than diverging silently. Each amendment is surfaced to the user with a short note: "I'm amending SPEC §X.Y because <reason>; new text: <diff>. OK to continue?"

If during coding a substantially larger change reveals itself than the contract covered:

- Stop coding.
- Surface the discovery to the user.
- Renegotiate the contract (return to step 5) or fork into a new session.

Never quietly grow scope past what the contract authorized.

### 8. Pre-completion reconciliation

Before declaring the work done, diff the implementation against the feature / spec definitions. Three outcomes:

- **Aligned.** Proceed to commit.
- **Code is wrong.** Fix the code to match the canon.
- **Definition was wrong or incomplete.** Update the canon to match the code, AND explicitly tell the user:

  > "The implementation diverged from the original feature definition. I have updated the canon to match. Here is what changed: [diff]. Confirm?"

  The user can accept, push back, or ask to revert the code instead. Never silently update canon to match drifted code.

This is the discipline that keeps the spec meaningful. Without it, the spec rots into a paraphrase of whatever the code currently does.

### 9. Commit

Commit canon and code together. Reference stable IDs in the commit message:

```
Implement F-007 offline call logging

Targets REQ-12 (offline operation for sales reps).
Spec amendments: SPEC-call-log §3.2 (sync conflict resolution).
ADR-0014: chose last-write-wins with visible warning.
```

If the work spans multiple natural commits, keep each one referencing the IDs it touches. The commit history is part of the corpus future agents will read.

After committing, populate the session card's `commits:` field with the resulting SHAs so the session is linkable to the actual git changes.

### 10. Suggest closing the session

When the work is committed and (where applicable) tested or deployed, suggest invoking `ispec-close-session`. Do not auto-close. Let the user confirm or continue working in the same session.

## Status transitions during the session

When applying canon updates, follow these lifecycle states:

- **Feature**: created as `proposed` at contract time; transitions to `active` when coding starts; transitions to `implemented` at commit time after reconciliation.
- **Requirement**: created as a card; satisfaction is derived from the features that target it (do not write satisfaction directly).
- **ADR**: created as `proposed` at contract time; transitions to `accepted` at commit time after reconciliation confirms the implementation realizes the decision.
- **Session**: stays `open` throughout; transitions on close.

## Deferral handled mid-session

If the contract step (4) or coding step (6) surfaces work the user wants to defer, capture it inline:

- Create a feature card with `status: deferred` and `discovered_during: <this-session-id>`.
- If the deferred work has a draft requirement / ADR, preserve them as `proposed` linked to the deferred feature.
- Add the cards to `deferred_cards` in the session.

The user can also invoke the `ispec-defer` skill explicitly to force smaller scope.

## Anti-patterns to resist

- **Skipping the contract step.** Even for "small" changes, the contract is where scope is agreed. Without it, the agent silently grows or shrinks scope.
- **Committing canon before code is aligned.** Produces history that says "we said X, then immediately changed to Y."
- **Silently updating spec to match drifted code.** The pre-completion reconciliation step exists for exactly this; the user must always be told when canon changes.
- **Treating the trigger as the entire scope.** Investigation often reveals the trigger is the tip of a larger issue; surface and decompose, do not silently expand.
- **Creating new cards when existing ones fit.** Search first; reuse stable IDs and slugs where possible.
