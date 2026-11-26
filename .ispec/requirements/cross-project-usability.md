---
type: requirement
stable_id: REQ-3
slug: cross-project-usability
stakeholders: [STK-1, STK-2]
satisfaction: satisfied
---

## Statement

The system shall be usable from any project via a shared installation of skills and tooling, without requiring the canon-using project to host its own copy of ispec's source.

## Rationale

ispec is a tool, not just a project. Each project that adopts ispec needs access to the same workflow skills (`ispec-start-session`, `ispec-bootstrap`, `ispec-retrofit`, `ispec-defer`, `ispec-close-session`) and the same lint binary. If those have to be copied into every adopting project, every change to ispec requires a sync step in every project, and skills become per-project forks.

The natural model is install-once: ispec's skills live at the user-global skill location, the lint binary is on `$PATH`, and individual projects only carry their own canon (`.ispec/`).

## Currently

Satisfied. [[F-3]] distributes ispec as a single Go binary with its workflow
skills embedded, installable with `go install
github.com/lheiskan/ispec/cmd/ispec@latest`; [[F-5]] installs the skills and
the session-start hook from the binary itself. No ispec checkout is required
to install, and none is referenced afterwards. A project carries only its own
`.ispec/` canon.

The limitations that held this to partially-satisfied under the previous
symlink install — a checkout the user had to keep in place, no drift
detection, a manual rebuild step — are gone: skills cannot drift from the
binary because they are the binary, and `ispec harness check` reports a stale
install.

What remains is narrower than the requirement's statement and does not block
it: no per-project version pinning, so every project on a machine shares the
installed binary; and installing needs a Go toolchain, since there are no
prebuilt binaries or package-manager distribution. Both are tracked on [[F-3]].
