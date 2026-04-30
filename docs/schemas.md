# Card schemas

Reference for the frontmatter expected on each card type. Lint enforces these. Body content is free-form markdown.

## Common conventions

- **File location**: `.ispec/<type>s/<slug>.md` (e.g. `.ispec/features/offline-call-logging.md`). ADRs live in `.ispec/adr/<NNNN>-<slug>.md` (numeric prefix). Sessions live in `.ispec/sessions/<YYYY-MM-DD>-<slug>.md`. The default location is `.ispec/` but is configurable via `ispec-lint -canon-dir <subdir>`.
- **Slug**: kebab-case, must match the filename (without extension and any numeric prefix).
- **Stable ID**: typed prefix + counter (`STK-7`, `REQ-12`, `F-23`, `SPEC-auth.refresh`, `ADR-0014`). Sessions are identified by slug only — they don't have stable IDs because they're chronological by nature.
- **Cross-repo references**: prefix with the repo name and a colon, e.g. `[[deploy-tool:F-12]]`, `depends_on: [deploy-tool:F-12]`. Sibling repos are listed in `.ispec/repos.yml` (when cross-repo lint is enabled).
- **Wikilinks**: `[[<stable-id>]]` or `[[<slug>]]` in body prose. Lint resolves both forms.
- **Frontmatter format**: standard YAML, fenced by `---` lines at the top of the file.

## Hard relations vs. soft references

- **Hard relations** (load-bearing graph edges) are expressed as frontmatter array fields with stable IDs (`targets`, `depends_on`, `satisfies_features`, etc.). Lint validates these strictly.
- **Soft references** (mentions in passing) are expressed as `[[wikilinks]]` in body prose. Lint validates that they resolve, but doesn't require them.

## stakeholder

```yaml
---
type: stakeholder
stable_id: STK-NN          # required, unique
slug: <kebab-case>          # required, matches filename
status: active | retired    # required
---
```

Body: identity, goals / jobs-to-be-done, constraints, communication channel. Free-form markdown.

## requirement

```yaml
---
type: requirement
stable_id: REQ-NN           # required, unique
slug: <kebab-case>          # required
stakeholders: [STK-NN, ...] # required, at least one
satisfaction: unsatisfied | partially-satisfied | satisfied
                            # optional — derived from features that target it;
                            # may be set explicitly during retrofit/bootstrap
---
```

Body: the requirement statement in domain language plus rationale.

## feature

```yaml
---
type: feature
stable_id: F-NN             # required, unique
slug: <kebab-case>          # required
status: proposed | active | implemented | deferred | abandoned
                            # required
targets: [REQ-NN, ...]      # required, at least one
depends_on: [F-NN, ...] | [<repo>:F-NN, ...]
                            # optional
discovered_during: <session-slug>
                            # optional
paths: [<glob>, ...]        # optional, code paths implementing this feature
---
```

Body: feature description, behavior, edge cases.

## spec section

```yaml
---
type: spec
stable_id: SPEC-<concept>   # required, unique (hierarchical names allowed: SPEC-auth.refresh)
slug: <kebab-case>          # required
audience: [end-user | operator | developer | admin]
                            # required, at least one
satisfies_features: [F-NN, ...]
                            # required, at least one
paths: [<glob>, ...]        # optional
---
```

Body: precise description of what the system does in this area. Atemporal — describes what *is*, not how it got there.

## adr

```yaml
---
type: adr
stable_id: ADR-NNNN         # required, unique, 4-digit zero-padded
slug: <kebab-case>          # required
status: proposed | accepted | superseded | deprecated
                            # required
supersedes: [ADR-NNNN, ...] # optional
relates_to: [<id>, ...]     # optional, any card type
---
```

Body: context, decision, consequences. Typically also alternatives considered and rationale for the choice.

## session

```yaml
---
type: session
slug: <YYYY-MM-DD>-<short>  # required, matches filename
trigger: <one-line>         # required
classification: <category>  # required (e.g. new-feature, bug-bad-impl, bug-corner-case, refactor, ops, bootstrap, retrofit)
status: open | resolved | deferred | abandoned
                            # required
affected_cards: [<id>, ...] # optional, populated as work progresses
deferred_cards: [<id>, ...] # optional
commits: [<sha>, ...]       # optional
outcome: <multi-line string>
                            # required when status = resolved
---
```

Body: trigger details, design dialogue, discoveries, decisions made during the session. Chronological — this is the session log, not canon.

## Lint rules (v0.1)

The MVP lint binary checks:

1. Every card has parseable YAML frontmatter.
2. Required fields present per type.
3. `type` is one of the recognized values.
4. `status` (where required) is from the type's allowed set.
5. `slug` matches the filename (without extension and any numeric prefix).
6. Stable IDs are unique within the repo.
7. All ID references in frontmatter resolve to existing cards (`targets`, `satisfies_features`, `depends_on` (in-repo), `supersedes`, `relates_to`, `stakeholders`, `discovered_during`, `affected_cards`, `deferred_cards`).
8. All `[[wikilinks]]` in body prose resolve to existing cards (by stable ID or slug).
9. Cross-repo references (`<repo>:<id>` form) are noted but not resolved (until cross-repo config is implemented).

Future lint additions (not v0.1):

- Status invariants across linked cards (a feature `implemented` whose linked ADR is `proposed` is inconsistent).
- `paths:` glob validation against the actual filesystem.
- Audience-leakage checks in generated docs.
- Cross-repo reference resolution.
