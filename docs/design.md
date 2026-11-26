# ispec — design

A system for capturing and maintaining software requirements, specifications, and decisions in AI-assisted software development projects. The goal is to evolve a project's canonical documentation in parallel with the code, so that an agent holding the canon, the git history, and the source has the context it needs to reimplement parts of the system, explain why it works the way it does, or produce user-facing and operator-facing documentation.

ispec captures that context; it does not produce those downstream artifacts. See "Consumers of the canon" below and ADR-0006.

## Motivation

Prompt-driven development tends to produce a chronological narrative (how the design evolved) rather than an atemporal reference (what the design is). The atemporal reference is what every downstream use needs: regeneration, training, user docs, operator runbooks, API documentation. ispec is an opinionated set of artifacts and workflows that keep the atemporal reference accurate as the system grows — so that whatever consumes it is working from something true.

## Core artifacts

Six types of artifact, all stored as markdown cards with frontmatter and `[[wikilinks]]` between them. Same primitives as a Zettelkasten store, but committed to the repo as project canon.

### Stakeholders

A stakeholder is anyone affected by or interacting with the system. Broader than "users":

- End users (one card per persona)
- Operators (SREs, on-call, deployment automation)
- Support / customer success
- Developers maintaining or extending the system
- Compliance / legal / regulatory bodies
- Third-party integrators and APIs depended on
- The AI agents working on the codebase
- Affected non-users (data subjects, regulators)

Each stakeholder card carries: identity, goals / jobs-to-be-done, constraints, communication channel, and links to the requirements they drive.

Surfaced separately from requirements because:

- They evolve on a different cadence (a new stakeholder is a bigger event than a requirement tweak).
- They drive audience tags on spec sections, which record who a section is written for.
- They make "who would push back if X changed" explicit.

### Requirements

A requirement describes the problem in domain language. It answers "should we build this? is this still the right thing?" — independent of solution.

Frontmatter: `type: requirement`, `stable_id: REQ-N`, `slug`, `stakeholders: [REQ-driving stakeholders]`, `satisfaction: unsatisfied | partially-satisfied | satisfied`.

Satisfaction is *derived* from the features that target the requirement, not authored directly.

### Features

A feature is a unit of solution work. It targets one or more requirements and is the level at which work is planned, implemented, and reconciled.

Frontmatter: `type: feature`, `stable_id: F-N`, `slug`, `targets: [REQ-N, ...]`, `status: proposed | active | implemented | deferred | abandoned`, `discovered_during: <session-id>` (optional, for items captured mid-session).

The `[[wikilink]]` between feature and requirement is the load-bearing edge. Linting checks every feature has at least one target.

### Spec sections

A spec section describes part of the solution precisely enough to be implemented from scratch. Atemporal — describes what *is*, not how it got there. Organized by concept (auth, payments, observability), not by feature delivery order.

Frontmatter: `type: spec`, `stable_id: SPEC-N` (or hierarchical, e.g. `SPEC-auth.refresh`), `slug`, `audience: [end-user | operator | developer | admin]`, `paths: [glob, ...]` (which code paths implement this section), `satisfies_features: [F-N, ...]`.

The `paths` field plus stable IDs make the bidirectional code↔spec link grep-able and lintable.

### ADRs

An Architecture Decision Record captures *why* a decision was made: context, alternatives considered, decision, consequences. ADRs span audiences (a single decision often affects product, ops, and developers) and are kept in a flat directory.

Frontmatter: `type: adr`, `stable_id: ADR-NNNN`, `slug`, `status: proposed | accepted | superseded | deprecated`, `supersedes: [ADR-...]` (optional), `relates_to: [feature/spec/requirement IDs]`.

A `proposed` ADR is one where the decision has been drafted but not yet realized in code; `accepted` is realized.

### Sessions

A session is one iteration of the dev loop: a trigger (new requirement, bug, refactor, ops issue), the work that addressed it, and the outcome. The session is the chronological record; it does not live in canon, it documents the path the canon took.

Frontmatter: `type: session`, `slug: <YYYY-MM-DD>-<short-description>`, `trigger`, `classification`, `status: open | resolved | deferred | abandoned`, `affected_cards: [...]`, `deferred_cards: [...]`, `outcome`.

Sessions form the chronological narrative; canon (requirements, features, spec, ADRs, stakeholders) forms the atemporal reference.

## Data model

All artifacts are markdown cards in a flat-ish layout under `.ispec/`, navigable as a graph.

