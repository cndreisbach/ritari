// Package lint reports conformance errors and hygiene warnings (§8).
package lint

import (
	"fmt"
	"path"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/cndreisbach/ritari/internal/bundle"
	"github.com/cndreisbach/ritari/internal/note"
)

// Severities.
const (
	Error   = "error"
	Warning = "warning"
)

// Rule names.
const (
	Unparseable           = "unparseable"
	MissingType           = "missing-type"
	UnknownType           = "unknown-type"
	MissingTitle          = "missing-title"
	MissingDescription    = "missing-description"
	TypeDirectoryMismatch = "type-directory-mismatch"
	DuplicateTitle        = "duplicate-title"
	DanglingLink          = "dangling-link"
	StaleIndex            = "stale-index"
	UnknownCategory       = "unknown-category"
	StaleVerification     = "stale-verification"
	Orphan                = "orphan"
	ReservedCasing        = "reserved-casing"
)

// Rules lists every rule name, sorted.
var Rules = []string{
	DanglingLink, DuplicateTitle, MissingDescription, MissingTitle, MissingType, Orphan,
	ReservedCasing, StaleIndex, StaleVerification, TypeDirectoryMismatch, UnknownCategory, UnknownType, Unparseable,
}

// Finding is one problem.
type Finding struct {
	Severity string
	Rule     string
	ID       string
	Path     string
	Message  string
}

// Options filter and tune a run.
type Options struct {
	Conformance bool
	Rules       []string
	Strict      bool
	Categories  []string
}

// Report is the outcome of a run.
type Report struct {
	Findings []Finding
	Errors   int
	Warnings int
	// Failure is the reason the run fails, or "".
	Failure string
}

// Run lints the bundle.
func Run(b *bundle.Bundle, o Options) (Report, error) {
	for _, r := range o.Rules {
		if !slices.Contains(Rules, r) {
			return Report{}, fmt.Errorf("unknown rule %q (known: %s)", r, strings.Join(Rules, ", "))
		}
	}
	all, err := findings(b, o.Categories)
	if err != nil {
		return Report{}, err
	}
	var rep Report
	for _, f := range all {
		if o.Conformance && f.Severity != Error {
			continue
		}
		if len(o.Rules) > 0 && !slices.Contains(o.Rules, f.Rule) {
			continue
		}
		rep.Findings = append(rep.Findings, f)
		if f.Severity == Error {
			rep.Errors++
		} else {
			rep.Warnings++
		}
	}
	switch {
	case rep.Errors > 0:
		rep.Failure = fmt.Sprintf("%d conformance error(s)", rep.Errors)
	case o.Strict && rep.Warnings > 0:
		rep.Failure = fmt.Sprintf("%d warning(s) with --strict", rep.Warnings)
	}
	return rep, nil
}

