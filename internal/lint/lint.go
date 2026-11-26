// Package lint parses and validates an ispec card store.
//
// It owns the card schema: what fields exist, which are required per type,
// which values are legal, and how references between cards resolve. It does
// not own presentation — callers decide how issues are rendered — and it does
// not own the command line.
//
// Package status builds on Load to report on a canon without validating it,
// so the parsing half is exported separately from the checking half.
package lint

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// DefaultCanonDir is the canon directory name relative to a project root.
const DefaultCanonDir = ".ispec"

// ErrNoCanonDir reports a missing or non-directory canon path. Callers
// distinguish it because "this project has no canon" is a different outcome
// from "this project's canon is broken": the former is how every pre-bootstrap
// project looks, and it must not be reported as a lint failure.
var ErrNoCanonDir = errors.New("no canon directory")

// Card holds the parsed frontmatter plus body of a single card.
type Card struct {
	Path     string `yaml:"-"`
	Type     string `yaml:"type"`
	StableID string `yaml:"stable_id"`
	Slug     string `yaml:"slug"`
	Status   string `yaml:"status"`

	Stakeholders      []string `yaml:"stakeholders"`
	Satisfaction      string   `yaml:"satisfaction"`
	Targets           []string `yaml:"targets"`
	DependsOn         []string `yaml:"depends_on"`
	DiscoveredDuring  string   `yaml:"discovered_during"`
	Paths             []string `yaml:"paths"`
	Audience          []string `yaml:"audience"`
	SatisfiesFeatures []string `yaml:"satisfies_features"`
	Supersedes        []string `yaml:"supersedes"`
	RelatesTo         []string `yaml:"relates_to"`
	Trigger           string   `yaml:"trigger"`
	Classification    string   `yaml:"classification"`
	AffectedCards     []string `yaml:"affected_cards"`
	DeferredCards     []string `yaml:"deferred_cards"`
	Commits           []string `yaml:"commits"`
	Outcome           string   `yaml:"outcome"`

	Body string `yaml:"-"`
}

// Issue is one problem found in a card store.
type Issue struct {
	Path     string
	Severity string // "error" or "warning"
	Msg      string
}

// Result is the outcome of a Check.
type Result struct {
	Cards    []*Card
	Issues   []Issue
	Errors   int
	Warnings int
}

var (
	fmRe          = regexp.MustCompile(`(?s)\A---\n(.*?)\n---\n?(.*)\z`)
	wikilinkRe    = regexp.MustCompile(`\[\[([^\]\n]+)\]\]`)
	crossRepoRe   = regexp.MustCompile(`^[a-z][a-z0-9_-]*:[A-Za-z]+-[0-9a-zA-Z._-]+$`)
	codeSpanRe    = regexp.MustCompile("`[^`\n]*`")
	fencedBlockRe = regexp.MustCompile("(?s)```.*?```")

	statusByType = map[string][]string{
		"stakeholder": {"active", "retired"},
		"feature":     {"proposed", "active", "implemented", "deferred", "abandoned"},
		"adr":         {"proposed", "accepted", "superseded", "deprecated"},
		"session":     {"open", "resolved", "deferred", "abandoned"},
	}
	audienceValues     = []string{"end-user", "operator", "developer", "admin"}
	satisfactionValues = []string{"unsatisfied", "partially-satisfied", "satisfied"}

	cardDirs = map[string]bool{
		"stakeholders": true,
		"requirements": true,
		"features":     true,
		"spec":         true,
		"adr":          true,
		"sessions":     true,
	}
)

// CanonPath joins a project root and a canon directory name.
func CanonPath(root, canonDir string) string {
	return filepath.Join(root, canonDir)
}

// Load reads every card under the canon directory. The returned issues cover
// parse and layout failures only; schema validation is Check's job.
//
// It returns ErrNoCanonDir if the canon path is absent or not a directory.
func Load(root, canonDir string) ([]*Card, []Issue, error) {
	canon := CanonPath(root, canonDir)
	if info, err := os.Stat(canon); err != nil || !info.IsDir() {
		return nil, nil, fmt.Errorf("%w at %s", ErrNoCanonDir, canon)
	}
	cards, issues := collectCards(canon)
	return cards, issues, nil
}

