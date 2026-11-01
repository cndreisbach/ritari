package screen

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// charNames names the invisible characters the screen reports. Ranges
// without a name here are named by charName.
var charNames = map[rune]string{
	0x034F: "COMBINING GRAPHEME JOINER",
	0x061C: "ARABIC LETTER MARK",
	0x115F: "HANGUL CHOSEONG FILLER",
	0x1160: "HANGUL JUNGSEONG FILLER",
	0x17B4: "KHMER VOWEL INHERENT AQ",
	0x17B5: "KHMER VOWEL INHERENT AA",
	0x180B: "MONGOLIAN FREE VARIATION SELECTOR ONE",
	0x180C: "MONGOLIAN FREE VARIATION SELECTOR TWO",
	0x180D: "MONGOLIAN FREE VARIATION SELECTOR THREE",
	0x180E: "MONGOLIAN VOWEL SEPARATOR",
	0x180F: "MONGOLIAN FREE VARIATION SELECTOR FOUR",
	0x200B: "ZERO WIDTH SPACE",
	0x200C: "ZERO WIDTH NON-JOINER",
	0x200D: "ZERO WIDTH JOINER",
	0x200E: "LEFT-TO-RIGHT MARK",
	0x200F: "RIGHT-TO-LEFT MARK",
	0x2028: "LINE SEPARATOR",
	0x2029: "PARAGRAPH SEPARATOR",
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
	0x2065: "INVISIBLE FORMAT CHARACTER",
	0x2066: "LEFT-TO-RIGHT ISOLATE",
	0x2067: "RIGHT-TO-LEFT ISOLATE",
	0x2068: "FIRST STRONG ISOLATE",
	0x2069: "POP DIRECTIONAL ISOLATE",
	0x2800: "BRAILLE PATTERN BLANK",
	0x3164: "HANGUL FILLER",
	0xFEFF: "ZERO WIDTH NO-BREAK SPACE (BOM)",
	0xFFA0: "HALFWIDTH HANGUL FILLER",
	0xFFF9: "INTERLINEAR ANNOTATION ANCHOR",
	0xFFFA: "INTERLINEAR ANNOTATION SEPARATOR",
	0xFFFB: "INTERLINEAR ANNOTATION TERMINATOR",
	0xFFFC: "OBJECT REPLACEMENT CHARACTER",
}

const (
	blackFlag = 0x1F3F4 // starts an emoji tag sequence (subdivision flags)
	cancelTag = 0xE007F
	softHyph  = 0x00AD
)

func isTag(r rune) bool { return r >= 0xE0000 && r <= 0xE007F }

// isVS reports whether r is a variation selector (VS1–VS16, or the
// supplement VS17–VS256).
func isVS(r rune) bool { return (r >= 0xFE00 && r <= 0xFE0F) || (r >= 0xE0100 && r <= 0xE01EF) }

// suspicious reports whether r is a character the screen considers invisible,
// before any allowance for its context.
func suspicious(r rune) bool {
	switch {
	case r == '\t' || r == '\n' || r == '\r':
		return false
	case r < 0x20 || r == 0x7F || (r >= 0x80 && r <= 0x9F):
		return true
	case isTag(r), isVS(r):
		return true
	case r >= 0x206A && r <= 0x206F: // deprecated format characters
		return true
	case r >= 0x1D173 && r <= 0x1D17A: // musical symbol format characters
		return true
	}
	_, ok := charNames[r]
	return ok
}

// ignorable reports whether normalize drops r before pattern matching: every
// suspicious character, and the soft hyphen, which is invisible but too common
// in copied text to report.
func ignorable(r rune) bool { return r == softHyph || suspicious(r) }

// joinerScripts are the scripts in which ZWNJ and ZWJ shape text legitimately.
var joinerScripts = []*unicode.RangeTable{
	unicode.Arabic, unicode.Syriac, unicode.Devanagari, unicode.Bengali, unicode.Gurmukhi,
	unicode.Gujarati, unicode.Oriya, unicode.Tamil, unicode.Telugu, unicode.Kannada,
	unicode.Malayalam, unicode.Sinhala, unicode.Myanmar, unicode.Khmer,
}

