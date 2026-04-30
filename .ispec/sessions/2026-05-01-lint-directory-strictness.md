---
type: session
slug: 2026-05-01-lint-directory-strictness
trigger: a retrofit in an adopter project revealed that lint silently skips cards in unrecognized directories under .ispec/
classification: bad-implementation + missed-corner-case
status: resolved
affected_cards:
  - F-1
deferred_cards: []
commits: []
outcome: |
  ispec-lint now errors on unrecognized directories directly under
  the canon root, with a "did you mean..." suggestion when the
  directory name is one trailing 's' away from a valid name (so
  `specs/` → suggests `spec/`). Other unknown names fall back to
  listing all valid card-type directories.

  schemas.md now has an explicit per-type table for directory
  naming, calling out that `spec` and `adr` stay singular (because
  their plurals read awkwardly) while the rest are pluralized. No
  more <type>s with-exceptions ambiguity.

  F-1 (lint-binary) body was amended under Behavior to document
  the new validation, and to clarify that files at the canon root
  itself (e.g. .ispec/index.md) are still skipped — only
  unrecognized *directories* error.

  Verified the new behavior via synthetic test: a canon root with
  `.ispec/specs/` and `.ispec/unknown-dir/` produces two errors
  with the expected suggestions, while the valid `stakeholders/`
  card still gets validated.
---

## Trigger

A retrofit session in an adopter project placed spec cards in `.ispec/specs/` (plural) — reasonable, since `schemas.md` says `.ispec/<type>s/<slug>.md` and lists `adr/` / `sessions/` as exceptions but doesn't mention `spec/` explicitly. Lint walks `.ispec/spec/` (singular) and silently skipped the 5 cards in `specs/`. The retrofit's own lint reported "0 errors" while a quarter of the canon was orphaned. The bug only surfaced when the session card's `affected_cards` listed the now-orphaned spec stable IDs as dangling references — without the session card, this would have been completely invisible.

Three sources of truth disagreed: schemas.md (`<type>s` pattern), design.md (`spec/` singular), and lint code (`spec/` singular). The agent followed schemas.md and got bitten by lint+design.

## Pre-decisions

- **Strict, not lenient.** Lint will error on unrecognized directories under the canon root, with a "did you mean..." suggestion when the name is close to a valid one. Accepting both `spec/` and `specs/` as aliases would cement the inconsistency; one canonical form is cleaner.
- **`spec/` and `adr/` stay singular.** "specs" reads fine but "adrs" is awkward. Singular for both is consistent.
- **F-1 body is amended.** This is a real expansion of lint's behavior, not a drive-by fix.
- **The adopter project's canon is left alone.** The `mv specs spec` there is for the user to do (or a separate session in that project) once the lint update lands.

## Contract

In ispec:

- Amend `cmd/ispec-lint/main.go`: error on unrecognized directories directly under the canon root; helpful suggestion when the dir name is `<valid>s`/`<valid>` (off-by-trailing-s case).
- Amend F-1 (lint-binary) body: add the new validation under Behavior; clarify that files at the canon root itself (e.g. `index.md`) are still skipped.
- Reconcile `docs/schemas.md`: explicit per-type directory list (no more `<type>s with exceptions` ambiguity).

Out of scope (deferred):

- Fixing the adopter project's `.ispec/specs/` → `.ispec/spec/`. User-side, or a future session in that project.
- Soften retrofit skill's "do not batch" rule to acknowledge batched-with-review is acceptable in auto mode.
- Enumerate `classification` values vs. accept free-form. Currently free-form; agents invent labels (`bug-bad-impl` vs. `bad-implementation`).
- Future-work prose in outcomes → structural deferred cards.
- Schema renames bundle (`satisfies_features`, `relates_to` factoring).

## Discoveries during the session

- **The dogfood loop is faster than the fix.** Between the initial survey of the adopter project (which found `specs/` plural) and the implementation phase of this session, the user had already moved that project's `.ispec/specs/` → `.ispec/spec/` based on the dangling-reference errors my first lint run produced. So when I tried to verify that the new lint catches the bug, the project was already clean. Verification ended up using a synthetic test directory instead. Validates that the dangling-reference-on-session-card-affected-cards is itself a useful detection mechanism — even in a world without the new directory check, the session card's references catch some structural issues.

- **The "off by trailing s" heuristic for suggestions is small but high-value.** Both `specs` and `adrs` are exactly the typo this catches; that covers the realistic case. Wider Levenshtein-style matching would catch arbitrary typos but isn't worth the complexity until we hit a real case beyond singular-vs-plural.

- **The adopter project has continued evolving in parallel.** A new ADR and a corresponding feature card exist now that didn't during the initial survey, and a follow-on session there is closing a gap the retrofit session flagged. That's the dogfood loop working as designed: the retrofit produced canon, the canon flagged a known limitation in REQ-1, a follow-on session is resolving it. Worth observing without action.

- **Findings 3-5 from the original survey remain valid and unaddressed in this session.** Specifically: the retrofit "do not batch" rule should soften; classification field is free-form across agents; future-work prose vs. deferred cards. None block; all are real. Captured for future sessions.

## Outcome

See frontmatter `outcome` field.
