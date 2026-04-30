---
type: session
slug: 2026-05-01-retrofit-findings
trigger: an adopter project's retrofit completed; retrofit skill produced no ispec-side feedback because the session-card body template lacks a "Discoveries about ispec" section
classification: bad-implementation + new-feature + missed-corner-case
status: resolved
affected_cards:
  - F-4
  - F-2
deferred_cards: []
commits: []
outcome: |
  ispec now has an explicit dogfood-loop mechanism. Every session
  card template (in retrofit, spec-session, bootstrap) prompts for
  a "Discoveries about ispec" or "Discoveries during the session"
  body section. The session-close skill explicitly checks that
  section is non-empty before resolving a session, asking the user
  one more time if it's blank.

  Retrofit captured as F-4. The retrofit/SKILL.md amendment is the
  feature's first real definition rather than a change to existing
  behaviour, so F-4's body describes the skill including the new
  discoveries-section discipline.

  Drive-by SKILL.md updates landed in spec-session, bootstrap, and
  session-close to keep the pattern consistent across all session
  types. Those skills' feature cards (other than F-2 spec-session,
  which was already captured) remain deferred per the previous
  decision to lazily fill in canon coverage.

  design.md "Non-ispec upstreams" subsection amended to explicitly
  cover the "internal sibling not yet retrofitted" case, with
  a downstream app and its deploy tool as the live example. The
  fallback (capture downstream
  ADR in prose; tighten to structural cross-repo wikilinks if the
  sibling is later retrofitted) is unchanged; the wording is clearer
  about when this case applies.

  All session card templates now include `commits: []` in their
  frontmatter; the close protocol in each skill explicitly mentions
  populating it from `git log` since the session opened.
---

## Trigger

The first adopter project's retrofit completed cleanly — 26 cards across 2 stakeholders, 5 requirements, 4 ADRs, 8 features, 6 spec sections, all lint-clean. But the resulting session card had no place to surface ispec-side observations: it had `## Reading notes` (project-side), stakeholder identification (project-side), and `## What's deliberately NOT being captured` (project-side scope decisions). The agent doing the retrofit had no template slot to record friction with ispec itself, so feedback that should have flowed back was lost.

This is a structural failure of the dogfood loop: ispec asks projects to capture canon for themselves but doesn't yet collect feedback about itself.

## Pre-decisions

- Capture retrofit as F-4 in this session, since we're amending its SKILL.md. Other workflow skills (bootstrap, defer, session-close) stay as deferred feature captures — handled as drive-bys, same pattern as the previous bootstrap session.
- The "Discoveries about ispec" template addition applies to every session, not just retrofit. Drive-by parallel updates to spec-session, bootstrap, and session-close SKILL files keep the pattern consistent.
- One-sided cross-repo (downstream uses ispec, upstream doesn't yet) is technically covered by the existing Non-ispec-upstreams fallback in design.md, but the language doesn't make the "internal-sibling-not-yet-retrofitted" case explicit. Brief amendment.

## Contract

In ispec:

- New F-4 retrofit-skill (paths: `.claude/skills/retrofit/SKILL.md`, targets: REQ-1 + REQ-3). Captures retrofit as a feature so the amendment has formal attachment.
- Amend `retrofit/SKILL.md`: expanded session-card body template with `## Reading notes`, `## Discoveries about ispec`, `## What's deliberately NOT being captured`, `## Outcome`; explicit step to populate `commits:` on close.
- Drive-by parallel update in `spec-session/SKILL.md` (F-2; drive-by SKILL change without F-2 card amendment), `bootstrap/SKILL.md`, and `session-close/SKILL.md`: same template + commits prompt at appropriate phases.
- Append to `docs/design.md` "Non-ispec upstreams" subsection: explicitly mention "internal sibling project not yet retrofitted onto ispec" (a downstream app depending on its deploy tool is the live example).

Out of scope (deferred to future sessions):

- Capturing bootstrap, defer, session-close as feature cards. Continues to defer; same pattern as previous sessions.
- Schema renames: `satisfies_features` (reads backwards), `relates_to` factoring into typed fields like `addresses_requirements:` and possibly merging with `supersedes:`.
- New frontmatter field for "external dependency on non-ispec sibling" — currently expressed in prose, which is fine for now.
- Skill quality improvement: turning "future work" mentions in outcome prose into structural deferred cards.

These are all real findings but lower-urgency than the missing-discoveries feedback loop, which actively prevents ispec from learning from each use.

## Discoveries during the session

- **The retrofit's reading-notes section worked well** — having a freeform body slot for project-side observations is genuinely useful and survived the roundtrip. The fix here was adding *another* body slot, not replacing the existing one.
- **Recursive proof point:** this very session populated its "Discoveries during the session" section with real content. The pattern works because the slot exists; without the slot, the temptation is to dump everything into `outcome` or skip it entirely.
- **The drive-by approach has a structural cost:** the session card's `affected_cards` lists only F-4 and F-2, but the work also touched `.claude/skills/bootstrap/SKILL.md` and `.claude/skills/session-close/SKILL.md`. Those edits are real but invisible to anyone walking the canon graph because the underlying skills aren't yet feature cards. Capturing bootstrap, defer, and session-close as features (deferred again here) would close this gap. Should escalate from "defer indefinitely" to "next time we touch one of these skills, capture as feature first."
- **Confirmed `relates_to` is overloaded.** Visible in the adopter project's canon: an ADR carrying `relates_to: [ADR-0002, REQ-5]`, mixing ADR-to-ADR and ADR-to-REQ relationships into one field. The lint accepts both because `relates_to` is "any-type", but semantically these are different relations. Worth a future schema-naming session bundling: `relates_to` factoring + `satisfies_features` rename + `addresses_requirements:` introduction. Not in scope here.
- **Session card `commits:` field was missing from templates.** The bootstrap template explicitly didn't include it; spec-session's didn't either. Fixed in this session as part of the consistency pass. The schema documented `commits:` as optional, but template absence meant it was being skipped in practice.

## Outcome

See frontmatter `outcome` field.
