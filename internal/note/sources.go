package note

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var (
	footnoteDefRe = regexp.MustCompile(`^\[\^([^\]\s]+)\]:`)
	footnoteRefRe = regexp.MustCompile(`\[\^([^\]\s]+)\]`)
	mdLinkRe      = regexp.MustCompile(`^\[(.*)\]\((\S*)\)$`)
	schemeURIRe   = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.-]*://\S+$`)
)

// numericID reports the positive integer value of a source id, or 0.
func numericID(id string) int {
	if id == "" || strings.TrimLeft(id, "0123456789") != "" {
		return 0
	}
	n, err := strconv.Atoi(id)
	if err != nil || n <= 0 {
		return 0
	}
	return n
}

// Followable reports whether a source resource is a URL or path (no
// whitespace) rather than a descriptor.
func Followable(resource string) bool {
	return resource != "" && !strings.ContainsAny(resource, " \t\n")
}

func uriLooking(tok string) bool {
	return schemeURIRe.MatchString(tok) || strings.HasPrefix(tok, "mailto:") || strings.HasPrefix(tok, "/")
}

// ParseSource interprets a --source value. ok is false for a blank value.
func ParseSource(value string) (Source, bool) {
	v := strings.TrimSpace(value)
	if v == "" {
		return Source{}, false
	}
	if m := mdLinkRe.FindStringSubmatch(v); m != nil && m[2] != "" {
		return Source{Resource: m[2], Title: strings.TrimSpace(unescapeBrackets(m[1]))}, true
	}
	if strings.HasPrefix(v, "<") && strings.HasSuffix(v, ">") && !strings.ContainsAny(v, " \t") {
		return Source{Resource: v[1 : len(v)-1]}, true
	}
	fields := strings.Fields(v)
	last := fields[len(fields)-1]
	if uriLooking(last) {
		if len(fields) == 1 {
			return Source{Resource: last}, true
		}
		title := strings.TrimSpace(strings.TrimSuffix(v, last))
		return Source{Resource: last, Title: title}, true
	}
	return Source{Resource: v}, true
}

// SameSource reports whether two sources name the same material.
func SameSource(a, b Source) bool {
	if Followable(a.Resource) || Followable(b.Resource) {
		return a.Resource == b.Resource
	}
	return strings.EqualFold(strings.TrimSpace(a.Resource), strings.TrimSpace(b.Resource))
}

// FootnoteDef renders a source's footnote definition line, `[^id]: …`.
func FootnoteDef(s Source) string {
	return "[^" + s.ID + "]: " + renderSourceText(s)
}

func renderSourceText(s Source) string {
	switch {
	case Followable(s.Resource) && s.Title != "":
		return "[" + escapeBrackets(s.Title) + "](" + s.Resource + ")"
	case Followable(s.Resource):
		return "[" + escapeBrackets(s.Resource) + "](" + s.Resource + ")"
	case s.Title != "" && s.Resource != "":
		return s.Title + " (" + s.Resource + ")"
	case s.Resource != "":
		return s.Resource
	}
	return s.Title
}

func escapeBrackets(s string) string {
	return strings.NewReplacer(`[`, `\[`, `]`, `\]`).Replace(s)
}

func unescapeBrackets(s string) string {
	return strings.NewReplacer(`\[`, `[`, `\]`, `]`).Replace(s)
}

// managedIDs are the ids of numbered sources, whose footnote definitions the
// tool derives.
func managedIDs(sources []Source) map[string]bool {
	ids := map[string]bool{}
	for _, s := range sources {
		if numericID(s.ID) > 0 {
			ids[s.ID] = true
		}
	}
	return ids
}

// StripDerived removes the footnote definitions derived from sources.
func StripDerived(body string, sources []Source) string {
	ids := managedIDs(sources)
	if len(ids) == 0 {
		return body
	}
	var out []string
	for _, l := range lines(body) {
		if m := footnoteDefRe.FindStringSubmatch(l); m != nil && ids[m[1]] {
			continue
		}
		out = append(out, l)
	}
	return join(out)
}

// derivedDefs renders the footnote definitions for numbered sources, in id order.
func derivedDefs(sources []Source) []string {
	var nums []Source
	for _, s := range sources {
		if numericID(s.ID) > 0 {
			nums = append(nums, s)
		}
	}
	slices.SortStableFunc(nums, func(a, b Source) int { return numericID(a.ID) - numericID(b.ID) })
	var out []string
	seen := map[string]bool{}
	for _, s := range nums {
		if seen[s.ID] {
			continue
		}
		seen[s.ID] = true
		out = append(out, FootnoteDef(s))
	}
	return out
}

// maxFootnote is the largest numeric footnote label used anywhere in body.
func maxFootnote(body string) int {
	max := 0
	for _, m := range footnoteRefRe.FindAllStringSubmatch(body, -1) {
		if n := numericID(m[1]); n > max {
			max = n
		}
	}
	return max
}

// AddSources merges new sources into existing ones, allocating ids per §3.7.
// It returns the updated list, the id assigned to each input (in order), and
// the number of entries added.
func AddSources(existing []Source, body string, inputs []Source) ([]Source, []string, int) {
	out := append([]Source(nil), existing...)
	next := maxFootnote(body)
	for _, s := range existing {
		if n := numericID(s.ID); n > next {
			next = n
		}
	}
	var ids []string
	added := 0
	for _, in := range inputs {
		found := ""
		for _, s := range out {
			if SameSource(s, in) {
				found = s.ID
				break
			}
		}
		if found == "" {
			next++
			in.ID = strconv.Itoa(next)
			out = append(out, in)
			found = in.ID
			added++
		}
		if found != "" && !slices.Contains(ids, found) {
			ids = append(ids, found)
		}
	}
	return out, ids, added
}
