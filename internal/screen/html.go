package screen

import (
	"regexp"
	"strings"
)

// voidElements never have content or a closing tag.
var voidElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true, "hr": true, "img": true,
	"input": true, "link": true, "meta": true, "param": true, "source": true, "track": true, "wbr": true,
}

// neverRendered are elements whose content a browser never shows.
var neverRendered = map[string]bool{"script": true, "style": true, "template": true, "noscript": true}

// rawText elements hold text up to their closing tag; nothing in them is markup.
var rawText = map[string]bool{"script": true, "style": true, "textarea": true, "title": true, "xmp": true}

// hiddenCSS recognizes a style that hides its element, matched against the
// lowercased style with all whitespace removed.
var hiddenCSS = regexp.MustCompile(`(?:^|[;{])(?:` +
	`display:none|visibility:(?:hidden|collapse)|` +
	`font-size:0(?:\.0+)?(?:px|pt|em|rem|%|vw|vh)?(?:[;!}]|$)|` +
	`opacity:0(?:\.0+)?%?(?:[;!}]|$)|` +
	`color:(?:transparent|rgba\([^)]*,0(?:\.0+)?\)|#[0-9a-f]{6}00(?:[;!}]|$)|#[0-9a-f]{3}0(?:[;!}]|$))|` +
	`(?:max-)?(?:height|width):0(?:px|em|rem|%)?(?:[;!}]|$)|` +
	`clip:rect\(0|clip-path:(?:inset\(50%|circle\(0)|` +
	`transform:scale\(0(?:[,)])|` +
	`(?:left|top|right|bottom|text-indent|margin-left|margin-top):-\d{3,})`)

func hidesByAttrs(attrs []attr) bool {
	for _, a := range attrs {
		switch a.name {
		case "hidden":
			return true
		case "style":
			css := strings.Join(strings.Fields(strings.ToLower(a.value)), "")
			if hiddenCSS.MatchString(css) {
				return true
			}
		}
	}
	return false
}

// attr is one attribute of a start tag; start and end are its byte range,
// name to end of value.
type attr struct {
	name, value string
	start, end  int
}

// hidesInput reports a form field that never shows: <input type=hidden>.
func hidesInput(name string, attrs []attr) bool {
	if name != "input" {
		return false
	}
	for _, a := range attrs {
		if a.name == "type" && strings.EqualFold(strings.TrimSpace(a.value), "hidden") {
			return true
		}
	}
	return false
}

// linkAttrs name where a link or image points; they are scanned but kept by
// Strip, since a reader follows them.
var linkAttrs = map[string]bool{"href": true, "src": true, "srcset": true}

type openElem struct {
	name  string
	hides bool
}

// htmlSpans tokenizes the HTML in a body. masked is the body with code spans
// and code blocks blanked out (same byte offsets), so structure comes from
// masked and text from src. Comments and elements that hide their content
// become flagged spans; an element left open hides to the end of the body,
// as it does in a browser. Attribute values are scanned, not flagged.
func htmlSpans(src, masked string) []span {
	var out []span
	var stack []openElem
	hiding := 0 // number of hiding elements on the stack
	var hidden strings.Builder
	hiddenKind, hiddenStart := "", 0
	flush := func(end int) {
		// An element with nothing in it is still removed by Strip, but is
		// not reported.
		t := strings.TrimSpace(hidden.String())
		out = append(out, span{kind: hiddenKind, text: t, flag: t != "", strip: true, start: hiddenStart, end: end})
		hidden.Reset()
		hiddenKind = ""
	}
	text := func(from, to int) {
		if hiding > 0 && to > from {
			hidden.WriteString(src[from:to])
			hidden.WriteByte(' ')
		}
	}
	i, last := 0, 0
	for i < len(masked) {
		lt := strings.IndexByte(masked[i:], '<')
		if lt < 0 {
			break
		}
		i += lt
		if i > 0 && masked[i-1] == '\\' && (i < 2 || masked[i-2] != '\\') {
			i++
			continue
		}
		rest := masked[i:]
		switch {
		case strings.HasPrefix(rest, "<!--"):
			text(last, i)
			body, end := comment(src, masked, i)
			if strings.TrimSpace(body) != "" {
				out = append(out, span{kind: "an HTML comment", text: body, flag: true, strip: true, start: i, end: end})
			}
			i, last = end, end
			continue
		case strings.HasPrefix(rest, "<!") || strings.HasPrefix(rest, "<?"):
			// Doctypes, CDATA, and processing instructions are bogus
			// comments in an HTML body: never shown.
			text(last, i)
			end := strings.IndexByte(masked[i:], '>')
			stop := len(masked)
			if end >= 0 {
				stop = i + end + 1
			}
			inner := src[i+2 : max(i+2, stop-1)]
			if strings.TrimSpace(inner) != "" {
				out = append(out, span{kind: "an HTML comment", text: inner, flag: true, strip: true, start: i, end: stop})
			}
			i, last = stop, stop
			continue
		case strings.HasPrefix(rest, "</"):
			name, end, ok := tagName(masked, i+2)
			if !ok {
				i++
				continue
			}
			gt := strings.IndexByte(masked[end:], '>')
			if gt < 0 {
				i++
				continue
			}
			text(last, i)
			i = end + gt + 1
			last = i
			for k := len(stack) - 1; k >= 0; k-- {
				if stack[k].name != name {
					continue
				}
				for _, e := range stack[k:] {
					if e.hides {
						hiding--
					}
				}
				stack = stack[:k]
				if hiding == 0 && hiddenKind != "" {
					flush(i)
				}
				break
			}
			continue
		}
		name, end, ok := tagName(masked, i+1)
		if !ok {
			i++
			continue
		}
		attrs, close, selfClosing, ok := attributes(src, masked, end)
		if !ok {
			i++
			continue
		}
		text(last, i)
		open := i
		hides := neverRendered[name] || hidesByAttrs(attrs) || hidesInput(name, attrs)
		for _, a := range attrs {
			if a.name == "style" || strings.TrimSpace(a.value) == "" {
				continue
			}
			if hides || hiding > 0 {
				// Inside a hidden element, attribute text is part of what it hides.
				hidden.WriteString(a.value)
				hidden.WriteByte(' ')
				continue
			}
			from := a.start // take the space before the attribute with it
			for from > 0 && strings.IndexByte(" \t", masked[from-1]) >= 0 {
				from--
			}
			out = append(out, span{kind: "an HTML attribute", text: a.value, strip: !linkAttrs[a.name], start: from, end: a.end})
		}
		i, last = close, close
		if hides && hiding == 0 {
			hiddenStart = open
			hiddenKind = "a hidden HTML element"
			if neverRendered[name] {
				hiddenKind = "a <" + name + "> element"
			}
		}
		if rawText[name] && !selfClosing {
			endTag := strings.Index(strings.ToLower(masked[i:]), "</"+name)
			stop := len(masked)
			if endTag >= 0 {
				stop = i + endTag
			}
			if hides || hiding > 0 {
				hidden.WriteString(src[i:stop])
				hidden.WriteByte(' ')
			}
			gt := strings.IndexByte(masked[stop:], '>')
			if endTag < 0 || gt < 0 {
				i, last = len(masked), len(masked)
			} else {
				i, last = stop+gt+1, stop+gt+1
			}
			if hides && hiding == 0 {
				flush(i)
			}
			continue
		}
		if voidElements[name] || selfClosing {
			if hides && hiding == 0 {
				flush(i)
			}
			continue
		}
		stack = append(stack, openElem{name: name, hides: hides})
		if hides {
			hiding++
		}
	}
	text(last, len(src))
	if hiddenKind != "" {
		flush(len(src))
	}
	return out
}

