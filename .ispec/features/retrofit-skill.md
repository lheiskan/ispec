---
type: feature
stable_id: F-4
slug: retrofit-skill
status: implemented
targets: [REQ-1, REQ-3]
paths:
  - skills/ispec-retrofit/SKILL.md
---

# ispec-retrofit skill

Adopts ispec onto an existing project — one with substantial code, git history, and possibly existing human-authored documentation. Different shape from `ispec-bootstrap`: the first slice already exists, and the discipline shifts from "produce canon for what was just built" to "capture only the load-bearing canon and resist the urge to retroactively document everything."

## Behavior

The skill walks the agent through:

1. Confirm intent and scope (minimum-viable / standard / comprehensive). Default recommendation: standard.
2. Open a retrofit session card with body sections covering reading notes, ispec-side discoveries, scope decisions, and outcome.
3. Read the project: existing human docs (README, OPS.md, DESIGN.md), dependency files, schema / type files, handler / route files, deployment config, and recent + earliest commit history.
4. Propose initial canon at the chosen scope:
   - Stakeholders (primary user; operator if applicable).
   - ADR-0001 bundling load-bearing tooling decisions.
   - Domain-model spec section capturing entities and non-obvious semantics.
   - 1-3 requirements (minimum-viable) up through full coverage (comprehensive).
   - Feature cards for current major surfaces, each carrying a `paths:` glob.
   - At standard scope and above: audience-tagged spec sections per major subsystem and historical-pivot ADRs recovered from commit history.
5. Translate existing project docs into ispec form rather than preserving them as-is.
6. Iterate one card at a time with the user; capture ispec-side observations in the session card's "Discoveries about ispec" section as they surface.
7. Run `ispec lint` to verify consistency.
8. Close the retrofit session: populate `affected_cards`, `commits` (the SHAs landed during this work), `outcome`, set `status: resolved`.

## Scope discipline

The skill defaults to **standard** scope and resists comprehensive coverage. A retrofit is a one-shot session; if it spans multiple days, the scope was wrong and should be cut shorter.

## Self-improvement loop

The "Discoveries about ispec" section in the session card is the mechanism by which retrofit feeds back into ispec's own design. Friction encountered while applying ispec to a real project — schema gaps, ambiguous skill instructions, missing relations — gets captured there and becomes the trigger for follow-up spec-sessions in ispec.

## Limitations (deferred)

- No automated reading of existing documentation; the agent reads and translates by hand.
- No automated detection of cross-repo dependencies (when a project depends on a sibling tool that's also retrofittable, the agent has to surface that as a follow-up session).
- No mechanical scope enforcement — the standard / comprehensive distinction is judgment-based.
