package note

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComposeObservation(t *testing.T) {
	cases := map[string]string{
		"decision: Use JWT #security (ctx)": "- [decision] Use JWT #security (ctx) [^1]",
		"[constraint] 15m expiry":           "- [constraint] 15m expiry [^1]",
		"- plain fact":                      "- [note] plain fact [^1]",
		"Use JWT: fast":                     "- [note] Use JWT: fast [^1]",
		"fact: already cited [^3] [^1]":     "- [fact] already cited [^3] [^1]",
	}
	for in, want := range cases {
		got, err := ComposeObservation(in, []string{"1"})
		require.NoError(t, err, in)
		assert.Equal(t, want, got, in)
	}
	_, err := ComposeObservation("decision:", nil)
	assert.ErrorIs(t, err, ErrEmptyObservation)
}

func TestComposeFreshNote(t *testing.T) {
	n := &Note{FM: Frontmatter{Type: "decision", Title: "Auth uses JWT"}}
	src, _ := ParseSource("PR #1234 https://github.com/acme/api/pull/1234")
	res, err := n.Compose("", Additions{
		Obs:     []string{"decision: Use JWT for stateless auth #security"},
		Rels:    []string{"- depends_on [User Database Schema](/notes/user-database-schema.md)"},
		Sources: []Source{src},
	})
	require.NoError(t, err)
	assert.Equal(t, ComposeResult{AddedObs: 1, AddedRels: 1, AddedSources: 1, BodyChanged: true}, res)
	n.FillGenerated("human:jane@acme.com", "2026-07-29T14:30:00Z")
	assert.Equal(t, `---
type: decision
title: Auth uses JWT
sources:
  - id: "1"
    resource: https://github.com/acme/api/pull/1234
    title: "PR #1234"
generated: {by: human:jane@acme.com, at: 2026-07-29T14:30:00Z}
---

# Observations
- [decision] Use JWT for stateless auth #security [^1]

# Relations
- depends_on [User Database Schema](/notes/user-database-schema.md)

[^1]: [PR #1234](https://github.com/acme/api/pull/1234)
`, string(Render(n)))
}

func TestComposeWithProse(t *testing.T) {
	n := &Note{}
	_, err := n.Compose("Some prose.\n\n\n", Additions{Obs: []string{"a: b"}})
	require.NoError(t, err)
	assert.Equal(t, "Some prose.\n\n# Observations\n- [a] b\n", n.Body)
}

func TestMergeConverges(t *testing.T) {
	n, err := Parse([]byte(example))
	require.NoError(t, err)
	before := string(Render(n))

	res, err := n.Merge(Additions{
		Obs:     []string{"decision: Use JWT for stateless auth #security"},
		Rels:    []string{"- depends_on [User Database Schema](/references/user-database-schema.md)"},
		Sources: []Source{{Resource: "https://github.com/acme/api/pull/1234"}},
	})
	require.NoError(t, err)
	assert.Equal(t, ComposeResult{}, res)
	assert.Equal(t, before, string(Render(n)))
}

func TestMergeNewSourceUnionsRefs(t *testing.T) {
	n, err := Parse([]byte(example))
	require.NoError(t, err)
	res, err := n.Merge(Additions{
		Obs:     []string{"decision: Use JWT for stateless auth #security"},
		Sources: []Source{{Resource: "https://wiki/auth", Title: "Auth wiki"}},
	})
	require.NoError(t, err)
	assert.Equal(t, 0, res.AddedObs)
	assert.Equal(t, 1, res.AddedSources)
	assert.True(t, res.BodyChanged)
	out := string(Render(n))
	assert.Contains(t, out, "- [decision] Use JWT for stateless auth #security [^1] [^2]\n")
	assert.Contains(t, out, "\n[^1]: [PR #1234 — switch to JWT](https://github.com/acme/api/pull/1234)\n[^2]: [Auth wiki](https://wiki/auth)\n")
	assert.Contains(t, out, "  - id: \"2\"\n    resource: https://wiki/auth\n    title: Auth wiki\n")
}

func TestMergeAddsToSections(t *testing.T) {
	n, err := Parse([]byte(example))
	require.NoError(t, err)
	res, err := n.Merge(Additions{
		Obs:  []string{"fact: new one", "fact: new one"},
		Rels: []string{"- relates_to [Other](/notes/other.md)"},
	})
	require.NoError(t, err)
	assert.Equal(t, 1, res.AddedObs)
	assert.Equal(t, 1, res.AddedRels)
	assert.Equal(t, `Server-side sessions would have meant sticky routing, which we did not want.

# Observations
- [decision] Use JWT for stateless auth #security [^1]
- [constraint] Tokens expire after 15m (per security audit) [^1]
- [fact] new one

# Relations
- depends_on [User Database Schema](/references/user-database-schema.md)
- relates_to [Other](/notes/other.md)
`, n.Body)
}

