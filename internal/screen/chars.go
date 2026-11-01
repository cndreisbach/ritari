package screen

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// charNames names the invisible characters the screen reports. A character
// not listed here but in a flagged range is named by its code point alone.
var charNames = map[rune]string{
	0x061C: "ARABIC LETTER MARK",
	0x180E: "MONGOLIAN VOWEL SEPARATOR",
	0x200B: "ZERO WIDTH SPACE",
	0x200C: "ZERO WIDTH NON-JOINER",
	0x200D: "ZERO WIDTH JOINER",
	0x200E: "LEFT-TO-RIGHT MARK",
	0x200F: "RIGHT-TO-LEFT MARK",
	0x202A: "LEFT-TO-RIGHT EMBEDDING",
	0x202B: "RIGHT-TO-LEFT EMBEDDING",
	0x202C: "POP DIRECTIONAL FORMATTING",
	0x202D: "LEFT-TO-RIGHT OVERRIDE",
	0x202E: "RIGHT-TO-LEFT OVERRIDE",
	0x2060: "WORD JOINER",
	0x2061: "FUNCTION APPLICATION",
	0x2062: "INVISIBLE TIMES",
	0x2063: "INVISIBLE SEPARATOR",
	0x2064: "INVISIBLE PLUS",
	0x2066: "LEFT-TO-RIGHT ISOLATE",
	0x2067: "RIGHT-TO-LEFT ISOLATE",
	0x2068: "FIRST STRONG ISOLATE",
	0x2069: "POP DIRECTIONAL ISOLATE",
	0xFEFF: "ZERO WIDTH NO-BREAK SPACE (BOM)",
}

const (
	blackFlag = 0x1F3F4 // starts an emoji tag sequence (subdivision flags)
	cancelTag = 0xE007F
)

func isTag(r rune) bool { return r >= 0xE0000 && r <= 0xE007F }

// suspicious reports whether r is a character the screen considers invisible,
// before any allowance for its context.
func suspicious(r rune) bool {
	switch {
	case r == '\t' || r == '\n' || r == '\r':
		return false
	case r < 0x20 || r == 0x7F || (r >= 0x80 && r <= 0x9F):
		return true
	case isTag(r):
		return true
	}
	_, ok := charNames[r]
	return ok
}

// joinerScripts are the scripts in which ZWNJ and ZWJ shape text legitimately.
var joinerScripts = []*unicode.RangeTable{
	unicode.Arabic, unicode.Syriac, unicode.Devanagari, unicode.Bengali, unicode.Gurmukhi,
	unicode.Gujarati, unicode.Oriya, unicode.Tamil, unicode.Telugu, unicode.Kannada,
	unicode.Malayalam, unicode.Sinhala, unicode.Myanmar, unicode.Khmer,
}

func inJoinerScript(r rune) bool { return unicode.In(r, joinerScripts...) }

// emojiish reports whether r can sit on either side of a ZWJ in an emoji
// sequence: pictographs, skin-tone modifiers, and the emoji variation selector.
func emojiish(r rune) bool {
	return unicode.Is(unicode.So, r) || (r >= 0x1F3FB && r <= 0x1F3FF) || r == 0xFE0F
}

func rtl(r rune) bool {
	return unicode.In(r, unicode.Hebrew, unicode.Arabic, unicode.Syriac, unicode.Thaana, unicode.Nko)
}

// invisibleAt reports whether the character at rs[i] is invisible and not
// allowed by its context. hasRTL says whether the surrounding text contains a
// right-to-left script, which makes the directional marks legitimate.
func invisibleAt(rs []rune, i int, hasRTL bool) bool {
	r := rs[i]
	if !suspicious(r) {
		return false
	}
	prev, next := rune(0), rune(0)
	if i > 0 {
		prev = rs[i-1]
	}
	if i+1 < len(rs) {
		next = rs[i+1]
	}
	switch r {
	case 0x200C: // ZWNJ: Persian, Arabic, and Indic orthography
		return !inJoinerScript(prev) && !inJoinerScript(next)
	case 0x200D: // ZWJ: emoji sequences and Indic conjuncts
		if emojiish(prev) && emojiish(next) {
			return false
		}
		return !inJoinerScript(prev) && !inJoinerScript(next)
	case 0x200E, 0x200F, 0x061C: // directional marks in right-to-left text
		return !hasRTL
	}
	if isTag(r) {
		return !inFlagSequence(rs, i)
	}
	return true
}

