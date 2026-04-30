---
type: feature
stable_id: F-1
slug: lint-binary
status: implemented
targets: [REQ-1, REQ-2]
paths:
  - cmd/ispec-lint/**
  - go.mod
  - go.sum
---

## Description

A Go binary (`ispec-lint`) that walks the canon directory, parses card YAML frontmatter, validates per-type schema rules, and resolves cross-card references — both frontmatter ID lists (`targets`, `satisfies_features`, etc.) and body `[[wikilinks]]`.

## Behavior

- Reads `.ispec/` (or `-canon-dir <subdir>` override) under `-root` (default `.`).
- **Errors on unrecognized directories** directly under the canon root, with a "did you mean..." suggestion when the directory name is one trailing `s` away from a valid card-type directory. This catches the common typo of pluralizing `spec` → `specs` or singularizing `stakeholders` → `stakeholder`.
- Skips files at the canon root itself (`.ispec/index.md`, etc.) — those are not cards and don't trigger validation.
- Validates: parseability, required fields per type, valid `status` / `audience` / `satisfaction` values, slug-matches-filename, stable ID uniqueness.
- Resolves: every typed reference field plus every body wikilink, by stable ID or slug, with type-checking where applicable.
- Cross-repo references (`<repo>:<id>`) emit a warning rather than an error, since cross-repo resolution is not yet implemented.
- Exits 0 on clean, 1 on any error, 2 on missing canon directory.

## Limitations (deferred)

- No status invariant checks across linked cards (a feature `implemented` whose linked ADR is still `proposed` is currently allowed).
- No `paths:` glob validation against the actual filesystem.
- No cross-repo reference resolution; needs `.ispec/repos.yml` plus sibling repo discovery.
- No JSON output mode for tooling integration.
- No tests in the Go module yet.
