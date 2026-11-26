---
type: adr
stable_id: ADR-0005
slug: harness-registry-and-reversible-writes
status: accepted
---

## Context

Installing ispec into an agent harness means writing into configuration the
user owns: skill directories, and — for the session-start hook — the global
settings file that drives every session on the machine. Two questions needed
answering before writing any of it.

**Which harness?** Claude Code is the only one supported today. The cheap path
is to hardcode `~/.claude` throughout the installer. The cost arrives later:
every caller silently assumes one layout, and adding a second harness becomes
a rewrite rather than an addition. ispec's own README lists agent-agnostic
packaging as an unbuilt gap; the workflow instructions themselves contain
nothing Claude-specific.

**How invasively?** The settings file holds keys this binary has never heard
of, in an order the user reads, and ispec owns exactly one entry inside it.
Decoding it into a map and re-marshalling would reorder every key and rewrite
every number — turning a one-line install into an unreviewable diff.

## Decision

Harness integration is a named-target registry. `--target NAME` selects one;
an unknown name is an error listing what is accepted, never a silent fallback
to the default. Each target owns its own placement behind a `Target`
interface, so a second harness is a new file rather than a change to the
install flow.

Writes into user configuration obey four rules:

1. **Order-preserving.** The settings file is decoded to ordered key/raw-value
   pairs. Only the one member ispec owns is replaced; every other value is
   re-emitted from its original bytes, so unknown fields survive verbatim.
2. **Idempotent.** A second install reports `unchanged` and writes nothing.
   Ownership is detected and the existing entry repointed, never duplicated.
3. **Backed up once.** The first edit of a pre-existing file copies it aside.
   Only the first: a later install must not overwrite the pristine backup with
   an already-modified file.
4. **Reversible.** `uninstall` removes only what install wrote.

Ownership of a hook entry keys on the **binary name**, not the subcommand.
`harness session-start` is a convention shared with sibling tools built the
same way, so matching the subcommand alone would make ispec adopt another
tool's hook on install and delete it on uninstall.

A settings file that is not valid JSON is never modified; it is more likely a
config ispec does not understand than one it should replace.

The session-start directive **prints nothing when the project has no canon.**
The hook is installed user-globally and so fires in every session in every
repository, the overwhelming majority of which do not use ispec.

## Consequences

- Adding a harness is additive, and `--target cursor` fails loudly today
  instead of quietly installing Claude Code's layout.
- An install against a real settings file produces a diff a user can read:
  one appended hook group, nothing else touched.
- A sibling tool's identically-shaped hook survives both install and
  uninstall.
- ispec is silent in non-ispec projects, so the hook does not train the user
  to ignore it, and does not push a workflow onto unrelated work.
- Order-preserving JSON editing is machinery ispec has to carry
  (`internal/harness/jsonedit.go`). The alternative was a worse diff on every
  install, which is the thing most likely to make a user distrust an
  installer.
- The registry is overhead while there is one target. Accepted deliberately:
  the cost is one interface and a lookup, paid once.

## Alternatives considered

- **Hardcode `~/.claude`.** Rejected per above.
- **Marshal settings through `map[string]any`.** Rejected: reorders keys,
  rewrites numbers, and loses nothing but changes everything.
- **Append the hook without ownership detection.** Rejected: every install
  would add another duplicate group.
- **Let the hook announce ispec in every project.** Rejected: noise in the
  vast majority of sessions, and an implicit demand to adopt ispec in repos
  that have not.
