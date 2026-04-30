---
type: spec
stable_id: SPEC-data-model
slug: data-model
audience: [developer]
satisfies_features: [F-1, F-2]
---

## Card types

Six types, each a markdown file with YAML frontmatter under `.ispec/<type>s/`:

- **stakeholder** — who is affected by or interacts with the system.
- **requirement** — problem statement in domain language; what the system shall do for whom.
- **feature** — unit of solution work; targets one or more requirements.
- **spec** — atemporal description of part of the system; satisfies the documentation need of one or more features.
- **adr** — architecture decision record; context, decision, consequences.
- **session** — chronological record of one unit of work; trigger, classification, affected cards, outcome.

## Stable IDs

Typed prefix plus counter or concept name:

- `STK-NN` for stakeholders.
- `REQ-NN` for requirements.
- `F-NN` for features.
- `SPEC-<concept>` for spec sections (hierarchical names allowed: `SPEC-auth.refresh`).
- `ADR-NNNN` for ADRs (4-digit zero-padded).
- Sessions are slug-only (`<YYYY-MM-DD>-<short>`); they are inherently chronological and don't need a counter.

## References

Hard relations live in frontmatter as arrays of stable IDs:

- `stakeholders` (on requirements)
- `targets` (feature → requirements)
- `satisfies_features` (spec → features)
- `depends_on` (feature → features, possibly cross-repo)
- `supersedes` (ADR → ADRs)
- `relates_to` (ADR → any)
- `discovered_during` (any → session)
- `affected_cards`, `deferred_cards` (session → any)

Soft references appear in body prose as `[[wikilinks]]`, resolving by stable ID or slug. Lint validates that they resolve.

Cross-repo references prefix the repo name: `[[deploy-tool:F-12]]`, `depends_on: [deploy-tool:F-12]`. Sibling repos are listed in `.ispec/repos.yml` (when cross-repo lint is enabled).

## Lifecycle

- **Stakeholder**: `active | retired`.
- **Feature**: `proposed | active | implemented | deferred | abandoned`.
- **ADR**: `proposed | accepted | superseded | deprecated`.
- **Session**: `open | resolved | deferred | abandoned`.
- **Requirement**: no status; *satisfaction* (`unsatisfied | partially-satisfied | satisfied`) is derived from the features that target it.
- **Spec**: no status; spec sections are always current — when reality changes, the spec is amended.

See also `docs/schemas.md` for the full per-type frontmatter reference, and [[ADR-0001]] for the workflow that produces and maintains these cards.