// Check loads the canon, validates every card, and resolves references.
// Issues are sorted deterministically so output is diffable.
func Check(root, canonDir string) (Result, error) {
	cards, issues, err := Load(root, canonDir)
	if err != nil {
		return Result{}, err
	}
	for _, c := range cards {
		issues = append(issues, validateCard(c)...)
	}
	issues = append(issues, resolveReferences(cards)...)

	sort.SliceStable(issues, func(i, j int) bool {
		if issues[i].Path != issues[j].Path {
			return issues[i].Path < issues[j].Path
		}
		if issues[i].Severity != issues[j].Severity {
			return issues[i].Severity < issues[j].Severity
		}
		return issues[i].Msg < issues[j].Msg
	})

	res := Result{Cards: cards, Issues: issues}
	for _, i := range issues {
		if i.Severity == "error" {
			res.Errors++
		} else {
			res.Warnings++
		}
	}
	return res, nil
}

func parseCard(path string) (*Card, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	m := fmRe.FindSubmatch(data)
	if m == nil {
		return nil, fmt.Errorf("missing or malformed YAML frontmatter")
	}
	var c Card
	if err := yaml.Unmarshal(m[1], &c); err != nil {
		return nil, fmt.Errorf("frontmatter parse: %w", err)
	}
	c.Path = path
	c.Body = string(m[2])
	return &c, nil
}

