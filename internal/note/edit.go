package note

import (
	"fmt"
	"slices"
	"strings"
)

// EditBody applies one `note edit` operation to a body (§5.7).
func EditBody(body, op, content, section, find string, count int) (string, error) {
	switch op {
	case "append":
		if body != "" && !strings.HasSuffix(body, "\n") {
			body += "\n"
		}
		return body + content + "\n", nil
	case "prepend":
		return content + "\n" + body, nil
	case "find_replace":
		if find == "" {
			return "", fmt.Errorf("find_replace requires --find")
		}
		n := strings.Count(body, find)
		if n == 0 {
			return "", fmt.Errorf("text not found: %q (only the body is searched, not frontmatter; use `rit note write` or `rit note edit --title/--description/--tags/--resource/--meta` for frontmatter fields)", find)
		}
		if count > 0 && n != count {
			return "", fmt.Errorf("expected %d occurrences of %q, found %d", count, find, n)
		}
		return strings.ReplaceAll(body, find, content), nil
	case "replace_section", "insert_before_section", "insert_after_section":
		if strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(section), "#")) == "" {
			return "", fmt.Errorf("%s requires a section heading", op)
		}
		ls := lines(body)
		start, end, ok := findSection(ls, section)
		if !ok {
			return "", fmt.Errorf("section not found: %q", section)
		}
		c := lines(content)
		switch op {
		case "replace_section":
			ls = slices.Concat(ls[:start+1], []string{""}, c, []string{""}, ls[end:])
		case "insert_before_section":
			ls = slices.Concat(ls[:start], c, []string{""}, ls[start:])
		case "insert_after_section":
			ls = slices.Concat(ls[:end], c, []string{""}, ls[end:])
		}
		return join(ls), nil
	}
	return "", fmt.Errorf("unknown edit op: %q", op)
}

// AppendNarrative adds paragraphs at the end of the prose before the first
// heading, leaving everything under headings untouched.
func AppendNarrative(body string, paras []string) string {
	if len(paras) == 0 {
		return body
	}
	ls := lines(body)
	h := len(ls)
	for i, l := range ls {
		if _, ok := parseHeading(l); ok {
			h = i
			break
		}
	}
	out := trimBlank(ls[:h])
	for _, p := range paras {
		if len(out) > 0 {
			out = append(out, "")
		}
		out = append(out, p)
	}
	if h < len(ls) {
		out = append(append(out, ""), ls[h:]...)
	}
	return join(out)
}

// SetSection replaces the named section wholesale with `# name` and items,
// creating it at the end of the body if missing, or removing it when items is
// empty.
func SetSection(body, name string, items []string) string {
	ls := lines(body)
	var section []string
	if len(items) > 0 {
		section = append([]string{"# " + name}, items...)
	}
	before, after := trimBlank(ls), []string(nil)
	if start, end, ok := findSection(ls, name); ok {
		before, after = trimBlank(ls[:start]), ls[end:]
	}
	out := before
	for _, part := range [][]string{section, after} {
		if len(part) == 0 {
			continue
		}
		if len(out) > 0 {
			out = append(out, "")
		}
		out = append(out, part...)
	}
	return join(out)
}

func trimBlank(ls []string) []string {
	ls = slices.Clone(ls)
	for len(ls) > 0 && strings.TrimSpace(ls[len(ls)-1]) == "" {
		ls = ls[:len(ls)-1]
	}
	return ls
}

// Document is the part of the body that `replace_body` replaces and
// `note view --body` prints (§5.7): the body without its Observations and
// Relations sections or the footnote definitions derived from sources.
func (n *Note) Document() string {
	ls := lines(n.EditableBody())
	kept := sectionMask(ls, ObservationsSection, RelationsSection)
	var doc []string
	for i, l := range ls {
		if !kept[i] {
			doc = append(doc, l)
		}
	}
	return join(trimBlankEdges(doc))
}

// ReplaceDocument is the `replace_body` edit (§5.7): content becomes the
// note's document, followed by the note's existing Observations and Relations
// sections in their original order.
func (n *Note) ReplaceDocument(content string) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("replace_body requires --content (the new document)")
	}
	if fm, _, ok := split(content); ok {
		if parsed, err := parseFrontmatter(fm); err == nil && (parsed.Type != "" || parsed.Title != "") {
			return fmt.Errorf("replace_body content starts with a frontmatter block; pass only the document (rit note view --body prints it)")
		}
	}
	for _, h := range Headings(content) {
		if strings.EqualFold(h.Text, ObservationsSection) || strings.EqualFold(h.Text, RelationsSection) {
			return fmt.Errorf("replace_body content has a %q section, but the note keeps its own; add observations with rit note write --obs and relations with --rel", h.Text)
		}
	}
	derived := map[string]string{}
	for _, d := range derivedDefs(n.FM.Sources) {
		label := footnoteDefRe.FindStringSubmatch(d)[1]
		derived[label] = d
	}
	var doc []string
	for _, l := range lines(content) {
		if m := footnoteDefRe.FindStringSubmatch(l); m != nil {
			if want, ok := derived[m[1]]; ok {
				if strings.TrimSpace(l) != want {
					return fmt.Errorf("replace_body content defines footnote [^%s], but [^%s] is this note's source %s; renumber the document's own footnotes, or cite the source with [^%s] and leave its definition out", m[1], m[1], m[1], m[1])
				}
				continue
			}
		}
		doc = append(doc, l)
	}
	out := trimBlankEdges(doc)
	ls := lines(n.EditableBody())
	kept := sectionMask(ls, ObservationsSection, RelationsSection)
	for i := 0; i < len(ls); {
		if !kept[i] {
			i++
			continue
		}
		j := i
		for j < len(ls) && kept[j] {
			j++
		}
		out = append(append(out, ""), trimBlank(ls[i:j])...)
		i = j
	}
	n.Body = join(out)
	return nil
}

func trimBlankEdges(ls []string) []string {
	for len(ls) > 0 && strings.TrimSpace(ls[0]) == "" {
		ls = ls[1:]
	}
	return trimBlank(ls)
}
