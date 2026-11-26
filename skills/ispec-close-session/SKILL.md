---
name: ispec-close-session
description: Use this skill at the natural end of an `ispec-start-session` — when the work is committed, tests pass, deploy (if applicable) succeeded, and the agent or user judges that the unit of work is complete. Verifies alignment between code and canon, surfaces deferred items captured during the session, and marks the session card resolved. The retrospective counterpart to `ispec-start-session`.
---

# ispec-close-session

Closes a session opened by `ispec-start-session`. Performs final reconciliation, marks the session card resolved, and surfaces any deferred items for follow-up.

## When to invoke

At the natural end of a session — usually after `ispec-start-session` step 10 has suggested closing. Trigger conditions:

- All planned work is committed.
- Tests pass (where applicable).
- Deploy succeeded (where applicable).
- No open threads (TODO comments left in code that should not stay, draft cards in canon that have not been confirmed).

The agent should *suggest* closing when the conditions are met but never auto-close. The user confirms.

## Process

### 1. Verify alignment

Before closing, run a final pre-completion reconciliation pass:

- Diff the committed implementation against the feature / spec definitions one more time.
- Confirm every affected card is in its terminal state for this session (features `implemented` or `deferred`, ADRs `accepted` or `proposed`).
- Confirm there are no draft cards left in an inconsistent state (e.g. an ADR drafted but never linked to a feature).

If any divergence remains, return to `ispec-start-session` step 8 and resolve it before closing.

### 2. Update the session card

Fill in the session card's terminal state:

- `outcome`: a short paragraph describing what was achieved, in the user's domain language. Not a changelog of file edits — a description of the change that future agents and humans will read to understand why this session existed.
- `affected_cards`: list of every card created or amended during the session, with stable IDs.
- `deferred_cards`: list of every card captured as deferred, with stable IDs and a short rationale per item.
- `commits`: the git SHAs landed during the session (look at `git log` since the session opened).
- `status: resolved`.

Before resolving, check the "Discoveries during the session" (or "Discoveries about ispec" for bootstrap / retrofit sessions) body section. If it's empty, ask the user once whether anything felt friction-y, surprising, or wrong about ispec while doing this work — even minor observations belong there. The dogfood loop only works if findings flow back.

If the session was abandoned (work started but explicitly dropped without merging), set `status: abandoned` instead and capture the reason in `outcome`.

### 3. Surface deferred items

If the session captured deferred items, summarize them for the user:

> "These items were deferred during this session:
> - F-019 (sops/age key rotation procedure) — discovered while implementing secret management; ops-side, separable.
> - F-020 (admin UI for key rotation) — separable, lower priority.
>
> Would you like to schedule any of them next, or keep them deferred?"

Do not push. The user decides what to do; the surfacing is the point.

### 4. Confirm closure

Ask the user explicitly: "Mark session `<slug>` as resolved?" Wait for confirmation. The session card is updated to `status: resolved` only on user approval.

If the user wants to keep the session open for further work, leave it open and return control. The session can be closed later.

### 5. Optional: capture cross-project lessons

If the session involved non-trivial problem solving, debugging, or decisions that surface insights useful beyond this project, suggest recording them wherever the user keeps cross-project notes (a personal knowledge store, a team wiki). Project canon captures *what is* and *why for this project*; cross-project notes capture *what was learned that applies anywhere*. They are complementary, and ispec deliberately does not try to be the second one.

## Anti-patterns to resist

- **Auto-closing.** The user should always confirm closure. Auto-close is how sessions get marked resolved with unresolved threads.
- **Padding the outcome.** A one-paragraph honest summary is more useful than a bullet list of every file touched. Future readers want the *what and why* of the change, not the diff.
- **Closing with deferred items unsurfaced.** The whole point of capturing deferred cards is for them to come back into view. Surface them on close.