func collectCards(canonDir string) ([]*Card, []Issue) {
	var cards []*Card
	var issues []Issue
	_ = filepath.WalkDir(canonDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path == canonDir {
			return nil
		}
		rel, _ := filepath.Rel(canonDir, path)
		parts := strings.Split(rel, string(filepath.Separator))

		// Top-level directories under the canon root must be recognized
		// card-type directories. Unknown ones are an error, not a silent skip.
		if d.IsDir() && len(parts) == 1 {
			if !cardDirs[parts[0]] {
				suggestion := closestCardDir(parts[0])
				var msg string
				if suggestion != "" {
					msg = fmt.Sprintf("unrecognized card-type directory %q under canon root; did you mean %q?", parts[0], suggestion)
				} else {
					msg = fmt.Sprintf("unrecognized card-type directory %q under canon root; expected one of: %s", parts[0], sortedCardDirsList())
				}
				issues = append(issues, Issue{Path: path, Severity: "error", Msg: msg})
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			// Subdirectories below recognized card-type dirs are allowed.
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		// Files at the canon root itself (e.g. .ispec/index.md) are skipped.
		if !cardDirs[parts[0]] {
			return nil
		}
		c, perr := parseCard(path)
		if perr != nil {
			issues = append(issues, Issue{Path: path, Severity: "error", Msg: perr.Error()})
			return nil
		}
		cards = append(cards, c)
		return nil
	})
	return cards, issues
}

// closestCardDir returns a recognized card-type directory name that's one
// trailing-s away from the given name, or "" if no close match.
func closestCardDir(name string) string {
	if cardDirs[name] {
		return name
	}
	if strings.HasSuffix(name, "s") {
		trimmed := strings.TrimSuffix(name, "s")
		if cardDirs[trimmed] {
			return trimmed
		}
	}
	if cardDirs[name+"s"] {
		return name + "s"
	}
	return ""
}

// sortedCardDirsList returns the recognized card-type dir names as a
// comma-separated string in deterministic order, for error messages.
func sortedCardDirsList() string {
	names := make([]string, 0, len(cardDirs))
	for n := range cardDirs {
		names = append(names, n)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

func validateCard(c *Card) []Issue {
	var issues []Issue
	add := func(sev, msg string) {
		issues = append(issues, Issue{Path: c.Path, Severity: sev, Msg: msg})
	}

	if c.Type == "" {
		add("error", "missing type")
		return issues
	}
	switch c.Type {
	case "stakeholder", "requirement", "feature", "spec", "adr", "session":
	default:
		add("error", fmt.Sprintf("unknown type: %s", c.Type))
		return issues
	}

	if c.Slug == "" {
		add("error", "missing slug")
	} else {
		base := strings.TrimSuffix(filepath.Base(c.Path), ".md")
		if base != c.Slug && !strings.HasSuffix(base, "-"+c.Slug) {
			add("error", fmt.Sprintf("slug %q does not match filename %q", c.Slug, base))
		}
	}

	if c.Type != "session" && c.StableID == "" {
		add("error", "missing stable_id")
	}

	if statuses, ok := statusByType[c.Type]; ok {
		switch c.Status {
		case "":
			add("error", "missing status")
		default:
			if !slices.Contains(statuses, c.Status) {
				add("error", fmt.Sprintf("invalid status %q for type %s; expected one of %v", c.Status, c.Type, statuses))
			}
		}
	}

	switch c.Type {
	case "requirement":
		if len(c.Stakeholders) == 0 {
			add("error", "requirement must reference at least one stakeholder")
		}
		if c.Satisfaction != "" && !slices.Contains(satisfactionValues, c.Satisfaction) {
			add("error", fmt.Sprintf("invalid satisfaction %q; expected one of %v", c.Satisfaction, satisfactionValues))
		}
	case "feature":
		if len(c.Targets) == 0 {
			add("error", "feature must have at least one targets entry")
		}
	case "spec":
		if len(c.Audience) == 0 {
			add("error", "spec section must have at least one audience tag")
		}
		for _, a := range c.Audience {
			if !slices.Contains(audienceValues, a) {
				add("error", fmt.Sprintf("invalid audience %q; expected one of %v", a, audienceValues))
			}
		}
		if len(c.SatisfiesFeatures) == 0 {
			add("error", "spec section must satisfy at least one feature")
		}
	case "session":
		if c.Trigger == "" {
			add("error", "session must have trigger")
		}
		if c.Classification == "" {
			add("error", "session must have classification")
		}
		if c.Status == "resolved" && c.Outcome == "" {
			add("error", "resolved session must have outcome")
		}
	}
	return issues
}

func resolveReferences(cards []*Card) []Issue {
	var issues []Issue

	byID := map[string]*Card{}
	bySlug := map[string]*Card{}
	for _, c := range cards {
		if c.StableID != "" {
			if dup, ok := byID[c.StableID]; ok {
				issues = append(issues, Issue{
					Path:     c.Path,
					Severity: "error",
					Msg:      fmt.Sprintf("duplicate stable_id %q (also at %s)", c.StableID, dup.Path),
				})
			} else {
				byID[c.StableID] = c
			}
		}
		if c.Slug != "" {
			if _, ok := bySlug[c.Slug]; !ok {
				bySlug[c.Slug] = c
			}
		}
	}

	resolveRef := func(c *Card, ref, expectedType, fieldName string) {
		ref = strings.TrimSpace(ref)
		if ref == "" {
			return
		}
		if crossRepoRe.MatchString(ref) {
			issues = append(issues, Issue{
				Path:     c.Path,
				Severity: "warning",
				Msg:      fmt.Sprintf("%s: cross-repo reference %q (not validated; cross-repo lint not implemented)", fieldName, ref),
			})
			return
		}
		target, ok := byID[ref]
		if !ok {
			target, ok = bySlug[ref]
		}
		if !ok {
			issues = append(issues, Issue{
				Path:     c.Path,
				Severity: "error",
				Msg:      fmt.Sprintf("%s: dangling reference %q", fieldName, ref),
			})
			return
		}
		if expectedType != "" && target.Type != expectedType {
			issues = append(issues, Issue{
				Path:     c.Path,
				Severity: "error",
				Msg:      fmt.Sprintf("%s: %q is a %s, expected %s", fieldName, ref, target.Type, expectedType),
			})
		}
	}

	for _, c := range cards {
		for _, r := range c.Stakeholders {
			resolveRef(c, r, "stakeholder", "stakeholders")
		}
		for _, r := range c.Targets {
			resolveRef(c, r, "requirement", "targets")
		}
		for _, r := range c.SatisfiesFeatures {
			resolveRef(c, r, "feature", "satisfies_features")
		}
		for _, r := range c.Supersedes {
			resolveRef(c, r, "adr", "supersedes")
		}
		for _, r := range c.DependsOn {
			resolveRef(c, r, "feature", "depends_on")
		}
		for _, r := range c.RelatesTo {
			resolveRef(c, r, "", "relates_to")
		}
		for _, r := range c.AffectedCards {
			resolveRef(c, r, "", "affected_cards")
		}
		for _, r := range c.DeferredCards {
			resolveRef(c, r, "", "deferred_cards")
		}
		if c.DiscoveredDuring != "" {
			resolveRef(c, c.DiscoveredDuring, "session", "discovered_during")
		}

		// Strip fenced code blocks and inline code before scanning for
		// wikilinks; otherwise documentation snippets like `[[wikilinks]]` or
		// `[[deploy-tool:F-12]]` are spuriously treated as references.
		clean := fencedBlockRe.ReplaceAllString(c.Body, "")
		clean = codeSpanRe.ReplaceAllString(clean, "")
		for _, m := range wikilinkRe.FindAllStringSubmatch(clean, -1) {
			ref := strings.TrimSpace(m[1])
			if i := strings.Index(ref, "|"); i >= 0 {
				ref = strings.TrimSpace(ref[:i])
			}
			resolveRef(c, ref, "", "wikilink")
		}
	}
	return issues
}
