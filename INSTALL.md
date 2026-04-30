# Installing ispec

Make ispec's workflow skills and the `ispec-lint` binary available to any Claude Code session in any project, via symlinks into the ispec checkout.

## Prerequisites

- Go 1.21 or later (only required for building `ispec-lint`).
- A directory on your `$PATH` for the binary. The default is `~/.local/bin`; configurable via `BIN_DST`.
- Claude Code or another Claude-Code-compatible client that reads skills from `~/.claude/skills/`.

## Install

From the ispec checkout directory:

```sh
make install
```

This:

1. Builds `ispec-lint` in the project root.
2. Symlinks each workflow skill from `<ispec-checkout>/.claude/skills/<skill>` into `~/.claude/skills/<skill>`. The skills installed are `bootstrap`, `defer`, `retrofit`, `session-close`, and `spec-session`.
3. Symlinks `<ispec-checkout>/ispec-lint` into `~/.local/bin/ispec-lint`.

Existing entries (real files or other symlinks) at install destinations are left in place; the install reports `skipping` for those. To replace, run `make uninstall` first, then `make install`.

## Verify

```sh
ispec-lint -h
ls ~/.claude/skills/spec-session/SKILL.md
```

In a fresh Claude Code session in any project, the workflow skills are now discoverable by name (`/spec-session`, `/bootstrap`, etc., depending on how your client surfaces them).

To check whether a project has an ispec canon and lint clean:

```sh
ispec-lint -root <project>
```

Exits 0 if clean, 1 if errors, 2 if `<project>/.ispec/` is missing.

## Update

After pulling new ispec changes, rebuild the binary:

```sh
git pull
make build
```

Skill changes propagate immediately via the symlinks; only the binary needs an explicit rebuild.

## Uninstall

```sh
make uninstall
```

Removes ispec's symlinks. Does not delete ispec's source or the built binary.

## Custom install locations

```sh
make install SKILLS_DST=/custom/skills BIN_DST=/custom/bin
```

Both variables default to `~/.claude/skills` and `~/.local/bin` respectively. `make uninstall` honors the same overrides.

## Notes

- Symlinks point at the live ispec checkout; do not move or delete the checkout while installs are active.
- There is no per-project version pinning. Every project sees the same live ispec; intentional for v0.1 (see `.ispec/adr/0003-install-via-symlinks.md`). A versioned distribution path is planned for once ispec stabilizes.