```
.ispec/
  stakeholders/<slug>.md
  requirements/<slug>.md
  features/<slug>.md
  spec/<slug>.md            # one card per concept, can be hierarchical
  adr/<NNNN>-<slug>.md
  sessions/<YYYY-MM-DD>-<slug>.md
  index.md                   # curated entry-point map (concept → cards)
```

The `.ispec/` location mirrors the `.git/` / `.claude/` convention: tool-managed source-of-truth, hidden from default human listings, parsed by tooling. `docs/` remains the project's own — human-authored long-form content (this design doc, worked examples, README), plus anything the project derives from the canon using other tools. The default canon directory can be overridden with `ispec lint -canon-dir <subdir>` if a project needs a different layout.

Properties:

- **Stable IDs** (`REQ-7`, `F-12`, `ADR-0014`, `SPEC-auth.refresh`) are the durable cross-reference primitive. Linting catches dangling refs.
- **Slugs** are the human-readable name, used for filenames and casual reference. A `supersedes:` field handles slug renames cleanly.
- **`[[wikilinks]]`** for prose-level cross-references. Cards link freely.
- **Audience tags** on spec sections record who the section is for, for the benefit of whatever reads it.
- **Status fields** drive lifecycle and lint.

## Consumers of the canon

The cards are the output, not an intermediate representation. ispec does not
generate documents from them — that is out of scope per ADR-0006, and
deliberately so: the valuable and hard-to-replace part of ispec is being
present when a decision is made and getting it recorded accurately. Turning
an accurate record into readable prose is something a general-purpose agent
already does well, *given* an accurate record.

So the question is not what ispec emits, but what a consumer can rely on:

- **Every card is independently readable.** Markdown with YAML frontmatter,
  no project-specific encoding, meaningful on its own.
- **References resolve.** `ispec lint` fails on a dangling stable ID or
  wikilink, so a consumer walking the graph will not hit a broken edge.
- **Audience is declared.** Spec sections carry `audience:` tags, so a
  consumer producing user-facing output can filter to what is meant for
  users. ispec validates the tags; acting on them is the consumer's job.
- **Code is linked both ways.** Spec sections carry `paths:` globs, so a
  consumer can find the implementation of a concept, or the concept behind a
  file.
- **Intent is separated from history.** The five atemporal card types say
  what the project is; session cards say how it got that way. A consumer
  wanting current state does not have to read the narrative to find it.
- **Deferred work is enumerable.** Features with `status: deferred` carry
  `discovered_during:`, and `ispec status` lists them. No backlog document is
  emitted because the cards already are one.

What a consumer does with this — a full `SPEC.md`, a user manual, an operator
runbook, API docs, an onboarding brief — is up to the consumer.

## The session loop

Skill-driven, not slash-command-driven. The agent classifies the trigger and walks the loop, asking for confirmation at the points that matter.

```
1. Trigger arrives (user prompt: feature, bug, refactor, ops issue)
2. Classify the trigger:
   - Bad implementation (code diverges from spec) — no canon change
   - Missed corner case (spec was incomplete) — amend spec section / add corner-case card
   - New feature, no requirement — draft requirement + feature card
   - New feature, requirement exists — draft feature card linked to requirement
   - Refactor / non-functional — usually no canon change; sometimes an ADR
3. Search existing cards for related context (avoid duplicating, reuse stable IDs).
4. Open a session card.
5. Draft the contract:
   - What feature(s) will be implemented
   - What requirement(s) they target
   - What spec amendments are proposed (as diff)
   - What ADRs will be drafted
   - What is explicitly out of scope (deferred, with rationale)
6. Wait for user approval of the contract.
7. Apply canon updates to the working tree (uncommitted).
8. Implement.
   - Direction can change. Mid-coding spec amendments amend the (still-uncommitted) canon.
   - Surface every amendment to the user; never diverge silently.
9. Pre-completion reconciliation:
   - Diff implementation against feature definition.
   - Aligned: proceed.
   - Code is wrong: fix the code.
   - Definition was wrong / incomplete: update the canon AND tell the user
     "the implementation diverged from the original feature definition;
     I've updated it. Here's what changed: [diff]. Confirm?"
10. Commit canon and code together. Reference stable IDs in the commit message.
11. Suggest closing the session (`ispec-close-session` skill).
```

The "tell the user when the definition changed" step is the discipline that keeps the spec meaningful. Without it, the agent silently updates spec to match whatever the code does, and the spec stops being a contract.

## Status lifecycles

Each artifact type has its own lifecycle. Status fields are what consistency lint operates on.

