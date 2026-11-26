# Installing ispec

ispec is a single Go binary with its workflow skills embedded. Installing it
needs no checkout, and nothing in a checkout is referenced afterwards.

## Prerequisites

- **Go 1.21 or later.** There are no prebuilt binaries yet.
- A directory on your `$PATH` for the binary — whatever `go env GOBIN` reports,
  or `$(go env GOPATH)/bin` if `GOBIN` is unset.
- An agent harness ispec knows how to install into. `ispec harness help` lists
  them; Claude Code is the only one today.

## Install

```sh
go install github.com/lheiskan/ispec/cmd/ispec@latest
ispec harness install
```

The first command fetches, builds and installs the binary. The second writes
the integration into your harness — for Claude Code, that is:

| What | Where |
|---|---|
| The five workflow skills | `~/.claude/skills/<name>/SKILL.md` |
| A session-start hook | a group appended to `.hooks.SessionStart` in `~/.claude/settings.json` |

Both steps are safe to repeat. `install` is idempotent: it reports
`unchanged` and writes nothing when the install is already current.

To see what it would do without writing anything:

```sh
ispec harness install --dry-run
```

### What the install does to settings.json

It appends one hook group and changes nothing else. The file is decoded to
ordered key/value pairs and only `.hooks.SessionStart` is rewritten, so member
order and any fields ispec does not model survive verbatim. Before the first
edit, the original is copied to `~/.claude/settings.json.ispec-bak`.

A settings file that is not valid JSON is refused rather than overwritten.

If another tool has installed its own `harness session-start` hook, it is left
alone — ispec identifies its own hook by the binary name, not the subcommand.

## Verify

```sh
ispec harness check
```

Reports each item as `unchanged`, `missing`, or `stale`, and exits non-zero if
anything needs installing — so it works in a script. Then, in any project with
a canon:

```sh
ispec status
```

## From a checkout

If you are working on ispec itself:

```sh
make install     # go install, then have the installed copy install the harness
make check       # ispec harness check
make build       # build ./ispec locally without installing
make test        # go test ./...
```

`make install` deliberately runs `go install` first and then invokes the
*installed* binary, so the hook records a stable absolute path rather than one
pointing into your working tree.

Because the skills are embedded at build time, **editing a skill requires a
rebuild before the change takes effect.** The development loop is `make
install` — rebuild, reinstall, and the harness integration is refreshed from
the new binary.

Note that `go run ./cmd/ispec harness install` is *refused*: `go run` builds
into Go's build cache, so the hook would record a path that disappears when
the cache is cleaned, breaking every session on the machine. The installer
detects this and tells you to install a real binary instead.

`ispec harness check` tells you when an installed skill has fallen behind the
binary.

## Update

```sh
go install github.com/lheiskan/ispec/cmd/ispec@latest
ispec harness install
```

The second step is needed because the skills live in the binary: a new binary
carries new skills, but the copies under `~/.claude/skills/` are only refreshed
when you ask. `ispec harness check` reports the gap.

## Uninstall

```sh
ispec harness uninstall
```

Removes the skills it installed and its own hook group, and nothing else. The
binary itself is left in place — delete it from `$(go env GOPATH)/bin` if you
want it gone.

## Custom locations

```sh
ispec harness install --target claude-code --dir /custom/claude-dir
```

`--dir` overrides the harness's config directory; `$CLAUDE_DIR` does the same
for the Claude Code target. The binary's own location is Go's business — set
`GOBIN` before `go install` to place it elsewhere.

## Notes

- No per-project version pinning: every project on the machine shares the
  installed binary. `go install ...@v1.2.3` pins a version globally if you
  need one.
- The session-start hook prints nothing in projects that have no `.ispec/`
  canon, so installing it does not affect your other work.
