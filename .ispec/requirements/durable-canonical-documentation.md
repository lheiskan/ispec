---
type: requirement
stable_id: REQ-1
slug: durable-canonical-documentation
stakeholders: [STK-1]
satisfaction: satisfied
---

## Statement

The system shall enable production of canonical project documentation — requirements, decisions, specifications, stakeholders — that is *atemporal* (describes what the project *is*), independent of the order in which work was done.

## Rationale

Prompt-driven development produces a chronological narrative by default: prompts, code edits, commit messages. Downstream uses — regeneration, agent training, user-doc production — need an atemporal reference. Without an explicit canon, the project drifts into "code is the spec" territory, and consumers must reverse-engineer intent from implementation.

Those downstream uses are performed by other tools, per [[ADR-0006]]. This requirement is about making the material they work from exist and stay accurate, not about producing their output.

## Currently

Satisfied. The artifact taxonomy, the workflow skills, and the lint that
verifies structural integrity all exist, and the canon directory is in place:
a project adopting ispec can produce and maintain atemporal canonical
documentation, and [[F-1]] catches it when that documentation's reference
graph breaks.

This card previously read `partially-satisfied` on the grounds that view
generation — aggregating the canon into a full `SPEC.md` and similar — was
unimplemented. [[ADR-0006]] removes that from scope: the requirement asks the
system to *enable production of* canonical documentation, which it does.
Rendering it into other documents is a consumer's job.

See [[F-1]] for the lint binary, [[F-2]] for the `ispec-start-session` skill
that drives canon authoring, and [[F-4]] for adoption on an existing project.
