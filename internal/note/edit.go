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
