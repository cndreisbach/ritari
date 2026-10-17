// Package search scores and filters concepts in a bundle (§9.11).
package search

import (
	"fmt"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/cndreisbach/ritari/internal/bundle"
	"github.com/cndreisbach/ritari/internal/note"
)

// Field weights.
const (
	wTitle       = 10
	wDescription = 6
	wTags        = 5
	wObservation = 3
	wBody        = 1
)

// Options are search inputs.
type Options struct {
	Query           string
	Type            string
	Category        string
	Tags            []string
	Relation        string
	Meta            []string // raw key=value specs
	IncludeDaily    bool
	IncludeArchived bool
	Limit           int
}

// Hit is one search result.
type Hit struct {
	Concept *bundle.Concept
	Score   int
	Matches []string // `[category] content` of matched observations
}

// Tokenize lowercases s and splits on every non-ASCII-alphanumeric character.
func Tokenize(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	})
}

// ParseMeta parses --meta specs into key → alternatives.
func ParseMeta(specs []string) (map[string][]string, error) {
	out := map[string][]string{}
	raw := map[string]string{}
	for _, s := range specs {
		k, v, ok := strings.Cut(s, "=")
		k = strings.TrimSpace(k)
		if !ok || k == "" {
			return nil, fmt.Errorf("--meta %q is not key=value", s)
		}
		var alts []string
		for _, a := range strings.Split(v, "|") {
			if a = strings.TrimSpace(a); a != "" {
				alts = append(alts, a)
			}
		}
		if len(alts) == 0 {
			return nil, fmt.Errorf("--meta %q has no value", s)
		}
		if prev, dup := raw[k]; dup {
			return nil, fmt.Errorf("--meta %s given twice (%s and %s); use %s=a|b for alternatives", k, prev, v, k)
		}
		raw[k] = v
		out[k] = alts
	}
	return out, nil
}

// fieldValues returns the values of a frontmatter field for --meta filtering.
func fieldValues(c *bundle.Concept, key string) []string {
	fm := &c.Note.FM
	nonEmpty := func(s string) []string {
		if s == "" {
			return nil
		}
		return []string{s}
	}
	switch strings.ToLower(key) {
	case "type":
		return nonEmpty(fm.Type)
	case "title":
		return nonEmpty(fm.Title)
	case "description":
		return nonEmpty(fm.Description)
	case "resource":
		return nonEmpty(fm.Resource)
	case "tags":
		return fm.Tags
	case "trust":
		return []string{fm.Trust()}
	case "generated.by":
		if fm.Generated != nil {
			return nonEmpty(fm.Generated.By)
		}
		return nil
	case "timestamp":
		return nonEmpty(fm.ContentTime())
	}
	v := fm.ExtraValue(key)
	if v == nil {
		return nil
	}
	switch v.Kind {
	case yaml.ScalarNode:
		return nonEmpty(v.Value)
	case yaml.SequenceNode:
		var out []string
		for _, e := range v.Content {
			if e.Kind == yaml.ScalarNode {
				out = append(out, e.Value)
			}
		}
		return out
	}
	return nil
}

func anyEqualFold(have, want []string) bool {
	for _, h := range have {
		for _, w := range want {
			if strings.EqualFold(strings.TrimSpace(h), w) {
				return true
			}
		}
	}
	return false
}

func passes(c *bundle.Concept, o Options, meta map[string][]string) bool {
	fm := &c.Note.FM
	if c.IsDaily() && !o.IncludeDaily {
		return false
	}
	if fm.Archived && !o.IncludeArchived {
		return false
	}
	for k, alts := range meta {
		if !anyEqualFold(fieldValues(c, k), alts) {
			return false
		}
	}
	if o.Type != "" && !strings.EqualFold(strings.TrimSpace(fm.Type), strings.TrimSpace(o.Type)) {
		return false
	}
	if o.Category != "" {
		found := false
		for _, ob := range note.Observations(c.Note.Body) {
			if strings.EqualFold(ob.Category, o.Category) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if o.Relation != "" {
		found := false
		for _, r := range note.Relations(c.Note.Body, c.ID) {
			if strings.EqualFold(r.Type, o.Relation) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	for _, t := range o.Tags {
		if !anyEqualFold(fm.Tags, []string{strings.TrimSpace(t)}) {
			return false
		}
	}
	return true
}

// count is the number of tokens a term matches: exactly, or by prefix when the
// term has at least 3 characters.
func count(term string, tokens []string) int {
	n := 0
	for _, t := range tokens {
		if t == term || len(term) >= 3 && strings.HasPrefix(t, term) {
			n++
		}
	}
	return n
}

// Run executes a search over b.
func Run(b *bundle.Bundle, o Options) ([]Hit, error) {
	meta, err := ParseMeta(o.Meta)
	if err != nil {
		return nil, err
	}
	terms := Tokenize(o.Query)
	var hits []Hit
	for _, c := range b.Concepts {
		if !passes(c, o, meta) {
			continue
		}
		if len(terms) == 0 {
			hits = append(hits, Hit{Concept: c})
			continue
		}
		if h, ok := score(c, terms); ok {
			hits = append(hits, h)
		}
	}
	if len(terms) == 0 {
		sort.SliceStable(hits, func(i, j int) bool { return hits[i].Concept.ModTime.After(hits[j].Concept.ModTime) })
	} else {
		sort.SliceStable(hits, func(i, j int) bool {
			if hits[i].Score != hits[j].Score {
				return hits[i].Score > hits[j].Score
			}
			return hits[i].Concept.ID < hits[j].Concept.ID
		})
	}
	if o.Limit > 0 && len(hits) > o.Limit {
		hits = hits[:o.Limit]
	}
	return hits, nil
}

func score(c *bundle.Concept, terms []string) (Hit, bool) {
	fm := &c.Note.FM
	title := Tokenize(c.Title())
	desc := Tokenize(fm.Description)
	tags := Tokenize(strings.Join(fm.Tags, " "))
	body := Tokenize(c.Note.Body)
	obs := note.Observations(c.Note.Body)
	obsTokens := make([][]string, len(obs))
	for i, ob := range obs {
		obsTokens[i] = Tokenize(ob.Category + " " + ob.Content + " " + strings.Join(ob.Tags, " ") + " " + ob.Context)
	}
	h := Hit{Concept: c}
	matched := make([]bool, len(obs))
	for _, term := range terms {
		s := count(term, title)*wTitle + count(term, desc)*wDescription + count(term, tags)*wTags + count(term, body)*wBody
		for i, ot := range obsTokens {
			if n := count(term, ot); n > 0 {
				s += n * wObservation
				matched[i] = true
			}
		}
		if s == 0 {
			return Hit{}, false
		}
		h.Score += s
	}
	for i, m := range matched {
		if m {
			h.Matches = append(h.Matches, obs[i].Summary())
		}
	}
	return h, true
}
