// Package screen finds content hidden from a human reader of a note, and
// hidden text that reads like an instruction to a model (§8.1).
package screen

import (
	"fmt"
	"regexp"
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
	kind string // "HTML comment", "hidden HTML element", "link title", ...
	text string
	// flag says whether the span is hidden content on its own; link titles
	// and unused reference definitions are only scanned for instructions.
	flag bool
}

var md = goldmark.New(goldmark.WithExtensions(extension.Footnote))

// Note screens a parsed note: its frontmatter text fields and its body.
func Note(n *note.Note) []Finding {
	var out []Finding
	for _, f := range fields(&n.FM) {
		out = append(out, textFindings(f.name, f.value)...)
	}
	out = append(out, textFindings("body", n.Body)...)
	for _, s := range hiddenSpans(n.Body) {
		if s.flag {
			out = append(out, Finding{HiddenContent, fmt.Sprintf("body has %s, not shown when rendered: %s", s.kind, excerpt(s.text))})
		}
		if cats := Classify(s.text); len(cats) > 0 {
			out = append(out, Finding{HiddenInstruction, fmt.Sprintf("hidden text in %s reads like %s: %s", s.kind, strings.Join(cats, ", "), excerpt(s.text))})
		}
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
	for _, t := range smuggled(s) {
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

// fields lists the frontmatter's free-text values by name.
func fields(fm *note.Frontmatter) []field {
	out := []field{{"title", fm.Title}, {"description", fm.Description}}
	for _, t := range fm.Tags {
		out = append(out, field{"tags", t})
	}
	for _, s := range fm.Sources {
		out = append(out, field{"sources", s.Title}, field{"sources", s.Resource})
	}
	for _, e := range fm.Extra {
		var vals []string
		scalars(e.Value, &vals)
		for _, v := range vals {
			out = append(out, field{e.Key, v})
		}
	}
	return out
}

func scalars(n *yaml.Node, out *[]string) {
	if n == nil {
		return
	}
	if n.Kind == yaml.ScalarNode {
		*out = append(*out, n.Value)
		return
	}
	for _, c := range n.Content {
		scalars(c, out)
	}
}

var (
	commentRe    = regexp.MustCompile(`(?s)<!--(.*?)(?:-->|$)`)
	openTagRe    = regexp.MustCompile(`(?is)<([a-z][a-z0-9-]*)\b([^>]*)>`)
	hiddenAttrRe = regexp.MustCompile(`(?is)(?:^|\s)hidden(?:\s|=|/|$)|style\s*=\s*["'][^"']*(?:display\s*:\s*none|visibility\s*:\s*hidden|font-size\s*:\s*0(?:[a-z%]*)?\s*(?:;|["']|!)|opacity\s*:\s*0(?:\.0+)?\s*(?:;|["']|!))`)
	tagRe        = regexp.MustCompile(`(?s)<[^>]*>`)
)

// hiddenSpans returns the parts of a body that do not show when it is
// rendered: HTML comments, hidden HTML elements, link titles, and unused
// reference definitions. Code spans and code blocks are never HTML.
func hiddenSpans(body string) []span {
	src := []byte(body)
	ctx := parser.NewContext()
	doc := md.Parser().Parse(text.NewReader(src), parser.WithContext(ctx))
	var out []span
	used := map[string]bool{}
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch v := n.(type) {
		case *ast.HTMLBlock:
			var b strings.Builder
			for i := 0; i < v.Lines().Len(); i++ {
				seg := v.Lines().At(i)
				b.Write(seg.Value(src))
			}
			if v.HasClosure() {
				b.Write(v.ClosureLine.Value(src))
			}
			out = append(out, htmlSpans(b.String())...)
		case *ast.RawHTML:
			raw := segmentsText(v.Segments, src)
			if strings.HasPrefix(raw, "<!--") {
				out = append(out, htmlSpans(raw)...)
				break
			}
			if m := openTagRe.FindStringSubmatch(raw); m != nil && hiddenAttrRe.MatchString(m[2]) {
				out = append(out, span{kind: "a hidden HTML element", text: inlineUntilClose(v, strings.ToLower(m[1]), src), flag: true})
			}
		case *ast.Link:
			used[linkKey(v.Destination, v.Title)] = true
			if len(v.Title) > 0 {
				out = append(out, span{kind: "a link title", text: string(v.Title)})
			}
		case *ast.Image:
			used[linkKey(v.Destination, v.Title)] = true
			if len(v.Title) > 0 {
				out = append(out, span{kind: "an image title", text: string(v.Title)})
			}
		}
		return ast.WalkContinue, nil
	})
	for _, r := range ctx.References() {
		if used[linkKey(r.Destination(), r.Title())] {
			continue
		}
		t := strings.TrimSpace(string(r.Label()) + " " + string(r.Title()))
		out = append(out, span{kind: "an unused reference definition", text: t})
	}
	return out
}

func linkKey(dest, title []byte) string { return string(dest) + "\x00" + string(title) }

func segmentsText(segs *text.Segments, src []byte) string {
	var b strings.Builder
	for i := 0; i < segs.Len(); i++ {
		seg := segs.At(i)
		b.Write(seg.Value(src))
	}
	return b.String()
}

// htmlSpans finds comments and hidden elements in a run of raw HTML.
func htmlSpans(h string) []span {
	var out []span
	for _, m := range commentRe.FindAllStringSubmatch(h, -1) {
		out = append(out, span{kind: "an HTML comment", text: m[1], flag: true})
	}
	rest := commentRe.ReplaceAllString(h, "")
	for _, loc := range openTagRe.FindAllStringSubmatchIndex(rest, -1) {
		attrs := rest[loc[4]:loc[5]]
		if !hiddenAttrRe.MatchString(attrs) {
			continue
		}
		name := strings.ToLower(rest[loc[2]:loc[3]])
		inner := rest[loc[1]:]
		if i := strings.Index(strings.ToLower(inner), "</"+name); i >= 0 {
			inner = inner[:i]
		}
		out = append(out, span{kind: "a hidden HTML element", text: tagRe.ReplaceAllString(inner, " "), flag: true})
	}
	return out
}

// inlineUntilClose collects the text after an inline opening tag up to its
// closing tag (or the end of the paragraph).
func inlineUntilClose(open ast.Node, name string, src []byte) string {
	var b strings.Builder
	for n := open.NextSibling(); n != nil; n = n.NextSibling() {
		if r, ok := n.(*ast.RawHTML); ok {
			raw := strings.ToLower(segmentsText(r.Segments, src))
			if strings.HasPrefix(raw, "</"+name) {
				break
			}
			continue
		}
		_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
			if t, ok := c.(*ast.Text); ok && entering {
				b.Write(t.Segment.Value(src))
				if t.SoftLineBreak() || t.HardLineBreak() {
					b.WriteByte(' ')
				}
			}
			return ast.WalkContinue, nil
		})
	}
	return b.String()
}

const excerptLen = 80

// excerpt quotes up to excerptLen characters of s on one line, with
// invisible characters escaped.
func excerpt(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if rs := []rune(s); len(rs) > excerptLen {
		s = string(rs[:excerptLen]) + "…"
	}
	return quote(s)
}

func quote(s string) string { return `"` + Escape(s) + `"` }
