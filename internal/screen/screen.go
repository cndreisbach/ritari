// Package screen finds content hidden from a human reader of a note, and
// hidden text that reads like an instruction to a model (§8.1).
package screen

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.yaml.in/yaml/v3"

	"github.com/cndreisbach/ritari/internal/note"
)

// Rule names, shared with lint.
const (
	HiddenContent     = "hidden-content"
	HiddenInstruction = "hidden-instruction"
)

// Finding is one screening result.
type Finding struct {
	Rule    string
	Message string
}

// String renders a finding as `<rule>: <message>`, the form write warnings use.
func (f Finding) String() string { return f.Rule + ": " + f.Message }

// span is text a reader of the rendered note does not see.
type span struct {
	kind string // "an HTML comment", "a hidden HTML element", "a link title", ...
	text string
	// flag says whether the span is hidden content on its own; titles,
	// attribute values, and unused reference definitions are only scanned
	// for instructions.
	flag bool
	// strip says Strip removes the span's byte range [start, end), markup
	// included. A span may carry a range and no text (an empty hidden
	// element, a link title found only for removal).
	strip      bool
	start, end int
}

var md = goldmark.New(goldmark.WithExtensions(extension.Footnote))

// Note screens a parsed note: its frontmatter and its body.
func Note(n *note.Note) []Finding {
	return noteFindings(n, hiddenSpans(n.Body))
}

// NoteStripped screens a note and returns its body with what a reader would
// not see removed (see Strip), from one parse of the body.
func NoteStripped(n *note.Note) ([]Finding, string) {
	spans := hiddenSpans(n.Body)
	return noteFindings(n, spans), strip(n.Body, spans)
}

func noteFindings(n *note.Note, spans []span) []Finding {
	var out []Finding
	for _, f := range fields(n) {
		out = append(out, textFindings(f.name, f.value)...)
	}
	out = append(out, textFindings("body", n.Body)...)
	for _, s := range spans {
		if s.flag {
			out = append(out, Finding{HiddenContent, fmt.Sprintf("body has %s, not shown when rendered: %s", s.kind, excerpt(s.text))})
		}
		if cats := Classify(s.text); len(cats) > 0 {
			out = append(out, Finding{HiddenInstruction, fmt.Sprintf("hidden text in %s reads like %s: %s", s.kind, strings.Join(cats, ", "), excerpt(s.text))})
		}
	}
	return out
}

// Worst names the most serious rule among findings: hidden-instruction,
// then hidden-content, or "" when there are none (§8.1).
func Worst(fs []Finding) string {
	worst := ""
	for _, f := range fs {
		if f.Rule == HiddenInstruction {
			return HiddenInstruction
		}
		worst = HiddenContent
	}
	return worst
}

// Strip removes what a reader would not see from a body (§8.1): HTML
// comments, hidden and never-rendered elements (markup included), link and
// image titles, HTML attributes other than href, src, and srcset, unused
// reference definitions, disallowed invisible characters, and soft hyphens.
// Line breaks inside a removed span are kept, so lines never merge.
func Strip(body string) string { return strip(body, hiddenSpans(body)) }

func strip(body string, spans []span) string {
	var ranges [][2]int
	for _, s := range spans {
		if s.strip && s.end > s.start {
			ranges = append(ranges, [2]int{s.start, min(s.end, len(body))})
		}
	}
	sort.Slice(ranges, func(i, j int) bool { return ranges[i][0] < ranges[j][0] })
	var b strings.Builder
	at := 0
	lastSpace := false
	for _, r := range ranges {
		if r[1] <= at {
			continue
		}
		if r[0] > at {
			chunk := body[at:r[0]]
			b.WriteString(chunk)
			lastSpace = strings.HasSuffix(chunk, " ")
		}
		removed := body[max(at, r[0]):r[1]]
		at = r[1]
		if strings.Contains(removed, "\n") {
			// Keep the lines on either side apart, without adding a blank
			// line that would split a paragraph.
			out := b.String()
			if !strings.HasSuffix(out, "\n") && !strings.HasPrefix(body[at:], "\n") && out != "" {
				b.WriteByte('\n')
			}
			lastSpace = false
			continue
		}
		// Removing a span from between two words leaves one space, not two;
		// removing all of a line leaves no blank line in its place.
		out := b.String()
		switch {
		case lastSpace && at < len(body) && body[at] == ' ':
			at++
		case (out == "" || strings.HasSuffix(out, "\n")) && at < len(body) && body[at] == '\n':
			at++
		}
	}
	b.WriteString(body[at:])
	return StripInvisible(b.String())
}

