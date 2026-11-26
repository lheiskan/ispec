---
type: session
slug: 2026-11-19-single-binary-distribution
trigger: the symlink install story is too weak to put in front of a public audience; an embedded-asset single binary is a better model
classification: new-feature + refactor
status: resolved
affected_cards:
  - REQ-3
  - F-1
  - F-2
  - F-3
  - F-4
  - F-5
  - F-6
  - ADR-0002
  - ADR-0003
  - ADR-0004
  - ADR-0005
  - SPEC-distribution
deferred_cards: []
commits: []
outcome: |
  ispec is now distributed as one Go binary that carries its own workflow
  skills, installed with `go install github.com/lheiskan/ispec/cmd/ispec@latest`
  followed by `ispec harness install`. No checkout is needed to install and
  none is referenced afterwards, which is what [[REQ-3]] asked for and what
  the symlink install could never give: it required the user to keep a
  checkout in place forever, on the right branch, and silently skipped any
  install destination that already existed.

  The unlock was `go:embed`. [[ADR-0003]] had rejected real distribution on
  the grounds that markdown skills are not binary content and would need a
  second delivery mechanism; embedding makes them binary content, so one
  artifact carries both halves. [[ADR-0004]] records that and supersedes
  ADR-0003.

  `ispec-lint` became `ispec`, with `lint` as a subcommand alongside the new
  `status` and the `harness` namespace. The validation engine moved into
  `internal/lint` so that `status` and the session-start directive can read a
  canon directly rather than shelling out to a lint run.

  The session-start hook is the behavioural change an adopter will notice
  most: a session opened in a project that has a canon is now told what state
  that canon is in — card counts, lint result, and any session left open —
  without the user remembering to ask. In projects with no canon it prints
  nothing at all, which [[ADR-0005]] records as a deliberate position rather
  than an implementation detail: the hook is user-global and fires in every
  session on the machine, and ispec does not advertise itself in repositories
  that have not adopted it.

  [[ADR-0005]] also governs how the installer touches configuration the user
  owns. Writes preserve member order and fields ispec does not model, are
  idempotent, back the original up once, and are reversible. Two hazards were
  found by testing rather than by reading: hook ownership has to key on the
  binary name, because `harness session-start` is a convention shared with
  sibling tools and subcommand-matching would have made ispec adopt another
  tool's hook on install and delete it on uninstall; and a symlinked skill
  *directory* means the destination file path resolves into the user's
  checkout, so an install over ispec's own older layout would have modified
  tracked source. Installing from a `go run` binary is refused outright,
  since the recorded hook path would vanish with the build cache.

  The five skills were renamed with an `ispec-` prefix. The install now
  writes real directories into a shared `~/.claude/skills/`, where
  `bootstrap` and `defer` were collisions waiting to happen. The trigger was
  writing the public README — a naming problem three sessions of internal
  work had not surfaced.

  This session also added the module's first tests, covering
  `internal/harness`, on the grounds that it is the part of ispec that writes
  to a user's global configuration.
---

## Trigger

The repo is about to go public. The install story is `make install`, which
symlinks five skill directories and the `ispec-lint` binary out of a live
checkout the user must then keep forever. That is a weak first impression and
a real footgun: INSTALL.md has to warn "do not move or delete the checkout."

A better model: embed the agent-facing markdown into the binary with
`go:embed`, then install from the binary itself under a single `harness`
namespace — no checkout retained, idempotent, `--dry-run`, a `check` verb that
reports staleness, and an `uninstall` that removes only what install wrote.

The decisive detail is that embedding dissolves the objection [[ADR-0003]]
recorded against real distribution: "the skills aren't binary content — they're
markdown — so a Go-only distribution doesn't cover everything." With
`go:embed` the markdown *is* binary content, so one `go install` covers
skills and tooling together.

## Pre-decisions

- **Supersede [[ADR-0003]] rather than amend it.** The symlink decision was
  explicitly scoped to v0.1 and predicted its own supersession.
