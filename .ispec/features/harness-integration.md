---
type: feature
stable_id: F-5
slug: harness-integration
status: implemented
targets: [REQ-3]
paths:
  - internal/harness/**
  - skills/**
---

## Description

Installation of ispec's workflow skills and session-start hook into an LLM
agent harness, as a named target rather than an assumed one. Per [[ADR-0005]].

## Behavior

- `ispec harness install` — writes each embedded skill to the target's skill
  location and merges a session-start hook into the target's settings.
  Idempotent: a second run reports `unchanged` and writes nothing.
- `ispec harness check` — reports per item whether it is installed, missing,
  or stale (installed bytes differ from the binary's own). Exits non-zero if
  anything needs installing, so it works as a gate.
- `ispec harness uninstall` — removes only what install wrote.
- `ispec harness session-start` — prints the directive; what the installed
  hook runs. Reports canon status via [[F-6]] and names the one or two things
  that change what the agent should do next, most-blocking first: open
  sessions, then a canon that does not lint clean.
- `--target NAME` selects the harness (`claude-code` today). An unknown name
  is an error listing what is accepted. `--dir PATH` overrides the target's
  config directory; `--dry-run` reports without writing.

Safety properties, all per [[ADR-0005]] and covered by tests in
`internal/harness/harness_test.go`:

- Editing the settings file preserves member order and any field this binary
  does not model.
- A pre-existing settings file is backed up once, before the first edit.
- A settings file that is not valid JSON is refused, not overwritten.
- Hook ownership keys on the binary name, so a sibling tool's identically
  shaped `harness session-start` hook survives install and uninstall.
- The directive prints nothing in a project with no canon.

## Limitations (deferred)

- Claude Code is the only target. The registry exists so a second one is
  additive, but no second one is written.
- No MCP server. ispec's canon lives inside the repository the agent already
  has open, so file access already reaches every card; a tool surface would
  reimplement that behind a schema costing tokens in every session. Revisit
  only for work file tools genuinely cannot do — cross-repo resolution being
  the real case.
- The hook covers session start only. No pre-commit or pre-tool hook checks
  canon alignment.
- `check` compares skill bytes exactly, so a user who edits an installed
  skill sees `stale` with no indication of what differs.
