---
name: ispec-retrofit
description: Use this skill to bootstrap ispec on an EXISTING project that has substantial code, commit history, and possibly existing documentation. Different from `ispec-bootstrap` — there is already material to read; the work is to identify the load-bearing canon hidden in the code, docs, and git history, and capture it without trying to retroactively document every feature ever shipped. Always scope retrofit explicitly to avoid an infinite project.
---

# ispec-retrofit

Initialize ispec canon for an existing project. Reads the codebase, existing docs (README, OPS.md, etc.), dependency files, and git history; proposes a deliberately-scoped initial canon set; iterates with the user.

## When to invoke

- An existing repository with code and git history but no ispec canon (no `.ispec/stakeholders/`, etc.).
- The user has asked to start using ispec on this project.
- Explicit user invocation.

If invoked on a project that already has ispec canon, refuse — hand off to `ispec-start-session` instead.

## Why retrofit is its own skill

Bootstrap (for brand-new projects) defers canon authoring until the first slice exists. Retrofit's first slice already exists — and so do the next 50 slices. The danger isn't that the canon will be wrong; it's that retroactively documenting every feature ever shipped is an infinite project that will collapse under its own weight.

The retrofit discipline is:

- Capture only **load-bearing canon** — the entities, requirements, decisions, and features that future work will refer to. Skip incidental work.
- Use existing documentation (README, OPS.md, design notes) as input but rewrite into ispec form. Don't just `git mv` — translate.
- Recover what's recoverable from git history, accept what's lost.
- Hand off to `ispec-start-session` for future work as soon as the project has enough canon to be useful, even if coverage is incomplete.

## Process

### 1. Confirm intent and scope

Ask the user:

- "How deep do you want this retrofit to go?" Options:
  - **Minimum viable**: 1 stakeholder, 1 ADR (core tooling), 1 domain-model spec section, 1-3 requirements, 3-5 feature cards covering current major surfaces. Done in one session.
  - **Standard**: above + audience-tagged spec sections covering main subsystems, major historical ADRs (pivots, framework changes), feature cards for current implemented surfaces.
  - **Comprehensive**: above + retroactive feature cards for historical work, full spec section coverage, ADRs for every load-bearing dependency.
- Default recommendation: **standard**. Comprehensive should be reserved for projects where the cost of poor canon is high (multi-author, regulated, long-lived).

### 2. Open a retrofit session card

`.ispec/sessions/<YYYY-MM-DD>-retrofit.md`:

```yaml
---
type: session
slug: <YYYY-MM-DD>-retrofit
trigger: retrofit existing project to use ispec
classification: retrofit
status: open
affected_cards: []
deferred_cards: []
commits: []
outcome: ""
---

## Reading notes

(observations about the project as the agent reads it — domain semantics, pivots in git history, what's in existing docs vs. what's only in code)

## Discoveries about ispec

(things that surprised, frustrated, or didn't fit while applying ispec to this project — schema gaps, ambiguous skill instructions, missing relations. These flow back to ispec as the trigger for follow-up spec-sessions in the ispec repo.)

## What's deliberately NOT being captured

(scope decisions: parts of the project that don't need canon at this scope, with brief rationale)

## Outcome

(filled at close)
```

The "Discoveries about ispec" section is the load-bearing dogfood-loop mechanism. Capture observations as they surface during the retrofit; do *not* try to fix ispec mid-retrofit (unless something is blocking). The findings travel back to ispec for processing in their own session.

### 3. Read the project

Before drafting any cards, read:

