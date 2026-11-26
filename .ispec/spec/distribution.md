---
type: spec
stable_id: SPEC-distribution
slug: distribution
audience: [developer]
satisfies_features: [F-3, F-5, F-6]
---

## Artifact

One Go binary, `ispec`, built from `cmd/ispec`. It carries its agent-facing
assets: `embed.go` at the repository root holds a single `//go:embed
skills/*/SKILL.md` directive, read at build time.

`skills/` sits at the repository root, not under `.claude/`. `go:embed` skips
dot-prefixed directories, and the assets are harness-neutral — placement is
the installer's decision, not a property of where the source lives.

The module path is `github.com/lheiskan/ispec`, which is what makes
`go install github.com/lheiskan/ispec/cmd/ispec@latest` resolve without a
clone.

## Command surface

```
ispec lint    [-root DIR] [-canon-dir NAME]
ispec status  [-root DIR] [-canon-dir NAME]
ispec harness install|check|uninstall|session-start
              [--target NAME] [--dir PATH] [--dry-run]
```

`-root` defaults to `.`, `-canon-dir` to `.ispec`.

Day-to-day canon work does not go through this binary. The cards are ordinary
markdown inside the repository the agent already has open, so reading and
writing them is plain file access. The binary covers what file access cannot
do: validating the reference graph, summarising state across all cards, and
installing itself.

### Exit codes

| Code | Meaning |
|---|---|
| 0 | clean |
| 1 | problems found — lint errors, or `harness check` found work to do |
| 2 | no canon directory at the given root |

2 is distinct on purpose: an unbootstrapped project must be distinguishable
from a failing one, since every project looks like the former before adoption.

## Package layout

| Package | Owns |
|---|---|
| `internal/lint` | the card schema: fields, required-per-type, legal values, reference resolution |
| `internal/status` | facts about a canon — counts, open sessions, deferred work, lint state. Computes; does not advise |
| `internal/harness` | placement per harness, and the session-start directive |
| `cmd/ispec` | flag parsing, output, exit codes |

The validation engine is a package rather than a `main`, so `status` and the
session-start directive read a canon directly instead of shelling out.

## Harness integration

`--target NAME` selects a harness from a registry; an unknown name is an error
listing what is accepted, never a fallback to the default. Each target
implements `Install`, `Check`, `Uninstall` and reports its own default config
directory.

### claude-code

| What | Where |
|---|---|
| Skills | `<dir>/skills/<name>/SKILL.md`, one directory per skill |
| Session-start hook | a group appended to `.hooks.SessionStart` in `<dir>/settings.json` |

`<dir>` is `$CLAUDE_DIR`, else `~/.claude`, else `--dir`. The hook group is:

```json
{
  "matcher": "",
  "hooks": [
    { "type": "command", "command": "<abs-path-to-ispec> harness session-start" }
  ]
}
```

The command is absolute because a hook runs in a non-interactive shell whose
`$PATH` may not contain the install prefix. It is quoted when the path
contains whitespace.

### Write rules

Per [[ADR-0005]]:

1. The settings file is decoded to ordered key/raw-value pairs. Only
   `.hooks.SessionStart` is rewritten; every other member is re-emitted from
   its original bytes, so member order and unmodelled fields survive.
2. Install is idempotent. An existing ispec group is repointed, not
   duplicated.
3. A pre-existing settings file is copied to `settings.json.ispec-bak` before
   the first edit, once only.
4. Uninstall removes only what install wrote.
5. A settings file that is not valid JSON is refused.
6. Writes go through a temporary file in the destination directory, then
   rename.
7. Symlinked install destinations get no special handling. A skill directory
   that is a symlink is written through, so the file behind it is replaced.
   The case is not defended against: a dotfiles setup symlinks the harness
   config directory as a whole rather than individual skills, and such a
   setup keeps the files under version control anyway.

Hook ownership keys on the **binary name**, not the subcommand:
`harness session-start` is a convention shared with sibling tools, so
subcommand-only matching would make ispec adopt and then delete another
tool's hook.

## Session-start directive

`ispec harness session-start` reads the canon at `-root` and prints:

1. The [[F-6]] status lines.
2. Guidance, most-blocking first — open sessions, then a canon that does not
   lint clean — ending with the standing instruction to invoke `ispec-start-session`
   for substantive work.

It prints **nothing** when the project has no canon, and exits 0 regardless.
The hook is installed user-globally and fires in every session in every
repository; announcing ispec in the ones that do not use it would be noise,
and an implicit demand to adopt it. A broken canon is reported as a fact
rather than raised as an error: a hook that fails loudly at session start
degrades the session it exists to help.