// comment returns the text of the comment starting at i and the offset just
// past it. `<!-->` and `<!--->` are empty; an unclosed comment runs to the end.
func comment(src, masked string, i int) (string, int) {
	start := i + 4
	for _, empty := range []string{">", "->"} {
		if strings.HasPrefix(masked[start:], empty) {
			return "", start + len(empty)
		}
	}
	rest := masked[start:]
	end, n := len(rest), 0
	if k := strings.Index(rest, "-->"); k >= 0 {
		end, n = k, 3
	}
	if k := strings.Index(rest, "--!>"); k >= 0 && k < end {
		end, n = k, 4
	}
	return src[start : start+end], start + end + n
}

func isNameByte(c byte, first bool) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z':
		return true
	case first:
		return false
	}
	return c >= '0' && c <= '9' || c == '-' || c == '_' || c == '.'
}

// tagName reads a tag name at i. It must be followed by whitespace, `/`,
// or `>`, which keeps autolinks (`<https://…>`) and addresses out.
func tagName(s string, i int) (string, int, bool) {
	j := i
	for j < len(s) && isNameByte(s[j], j == i) {
		j++
	}
	if j == i || j >= len(s) {
		return "", 0, false
	}
	switch s[j] {
	case ' ', '\t', '\n', '\r', '\f', '/', '>':
	default:
		return "", 0, false
	}
	return strings.ToLower(s[i:j]), j, true
}

// attributes reads the attributes of a start tag from offset i, returning
// them, the offset just past `>`, and whether the tag closed itself.
func attributes(src, masked string, i int) ([]attr, int, bool, bool) {
	var out []attr
	for i < len(masked) {
		for i < len(masked) && strings.IndexByte(" \t\n\r\f", masked[i]) >= 0 {
			i++
		}
		if i >= len(masked) {
			break
		}
		switch masked[i] {
		case '>':
			return out, i + 1, false, true
		case '/':
			if i+1 < len(masked) && masked[i+1] == '>' {
				return out, i + 2, true, true
			}
			i++
			continue
		case '<':
			return nil, 0, false, false
		}
		start := i
		j := i
		for j < len(masked) && strings.IndexByte(" \t\n\r\f/>=", masked[j]) < 0 {
			j++
		}
		name := strings.ToLower(masked[i:j])
		i = j
		for i < len(masked) && strings.IndexByte(" \t\n\r\f", masked[i]) >= 0 {
			i++
		}
		if i >= len(masked) || masked[i] != '=' {
			out = append(out, attr{name: name, start: start, end: j})
			continue
		}
		i++
		for i < len(masked) && strings.IndexByte(" \t\n\r\f", masked[i]) >= 0 {
			i++
		}
		if i >= len(masked) {
			break
		}
		var value string
		if q := masked[i]; q == '"' || q == '\'' {
			k := strings.IndexByte(masked[i+1:], q)
			if k < 0 {
				return nil, 0, false, false
			}
			value = src[i+1 : i+1+k]
			i += k + 2
		} else {
			j := i
			for j < len(masked) && strings.IndexByte(" \t\n\r\f>", masked[j]) < 0 {
				j++
			}
			value = src[i:j]
			i = j
		}
		out = append(out, attr{name: name, value: value, start: start, end: i})
	}
	return nil, 0, false, false
}
