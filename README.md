# ispec

Agent workflow for capturing intent — the who, what and why behind a codebase — as part of the development session itself.

## Installation

A single Go binary contains an install routine as well as the agent workflows. Requires Go 1.21+.

```sh
go install github.com/lheiskan/ispec/cmd/ispec@latest
ispec harness install     # write the skills and the session-start hook
ispec harness check       # confirm the install is current
```

The install is user-global: the workflow skills go under `~/.claude/skills/`,
and one session-start hook is appended to `~/.claude/settings.json` — nothing
else in that file is touched, and the original is backed up first. To remove
only what the install wrote:

```sh
ispec harness uninstall
```

See [INSTALL.md](INSTALL.md) for additional info on building, installing and upgrading.

## Usage

Workflows trigger automatically in the agent. Before starting new work items, agent will summarize the user ask and present a set of card artifacts that reflect the goal of the work. Once user has approved the artifacts, agent proceeds to write the corresponding code. Work takes place in an ispec session. Once work is completed, code and the artifacts are committed and the session is closed.

New projects start in a bootstrap (`/ispec-bootstrap`) stage. This stage does not include the artifact creation. This stage is intended for quick iteration and experimentation to gain an understanding of what to build and how. A good place to transition from the bootstrap phase into using regular ispec sessions is after the first functional slice of the project is complete. Once this stage is reached, ask the agent to create the first canon for the project and exit the bootstrap phase.

For existing projects you can ask the agent to use `ispec-retrofit` to create an initial canon and start using ispec sessions for future work.

## The model

Six artifact types, all plain markdown cards with YAML frontmatter and `[[wikilinks]]` between them, committed to the repo as project canon:

| Type | Answers |
|---|---|
| **Stakeholder** | who is affected — users, operators, support, integrators, regulators, the agents working on the code |
| **Requirement** | what problem must be solved, in domain language, independent of solution |
| **Feature** | a unit of solution work, targeting one or more requirements |
| **Spec section** | what the solution *is*, precisely enough to reimplement, organized by concept |
| **ADR** | why a decision was made: context, alternatives, consequences |
| **Session** | one iteration of the dev loop: trigger, work, outcome |

```
.ispec/
  stakeholders/<slug>.md
  requirements/<slug>.md
  features/<slug>.md
  spec/<slug>.md
  adr/<NNNN>-<slug>.md
  sessions/<YYYY-MM-DD>-<slug>.md
```

Stable IDs (`REQ-7`, `F-12`, `ADR-0014`, `SPEC-auth.refresh`) are the durable cross-reference primitive; the lint binary catches dangling ones. Spec sections carry `paths:` globs pointing at the code that implements them, which makes the code↔spec link both grep-able and checkable.

## The loop

Work is driven by skills, not by the user remembering a checklist. The user says "fix the mail bug"; the workflow classifies the trigger, drafts a contract, waits for approval, implements, then reconciles.

Two steps carry most of the value:

- **The contract**, before coding: which features, targeting which requirements, which spec amendments, which ADRs, and what is explicitly *out* of scope. Approved by the user before implementation starts. This is what forces an explicit conversation when investigation reveals more work than the trigger suggested.
- **Pre-completion reconciliation**, after coding: diff the implementation against the feature definition. If the code is wrong, fix the code. If the definition was wrong, update the canon **and say so** — "the implementation diverged from the original definition; here's what I changed; confirm?"

That second rule is the discipline that keeps a spec meaningful. Without it, canon silently converges on whatever the code happens to do.

Work surfaced mid-session that doesn't belong in this iteration gets deferred as a card carrying `discovered_during:`, with cascading status updates across the requirements and ADRs it touched. A backlog falls out as a side effect — no separate tracker.

## The skills

In a project with no canon yet:

- **`/ispec-bootstrap`** — a brand-new project. Runs in three phases: ship the first working slice with no canon (the shape isn't clear yet and the canon would be wrong), capture the load-bearing canon retrospectively once it is, then graduate to the steady-state loop.
- **`/ispec-retrofit`** — an existing project with code, history, and maybe docs. Reads the codebase and git history and proposes a deliberately scoped initial canon. Scope it explicitly; retroactively documenting everything ever shipped is an infinite project.

Then, per unit of work:

- **`/ispec-start-session`** — the main loop. Invoke at the start of any substantive change.
- **`/ispec-close-session`** — final reconciliation; marks the session resolved.
- **`/ispec-defer`** — force smaller scope than the agent proposed, or park discovered work.

You rarely type the last three: once the harness integration is installed, a
session opened in a project that has a canon is told what state that canon is
in — including any session left open — and to run `/ispec-start-session` for
substantive work. In projects without a canon, ispec says nothing.

## The CLI

Day-to-day canon work does not go through the binary: the cards are ordinary
markdown in the repository the agent already has open, so reading and writing
them is plain file access. The binary covers what file access cannot do.

```sh
ispec lint     # validate schemas and resolve every cross-card reference
ispec status   # card counts, open sessions, deferred work, lint state
```

Both take `-root <project>` and `-canon-dir <name>`, and exit 0 clean, 1 on
problems, 2 if there is no canon directory — so an unbootstrapped project is
distinguishable from a failing one.

## Status

Early. The data model, schemas, workflow skills, and CLI exist and are used on real projects; ispec's own canon lives in [`.ispec/`](.ispec/) and its lint runs clean against itself.

## Docs

- [`docs/design.md`](docs/design.md) — artifact taxonomy, data model, session loop, status lifecycles, defer mechanism, cross-repo sessions, bootstrap path, lint rules, open questions.
- [`docs/schemas.md`](docs/schemas.md) — frontmatter schema per card type; what the lint binary checks.
- [`docs/worked-examples/cross-repo-session.md`](docs/worked-examples/cross-repo-session.md) — a cross-repo session walked end to end.

## License

MIT. See [LICENSE](LICENSE).
