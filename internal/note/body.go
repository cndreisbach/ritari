package note

import (
	"regexp"
	"strings"
	"unicode"
)

var headingRe = regexp.MustCompile(`^(#+)[ \t]+(.*)$`)

// The tool's conventional section names (§3.3).
const (
	ObservationsSection = "Observations"
	RelationsSection    = "Relations"
)

// Heading is an ATX heading line in a body.
type Heading struct {
	Line  int // index into the body's lines
	Level int
	Text  string
}

// lines splits a body into lines without their trailing newlines. An empty
// body has no lines.
func lines(body string) []string {
	if body == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(body, "\n"), "\n")
}

func join(ls []string) string {
	if len(ls) == 0 {
		return ""
	}
	return strings.Join(ls, "\n") + "\n"
}

func parseHeading(line string) (Heading, bool) {
	m := headingRe.FindStringSubmatch(line)
	if m == nil {
		return Heading{}, false
	}
	return Heading{Level: len(m[1]), Text: headingText(m[2])}, true
}

// headingText strips an optional closing sequence of #s and surrounding space.
func headingText(s string) string {
	s = strings.TrimSpace(s)
	trimmed := strings.TrimRight(s, "#")
	if trimmed != s && (trimmed == "" || strings.HasSuffix(trimmed, " ") || strings.HasSuffix(trimmed, "\t")) {
		s = strings.TrimSpace(trimmed)
	}
	return s
}

// Headings lists the ATX headings in a body.
func Headings(body string) []Heading {
	var hs []Heading
	for i, l := range lines(body) {
		if h, ok := parseHeading(l); ok {
			h.Line = i
			hs = append(hs, h)
		}
	}
	return hs
}

// sectionNameKey normalizes a section name for matching: leading #s and spaces
// are ignored, case-insensitively.
func sectionNameKey(name string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(name), "#")))
}

// findSection locates the first section whose heading matches name. It returns
// the heading's line index and the index one past the section's last line.
func findSection(ls []string, name string) (start, end int, ok bool) {
	key := sectionNameKey(name)
	for i, l := range ls {
		h, isH := parseHeading(l)
		if !isH || strings.ToLower(h.Text) != key {
			continue
		}
		end = len(ls)
		for j := i + 1; j < len(ls); j++ {
			if h2, isH2 := parseHeading(ls[j]); isH2 && h2.Level <= h.Level {
				end = j
				break
			}
		}
		return i, end, true
	}
	return 0, 0, false
}

// Narrative is the prose before the first heading.
func Narrative(body string) string {
	ls := lines(body)
	for i, l := range ls {
		if _, ok := parseHeading(l); ok {
			return join(ls[:i])
		}
	}
	return body
}

// Observation is one `- [category] content` fact.
type Observation struct {
	Category string
	Content  string
	Tags     []string
	Context  string
	Refs     []string
	// Raw is the text after `[category]`, trimmed, exactly as written.
	Raw string
	// Index is the observation's line index in the body.
	Index int
}

// Line renders the observation as written, minus its bullet:
// `[category] content #tag (context) [^1]`.
func (o Observation) Line() string {
	return "[" + o.Category + "] " + o.Raw
}

// Summary renders `[category] content`.
func (o Observation) Summary() string {
	return "[" + o.Category + "] " + o.Content
}

var (
	obsRe      = regexp.MustCompile(`^\s*[-*]\s+\[([^\]]*)\]\s+(.*)$`)
	trailRefRe = regexp.MustCompile(`\s*\[\^([A-Za-z0-9_-]+)\]\s*$`)
	trailCtxRe = regexp.MustCompile(`\s*\(([^()]*)\)\s*$`)
	tagRe      = regexp.MustCompile(`#([\p{L}\p{N}_/-]+)`)
	wsRe       = regexp.MustCompile(`\s+`)
)

// ParseObservation parses a single line, reporting whether it is an observation.
func ParseObservation(line string) (Observation, bool) {
	m := obsRe.FindStringSubmatch(line)
	if m == nil {
		return Observation{}, false
	}
	cat := strings.TrimSpace(m[1])
	if cat == "" || cat == "x" || cat == "X" || strings.HasPrefix(cat, "^") {
		return Observation{}, false
	}
	rest, refs := stripRefs(m[2])
	o := Observation{Category: cat, Refs: refs, Raw: strings.TrimSpace(m[2])}
	if cm := trailCtxRe.FindStringSubmatchIndex(rest); cm != nil {
		o.Context = strings.TrimSpace(rest[cm[2]:cm[3]])
		rest = rest[:cm[0]]
		if len(refs) == 0 {
			rest, o.Refs = stripRefs(rest)
		}
	}
	rest = tagRe.ReplaceAllStringFunc(rest, func(tok string) string {
		word := tok[1:]
		if !isTag(word) {
			return tok
		}
		o.Tags = append(o.Tags, word)
		return " "
	})
	o.Content = strings.TrimSpace(wsRe.ReplaceAllString(rest, " "))
	return o, true
}

// stripRefs removes trailing `[^label]` references, returning them in written order.
func stripRefs(s string) (string, []string) {
	var refs []string
	for {
		m := trailRefRe.FindStringSubmatchIndex(s)
		if m == nil {
			break
		}
		refs = append([]string{s[m[2]:m[3]]}, refs...)
		s = s[:m[0]]
	}
	return s, refs
}

// isTag reports whether a `#word` is a tag: words of only decimal digits, or
// only hex digits including at least one decimal digit, are not.
func isTag(word string) bool {
	hasDigit, allHex := false, true
	for _, r := range word {
		switch {
		case r >= '0' && r <= '9':
			hasDigit = true
		case r >= 'a' && r <= 'f', r >= 'A' && r <= 'F':
		default:
			allHex = false
		}
	}
	return !allHex || !hasDigit
}

// Observations lists the observation lines in a body's Observations sections
// (§3.4). Observation-shaped lines anywhere else are ordinary prose.
func Observations(body string) []Observation {
	var out []Observation
	ls := lines(body)
	in := inObservations(ls)
	for i, l := range ls {
		if !in[i] {
			continue
		}
		if o, ok := ParseObservation(l); ok {
			o.Index = i
			out = append(out, o)
		}
	}
	return out
}

// inObservations reports, for each line, whether it lies inside a section
// headed `Observations` (any level, case-insensitively). Heading lines are
// never inside.
func inObservations(ls []string) []bool {
	in := make([]bool, len(ls))
	level := 0 // the open Observations section's heading level, or 0
	for i, l := range ls {
		if h, ok := parseHeading(l); ok {
			if level > 0 && h.Level <= level {
				level = 0
			}
			if level == 0 && strings.EqualFold(h.Text, ObservationsSection) {
				level = h.Level
			}
			continue
		}
		in[i] = level > 0
	}
	return in
}

// lineKey is a line's text ignoring indentation and trailing source refs.
func lineKey(line string) string {
	s, _ := stripRefs(line)
	return strings.TrimSpace(s)
}

func indentOf(line string) string {
	return line[:len(line)-len(strings.TrimLeftFunc(line, unicode.IsSpace))]
}
