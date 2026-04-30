// ispec-lint validates the structure and consistency of an ispec card store.
//
// Usage:
//
//	ispec-lint [-root <project-root>] [-canon-dir <subdir>]
//
// Walks <project-root>/<canon-dir> (default .ispec), parses YAML frontmatter
// on every .md file under recognized card-type directories (stakeholders,
// requirements, features, spec, adr, sessions), validates per-type schema
// rules, and resolves cross-card references (frontmatter ID lists and body
// [[wikilinks]]).
//
// Cross-repo references of the form <repo>:<id> are noted as warnings but not
// resolved; cross-repo lint is not yet implemented.
//
// Exits 0 if no errors, 1 if any errors, 2 on usage / IO failure.
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

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

type Issue struct {
	Path     string
	Severity string
	Msg      string
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

func collectCards(docsDir string) ([]*Card, []Issue) {
	var cards []*Card
	var issues []Issue
	_ = filepath.WalkDir(docsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		rel, _ := filepath.Rel(docsDir, path)
		top := strings.SplitN(rel, string(filepath.Separator), 2)[0]
		if !cardDirs[top] {
			// Files like docs/index.md, docs/design.md, docs/build/* are not cards.
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
			if !contains(statuses, c.Status) {
				add("error", fmt.Sprintf("invalid status %q for type %s; expected one of %v", c.Status, c.Type, statuses))
			}
		}
	}

	switch c.Type {
	case "requirement":
		if len(c.Stakeholders) == 0 {
			add("error", "requirement must reference at least one stakeholder")
		}
		if c.Satisfaction != "" && !contains(satisfactionValues, c.Satisfaction) {
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
			if !contains(audienceValues, a) {
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

func contains(set []string, v string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
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

func main() {
	rootFlag := flag.String("root", ".", "project root containing the canon directory")
	canonDirFlag := flag.String("canon-dir", ".ispec", "directory under root containing card files")
	flag.Parse()

	canonDir := filepath.Join(*rootFlag, *canonDirFlag)
	if info, err := os.Stat(canonDir); err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "ispec-lint: no canon directory at %s\n", canonDir)
		os.Exit(2)
	}

	cards, issues := collectCards(canonDir)
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

	var errs, warns int
	for _, i := range issues {
		rel, err := filepath.Rel(*rootFlag, i.Path)
		if err != nil {
			rel = i.Path
		}
		fmt.Printf("%s: [%s] %s\n", rel, i.Severity, i.Msg)
		if i.Severity == "error" {
			errs++
		} else {
			warns++
		}
	}
	fmt.Printf("\n%d card(s) checked. %d error(s), %d warning(s).\n", len(cards), errs, warns)
	if errs > 0 {
		os.Exit(1)
	}
}