- **Stakeholder**: `active | retired`
- **Requirement satisfaction** (derived): `unsatisfied | partially-satisfied | satisfied`
- **Feature**: `proposed | active | implemented | deferred | abandoned`
- **ADR**: `proposed | accepted | superseded | deprecated`
- **Session**: `open | resolved | deferred | abandoned`

Cross-status invariants the lint catches:

- A feature marked `implemented` whose linked ADR is `proposed` → inconsistent.
- A requirement marked `satisfied` with no `implemented` feature targeting it → inconsistent.
- A spec section whose `paths` glob matches no files in the repo → stale.
- A `[[wikilink]]` to a non-existent slug or stable ID → dangling.

## Defer mechanism

Discovery during a session is normal: a bug investigation reveals that a structural gap is the real cause; a feature design surfaces follow-on features that don't fit this iteration. Captured deferred items become a backlog as a side effect — no separate tracker.

Defer is a status transition with cascading effects:

- **Feature** → `deferred`. Carries `discovered_during: <session-id>`.
- **Linked requirement** → if no other active or implemented feature satisfies it, satisfaction becomes `unsatisfied`. The requirement itself is preserved (it's still a real need).
- **Linked ADR** → if specifically about implementing this feature and not yet realized in code, demote `accepted` → `proposed`. ADRs that capture broader, surviving decisions stay `accepted`.

The reverse cascade runs when a deferred item is later picked up.

A deferred card with associated draft requirement and ADR preserves the research that was done — tentative decisions stay visible without pretending to be canon.

`/ispec-defer` (or a thin skill) is the explicit user-facing command to force smaller scope when the agent's grouping instinct is more ambitious than the user wants right now. The agent is also instructed to surface deferral candidates proactively when the contract step reveals more work than the trigger suggested.

## Lint rules

A `spec-lint` step (CI or pre-commit) checks:

- Every stable ID referenced exists.
- Every feature has at least one `targets` requirement.
- Every spec section has at least one `audience` tag.
- Status invariants across linked cards (see above).
- `paths` globs in spec sections match real files.
- Every commit touching code referenced by a spec section also touches that spec section, OR carries an explicit `spec-unchanged: <reason>` trailer (escape hatch for trivial refactors).

The last one is the most aggressive rule and the one that most prevents drift. Use with judgment.

## Skill-driven workflow

Skills frame the agent's behavior throughout. The user works normally; skills steer.

- **`ispec-start-session`** — the main loop. Triggers when work begins on a substantive change (feature, bug, refactor, ops issue). Walks classification, contract, implementation, reconciliation.
- **`ispec-close-session`** — the retrospective counterpart to `ispec-start-session`. Runs at natural session boundaries. Verifies alignment, marks session resolved, surfaces deferred items.
- **`ispec-defer`** — explicit user command to capture discovered items for later, with cascading status updates.

Skills compose. Phase boundaries are real but internal to the agent's process, not user-visible knobs.

## Cross-repo sessions

Many real units of work cross repository boundaries: a bug or feature in a downstream app surfaces a need against an upstream library, infrastructure tool, or platform. If both repos use ispec, the session can span them; if not, the gap is captured locally as an external-dependency ADR.

### Repo-prefixed identifiers

Stable IDs are scoped per repo, but cross-repo references add a repo prefix:

- Wikilinks: `[[<repo>:<stable-id>]]` or `[[<repo>:<slug>]]`, e.g. `[[deploy-tool:F-004]]`.
- `affected_cards` entries: `[web-app:F-007, deploy-tool:REQ-15, deploy-tool:F-004]`.
- `depends_on` edges: `depends_on: [deploy-tool:F-004]` on a downstream feature.

Sibling repos are configured in `.ispec/repos.yml` (or equivalent) so the lint step can verify cross-repo refs the same way it verifies in-repo ones.

### Session card lives in the originating repo

The session card lives where the trigger arrived. Its `affected_cards` lists cards from every repo touched. This keeps the chronological narrative single-sourced even when canon updates land in multiple repos.

### Upstream tools see downstream apps as stakeholders

A platform / library / infrastructure tool's stakeholder set includes "client applications" — the apps that depend on it. When a session in a client app surfaces a need against the upstream, the upstream's stakeholder list is amended (or an existing entry referenced) and a requirement is drafted in the upstream repo expressing the *general* need, not the client-specific instance. The originating session is traced via:

```yaml
originated_during: <repo>:<session-id>
```

This lets future contributors to the upstream repo answer "why do we have this requirement?" without needing the client repo's history.

### Dependency edges in canon

Cross-repo dependencies are first-class in the canon, not just implicit in the build:

- A downstream feature carries `depends_on: [<upstream-repo>:<feature-id>]`.
- Lint flags dangling cross-repo refs the same way it flags in-repo ones.
- When the upstream feature changes status (e.g. `accepted` → `superseded`), downstream features that depend on it surface in lint output.

### Session protocol when cross-repo work surfaces

During the cross-repo detection step (`ispec-start-session` step 4) and contract step (step 5), if investigation reveals work in another repo:

1. The agent surfaces the cross-repo nature explicitly: "this requires changes in the deploy tool as well as the app; we can extend this session to cover both, or defer the deploy-tool part."
2. User decides scope.
3. If extended, the contract grows to cover canon updates in both repos. The agent applies updates to each repo's working tree; commits per repo (git commits don't span repos); each commit references the same originating session ID.
4. The session card in the originating repo collects the cross-repo commit refs in `affected_cards`.
5. Pre-completion reconciliation runs across both repos before close.