- **README.md, OPS.md, DESIGN.md, NOTES.md, docs/** — anything human-authored explaining what this is. These often contain the project's motivation and operational shape in concentrated form.
- **Dependency files** — `go.mod`, `package.json`, `requirements.txt`, `Dockerfile`, `docker-compose.yml`. These reveal load-bearing tooling decisions.
- **Schema / type files** — SQL migrations, ORM models, struct definitions, type files. These reveal the domain model with semantics.
- **Handler / route / API files** — these show what the system actually does, in user-facing terms.
- **Deployment config** — systemd units, Helm charts, CI workflows. These define the operator-facing surface.
- **`git log --oneline | head -50`** — recent activity to see what's actively maintained vs. legacy.
- **`git log --reverse --oneline | head -20`** — earliest commits to recover original intent.
- **Pivots in the log** — search commit messages for "rewrite", "rename", "switch from", "drop", "replace". These are the highest-value historical ADRs.

Take notes mentally; don't write cards yet. The reading phase is to build a holistic picture.

### 4. Propose initial canon

For **minimum viable** scope:

- **One stakeholder card** for the primary user / customer. If the project has obvious operator stakeholders (e.g. self-hosted infrastructure), add one operator stakeholder card too.
- **ADR-0001: core technology choices** — bundle the load-bearing 2-5 tooling decisions visible in dependency files and deployment config. One short paragraph per decision.
- **A domain-model spec section** capturing entities, relationships, and non-obvious semantics. This is the highest-value retrofit artifact.
- **1-3 requirement cards** stating the high-level "the system shall X for stakeholder Y" — drawn from the README's motivation section if it exists, or distilled from what the code clearly does.
- **3-5 feature cards** for the project's current major surfaces. Each targets a requirement; each carries a `paths:` glob pointing at the implementation.

For **standard** scope, additionally:

- **Audience-tagged spec sections** for each major subsystem (e.g. auth, payments, observability). Use `audience: end-user` for user-facing behavior, `audience: operator` for ops surfaces.
- **Major historical ADRs** — pivots and framework changes recovered from commit history, with `status: accepted` (or `superseded` if later overridden). Don't try to recover every decision; aim for the 3-5 most consequential.
- **Feature cards for currently-implemented surfaces** beyond just the major ones. Skip features that are obviously stable and uncontroversial; focus on ones that future work might touch.

For **comprehensive** scope: continue extending until coverage feels complete enough for the project's risk profile.

### 5. Translate existing docs

If the project has existing documentation (README, OPS.md, etc.), translate sections into ispec cards rather than leaving the docs as-is:

- A README's "What this is" section → motivation in stakeholder card or requirement card.
- A README's "Architecture" section → spec sections.
- An OPS.md's deployment/operations content → operator-audience spec sections, possibly an ADR or two.
- Migration / pivot notes → ADRs with `status: superseded` or `accepted`.

Once the canon is drafted, the original docs can be retired, simplified, or pointed at the canon (e.g. README becomes a 50-line pointer to `.ispec/`).

### 6. Iterate with user, one card at a time

For each proposed card:

- Show the agent's draft.
- Get user edits or confirmation.
- Write the card to disk.
- Move to the next.

Do not batch — the agent's drafts will need correcting on domain semantics that aren't obvious from code, and the per-card dialogue is where that correction happens.

### 7. Run lint

After all cards are drafted, run `ispec lint`. Fix any failures (most commonly: dangling references where one card mentioned another that wasn't drafted yet).

### 8. Close the retrofit session

Update the session card:

- `affected_cards`: every card created.
- `commits`: the git SHAs landed during this retrofit (look at `git log` since the session opened).
- `outcome`: paragraph summarizing what was retrofitted, what scope was chosen, what was deliberately *not* captured (and why).
- `status: resolved`.

Confirm the "Discoveries about ispec" section is non-empty if there were any observations during the retrofit — even minor ones. If it's empty, ask the user one more time whether anything felt friction-y; ispec only improves if findings make it back.

Tell the user: "Retrofit complete at `<scope>` level. The project has initial canon and graduates to steady-state. Future work will use the `ispec-start-session` skill, which can lazily fill in additional canon as needed."

### 9. Suggest follow-ups

Common deferred items from retrofit:

- "Audience-tagged spec sections for the X subsystem" (if minimum viable was chosen).
- "ADRs for historical pivots in subsystem Y" (if not recovered during retrofit).
- "Feature cards for legacy surfaces in module Z" (if scope was minimum viable / standard).

Capture these as deferred feature cards or as a list in the session outcome, so the user has a roadmap for filling in canon as work touches each area.

## Anti-patterns to resist

- **Trying to document everything.** Retrofit is not "write the project's full history into ispec." It is "capture enough load-bearing canon that future work has a reference." Default to under-capturing; expand lazily.
- **Cargo-culting existing doc structure.** A README's section organization is for humans skimming; ispec's organization is for AI agents loading just-the-relevant-cards. Translate, don't copy.
- **Inventing requirements that nobody actually has.** If you can't tie a requirement to a stakeholder who genuinely benefits from it, drop it. The retrofit will look incomplete, and that's honest.
- **Treating recovered historical ADRs as having full alternatives sections.** Often you only have "the choice was made"; you don't know what alternatives were considered. That's fine — write the ADR with what you can recover, mark `status: accepted`, move on. Pretending to have full context is worse than admitting you don't.
- **Skipping lint to ship faster.** Lint catches dangling references that turn into rot within weeks. Run it.

## Notes for the agent

- The retrofit dialogue is the highest-stakes domain learning the agent will do on this project. Listen carefully when the user corrects domain language — that correction is canon.
- Retrofit is a one-shot session. It should not span multiple days of work. If it does, scope was wrong; cut it shorter.
- Use existing project docs as ground truth for the project's own vocabulary. If the README calls something a "container" and the code calls it a "storage_unit", ask the user which is canonical and use that consistently.