- **Reuse [[REQ-3]].** The requirement ("usable from any project via a shared
  installation, without the adopting project hosting ispec's source") is
  unchanged; this work moves it from partially-satisfied toward satisfied.
- **Amend [[F-3]] in place.** The feature's identity is "how ispec is
  installed"; only its mechanism changes, so the stable ID stays.
- **No MCP server.** ispec's canon lives in `.ispec/` inside the repo the
  agent already has open, so Read/Grep/Edit already reach every card. An MCP
  tool surface would reimplement file access behind a schema that costs tokens
  in every session. A tool whose store sits outside the repository needs a
  tool surface to reach it at all; ispec does not have that problem.
- **`ispec status` is in scope** because the session-start hook needs
  something to print. `next-id` and cross-repo resolution are not.
- **Proper module path from the start.** `go.mod` is already
  `github.com/lheiskan/ispec`, so `go install github.com/lheiskan/ispec/cmd/ispec@latest`
  works with no checkout at any point. A module declared with a bare name
  rather than its host path cannot do this: `go install <path>@latest` has
  nothing to resolve, leaving `go install ./cmd/<name>` from a clone as the
  only option.

## Discoveries during the session

- **The `harness session-start` subcommand name is a shared convention, and
  that makes it dangerous.** Any tool that installs an agent hook this way
  plausibly spells it `<its-binary> harness session-start`, exactly as ispec
  does. Detecting "our" hook by matching the subcommand — the obvious
  implementation, and a safe one for a tool that only has to recognise itself
  — would make ispec adopt a sibling tool's hook on install and delete it on
  uninstall. Ownership has to key on the binary name. Any future tool copying
  this layout inherits the same trap.

- **Writing a file is not enough to be symlink-safe.** The contract flagged
  that a skill destination might be a symlink into the checkout, and the fix
  was stated as one line: unlink before write. That was wrong. The old install
  symlinked the skill *directory*, so the destination *file path* resolves
  through the link, and `os.Remove(dest)` deletes the file inside the user's
  repository before writing a new one there. The real fix touches install,
  check and uninstall, and the distinction that matters is ownership: a
  directory ispec owns gets replaced, while the harness's own config directory
  is followed, because a user may legitimately symlink `~/.claude` into a
  dotfiles repo. Found by testing, not by reading.

- **A user-global hook forces a question ispec had not asked: what does ispec
  look like in a project that does not use it?** The hook fires in every
  session in every repository. The answer — print nothing, exit 0 — is now
  [[ADR-0005]], but it is a genuine design position (ispec does not advertise
  itself) that only surfaced because the delivery mechanism changed. A tool
  with a single global store never faces the question, because it is always
  relevant.

- **Preserving a user's config costs more code than editing it.**
  `internal/harness/jsonedit.go` exists only so that installing appends one
  hook group without reordering the user's keys or dropping fields ispec does
  not model. It is the single largest piece of this session that does nothing
  for ispec's own functionality. Worth it: the install diff is the first thing
  a new user inspects, and a 200-line reshuffle of their settings file is how
  an installer loses trust.

- **`filepath` is the wrong tool for parsing a path that came from a file.**
  `filepath.ToSlash` follows the *host* convention, so on macOS it left a
  backslashed Windows path intact and the whole string read as one filename.
  A settings file can be written on one platform and read on another; path
  strings out of config need manual separator handling.

- **Skill names were wrong in a way only the README surfaced.** The rename to
  `ispec-*` came from the user's own README draft, not from the contract.
  `bootstrap` and `defer` are plausible names for any tool's skill, and the
  install now writes real directories into a shared `~/.claude/skills/`, so
  the generic names were a collision waiting to happen. Worth noting that the
  trigger was writing user-facing documentation — the public-readme pass found
  a naming problem that three sessions of internal work had not.

- **Historical session cards were deliberately left naming the old skills.**
  Sessions are the chronological record; rewriting them to say
  `ispec-start-session` would be anachronistic. The rename is recorded here
  instead. This is the first time ispec's own canon has had to distinguish
  "atemporal card, update it" from "session record, leave it", and the design
  doc does not say which wins.

- **`defer` and `ispec-close-session` still have no feature cards.** Only
  `ispec-start-session` ([[F-2]]) and `ispec-retrofit` ([[F-4]]) are covered.
  Noticed while updating `paths:` across the feature cards; out of scope here.

## Outcome

(filled at session close)

## Note on commits

This session's `commits` list is empty because the repository's history was
squashed to a single commit when it was prepared for publication. The work
described here is part of that initial commit. A one-commit history cannot
hold a card that cites its own SHA: writing the SHA changes the commit, which
changes the SHA.

Worth recording that the squash cost the git-history half of ispec's corpus —
the design doc treats commit messages as material future agents read — while
the session cards survived it intact, because they are files in the canon
rather than metadata about it. The chronological record moved entirely into
`.ispec/sessions/`.
