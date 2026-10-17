package note

import (
	"errors"
	"slices"
	"strings"
)

// Render serializes a note: frontmatter, a blank line, the body with its
// derived footnote definitions re-appended at the end.
func Render(n *Note) []byte {
	body := strings.TrimRight(StripDerived(n.Body, n.FM.Sources), "\n")
	if defs := derivedDefs(n.FM.Sources); len(defs) > 0 {
		if body != "" {
			body += "\n\n"
		}
		body += strings.Join(defs, "\n")
	}
	out := renderFrontmatter(&n.FM)
	if body != "" {
		out += "\n" + body + "\n"
	}
	return []byte(out)
}

// EditableBody is the body without its derived footnote definitions or
// trailing blank lines; body operations work on this.
func (n *Note) EditableBody() string {
	b := strings.TrimRight(StripDerived(n.Body, n.FM.Sources), "\n")
	if b == "" {
		return ""
	}
	return b + "\n"
}

// FillGenerated completes a blank `generated.by`/`generated.at`.
func (n *Note) FillGenerated(actor, now string) {
	if n.FM.Generated == nil {
		n.FM.Generated = &Stamp{}
	}
	if n.FM.Generated.By == "" {
		n.FM.Generated.By = actor
	}
	if n.FM.Generated.At == "" {
		n.FM.Generated.At = now
	}
}

// Restamp records actor and now as the producer of the current content.
func (n *Note) Restamp(actor, now string) {
	n.FM.Generated = &Stamp{By: actor, At: now}
}

// ErrEmptyObservation is returned for an --obs value with no text.
var ErrEmptyObservation = errors.New("observation has no text")

// ComposeObservation turns an --obs value into a canonical observation line,
// appending refs (unioned with any refs already written at its end).
func ComposeObservation(input string, refs []string) (string, error) {
	s := strings.TrimSpace(input)
	if r, ok := strings.CutPrefix(s, "- "); ok {
		s = r
	} else if r, ok := strings.CutPrefix(s, "* "); ok {
		s = r
	}
	s = strings.TrimSpace(s)
	cat, text := "note", s
	if strings.HasPrefix(s, "[") {
		if i := strings.Index(s, "]"); i > 1 {
			cat, text = strings.TrimSpace(s[1:i]), strings.TrimSpace(s[i+1:])
		}
	} else if before, after, ok := strings.Cut(s, ":"); ok && before != "" && !strings.ContainsAny(before, " \t") {
		cat, text = before, strings.TrimSpace(after)
	}
	text, existing := stripRefs(text)
	text = strings.TrimSpace(text)
	if text == "" || cat == "" {
		return "", ErrEmptyObservation
	}
	line := "- [" + cat + "] " + text
	for _, r := range unionRefs(existing, refs) {
		line += " [^" + r + "]"
	}
	return line, nil
}

func unionRefs(a, b []string) []string {
	var out []string
	for _, r := range append(append([]string(nil), a...), b...) {
		if !slices.Contains(out, r) {
			out = append(out, r)
		}
	}
	return out
}

// ComposeBody builds a fresh body from prose, observation lines, and relation lines.
func ComposeBody(prose string, obsLines, relLines []string) string {
	var parts []string
	if p := strings.TrimRight(prose, "\n"); strings.TrimSpace(p) != "" {
		parts = append(parts, p)
	}
	if len(obsLines) > 0 {
		parts = append(parts, "# Observations\n"+strings.Join(dedupe(obsLines), "\n"))
	}
	if len(relLines) > 0 {
		parts = append(parts, "# Relations\n"+strings.Join(dedupe(relLines), "\n"))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "\n\n") + "\n"
}

func dedupe(ls []string) []string {
	var out []string
	for _, l := range ls {
		if !slices.Contains(out, l) {
			out = append(out, l)
		}
	}
	return out
}

// MergeBody adds observation and relation lines to a body without duplicating
// what is already there (§5.3). It returns the new body and the counts added.
func MergeBody(body string, obsLines, relLines []string) (string, int, int) {
	ls := lines(body)
	var newObs []string
	for _, l := range dedupe(obsLines) {
		key := lineKey(l)
		_, refs := stripRefs(l)
		found := false
		for i, existing := range ls {
			if lineKey(existing) != key {
				continue
			}
			found = true
			_, have := stripRefs(existing)
			union := unionRefs(have, refs)
			if len(union) > len(have) {
				line := indentOf(existing) + key
				for _, r := range union {
					line += " [^" + r + "]"
				}
				ls[i] = line
			}
			break
		}
		if !found && !containsTrimmed(ls, l) && !slices.Contains(newObs, l) {
			newObs = append(newObs, l)
		}
	}
	ls = appendToSection(ls, "Observations", newObs)

	var newRels []string
	for _, l := range dedupe(relLines) {
		if !containsTrimmed(ls, l) {
			newRels = append(newRels, l)
		}
	}
	ls = appendToSection(ls, "Relations", newRels)
	return join(ls), len(newObs), len(newRels)
}

func containsTrimmed(ls []string, l string) bool {
	t := strings.TrimSpace(l)
	for _, x := range ls {
		if strings.TrimSpace(x) == t {
			return true
		}
	}
	return false
}

// appendToSection appends lines at the end of the named section (before its
// trailing blank lines), creating `# name` at the end of the body if missing.
func appendToSection(ls []string, name string, add []string) []string {
	if len(add) == 0 {
		return ls
	}
	if _, end, ok := findSection(ls, name); ok {
		at := end
		for at > 0 && strings.TrimSpace(ls[at-1]) == "" {
			at--
		}
		return slices.Concat(ls[:at], add, ls[at:])
	}
	for len(ls) > 0 && strings.TrimSpace(ls[len(ls)-1]) == "" {
		ls = ls[:len(ls)-1]
	}
	if len(ls) > 0 {
		ls = append(ls, "")
	}
	ls = append(ls, "# "+name)
	return append(ls, add...)
}

// Additions are the --obs, --rel, and --source values of one write.
type Additions struct {
	Obs     []string
	Rels    []string // canonical relation lines
	Sources []Source
}

// ComposeResult reports what a composition or merge did.
type ComposeResult struct {
	AddedObs     int
	AddedRels    int
	AddedSources int
	BodyChanged  bool
}

// Compose sets a fresh body on n from prose and additions.
func (n *Note) Compose(prose string, add Additions) (ComposeResult, error) {
	srcs, ids, addedSrc := AddSources(nil, prose, add.Sources)
	obs, err := composeAll(add.Obs, ids)
	if err != nil {
		return ComposeResult{}, err
	}
	n.FM.Sources = srcs
	n.Body = ComposeBody(prose, obs, add.Rels)
	return ComposeResult{AddedObs: len(add.Obs), AddedRels: len(add.Rels), AddedSources: addedSrc, BodyChanged: true}, nil
}

// Merge adds additions to n's existing body (§5.3).
func (n *Note) Merge(add Additions) (ComposeResult, error) {
	before := n.EditableBody()
	srcs, ids, addedSrc := AddSources(n.FM.Sources, n.Body, add.Sources)
	obs, err := composeAll(add.Obs, ids)
	if err != nil {
		return ComposeResult{}, err
	}
	body, ao, ar := MergeBody(before, obs, add.Rels)
	n.FM.Sources = srcs
	n.Body = body
	return ComposeResult{AddedObs: ao, AddedRels: ar, AddedSources: addedSrc, BodyChanged: body != before}, nil
}

func composeAll(inputs, refs []string) ([]string, error) {
	var out []string
	for _, in := range inputs {
		l, err := ComposeObservation(in, refs)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, nil
}