// StripInvisible removes disallowed invisible characters and soft hyphens
// from s, leaving allowed characters (an emoji ZWJ, a Persian ZWNJ) in place.
func StripInvisible(s string) string {
	if !strings.ContainsFunc(s, ignorable) {
		return s
	}
	rs := []rune(s)
	bad := flagged(rs)
	var b strings.Builder
	for i, r := range rs {
		if !bad[i] && r != softHyph {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// New returns the findings in after that are not in before, counting
// repeats: what a write introduced (§8.1).
func New(before, after []Finding) []Finding {
	seen := map[Finding]int{}
	for _, f := range before {
		seen[f]++
	}
	var out []Finding
	for _, f := range after {
		if seen[f] > 0 {
			seen[f]--
			continue
		}
		out = append(out, f)
	}
	return out
}

// Warning joins findings into one write warning, or "".
func Warning(fs []Finding) string {
	parts := make([]string, len(fs))
	for i, f := range fs {
		parts[i] = f.String()
	}
	return strings.Join(parts, "; ")
}

// textFindings reports invisible characters, smuggled tag text, and
// mixed-script words in one field.
func textFindings(field, s string) []Finding {
	var out []Finding
	if rs, total := invisibles(s); total > 0 {
		names := make([]string, len(rs))
		for i, r := range rs {
			names[i] = charName(r)
		}
		out = append(out, Finding{HiddenContent, fmt.Sprintf("%s has %d invisible character(s): %s", field, total, strings.Join(names, ", "))})
	}
	if t := smuggled(s); t != "" {
		if cats := Classify(t); len(cats) > 0 {
			out = append(out, Finding{HiddenInstruction, fmt.Sprintf("hidden text in tag characters in %s reads like %s: %s", field, strings.Join(cats, ", "), excerpt(t))})
		}
	}
	for _, w := range mixedScript(s) {
		out = append(out, Finding{HiddenContent, fmt.Sprintf("%s has a word mixing %s letters: %s", field, scriptsOf(w), quote(w))})
	}
	return out
}

type field struct{ name, value string }

// fields lists every scalar in the note's frontmatter, named by its
// top-level key, and every key, named "frontmatter key". It reads the
// frontmatter as rendered, so recognized and unrecognized fields alike are
// covered.
func fields(n *note.Note) []field {
	fmText, _, ok := note.SplitFrontmatter(note.Render(n))
	if !ok {
		return nil
	}
	var doc yaml.Node
	if yaml.Unmarshal([]byte(fmText), &doc) != nil || len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return nil
	}
	var out []field
	var walk func(name string, n *yaml.Node)
	walk = func(name string, n *yaml.Node) {
		switch n.Kind {
		case yaml.ScalarNode:
			out = append(out, field{name, n.Value})
		case yaml.MappingNode:
			for i := 0; i+1 < len(n.Content); i += 2 {
				out = append(out, field{"frontmatter key", n.Content[i].Value})
				walk(name, n.Content[i+1])
			}
		default:
			for _, c := range n.Content {
				walk(name, c)
			}
		}
	}
	root := doc.Content[0]
	for i := 0; i+1 < len(root.Content); i += 2 {
		k := root.Content[i].Value
		out = append(out, field{"frontmatter key", k})
		walk(k, root.Content[i+1])
	}
	return out
}

// hiddenSpans returns the parts of a body that do not show when it is
// rendered: HTML comments and hidden elements (§8.1), and the text that is
// only scanned for instructions: link and image titles, HTML attribute
// values, and unused reference definitions. Code spans and code blocks are
// never HTML.
func hiddenSpans(body string) []span {
	src := []byte(body)
	ctx := parser.NewContext()
	doc := md.Parser().Parse(text.NewReader(src), parser.WithContext(ctx))
	masked := []byte(body)
	mask := func(seg text.Segment) {
		for i := seg.Start; i < seg.Stop && i < len(masked); i++ {
			if masked[i] != '\n' {
				masked[i] = ' '
			}
		}
	}
	var out []span
	used := map[string]bool{}
	var links, defs []ast.Node
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch v := n.(type) {
		case *ast.LinkReferenceDefinition:
			defs = append(defs, n)
		case *ast.CodeSpan:
			for c := v.FirstChild(); c != nil; c = c.NextSibling() {
				if t, ok := c.(*ast.Text); ok {
					mask(t.Segment)
				}
			}
			return ast.WalkSkipChildren, nil
		case *ast.CodeBlock, *ast.FencedCodeBlock:
			lines := n.Lines()
			for i := 0; i < lines.Len(); i++ {
				mask(lines.At(i))
			}
			if f, ok := v.(*ast.FencedCodeBlock); ok && f.Info != nil {
				mask(f.Info.Segment)
			}
		case *ast.Link:
			links = append(links, n)
			used[linkKey(v.Destination, v.Title)] = true
			if len(v.Title) > 0 {
				out = append(out, span{kind: "a link title", text: string(v.Title)})
			}
		case *ast.Image:
			links = append(links, n)
			used[linkKey(v.Destination, v.Title)] = true
			if len(v.Title) > 0 {
				out = append(out, span{kind: "an image title", text: string(v.Title)})
			}
		}
		return ast.WalkContinue, nil
	})
	out = append(out, htmlSpans(body, string(masked))...)
	for _, r := range ctx.References() {
		if used[linkKey(r.Destination(), r.Title())] {
			continue
		}
		t := strings.TrimSpace(string(r.Label()) + " " + string(r.Title()))
		out = append(out, span{kind: "an unused reference definition", text: t})
	}
	out = append(out, titleRemovals(string(masked), links, src)...)
	return append(out, refDefRemovals(string(masked), defs)...)
}