// spaceScripts write words without spaces and use ZWSP to mark word breaks.
var spaceScripts = []*unicode.RangeTable{unicode.Thai, unicode.Lao, unicode.Khmer, unicode.Myanmar, unicode.Tibetan}

func inJoinerScript(r rune) bool { return unicode.In(r, joinerScripts...) }

// emoji reports whether r is an emoji code point that can sit on either side
// of a ZWJ or take a variation selector: the pictograph blocks, the
// miscellaneous symbols and dingbats, and skin-tone modifiers. ©, ®, and ™
// are not included.
func emoji(r rune) bool {
	return (r >= 0x1F000 && r <= 0x1FAFF) || (r >= 0x2600 && r <= 0x27BF) || (r >= 0x2B00 && r <= 0x2BFF) ||
		(r >= 0x2190 && r <= 0x21FF) || (r >= 0x2300 && r <= 0x23FF) || r == 0x00A9 || r == 0x00AE ||
		r == 0x2122 || r == 0x203C || r == 0x2049 || r == 0x3030 || r == 0x303D || r == 0x3297 || r == 0x3299
}

// emojiJoinable is emoji without the text-style symbols (©, ®, ™, arrows),
// which never join with ZWJ; it accepts the emoji variation selector.
func emojiJoinable(r rune) bool {
	return r == 0xFE0F || (r >= 0x1F000 && r <= 0x1FAFF) || (r >= 0x2600 && r <= 0x27BF) || (r >= 0x2B00 && r <= 0x2BFF)
}

func rtl(r rune) bool {
	return unicode.In(r, unicode.Hebrew, unicode.Arabic, unicode.Syriac, unicode.Thaana, unicode.Nko)
}

// flagged marks, for each rune of rs, whether it is an invisible character
// not allowed by its context. It runs in one pass, so long runs stay linear.
func flagged(rs []rune) []bool {
	out := make([]bool, len(rs))
	hasRTL := false
	for _, r := range rs {
		if rtl(r) {
			hasRTL = true
			break
		}
	}
	allowedTags := flagSequences(rs)
	at := func(i int) rune {
		if i < 0 || i >= len(rs) {
			return 0
		}
		return rs[i]
	}
	for i, r := range rs {
		if !suspicious(r) {
			continue
		}
		prev, next := at(i-1), at(i+1)
		bad := true
		switch {
		case r == 0x200B: // ZWSP: word breaks in Thai, Lao, Khmer, Myanmar, Tibetan
			bad = !unicode.In(prev, spaceScripts...) && !unicode.In(next, spaceScripts...)
		case r == 0x200C: // ZWNJ: Persian, Arabic, and Indic orthography
			bad = !inJoinerScript(prev) && !inJoinerScript(next)
		case r == 0x200D: // ZWJ: emoji sequences and Indic conjuncts
			ok := (emojiJoinable(prev) && emojiJoinable(next)) || inJoinerScript(prev) || inJoinerScript(next)
			bad = !ok
		case r == 0x200E, r == 0x200F, r == 0x061C: // directional marks in right-to-left text
			bad = !hasRTL
		case r >= 0x180B && r <= 0x180F && r != 0x180E: // Mongolian variation selectors
			bad = !unicode.Is(unicode.Mongolian, prev)
		case r == 0xFE0E || r == 0xFE0F: // text/emoji presentation, once, after an emoji or keycap base
			base := emoji(prev) || prev == '#' || prev == '*' || (prev >= '0' && prev <= '9')
			bad = isVS(prev) || !base
		case isVS(r): // standardized and ideographic variation sequences, once, after a base
			base := unicode.In(prev, unicode.Han, unicode.Mongolian, unicode.Sm) || emoji(prev)
			bad = isVS(prev) || !base
		case isTag(r):
			bad = !allowedTags[i]
		}
		out[i] = bad
	}
	return out
}