// inFlagSequence reports whether the tag character at rs[i] belongs to an
// emoji subdivision flag: U+1F3F4, then tag characters, then CANCEL TAG.
func inFlagSequence(rs []rune, i int) bool {
	start := i
	for start > 0 && isTag(rs[start-1]) && rs[start-1] != cancelTag {
		start--
	}
	if start == 0 || rs[start-1] != blackFlag {
		return false
	}
	for j := i; j < len(rs); j++ {
		if rs[j] == cancelTag {
			return true
		}
		if !isTag(rs[j]) {
			return false
		}
	}
	return false
}

func hasRTLText(rs []rune) bool {
	for _, r := range rs {
		if rtl(r) {
			return true
		}
	}
	return false
}

// charName describes a code point for a message: `U+202E RIGHT-TO-LEFT OVERRIDE`.
func charName(r rune) string {
	n := charNames[r]
	switch {
	case n != "":
	case isTag(r):
		n = "TAG CHARACTER"
	default:
		n = "CONTROL CHARACTER"
	}
	return fmt.Sprintf("U+%04X %s", r, n)
}

// invisibles returns the distinct disallowed invisible characters in s, in
// code-point order, and how many there are in all.
func invisibles(s string) (distinct []rune, total int) {
	rs := []rune(s)
	hasRTL := hasRTLText(rs)
	seen := map[rune]bool{}
	for i := range rs {
		if invisibleAt(rs, i, hasRTL) {
			total++
			if !seen[rs[i]] {
				seen[rs[i]] = true
				distinct = append(distinct, rs[i])
			}
		}
	}
	sort.Slice(distinct, func(a, b int) bool { return distinct[a] < distinct[b] })
	return distinct, total
}

// smuggled decodes runs of tag characters that mirror printable ASCII
// (U+E0020–U+E007E), the "ASCII smuggling" trick, outside emoji flags.
func smuggled(s string) []string {
	rs := []rune(s)
	var out []string
	var b strings.Builder
	flush := func() {
		if t := strings.TrimSpace(b.String()); t != "" {
			out = append(out, t)
		}
		b.Reset()
	}
	for i, r := range rs {
		if r >= 0xE0020 && r <= 0xE007E && !inFlagSequence(rs, i) {
			b.WriteRune(r - 0xE0000)
			continue
		}
		flush()
	}
	flush()
	return out
}

// Escape rewrites the disallowed invisible characters in s as visible
// `\u{XXXX}` escapes, for text-mode output (§8.1). Allowed characters (an
// emoji ZWJ, a Persian ZWNJ) are left alone, so emoji and words still render.
func Escape(s string) string {
	if !strings.ContainsFunc(s, suspicious) {
		return s
	}
	rs := []rune(s)
	hasRTL := hasRTLText(rs)
	var b strings.Builder
	for i, r := range rs {
		if invisibleAt(rs, i, hasRTL) {
			fmt.Fprintf(&b, `\u{%04X}`, r)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// mixedScript returns the words in s that mix Latin letters with Cyrillic or
// Greek ones, the shape of a homoglyph spoof ("pаypal" with a Cyrillic а).
func mixedScript(s string) []string {
	var out []string
	seen := map[string]bool{}
	for _, w := range strings.FieldsFunc(s, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsMark(r) }) {
		var latin, other bool
		for _, r := range w {
			switch {
			case r == 0x00B5: // MICRO SIGN, as in µm
			case unicode.Is(unicode.Latin, r):
				latin = true
			case unicode.In(r, unicode.Cyrillic, unicode.Greek):
				other = true
			}
		}
		if latin && other && !seen[w] {
			seen[w] = true
			out = append(out, w)
		}
	}
	return out
}

func scriptsOf(w string) string {
	var names []string
	add := func(n string) {
		for _, x := range names {
			if x == n {
				return
			}
		}
		names = append(names, n)
	}
	for _, r := range w {
		switch {
		case r == 0x00B5:
		case unicode.Is(unicode.Latin, r):
			add("Latin")
		case unicode.Is(unicode.Cyrillic, r):
			add("Cyrillic")
		case unicode.Is(unicode.Greek, r):
			add("Greek")
		}
	}
	return strings.Join(names, " and ")
}