### Non-ispec upstreams

When the upstream is an external library, third-party platform, or any repo that does not use ispec, cross-repo canon updates aren't possible. The same fallback applies to **internal sibling projects that simply have not yet been retrofitted onto ispec** — which is the more common case in practice. A typical case: a downstream app's ADRs and spec sections reference its deploy tool extensively in prose because that tool is the upstream dependency, but the tool doesn't yet have its own canon, so structural cross-repo wikilinks aren't available.

Fall back:

- Capture an ADR in the *downstream* repo documenting the dependency choice. For external dependencies, add `external_dependency: <library/platform>` metadata. For internal-not-yet-ispec siblings, the ADR explains the contract from the downstream side; if and when the upstream is later retrofitted, the prose references can be tightened to structural `[[<repo>:<id>]]` wikilinks.
- If a workaround was needed because the upstream gap can't be filled in this session, capture a deferred feature linked to "upstream PR / waiting on external" (or "upstream retrofit / waiting on sibling").
- If the workaround is meant to be temporary, the deferred feature has a clear "remove once X" condition; a future session picking it up reverses the workaround.

This keeps the downstream canon honest about its dependencies and constraints without pretending to control the upstream, and without blocking adoption on whether sibling tools are also ready.

## Bootstrap path

Forcing a brand-new project to author a complete requirement, feature, ADR, and spec section before commit 1 would kill the velocity that makes most projects work. A survey of five real private projects shows a consistent pattern:

- The first commit ships a working feature, not scaffolding. Data model, handlers, deployment artifacts all land together.
- Architecture choices appear in `go.mod` / `package.json` / `requirements.txt` on day 1 but are never justified in commit messages.
- Implicit requirements are baked into the data model from the start. One project's first commit included roughly 2600 lines of seed data carrying multi-year payment history; the requirement "track payment state per member over time" was never written down anywhere.
- Retrospective design docs appear roughly 2-7 days in, after the shape is clear. One project's architecture overview landed on day 6; another's README only ever covered deployment. The one project with a day-1 README was the outlier — it had been designed in isolation before any of it was committed.
- Pivots happen visibly but invisibly: in one project, a single "dropped the members db" commit is the entire record of a meaningful change of direction.

The bootstrap path therefore runs in three phases.

### Phase 0 — Spike (day 1, sometimes day 2)

Just ship the first working slice. No ispec canon yet — most attempts to author canon at this point will turn out wrong because the shape isn't clear. The only ispec artifact at this stage is a single bootstrap session card capturing:

