---
type: requirement
stable_id: REQ-3
slug: cross-project-usability
stakeholders: [STK-1, STK-2]
satisfaction: partially-satisfied
---

## Statement

The system shall be usable from any project via a shared installation of skills and tooling, without requiring the canon-using project to host its own copy of ispec's source.

## Rationale

ispec is a tool, not just a project. Each project that adopts ispec needs access to the same workflow skills (spec-session, bootstrap, retrofit, defer, session-close) and the same lint binary. If those have to be copied into every adopting project, every change to ispec requires a sync step in every project, and skills become per-project forks.

The natural model is install-once: ispec's skills live at the user-global skill location, the lint binary is on `$PATH`, and individual projects only carry their own canon (`.ispec/`).

## Currently

Partially satisfied. [[F-3]] establishes the install path via Makefile and INSTALL.md, using symlinks to the ispec checkout. Limitations: no per-project version pinning (symlinks always point at the live checkout); no detection of skill-version drift; no automated update mechanism; the `ispec-lint` binary requires manual rebuild after ispec source changes.
