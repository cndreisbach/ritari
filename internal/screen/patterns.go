package screen

import (
	"regexp"
	"slices"
	"strings"
	"unicode"
)

// pattern is one entry in the pattern bank: a category and the expression
// that recognizes it in normalized text (lowercase, single spaces).
type pattern struct {
	category string
	re       *regexp.Regexp
}

// Categories of the pattern bank (after okf-guard's).
const (
	InstructionOverride = "instruction override"
	FakeSystemMarker    = "fake system marker"
	FakeRole            = "fake role"
	Jailbreak           = "jailbreak"
	DataExfiltration    = "data exfiltration"
)

const (
	earlier  = `(?:previous|prior|above|earlier|preceding|foregoing|original|all|any|your|the|these|those|other)`
	guidance = `(?:instructions?|prompts?|directions?|directives?|rules|guidelines|context|messages?|commands?|guidance|constraints|system prompt)`
	sendVerb = `(?:send|post|upload|exfiltrate|leak|forward|e-?mail|transmit|submit|share|copy)`
	secret   = `(?:api[ _-]?keys?|passwords?|passphrases?|secrets?|tokens?|credentials?|private keys?|ssh keys?|\.env\b|environment variables?|cookies?|session ids?)`
)

var bank = []pattern{
	{InstructionOverride, re(`\b(?:ignore|disregard|forget|override|overrule|bypass)\b(?: \S+){0,3}? ` + earlier + `(?: \S+){0,2}? ` + guidance + `\b`)},
	{InstructionOverride, re(`\b(?:new|updated|revised|real|actual|hidden|secret) (?:instructions?|directives?|system prompt)\b ?:`)},
	{InstructionOverride, re(`\b(?:do not|don't|never) (?:tell|inform|mention (?:this )?to|reveal (?:this )?to|alert|show) the user\b`)},
	{FakeSystemMarker, re(`\[/?(?:system|inst|sys)\]|<<\s*/?sys\s*>>|<\|?\s*(?:system|im_start|im_end|endoftext|start_header_id|end_header_id|eot_id)\s*\|?>|</?(?:system|system[_-]prompt|instructions?)>`)},
	{FakeSystemMarker, re(`(?:^|[\n#*])\s*system(?: message| prompt| override)?\s*:`)},
	{FakeSystemMarker, re(`\b(?:begin|start|end) (?:of )?(?:the )?system (?:prompt|message|instructions?)\b`)},
	{FakeRole, re(`\byou are (?:now|no longer|actually)\b`)},
	{FakeRole, re(`\bfrom now on,? (?:you|your|act|respond|always)\b`)},
	{FakeRole, re(`\b(?:pretend|imagine) (?:to be|you are|that you are)\b`)},
	{FakeRole, re(`\b(?:act|behave|respond) as (?:if you were |though you were )?(?:an? |the )?(?:unrestricted|unfiltered|different|new|evil|admin|root|system|developer)\b`)},
	{FakeRole, re(`(?:^|\n)\s*(?:assistant|user|human|ai)\s*:`)},
	{Jailbreak, re(`\b(?:jailbreak|jailbroken|developer mode|god mode|dan mode|do anything now)\b`)},
	{Jailbreak, re(`\b(?:bypass|disable|turn off|circumvent|evade) (?:your |the |all |any )?(?:safety|guardrails?|filters?|restrictions|content policy|moderation)\b`)},
	{DataExfiltration, re(`\b` + sendVerb + `\b.{0,80}?\b(?:to|at|into|via)\b.{0,30}?(?:https?://|\b[\w.+-]+@[\w-]+\.[\w.]+)`)},
	{DataExfiltration, re(`\b(?:curl|wget|invoke-webrequest|fetch)\b.{0,40}?https?://`)},
	{DataExfiltration, re(`\b(?:` + sendVerb + `|reveal|print|output|dump|read|cat|include|append)\b.{0,40}?\b` + secret)},
	{DataExfiltration, re(`!\[[^\]]*\]\(https?://[^)\s]*\?[^)\s]*=`)},
}

// re compiles a bank expression; a space in it matches any one space or
// newline, since normalized text keeps newlines.
func re(expr string) *regexp.Regexp {
	return regexp.MustCompile(strings.ReplaceAll(expr, " ", `\s`))
}

// normalize prepares text for the pattern bank: invisible characters are
// (and the soft hyphen) dropped (so "ig\u200bnore" still reads "ignore"), letters are lowercased,
// and runs of spaces collapse. Newlines survive, for the line-start patterns.
func normalize(s string) string {
	var b strings.Builder
	space := false
	for _, r := range s {
		switch {
		case ignorable(r) && r != '\n':
			continue
		case r == '\n':
			b.WriteByte('\n')
			space = false
			continue
		case unicode.IsSpace(r):
			if !space {
				b.WriteByte(' ')
			}
			space = true
			continue
		}
		space = false
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// Classify returns the pattern-bank categories that text matches, in bank
// order without repeats.
func Classify(text string) []string {
	t := normalize(text)
	var out []string
	for _, p := range bank {
		if !slices.Contains(out, p.category) && p.re.MatchString(t) {
			out = append(out, p.category)
		}
	}
	return out
}