- The trigger (one paragraph: what is being built and for whom, in the user's own words).
- The motivating problem (why this exists).
- Any tooling pre-decisions the user has already made or wants the agent to weigh in on.

The agent can ask clarifying questions and weigh in on tooling before coding starts. Output of the dialogue is captured into the session card body, not into a separate "design document."

### Phase 1 — Initial canon (day 2-7, when the shape is clear)

After the first slice is working — typically a handful of commits in — the project has enough definition to capture canon retrospectively. The `ispec-bootstrap` skill prompts capture of:

- **One stakeholder card** for the primary user / customer of the system. More can be added later.
- **ADR-0001: core technology choices.** A single ADR (multiple bullet points is fine) covering the load-bearing tooling decisions that are now visible — language, web framework, database driver, deployment runtime, key libraries. The agent reads dependency files (`go.mod`, `package.json`, `requirements.txt`, `Dockerfile`) and proposes the ADR. User edits and confirms. Even a one-paragraph rationale per choice is dramatically more useful than nothing, because the alternative (silent tooling lock-in) is what survives by default.
- **A domain-model card** — a spec section describing the entities, their relationships, and the meanings that aren't obvious from field names. This is the single most expensive thing to reverse-engineer later and the most valuable thing to capture early. (A field named `container` that means a storage unit rather than a person is the kind of detail that lives only in code by default.)
- **1-3 requirement cards** — the high-level "the system shall X for stakeholder Y" statements, written from the *current* understanding. They will be refined later.
- **Feature cards for the work already landed** — retroactive but cheap because the work is fresh. Each feature targets one of the new requirements.

The bootstrap session closes when these cards exist and pass lint. The project is now "bootstrapped" — future work uses the regular `ispec-start-session` skill.

### Phase 2 — Steady state

The full session loop applies. New triggers run through `ispec-start-session`; existing features get more detailed cards lazily, as work touches them; additional stakeholders, requirements, ADRs are added when they become relevant rather than speculatively.

### What not to try to capture on day 1

- A full requirements catalog. The real requirements are baked into the data model; trying to enumerate them up front produces a generic list that has nothing to do with the actual decisions made. Surface them when they're acted on.
- A complete spec. The system doesn't yet know its own shape; specifying it in detail is premature.
- Stakeholders beyond the primary one. They appear naturally as the project grows.
- Comprehensive ADRs covering every dependency. Most dependencies are uninteresting — only the load-bearing 2-3 deserve an ADR.
- A polished README. It will be wrong. Write it after the shape is clear.

### What to capture even if it feels premature

- The *motivation* — why this project exists, what problem it solves. Easy to forget once the work is underway. Goes in the bootstrap session card and gets refined into stakeholder + requirement cards in Phase 1.
- The *domain model with semantics* — not just field names, but what they mean and the relationships between entities. This is what's most expensive to recover later.
- *Pivot ADRs* — when direction changes early, capture the change with `supersedes:` to the prior ADR. The week-1 pivot is the most valuable ADR a project ever writes; it's also the one most likely to be skipped.

### Anti-patterns to resist

- **Authoring canon before the first slice exists.** The canon will be wrong. Wait until you can read the code.
- **Skipping Phase 1 because "it's still early."** The transition from one-off code to maintainable system happens silently; if Phase 1 is deferred indefinitely, the project enters a long stretch where every new contributor (human or AI) reverse-engineers domain semantics from the code. Capture once, the moment shape is clear; refine lazily afterwards.
- **Bootstrap as a checklist exercise.** The point is to capture what's actually load-bearing, not to fill out every artifact type. A bootstrap session that produces one stakeholder, one ADR, one domain-model spec, one requirement, and one feature card is enough to graduate to steady state.

## Project-local AI memory

The card store doubles as project-local memory for AI agents. An agent joining the project:

- Reads `docs/index.md` for entry points.
- Searches by topic to find relevant cards.
- Walks `[[wikilinks]]` to assemble context.
- Loads only what's relevant — the same retrieval pattern a Zettelkasten store uses.

The split with a personal cross-project knowledge store is clean: project cards are *canonical and shared* (committed to the repo, reviewed in PRs); personal notes are *unilateral and cross-project*. Same data model, different authority and scope.

## Open design questions

- **Bootstrap path.** How a brand-new project starts on day 1 with one stakeholder, one requirement, one session, and grows organically without front-loading bureaucracy.
- **CI integration.** What runs as a pre-commit hook, what runs as CI, what runs only on demand.
- **Multi-author / merge conflicts.** What happens when two sessions concurrently amend the same spec section; how stable IDs are reserved without collision.
- **Tool surface.** Whether a dedicated `ispec` CLI is worth building or whether plain markdown + skills + a small lint script is enough.

## Anti-patterns the system resists

- **Silent scope expansion.** The contract step forces an explicit conversation when investigation reveals more work than the trigger suggested.
- **Silent spec drift.** The pre-completion reconciliation step + the "tell the user when the definition changed" rule prevent the spec from rotting to match whatever the code does.
- **Lost discoveries.** Deferred cards with `discovered_during:` mean nothing surfaces during real work and gets forgotten.
- **Audience leakage.** Spec sections declare their audience, so a consumer producing user-facing output has what it needs to exclude operational detail. ispec supplies the tags; it cannot enforce the filtering, since it produces no documents (ADR-0006).
- **Documentation written for the past tense.** Atemporal canon plus a separate session log keeps the reference clean and the narrative complete, without conflating them.