func TestMergeCreatesMissingSections(t *testing.T) {
	n := &Note{Body: "Just prose.\n\n"}
	res, err := n.Merge(Additions{Obs: []string{"a: b"}, Rels: []string{"- x [Y](/y.md)"}})
	require.NoError(t, err)
	assert.Equal(t, 1, res.AddedObs)
	assert.Equal(t, "Just prose.\n\n# Observations\n- [a] b\n\n# Relations\n- x [Y](/y.md)\n", n.Body)
}

func TestMergeRefOnlyPreservesIndent(t *testing.T) {
	n := &Note{Body: "# Observations\n  - [a] b [^1]\n", FM: Frontmatter{Sources: []Source{{ID: "1", Resource: "https://one"}}}}
	res, err := n.Merge(Additions{Obs: []string{"a: b"}, Sources: []Source{{Resource: "https://two"}}})
	require.NoError(t, err)
	assert.Equal(t, 0, res.AddedObs)
	assert.Equal(t, "# Observations\n  - [a] b [^1] [^2]\n", n.Body)
}

func TestDerivedDefsMovedToEnd(t *testing.T) {
	src := "---\ntype: note\nsources:\n  - id: \"1\"\n    resource: https://a\n---\n\n[^1]: stale text\n\nprose [^1]\n\n[^note]: my own footnote\n"
	n, err := Parse([]byte(src))
	require.NoError(t, err)
	assert.Equal(t, "---\ntype: note\nsources:\n  - id: \"1\"\n    resource: https://a\n---\n\n\nprose [^1]\n\n[^note]: my own footnote\n\n[^1]: [https://a](https://a)\n", string(Render(n)))
}

func TestEditBody(t *testing.T) {
	body := "intro\n\n# Steps\n- [ ] one\n\n# Notes\nn\n"
	got, err := EditBody(body, "append", "tail", "", "", 0)
	require.NoError(t, err)
	assert.Equal(t, body+"tail\n", got)

	got, err = EditBody("no newline", "append", "tail", "", "", 0)
	require.NoError(t, err)
	assert.Equal(t, "no newline\ntail\n", got)

	got, err = EditBody(body, "prepend", "head", "", "", 0)
	require.NoError(t, err)
	assert.Equal(t, "head\n"+body, got)

	got, err = EditBody(body, "find_replace", "- [x] one", "", "- [ ] one", 1)
	require.NoError(t, err)
	assert.Contains(t, got, "- [x] one")
	_, err = EditBody(body, "find_replace", "x", "", "- [ ] one", 2)
	assert.EqualError(t, err, `expected 2 occurrences of "- [ ] one", found 1`)
	_, err = EditBody(body, "find_replace", "x", "", "absent", 0)
	assert.ErrorContains(t, err, "only the body is searched")
	_, err = EditBody(body, "find_replace", "x", "", "", 0)
	assert.Error(t, err)

	got, err = EditBody(body, "replace_section", "- [x] one\n- [ ] two", "steps", "", 0)
	require.NoError(t, err)
	assert.Equal(t, "intro\n\n# Steps\n\n- [x] one\n- [ ] two\n\n# Notes\nn\n", got)

	got, err = EditBody(body, "insert_before_section", "before", "# Notes", "", 0)
	require.NoError(t, err)
	assert.Equal(t, "intro\n\n# Steps\n- [ ] one\n\nbefore\n\n# Notes\nn\n", got)

	got, err = EditBody(body, "insert_after_section", "after", "Notes", "", 0)
	require.NoError(t, err)
	assert.Equal(t, body+"after\n\n", got)

	_, err = EditBody(body, "replace_section", "x", "", "", 0)
	assert.EqualError(t, err, "replace_section requires a section heading")
	_, err = EditBody(body, "replace_section", "x", "Nope", "", 0)
	assert.EqualError(t, err, `section not found: "Nope"`)
	_, err = EditBody(body, "explode", "", "", "", 0)
	assert.EqualError(t, err, `unknown edit op: "explode"`)
}

