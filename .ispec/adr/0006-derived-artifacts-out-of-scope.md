---
type: adr
stable_id: ADR-0006
slug: derived-artifacts-out-of-scope
status: accepted
---

## Context

ispec's first design carried a `## Generated views` section: the canon would
be traversed and filtered by frontmatter to emit `REQUIREMENTS.md`,
`SPEC.md`, a user manual, an operator runbook, a dev reference, and a
`BACKLOG.md`. Audience tags existed to route that generation.
`ispec-close-session` had a step to regenerate those documents on every
close; `ispec-defer` had a step to regenerate the backlog. [[REQ-1]] was held
at `partially-satisfied` specifically because none of it was built.

None of it ever was built, across the whole of ispec's development. That is
evidence worth reading rather than a backlog item to carry: the value kept
arriving from the capture side — the contract before coding, the
reconciliation after, the deferred card with `discovered_during:` — and never
from the prospect of a generated `SPEC.md`.

The underlying question is what ispec is *for*. Two answers were available:

1. **A documentation generator.** The canon is an intermediate representation,
   and ispec's output is a set of readable documents.
2. **An intent-capture system.** The canon *is* the output. Its value is that
   it records what a project is and why, accurately, at development time —
   when the intent is still known and cheap to state.

Answer 1 puts ispec in competition with every documentation tool, on their
terms, and makes the hard part prose generation — which is precisely what a
general-purpose agent is already good at, given good material.

Answer 2 makes ispec responsible for the thing nothing else does: being
present at the moment a decision is made, and getting it written down in a
structure that survives.

## Decision

Producing derived artifacts is **out of scope for ispec** — not deferred.
ispec captures intent at development time and keeps that capture structurally
verified. Any document derived from the canon — a full spec, a user manual, an
operator runbook, API documentation, a backlog view — is produced by other
tools, which ispec exists to supply with better context than they could
otherwise obtain.

Concretely:

- ispec's deliverables are the canon, the workflow skills that keep it
  accurate, and the lint that verifies its structure. Nothing else.
- `audience:` is retained. It was introduced to route ispec's own generation,
  but it is *intent about who a spec section is for* — something a consuming
  tool cannot infer from the code, and exactly the kind of context this
  decision says ispec should capture. ispec validates it and does not act on
  it.
- Deferred work is visible through the cards themselves and through
  `ispec status` ([[F-6]]). No backlog document is emitted.

## Consequences

- **ispec stops competing with documentation tooling** and starts feeding it.
  The scope is narrow enough to finish.
- **The canon must stand on its own as a read target**, since nothing
  downstream will reformat it. Cards are written to be read directly, by a
  human or an agent, which they already were.
- **ispec no longer prevents audience leakage structurally.** The original
  claim was that audience-filtered generation makes leaks impossible, because
  the generator physically cannot see operator content while producing
  user-facing docs. ispec gives a consumer the tags needed to achieve that,
  but cannot enforce it. This is a real capability given up, and the honest
  framing is that leak prevention is now the consumer's property, not ispec's.
- **No pipeline, no build step, no generated files to keep current.** The
  `ispec-close-session` and `ispec-defer` skills lose a step each, and
  sessions stop carrying an obligation nothing fulfilled.
- **[[REQ-1]] is satisfiable.** It asked that the system *enable production of*
  atemporal canonical documentation, which the canon and its workflows do.
- The decision is reversible in principle but unlikely to be revisited: a
  consumer that wants a generated document can read the canon and write one,
  which is the whole point.

## Alternatives considered

- **Keep generation as deferred work.** Rejected: it had been deferred since
  the first design commit without ever being the most valuable next thing,
  and carrying it made every requirement touching documentation permanently
  `partially-satisfied` against a target nobody wanted to hit.
- **Build a minimal `ispec render`.** Tempting, and the smallest version
  (concatenate spec cards in order) is nearly free. Rejected because the
  minimal version is not what anyone wants — the value in a generated
  document is editorial, and that is the part ispec would do worst.
- **Drop `audience:` along with generation.** Rejected: it is intent, not
  plumbing. A consumer deciding what belongs in a user manual needs it, and
  nothing in the code carries it.
