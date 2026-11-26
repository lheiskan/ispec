// Package ispec exists only to embed the repository's agent-facing assets.
//
// `go:embed` cannot reach outside its own package directory, so embedding
// skills/ from internal/harness would mean either a `../` (illegal) or moving
// the directory under internal/. A root package holding nothing but the
// directive is the cheapest way to keep skills/<name>/SKILL.md as the single
// source of truth.
//
// They are embedded rather than read from disk because `ispec harness install`
// has to work with no repository checkout present — that is the whole point of
// the distribution model (see ADR-0004). The directive reads at build time, so
// a freshly built binary cannot carry a stale copy.
//
// The directory is skills/ at the repository root rather than
// .claude/skills/: go:embed skips dot-prefixed directories by default, and
// more importantly the assets are harness-neutral. Which harness they get
// installed into is package harness's decision, not a fact baked into where
// they are stored.
package ispec

import "embed"

// Skills holds the bundled SKILL.md files, keyed skills/<name>/SKILL.md.
//
//go:embed skills/*/SKILL.md
var Skills embed.FS