func TestAppendNarrative(t *testing.T) {
	assert.Equal(t, "A\n\nB\n", AppendNarrative("A\n\n", []string{"B"}))
	assert.Equal(t, "B\n\n# Obs\n- x\n", AppendNarrative("# Obs\n- x\n", []string{"B"}))
	assert.Equal(t, "A\n\nB\n\n# H\n", AppendNarrative("A\n# H\n", []string{"B"}))
	assert.Equal(t, "B\n", AppendNarrative("", []string{"B"}))
}

func TestSetSection(t *testing.T) {
	assert.Equal(t, "p\n\n# Touched\n- a\n", SetSection("p\n", "Touched", []string{"- a"}))
	assert.Equal(t, "p\n\n# Touched\n- b\n\n# After\nx\n", SetSection("p\n\n# Touched\n- a\n\n# After\nx\n", "Touched", []string{"- b"}))
	assert.Equal(t, "p\n\n# After\nx\n", SetSection("p\n\n# Touched\n- a\n# After\nx\n", "Touched", nil))
	assert.Equal(t, "", SetSection("", "Touched", nil))
}

const reportNote = `---
type: task
title: Ship v2
description: Ship it.
sources:
  - id: "1"
    resource: https://example.com/pr/1
generated: {by: x, at: 2026-01-01T00:00:00Z}
state: open
---

Old plan [^1].

# Steps
- [ ] old

# Observations
- [state] open [^1]

# Relations
- part_of [API v2](/projects/api-v2.md)

[^1]: [https://example.com/pr/1](https://example.com/pr/1)
`

const reportDoc = "# Report\n\n| a | b |\n| - | - |\n| 1 | 2 |\n\n```sh\n# not a section\n- [fact] not an observation\n```\n\n- [WIP] prose bullet\n\nCites the PR [^1] and a note [^a].\n\n[^a]: My own footnote.\n"

func TestReplaceDocument(t *testing.T) {
	n, err := Parse([]byte(reportNote))
	require.NoError(t, err)
	require.NoError(t, n.ReplaceDocument("\n\n"+reportDoc+"\n\n"))
	assert.Equal(t, reportDoc, n.Document())
	assert.Equal(t, reportDoc+"\n# Observations\n- [state] open [^1]\n\n# Relations\n- part_of [API v2](/projects/api-v2.md)\n", n.Body)
	out := string(Render(n))
	assert.Contains(t, out, "description: Ship it.\n")
	assert.Contains(t, out, "state: open\n")
	assert.True(t, strings.HasSuffix(out, "\n\n[^1]: [https://example.com/pr/1](https://example.com/pr/1)\n"), out)
	assert.Equal(t, []string{"[state] open"}, summaries(Observations(n.Body)))

	// Round trip: replacing with the current document changes nothing.
	before := n.Body
	require.NoError(t, n.ReplaceDocument(n.Document()))
	assert.Equal(t, before, n.Body)
}

func TestReplaceDocumentAcceptsDerivedDefs(t *testing.T) {
	n, err := Parse([]byte(reportNote))
	require.NoError(t, err)
	require.NoError(t, n.ReplaceDocument("New [^1].\n\n[^1]: [https://example.com/pr/1](https://example.com/pr/1)\n"))
	assert.Equal(t, "New [^1].\n", n.Document())
}

func TestReplaceDocumentRefuses(t *testing.T) {
	n, err := Parse([]byte(reportNote))
	require.NoError(t, err)
	for content, msg := range map[string]string{
		"  \n":                             "replace_body requires --content",
		"Text\n\n## observations\n- [a] b": `replace_body content has a "observations" section`,
		"Text\n\n# Relations\n":            `replace_body content has a "Relations" section`,
		"Text [^1]\n\n[^1]: Mine.\n":       "defines footnote [^1], but [^1] is this note's source 1",
		"---\ntype: task\ntitle: Ship v2\n---\n\nBody\n": "starts with a frontmatter block",
	} {
		assert.ErrorContains(t, n.ReplaceDocument(content), msg, content)
	}
	assert.Equal(t, "Old plan [^1].\n\n# Steps\n- [ ] old\n", n.Document())
}

func TestReplaceDocumentWithoutSections(t *testing.T) {
	n := &Note{Body: "old\n"}
	require.NoError(t, n.ReplaceDocument("new"))
	assert.Equal(t, "new\n", n.Body)
}

func summaries(obs []Observation) []string {
	var out []string
	for _, o := range obs {
		out = append(out, o.Summary())
	}
	return out
}
