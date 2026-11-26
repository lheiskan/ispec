---
name: ispec-bootstrap
description: Use this skill when starting a brand-new project that intends to use ispec. Bootstrap runs in three phases — a Phase 0 spike where the first working slice is built without ispec canon, a Phase 1 retrospective capture where the load-bearing canon is authored once the shape is clear (typically after a few commits), and graduation to Phase 2 (the steady-state `ispec-start-session` loop). Do NOT force complete canon authoring before the first slice exists; the canon will be wrong if written before there is code to read.
---

# ispec-bootstrap

Initialize a new ispec-using project. Adapts to the empirical reality that most projects ship a working feature in their first commit and only have enough definition to write canon 2-7 days later.

## When to invoke

- Starting a brand-new repository that intends to adopt ispec.
- An existing repository without `.ispec/stakeholders/`, `.ispec/requirements/`, etc., where the user wants to begin using ispec.
- Explicit user invocation.

If invoked on a repo that already has ispec canon, do not overwrite — instead, hand off to `ispec-start-session` (the project is past bootstrap).

## Phase 0 — Spike (day 1)

Goal: ship the first working slice. Do not gate it on ispec canon. The single artifact at this stage is the bootstrap session card.

### 1. Capture the trigger

Have a brief dialogue with the user about:

- What is being built and for whom (one paragraph in the user's words).
- The motivating problem — what makes this worth doing.
- Any pre-decided tooling or constraints (language, deployment, must-use libraries).
- Open tooling questions where the user wants the agent's recommendation.

If the user offers tooling pre-decisions, accept them; do not relitigate. If they ask the agent to weigh in, surface 1-2 alternatives with one-line tradeoffs and let the user choose.

Output of the dialogue goes into the bootstrap session card body. Do NOT write a separate `DESIGN.md` or `SPEC.md` at this stage — it will be wrong.

### 2. Open the bootstrap session card

Create `.ispec/sessions/<YYYY-MM-DD>-bootstrap.md`:

```yaml
---
type: session
slug: <YYYY-MM-DD>-bootstrap
trigger: <one-line description of what is being built>
classification: bootstrap
status: open
affected_cards: []
deferred_cards: []
commits: []
outcome: ""
---

## Motivation
<paragraph from the user, in their own words>

## Pre-decisions
<list of tooling / architecture decisions the user has already made>

## Open questions (resolved during this session)
<list of questions the agent raised and how they were resolved>

## Discoveries about ispec
(filled as the bootstrap progresses — things that surprised, frustrated, or didn't fit while applying ispec to this project. Flow back as triggers for follow-up sessions in the ispec repo.)

## Outcome
(filled at close)
```

### 3. Implement the first working slice

Build the first end-to-end working version. This typically includes:

- The data model (the most important load-bearing artifact).
- Whatever handlers, UI, or interfaces are needed for it to do *something* observable.
- Minimum viable deployment if applicable (Dockerfile, systemd unit, build script).

Commit. The first commit message should be descriptive enough that a human reader can tell what was built ("v1: payment tracking with bank statement upload" beats "initial commit"), since this commit will become the basis for the Phase 1 canon authoring.

### 4. Iterate as needed

Continue with normal coding sessions for as long as the shape isn't yet clear. There is no fixed cutoff — Phase 0 ends when the project has enough definition to write canon that won't immediately be wrong. Common signals:

- The data model has stabilized (no major schema rewrites in the last few sessions).
- The first user-facing flow works end-to-end.
- A new contributor (human or AI) joining now would have to read the code to understand what the system does.

When those signals appear, transition to Phase 1.

## Phase 1 — Initial canon

Goal: capture the load-bearing canon retrospectively, once the shape is clear. This is one focused session, not an indefinite curation effort.

### 5. Read the code

Before drafting cards, read what's actually been built:

- Dependency files (`go.mod`, `package.json`, `requirements.txt`, `Dockerfile`) for the load-bearing tooling decisions.
- The data model (SQL schema, struct definitions, type files) for the domain entities and their semantics.
- Handler / route files for what the system actually does.
- Any deployment scaffolding for operational shape.
- Recent commit messages for any pivots or non-obvious choices.

Note especially: domain semantics that aren't obvious from field names, tooling choices that survived an alternative being considered, anything the agent had to ask "why is this like this?" while reading.

### 6. Draft the initial canon

Propose the following cards to the user, one at a time, with content drafted from the code review:

- **One stakeholder card** for the primary user / customer. Identity, goals, constraints, communication channel.
- **ADR-0001: core technology choices.** A single ADR covering the 2-5 load-bearing tooling decisions visible in dependency files. One short paragraph per decision: what was chosen, what alternatives existed, why this one. Mark `status: accepted`.
- **A domain-model spec section** describing the entities, their relationships, and the meanings that aren't obvious from field names. This is the highest-value card in the bootstrap. `audience: developer` (or `developer + admin` if some entities surface in the UI).
- **1-3 requirement cards** stating the high-level "the system shall X for stakeholder Y." Drafted from the work that has actually been done, not from imagined scope. The bootstrap session card's motivation paragraph is the source material.
- **Feature card(s) for the work already landed.** Retroactive but cheap because the work is fresh. Each feature targets one of the new requirements; each carries a `paths:` glob pointing at the files that implement it.

Present each card to the user, get confirmation or edits, write to disk. Do NOT batch — the dialogue per card is where the agent's draft gets sharpened by the user's actual understanding.

### 7. Lint

Run the consistency lint:

- Every feature has at least one `targets` requirement.
- Every spec section has at least one audience tag.
- Every `[[wikilink]]` resolves.
- Every `paths:` glob matches real files.

Fix any failures.

### 8. Close the bootstrap session

Update the session card:

- `affected_cards`: list of all cards created in Phase 1.
- `commits`: the git SHAs landed during Phase 1 (look at `git log` since the session opened).
- `outcome`: one paragraph summarizing what was bootstrapped.
- `status: resolved`.

Confirm the "Discoveries about ispec" section is non-empty if there were any observations during Phase 1 — even minor ones. If it's empty, ask the user one more time whether anything felt friction-y; ispec only improves if findings make it back.

Tell the user: "Bootstrap complete. The project now has initial canon and graduates to steady-state. Future work will use the `ispec-start-session` skill."

## Phase 2 — Steady state

After bootstrap closes, normal `ispec-start-session` applies. Existing features get more detailed cards lazily as work touches them; additional stakeholders, requirements, and ADRs are added when relevant.

## Anti-patterns to resist

- **Authoring complete canon before the first slice exists.** The canon will be wrong. Wait until there's code to read.
- **Skipping Phase 1 because "it's still early."** The transition from one-off code to maintainable system happens silently. If Phase 1 is indefinitely deferred, every new contributor will reverse-engineer domain semantics from code. Capture once, refine lazily.
- **Bootstrap as a checklist.** The goal is load-bearing canon, not full coverage. One stakeholder, one ADR, one domain-model spec, 1-3 requirements, and matching features is enough to graduate.
- **Writing a polished README in Phase 0.** It will be wrong. The README is generated (or hand-written) from the canon in Phase 1 or later, when the shape is known.
- **Ignoring early pivots.** When the project changes direction in the first weeks, capture an ADR with `supersedes:` to the prior ADR. The week-1 pivot is the highest-value ADR a project ever writes.

## Notes for the agent

- Be conversational in Phase 0. The user is still figuring out what they want; clarifying questions are useful, but excessive ceremony will frustrate them.
- Be precise in Phase 1. The cards being drafted will become the canonical reference; sloppy domain language will calcify and hurt later.
- The Phase 0 → Phase 1 transition is judgment-based, not algorithmic. The user can ask for it explicitly ("let's bootstrap the canon now"); the agent can suggest it ("the data model has been stable for a few sessions — want to capture initial canon?"). Either is fine.