// flagSequences marks the tag characters of rs that belong to an emoji
// subdivision flag: U+1F3F4, then 2–7 tag digits or lowercase tag letters,
// then CANCEL TAG. Any other tag run is smuggled text.
func flagSequences(rs []rune) []bool {
	ok := make([]bool, len(rs))
	for i := 0; i < len(rs); i++ {
		if rs[i] != blackFlag {
			continue
		}
		j := i + 1
		for j < len(rs) && ((rs[j] >= 0xE0030 && rs[j] <= 0xE0039) || (rs[j] >= 0xE0061 && rs[j] <= 0xE007A)) {
			j++
		}
		if n := j - i - 1; n >= 2 && n <= 7 && j < len(rs) && rs[j] == cancelTag {
			for k := i + 1; k <= j; k++ {
				ok[k] = true
			}
			i = j
		}
	}
	return ok
}

// charName describes a code point for a message: `U+202E RIGHT-TO-LEFT OVERRIDE`.
func charName(r rune) string {
	n := charNames[r]
	switch {
	case n != "":
	case isTag(r):
		n = "TAG CHARACTER"
	case isVS(r):
		n = "VARIATION SELECTOR"
	case r >= 0x206A && r <= 0x206F, r >= 0x1D173 && r <= 0x1D17A:
		n = "FORMAT CHARACTER"
	default:
		n = "CONTROL CHARACTER"
	}
	return fmt.Sprintf("U+%04X %s", r, n)
}

// invisibles returns the distinct disallowed invisible characters in s, in
// code-point order, and how many there are in all.
func invisibles(s string) (distinct []rune, total int) {
	if !strings.ContainsFunc(s, suspicious) {
		return nil, 0
	}
	rs := []rune(s)
	seen := map[rune]bool{}
	for i, bad := range flagged(rs) {
		if bad {
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

// smuggled decodes the tag characters of s that mirror printable ASCII
// (U+E0020–U+E007E), the "ASCII smuggling" trick, outside emoji flags. The
// runs are joined with a space wherever other characters separate them, so a
// message split across visible text still reads as one.
func smuggled(s string) string {
	if !strings.ContainsFunc(s, isTag) {
		return ""
	}
	rs := []rune(s)
	allowed := flagSequences(rs)
	var b strings.Builder
	gap := false
	for i, r := range rs {
		if r >= 0xE0020 && r <= 0xE007E && !allowed[i] {
			if gap && b.Len() > 0 {
				b.WriteByte(' ')
			}
			gap = false
			b.WriteRune(r - 0xE0000)
			continue
		}
		if !isTag(r) || !allowed[i] {
			gap = true
		}
	}
	return strings.TrimSpace(b.String())
}

// Escape rewrites the disallowed invisible characters in s as visible
// `\u{XXXX}` escapes, for text-mode output (§8.1). Allowed characters (an
// emoji ZWJ, a Persian ZWNJ) are left alone, so emoji and words still render.
func Escape(s string) string {
	if !strings.ContainsFunc(s, suspicious) {
		return s
	}
	rs := []rune(s)
	bad := flagged(rs)
	var b strings.Builder
	for i, r := range rs {
		if bad[i] {
			fmt.Fprintf(&b, `\u{%04X}`, r)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// greekConfusable are the Greek letters that look like Latin ones in common
// fonts. Other Greek letters (Δ, Ω, λ, π, µ) mixed with Latin are notation.
const greekConfusable = "ΑΒΕΖΗΙΚΜΝΟΡΤΥΧοικνρυχ"

// mixedScript returns the words in s that mix Latin letters with Cyrillic
// ones, or with Greek letters that pass for Latin: the shape of a homoglyph
// spoof ("pаypal" with a Cyrillic а).
func mixedScript(s string) []string {
	var out []string
	seen := map[string]bool{}
	for _, w := range strings.FieldsFunc(s, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsMark(r) }) {
		var latin, other bool
		for _, r := range w {
			switch {
			case unicode.Is(unicode.Latin, r):
				latin = true
			case unicode.Is(unicode.Cyrillic, r), strings.ContainsRune(greekConfusable, r):
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
