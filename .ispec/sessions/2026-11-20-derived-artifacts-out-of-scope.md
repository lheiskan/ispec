---
type: session
slug: 2026-11-20-derived-artifacts-out-of-scope
trigger: derived-artifact generation is not ispec's job — ispec captures intent so other tools have better context when they produce those artifacts
classification: scope-decision
status: resolved
affected_cards:
  - STK-1
  - STK-2
  - REQ-1
  - ADR-0002
  - ADR-0006
deferred_cards: []
commits: []
outcome: |
  ispec is now explicitly an intent-capture system rather than a
  documentation generator. Producing documents from the canon — a full
  SPEC.md, a user manual, an operator runbook, API docs, a backlog view — is
  out of scope and recorded as such in [[ADR-0006]], not carried as deferred
  work. What ispec delivers is the canon, the workflow skills that keep it
  accurate, and the lint that verifies its structure; what a consumer makes
  from that is the consumer's concern.

  The decision came from reviewing the README before publication, and the
  reasoning is in the evidence rather than in taste: view generation had been
  planned since ispec's first design commit and had never once been the most
  valuable next thing to build. Every bit of value had arrived from the
  capture side — the contract agreed before coding, the reconciliation after,
  the deferred card carrying `discovered_during:`. Treating a feature that
  stayed unbuilt through a project's entire life as a signal rather than a
  backlog item is what this session acted on.

  Narrowing the scope made [[REQ-1]] satisfiable: it asks the system to
  *enable production of* atemporal canonical documentation, which the canon
  and its workflows do. It had been held at `partially-satisfied` against a
  bar that turned out not to be its own. [[ADR-0002]] stopped justifying the
  `.ispec/` versus `docs/` split as source-versus-build-output and now rests
  on authority: `.ispec/` is tool-managed and lint-verified, `docs/` belongs
  to the project. [[STK-1]] wants a record good enough that other tools do
  regeneration and user-doc production well from it; [[STK-2]] consumes the
  canon as context rather than producing artifacts inside ispec's loop.

  Two skills lost a step. `ispec-close-session` no longer regenerates derived
  docs, and the anti-pattern that scolded an agent for skipping that step is
  gone with it. `ispec-defer` no longer emits a backlog document; the
  deferred cards already are the backlog and `ispec status` lists them, so
  the step now verifies the deferral is visible instead of producing a file.

  One capability was genuinely given up, and the canon says so rather than
  wording around it: audience-filtered generation would have made
  audience leakage structurally impossible, because a generator that cannot
  see operator-tagged content cannot leak it. ispec now supplies the tags and
  cannot enforce the filtering. Leak prevention became the consumer's
  property.
---

## Trigger

Reviewing the README before publication, the user reframed what ispec is for:

> might be that derived artifacts are done by other tools — ispec core feature
> is to capture additional 'intent' at development time, with the hope that
> this can provide better quality context information when other tools work on
> producing derived artifacts for the project. so i think all that stuff is out
> of scope currently and probably always.

This is a narrowing of purpose, not a deferral. ispec has carried
view generation as a planned-but-unbuilt feature since its first design
commit: a `## Generated views` section in `docs/design.md` listing six
documents to emit, a `Regenerate derived docs` step in
`ispec-close-session`, a `BACKLOG.md` regeneration step in `ispec-defer`,
and a [[REQ-1]] held at `partially-satisfied` specifically because that
generation did not exist.

Under the new framing, none of that is missing — it was never ispec's to
build. What ispec produces is a well-maintained, structurally-verified
record of intent; what a consumer does with it is the consumer's concern.

## Pre-decisions

- **A scope boundary, recorded as an ADR.** "We decided what this project is
  not" is a durable decision with consequences, and belongs in canon rather
  than only in a README edit.
- **`audience:` stays.** It was introduced to route ispec's own doc
  generation, but it survives the narrowing with its value intact: it is
  intent about who a spec section is *for*, which is exactly the kind of
  context a consuming tool cannot infer from code. ispec keeps validating it
  and stops claiming it will act on it.
- **`ispec status` replaces `BACKLOG.md`.** The deferred cards already *are*
  the backlog; [[F-6]] lists them. The defer skill does not need to emit a
  document to make deferred work visible.
- **No new feature card.** Nothing is being built. This session removes
  claims and records a boundary.

## Discoveries during the session

- **The classification taxonomy has no slot for a scope decision.**
  `ispec-start-session` offers six classifications — bad implementation,
  missed corner case, new feature (with or without an existing requirement),
  refactor, operational change. None covers deciding what a project *will not
  do*, which is among the most durable decisions a project makes and the one
  most likely to need an ADR. Classified as `scope-decision` rather than
  forced into "refactor". The taxonomy should probably grow this category.

- **One unbuilt feature had accumulated six dependents.** View generation was
  never implemented, but it had propagated into [[REQ-1]]'s satisfaction
  value, [[ADR-0002]]'s rationale for the `.ispec/` / `docs/` split, a goal on
  [[STK-1]], the identity of [[STK-2]], a step in `ispec-close-session`, a
  step in `ispec-defer`, a section of `docs/design.md`, a future-lint item in
  `docs/schemas.md`, and an anti-pattern claim. Removing it meant editing all
  of them. The reference graph made them *findable*, and lint caught the
  dangling [[ADR-0006]] reference the moment the session card cited it — but
  nothing flagged that an unimplemented capability was acquiring dependents.
  A "claims versus capabilities" check is the lint this wants, and ispec has
  no notion of it: cards can promise anything.

- **Giving up a capability honestly is harder than adding one.** The
  audience-leakage anti-pattern was a real structural guarantee and only
  generation could deliver it: a generator that physically cannot see
  operator-tagged content cannot leak it. With generation gone, ispec supplies
  tags and nothing more. The easy move was to soften the wording until the
  loss disappeared; the honest one was to state that leak prevention is now
  the consumer's property. Worth noting because the same pressure will apply
  to every future narrowing.

- **`harness check` validated itself.** Editing two SKILL.md files made the
  freshly built binary report exactly those two as `stale`, while the
  installed binary — still carrying the old embedded copies — reported
  everything `unchanged`. Both answers are correct from their own vantage.
  [[ADR-0004]] recorded "editing a skill now requires a rebuild" as the cost
  of embedding; this is the detection mechanism that makes the cost
  survivable, working as designed on its first real test.

- **A requirement became satisfied by removing scope, not by building.**
  [[REQ-1]] moved `partially-satisfied` → `satisfied` because the bar it was
  being held to turned out not to belong to it. That is a legitimate path to
  satisfaction and the derived-from-features rule handled it without
  contortion — every feature targeting REQ-1 was already `implemented`. But
  it is worth being suspicious of: "we decided the gap did not count" is also
  how a requirement gets quietly hollowed out, and nothing distinguishes the
  two cases mechanically.

- **`audience:` survived with its meaning changed rather than its data.** It
  was introduced as generator routing and is now captured intent about who a
  section is for. Same field, same values, same lint — different consumer.
  `docs/schemas.md` documents the field but not why it exists, so nothing in
  the schema records that shift.

## Outcome

(filled at session close)

## Note on commits

Empty for the same reason as
[[2026-11-19-single-binary-distribution]]: the repository's history was
squashed to a single commit when it was prepared for publication, and a
one-commit history cannot hold a card citing its own SHA. This session's work
is part of that initial commit.
