package note

import (
	"path"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

var md = goldmark.New(goldmark.WithExtensions(extension.Footnote))

// Link is a markdown link found in a body.
type Link struct {
	Text string
	Href string
}

// Relation is a typed edge: a list item holding an optional type token and one link.
type Relation struct {
	Type string
	Text string
	Href string
	// Target is the resolved concept ID for in-bundle links, else empty.
	Target string
}

type parsed struct {
	src []byte
	doc ast.Node
}

func parseMarkdown(body string) parsed {
	src := []byte(body)
	return parsed{src: src, doc: md.Parser().Parse(text.NewReader(src))}
}

// Links returns the distinct link destinations in a body, in order. Images and
// autolinks are ignored; links in code are not links.
func Links(body string) []Link {
	p := parseMarkdown(body)
	seen := map[string]bool{}
	var out []Link
	_ = ast.Walk(p.doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		l, ok := n.(*ast.Link)
		if !ok {
			return ast.WalkContinue, nil
		}
		href := string(l.Destination)
		if href == "" || seen[href] {
			return ast.WalkContinue, nil
		}
		seen[href] = true
		out = append(out, Link{Text: nodeText(l, p.src), Href: href})
		return ast.WalkContinue, nil
	})
	return out
}

func nodeText(n ast.Node, src []byte) string {
	var b strings.Builder
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if t, ok := c.(*ast.Text); ok {
				b.Write(t.Segment.Value(src))
				if t.SoftLineBreak() {
					b.WriteByte(' ')
				}
			}
		}
		return ast.WalkContinue, nil
	})
	return b.String()
}

var relLineRe = regexp.MustCompile(`^\s*[-*]\s+(?:(\S+)\s+)?\[(.*)\]\((.*)\)\s*$`)

// Relations returns the relation lines in a body. fromID is the concept ID of
// the note, used to resolve relative destinations.
func Relations(body, fromID string) []Relation {
	p := parseMarkdown(body)
	var out []Relation
	_ = ast.Walk(p.doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		li, ok := n.(*ast.ListItem)
		if !ok {
			return ast.WalkContinue, nil
		}
		if r, ok := relationOf(li, p.src, fromID); ok {
			out = append(out, r)
		}
		return ast.WalkContinue, nil
	})
	return out
}

func relationOf(li *ast.ListItem, src []byte, fromID string) (Relation, bool) {
	block := li.FirstChild()
	if block == nil || block.NextSibling() != nil || block.Lines().Len() != 1 {
		return Relation{}, false
	}
	if _, ok := block.(*ast.Paragraph); !ok {
		if _, ok := block.(*ast.TextBlock); !ok {
			return Relation{}, false
		}
	}
	var link *ast.Link
	var before strings.Builder
	for c := block.FirstChild(); c != nil; c = c.NextSibling() {
		switch v := c.(type) {
		case *ast.Text:
			if link != nil {
				if strings.TrimSpace(string(v.Segment.Value(src))) != "" {
					return Relation{}, false
				}
				continue
			}
			before.Write(v.Segment.Value(src))
		case *ast.Link:
			if link != nil {
				return Relation{}, false
			}
			link = v
		default:
			return Relation{}, false
		}
	}
	if link == nil {
		return Relation{}, false
	}
	typ := strings.TrimSpace(before.String())
	if strings.ContainsAny(typ, " \t") {
		return Relation{}, false
	}
	if typ == "" {
		typ = "links_to"
	}
	seg := block.Lines().At(0)
	lineStart := seg.Start
	for lineStart > 0 && src[lineStart-1] != '\n' {
		lineStart--
	}
	m := relLineRe.FindStringSubmatch(string(src[lineStart:seg.Stop]))
	txt := nodeText(link, src)
	if m != nil {
		txt = m[2]
	}
	href := string(link.Destination)
	target, _ := ResolveLink(href, fromID)
	return Relation{Type: typ, Text: txt, Href: href, Target: target}, true
}

// LinkKind classifies a link destination.
type LinkKind int

const (
	External LinkKind = iota
	Fragment
	InBundle
)

var schemeRe = regexp.MustCompile(`^[^/?#]*:`)

// Classify reports whether a destination is external, fragment-only, or in-bundle.
func Classify(href string) LinkKind {
	switch {
	case strings.HasPrefix(href, "//"):
		return External
	case strings.HasPrefix(href, "#"):
		return Fragment
	case schemeRe.MatchString(href):
		return External
	}
	return InBundle
}

// BundlePath resolves an in-bundle destination to a normalized,
// bundle-relative file path, as written (extension included). ok is false for
// external and fragment links, empty destinations, and links escaping the
// bundle.
func BundlePath(href, fromID string) (string, bool) {
	if Classify(href) != InBundle {
		return "", false
	}
	if i := strings.IndexByte(href, '#'); i >= 0 {
		href = href[:i]
	}
	if href == "" || escapes(href, fromID) {
		return "", false
	}
	p := href
	if !strings.HasPrefix(href, "/") {
		p = path.Join(path.Dir(fromID), href)
	}
	p = strings.TrimPrefix(path.Clean("/"+p), "/")
	if p == "" || p == "." {
		return "", false
	}
	return p, true
}

// IsReservedName reports whether a file name is one of OKF's reserved files
// (index.md, log.md), which are never concepts. Names match exactly (§3.1).
func IsReservedName(name string) bool {
	return name == "index.md" || name == "log.md"
}

// FileLink reports whether a destination is a file link (§3.6): one that names
// a file rather than a concept, because it lies under sources/, names a
// reserved file, or has an extension other than .md. It returns the file's
// bundle-relative path.
func FileLink(href, fromID string) (string, bool) {
	p, ok := BundlePath(href, fromID)
	if !ok || !isFilePath(p) {
		return "", false
	}
	return p, true
}

func isFilePath(p string) bool {
	if p == "sources" || strings.HasPrefix(p, "sources/") {
		return true
	}
	base := path.Base(p)
	ext := path.Ext(base)
	return IsReservedName(base) || ext != "" && !strings.EqualFold(ext, ".md")
}

// ResolveLink resolves an in-bundle destination to a concept ID. ok is false
// for external and fragment links, links escaping the bundle, and file links.
func ResolveLink(href, fromID string) (string, bool) {
	p, ok := BundlePath(href, fromID)
	if !ok || isFilePath(p) {
		return "", false
	}
	if strings.HasSuffix(strings.ToLower(p), ".md") {
		p = p[:len(p)-3]
	}
	return p, true
}

// escapes reports whether resolving href from fromID climbs above the root.
func escapes(href, fromID string) bool {
	var segs []string
	if !strings.HasPrefix(href, "/") {
		if d := path.Dir(fromID); d != "." {
			segs = strings.Split(d, "/")
		}
	}
	for _, s := range strings.Split(strings.TrimPrefix(href, "/"), "/") {
		switch s {
		case "", ".":
		case "..":
			if len(segs) == 0 {
				return true
			}
			segs = segs[:len(segs)-1]
		default:
			segs = append(segs, s)
		}
	}
	return false
}

// Targets returns the distinct in-bundle concept IDs a body links to.
func Targets(body, fromID string) []string {
	seen := map[string]bool{}
	var out []string
	for _, l := range Links(body) {
		if t, ok := ResolveLink(l.Href, fromID); ok && !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	return out
}
