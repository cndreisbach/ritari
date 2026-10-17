package vocab

import (
	"strings"
)

// Slug lowercases and trims s, replaces every run of characters outside
// [a-z0-9] with "-", and trims leading and trailing "-".
func Slug(s string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
			dash = false
		} else if !dash {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

// Plural pluralizes a type slug to get its directory name.
func Plural(s string) string {
	switch {
	case s == "":
		return ""
	case strings.HasSuffix(s, "s"), strings.HasSuffix(s, "x"), strings.HasSuffix(s, "ch"), strings.HasSuffix(s, "sh"):
		return s + "es"
	case strings.HasSuffix(s, "y") && len(s) > 1 && !strings.ContainsRune("aeiou", rune(s[len(s)-2])):
		return s[:len(s)-1] + "ies"
	}
	return s + "s"
}

// Humanize turns a directory path's last segment into a heading: split on - and
// _, capitalize each word.
func Humanize(dir string) string {
	if i := strings.LastIndexByte(dir, '/'); i >= 0 {
		dir = dir[i+1:]
	}
	words := strings.FieldsFunc(dir, func(r rune) bool { return r == '-' || r == '_' })
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}
