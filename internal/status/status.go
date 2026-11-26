// Package status reports facts about a project's canon.
//
// It computes; it does not advise. The prose that tells an agent what to do
// with these facts lives in package harness (for the session-start directive)
// and in the skills themselves. Keeping the split means the same numbers back
// both `ispec status` and the hook without the two drifting.
package status

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/lheiskan/ispec/internal/lint"
)

// cardTypes is the report order: the atemporal canon first, in the order the
// data model presents it, with sessions last because they are the
// chronological record rather than part of the reference.
var cardTypes = []string{"stakeholder", "requirement", "feature", "spec", "adr", "session"}

// Status is a snapshot of one project's canon.
type Status struct {
	// Counts is cards per type; absent types are zero.
	Counts map[string]int
	Total  int

	// OpenSessions holds the slugs of sessions still open. This is the
	// load-bearing field: an open session is unfinished work that a new
	// session must not silently start on top of.
	OpenSessions []string

	// DeferredFeatures holds stable IDs of features parked for later — the
	// backlog that falls out of the defer mechanism.
	DeferredFeatures []string

	// Unsatisfied holds stable IDs of requirements that no implemented
	// feature fully covers, partially-satisfied included.
	Unsatisfied []string

	// Errors and Warnings are lint counts over the same canon.
	Errors   int
	Warnings int
}

// Compute loads the canon and derives a Status. It propagates
// lint.ErrNoCanonDir unchanged so callers can distinguish "no canon here"
// from a real failure.
func Compute(root, canonDir string) (Status, error) {
	res, err := lint.Check(root, canonDir)
	if err != nil {
		return Status{}, err
	}

	s := Status{
		Counts:   map[string]int{},
		Total:    len(res.Cards),
		Errors:   res.Errors,
		Warnings: res.Warnings,
	}
	for _, c := range res.Cards {
		s.Counts[c.Type]++
		switch c.Type {
		case "session":
			if c.Status == "open" {
				s.OpenSessions = append(s.OpenSessions, c.Slug)
			}
		case "feature":
			if c.Status == "deferred" {
				s.DeferredFeatures = append(s.DeferredFeatures, identify(c))
			}
		case "requirement":
			if c.Satisfaction != "" && c.Satisfaction != "satisfied" {
				s.Unsatisfied = append(s.Unsatisfied, identify(c))
			}
		}
	}
	sort.Strings(s.OpenSessions)
	sort.Strings(s.DeferredFeatures)
	sort.Strings(s.Unsatisfied)
	return s, nil
}

// identify prefers the stable ID, falling back to the slug. A card missing
// both is already a lint error; reporting "?" keeps status readable rather
// than printing an empty entry that looks like a formatting bug.
func identify(c *lint.Card) string {
	switch {
	case c.StableID != "":
		return c.StableID
	case c.Slug != "":
		return c.Slug
	default:
		return "?"
	}
}

// Clean reports whether the canon has no lint errors.
func (s Status) Clean() bool { return s.Errors == 0 }

// Write renders the status as human-readable lines.
func (s Status) Write(w io.Writer) {
	fmt.Fprintf(w, "%d card(s)", s.Total)
	var parts []string
	for _, t := range cardTypes {
		if n := s.Counts[t]; n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n, plural(t, n)))
		}
	}
	if len(parts) > 0 {
		fmt.Fprintf(w, ": %s", strings.Join(parts, ", "))
	}
	fmt.Fprintln(w)

	fmt.Fprintf(w, "lint: %d error(s), %d warning(s)\n", s.Errors, s.Warnings)

	writeList(w, "open sessions", s.OpenSessions)
	writeList(w, "deferred features", s.DeferredFeatures)
	writeList(w, "requirements not yet satisfied", s.Unsatisfied)
}

func writeList(w io.Writer, label string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(w, "%s (%d): %s\n", label, len(items), strings.Join(items, ", "))
}

// plural pluralizes a card-type name for display. "spec" is left alone
// because "specs" is the exact typo the lint's directory check exists to
// catch, and printing it here would teach the wrong spelling.
func plural(t string, n int) string {
	if n == 1 || t == "spec" {
		return t
	}
	return t + "s"
}
