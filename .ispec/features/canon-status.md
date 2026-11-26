---
type: feature
stable_id: F-6
slug: canon-status
status: implemented
targets: [REQ-2]
paths:
  - internal/status/**
---

## Description

Reporting of the facts about a canon that determine what should happen next in
a session — chiefly whether work is already open. Surfaced as `ispec status`
and consumed by the session-start directive in [[F-5]].

## Behavior

- Card counts per type, and the total.
- **Open sessions**, by slug. The load-bearing field: an open session card is
  unfinished work that a new session must not silently start on top of.
- Deferred features, by stable ID — the backlog the defer mechanism produces.
- Requirements whose satisfaction is not `satisfied`.
- Lint error and warning counts over the same canon.
- Exits non-zero when the canon does not lint clean; exits 2 when there is no
  canon directory, distinguishing an unbootstrapped project from a failing
  one.

The package computes and does not advise. The prose telling an agent what to
do with these facts belongs to [[F-5]] and to the skills, so the same numbers
back both `ispec status` and the hook without the two drifting.

## Limitations (deferred)

- Requirement satisfaction is read from the card's own field rather than
  derived from the features that target it, so a stale hand-edited value is
  reported as written. Deriving it is the proper fix and is unimplemented.
- No JSON output mode.
- Sessions are not aged; a session left open for months reads the same as one
  opened this morning.
- No tests. The harness package is tested because it writes to user
  configuration; this one only reads.