func linkKey(dest, title []byte) string { return string(dest) + "\x00" + string(title) }

// linkTailRe matches the rest of an inline link or image from the end of
// its text: `](dest "title")`, allowing emphasis or code closers before the
// bracket. Group 1 is the title with its leading space.
var linkTailRe = regexp.MustCompile(`^[*_~` + "`" + `]*\]\(\s*(?:<[^>\n]*>|[^\s()<]*(?:\([^\s()]*\)[^\s()]*)*)(\s+(?:"(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*'|\((?:[^()\\]|\\.)*\)))\s*\)`)

// titleRemovals finds the title of each inline link and image the parser
// produced, right after its text, for Strip to remove. Text that only looks
// like a link is never touched.
func titleRemovals(masked string, links []ast.Node, src []byte) []span {
	var out []span
	for _, l := range links {
		pos := -1
		_ = ast.Walk(l, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
			if t, ok := c.(*ast.Text); ok && entering && t.Segment.Stop > pos {
				pos = t.Segment.Stop
			}
			return ast.WalkContinue, nil
		})
		if pos < 0 || pos > len(masked) {
			continue
		}
		if m := linkTailRe.FindStringSubmatchIndex(masked[pos:]); m != nil && m[2] >= 0 {
			out = append(out, span{strip: true, start: pos + m[2], end: pos + m[3]})
		}
	}
	return out
}

// refDefRemovals returns, for Strip to remove whole, the lines of each
// reference definition the parser found: they render as nothing. Lines the
// parser kept as paragraph text are never touched, whatever they look like;
// definitions in block quotes and lists, and ones spanning lines, are found.
func refDefRemovals(masked string, defs []ast.Node) []span {
	var out []span
	for _, d := range defs {
		lines := d.Lines()
		if lines.Len() == 0 {
			continue
		}
		start := lines.At(0).Start
		start = strings.LastIndexByte(masked[:start], '\n') + 1 // from the line's start, markers included
		end := lines.At(lines.Len() - 1).Stop
		if end < len(masked) {
			if nl := strings.IndexByte(masked[end:], '\n'); nl >= 0 {
				end += nl + 1
			} else {
				end = len(masked)
			}
		}
		out = append(out, span{strip: true, start: start, end: min(end, len(masked))})
	}
	return out
}

const excerptLen = 80

// excerpt quotes up to excerptLen characters of s on one line, with runs of
// whitespace collapsed and invisible characters escaped.
func excerpt(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if rs := []rune(s); len(rs) > excerptLen {
		s = string(rs[:excerptLen]) + "…"
	}
	return quote(s)
}

func quote(s string) string { return `"` + Escape(s) + `"` }