func findings(b *bundle.Bundle, categories []string) ([]Finding, error) {
	var out []Finding
	add := func(sev, rule string, c *bundle.Concept, msg string) {
		out = append(out, Finding{Severity: sev, Rule: rule, ID: c.ID, Path: c.Path, Message: msg})
	}
	casing := func(id, rel string) {
		if reserved, ok := b.ReservedCollision(rel); ok {
			out = append(out, Finding{Severity: Warning, Rule: ReservedCasing, ID: id, Path: rel,
				Message: fmt.Sprintf("%s collides with the reserved %s on this case-insensitive file system; rename it", path.Base(rel), reserved)})
		}
	}
	for _, p := range b.Problems {
		out = append(out, Finding{Severity: Error, Rule: Unparseable, ID: p.ID, Path: p.Path, Message: p.Err.Error()})
		casing(p.ID, p.Path)
	}
	g := b.Graph()
	titles := map[string][]*bundle.Concept{}
	for _, c := range b.Concepts {
		fm := &c.Note.FM
		live := !fm.Archived
		typ := strings.TrimSpace(fm.Type)
		if typ == "" {
			add(Error, MissingType, c, "frontmatter has no type (OKF requires one)")
		} else if _, ok := b.Vocab.Lookup(typ); !ok {
			add(Warning, UnknownType, c, fmt.Sprintf("type %q is not in this knowledge base's vocabulary (%s)", typ, strings.Join(b.Vocab.Names(), ", ")))
		}
		for _, t := range g.Dangling[c.ID] {
			add(Warning, DanglingLink, c, fmt.Sprintf("links to %s, which does not exist", t))
		}
		casing(c.ID, c.Path)
		if !live {
			continue
		}
		if strings.TrimSpace(fm.Title) == "" {
			add(Warning, MissingTitle, c, "no title")
		}
		if strings.TrimSpace(fm.Description) == "" {
			add(Warning, MissingDescription, c, "no description, so it appears bare in listings")
		}
		if typ != "" {
			if want := b.Vocab.DirFor(typ); want != c.Dir() {
				where := c.Dir() + "/"
				if c.Dir() == "" {
					where = "the bundle root"
				}
				add(Warning, TypeDirectoryMismatch, c, fmt.Sprintf("type %q belongs in %s/, but this note is in %s (fix: rit note move %s %s/%s)",
					typ, want, where, c.ID, want, path.Base(c.ID)))
			}
		}
		key := strings.ToLower(strings.TrimSpace(c.Title()))
		titles[key] = append(titles[key], c)
		if len(categories) > 0 {
			seen := map[string]bool{}
			for _, ob := range note.Observations(c.Note.Body) {
				cat := strings.ToLower(ob.Category)
				if seen[cat] || slices.ContainsFunc(categories, func(a string) bool { return strings.EqualFold(strings.TrimSpace(a), cat) }) {
					continue
				}
				seen[cat] = true
				add(Warning, UnknownCategory, c, fmt.Sprintf("observation category %q is not in the allowed list (%s)", ob.Category, strings.Join(categories, ", ")))
			}
		}
		if staleVerification(fm) {
			add(Warning, StaleVerification, c, "content changed after its last verification (generated.at is later than verified[].at)")
		}
		if !c.IsDaily() && !slices.ContainsFunc(g.Out[c.ID], func(t string) bool { return t != c.ID }) && !hasNonDailyInbound(b, g, c.ID) {
			add(Warning, Orphan, c, "no links to or from other notes")
		}
	}
	for _, cs := range titles {
		if len(cs) < 2 {
			continue
		}
		for _, c := range cs {
			var others []string
			for _, o := range cs {
				if o != c {
					others = append(others, o.ID)
				}
			}
			add(Warning, DuplicateTitle, c, fmt.Sprintf("title %q is also used by %s", c.Title(), strings.Join(others, ", ")))
		}
	}
	stale, err := b.SyncIndexes(true)
	if err != nil {
		return nil, err
	}
	for _, s := range stale {
		p, removed := strings.CutSuffix(s, " (removed)")
		msg := "generated listing is out of date (run: rit index)"
		if removed {
			msg = "generated listing is no longer needed (run: rit index)"
		}
		out = append(out, Finding{Severity: Warning, Rule: StaleIndex, Path: p, Message: msg})
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		if a.Severity != b.Severity {
			return a.Severity == Error
		}
		if a.Rule != b.Rule {
			return a.Rule < b.Rule
		}
		return a.Message < b.Message
	})
	return out, nil
}

func hasNonDailyInbound(b *bundle.Bundle, g *bundle.Graph, id string) bool {
	for _, from := range g.In[id] {
		if c, ok := b.Get(from); ok && !c.IsDaily() && from != id {
			return true
		}
	}
	return false
}

func staleVerification(fm *note.Frontmatter) bool {
	if len(fm.Verified) == 0 || fm.Generated == nil {
		return false
	}
	gen, err := time.Parse(time.RFC3339, strings.TrimSpace(fm.Generated.At))
	if err != nil {
		return false
	}
	var newest time.Time
	found := false
	for _, v := range fm.Verified {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(v.At))
		if err != nil {
			continue
		}
		if !found || t.After(newest) {
			newest, found = t, true
		}
	}
	return found && gen.After(newest)
}
