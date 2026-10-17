package note

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const example = `---
type: decision
title: Auth uses JWT
description: Stateless auth via short-lived JWTs, chosen over server sessions.
resource: https://github.com/acme/api/pull/1234
tags: [auth, security]
sources:
  - id: "1"
    resource: https://github.com/acme/api/pull/1234
    title: "PR #1234 — switch to JWT"
generated: {by: human:jane@acme.com, at: 2026-07-29T14:30:00Z}
verified:
  - {by: human:jane@acme.com, at: 2026-07-30T09:00:00Z}
state: shipped
---

Server-side sessions would have meant sticky routing, which we did not want.

# Observations
- [decision] Use JWT for stateless auth #security [^1]
- [constraint] Tokens expire after 15m (per security audit) [^1]

# Relations
- depends_on [User Database Schema](/references/user-database-schema.md)

[^1]: [PR #1234 — switch to JWT](https://github.com/acme/api/pull/1234)
`

func TestParseExample(t *testing.T) {
	n, err := Parse([]byte(example))
	require.NoError(t, err)
	fm := n.FM
	assert.Equal(t, "decision", fm.Type)
	assert.Equal(t, "Auth uses JWT", fm.Title)
	assert.Equal(t, "https://github.com/acme/api/pull/1234", fm.Resource)
	assert.Equal(t, []string{"auth", "security"}, fm.Tags)
	require.Len(t, fm.Sources, 1)
	assert.Equal(t, Source{ID: "1", Resource: "https://github.com/acme/api/pull/1234", Title: "PR #1234 — switch to JWT"}, fm.Sources[0])
	assert.Equal(t, &Stamp{By: "human:jane@acme.com", At: "2026-07-29T14:30:00Z"}, fm.Generated)
	assert.Equal(t, []Stamp{{By: "human:jane@acme.com", At: "2026-07-30T09:00:00Z"}}, fm.Verified)
	require.Len(t, fm.Extra, 1)
	assert.Equal(t, "state", fm.Extra[0].Key)
	assert.Equal(t, "state: shipped", fm.Extra[0].Raw)
	assert.True(t, len(n.Body) > 0 && n.Body[0] == 'S', "separator blank line is not part of the body")
}

func TestRenderRoundTrip(t *testing.T) {
	n, err := Parse([]byte(example))
	require.NoError(t, err)
	assert.Equal(t, example, string(Render(n)))
}

func TestParseNoFrontmatter(t *testing.T) {
	n, err := Parse([]byte("# Hello\n\nworld\n"))
	require.NoError(t, err)
	assert.Equal(t, "", n.FM.Type)
	assert.Equal(t, "# Hello\n\nworld\n", n.Body)
}

func TestParseUnclosedFrontmatterIsBody(t *testing.T) {
	src := "---\ntype: x\nno close\n"
	n, err := Parse([]byte(src))
	require.NoError(t, err)
	assert.Equal(t, src, n.Body)
	assert.Equal(t, "", n.FM.Type)
}

func TestParseTolerances(t *testing.T) {
	src := "  ---\r\ntype: note\r\ntags: auth\r\nverified: {by: human:a, at: 2026-01-01T00:00:00Z}\r\narchived: Yes\r\n---   \r\nbody\r\n"
	n, err := Parse([]byte(src))
	require.NoError(t, err)
	assert.Equal(t, []string{"auth"}, n.FM.Tags)
	assert.Equal(t, []Stamp{{By: "human:a", At: "2026-01-01T00:00:00Z"}}, n.FM.Verified)
	assert.True(t, n.FM.Archived)
	assert.Equal(t, "body\n", n.Body)
}

func TestParseUnparseable(t *testing.T) {
	for name, src := range map[string]string{
		"bad yaml":       "---\ntype: [unclosed\n---\n",
		"not a mapping":  "---\n- a\n- b\n---\n",
		"title mapping":  "---\ntitle: {a: b}\n---\n",
		"sources scalar": "---\nsources: foo\n---\n",
		"sources items":  "---\nsources: [a, b]\n---\n",
		"duplicate key":  "---\ntype: a\ntype: b\n---\n",
	} {
		_, err := Parse([]byte(src))
		assert.Error(t, err, name)
	}
}

func TestExtraKeysPreservedVerbatim(t *testing.T) {
	src := "---\ntype: note\nzeta: 1\nalpha:\n  nested: [1, 2]   # comment\ntimestamp: '2026-05-28T22:53:05+00:00'\n---\n"
	n, err := Parse([]byte(src))
	require.NoError(t, err)
	assert.Equal(t, "2026-05-28T22:53:05+00:00", n.FM.ContentTime())
	assert.Equal(t, "---\ntype: note\nalpha:\n  nested: [1, 2]   # comment\ntimestamp: '2026-05-28T22:53:05+00:00'\nzeta: 1\n---\n", string(Render(n)))
}

func TestRenderOrderAndQuoting(t *testing.T) {
	n := &Note{FM: Frontmatter{
		Title:     "true",
		Tags:      []string{"b", "", "a", "b"},
		Generated: &Stamp{By: "claude-code", At: "2026-07-29T14:30:00Z"},
		Archived:  true,
	}}
	n.FM.SetExtra("priority", "123")
	n.FM.SetExtra("beta", "x")
	assert.Equal(t, `---
type: note
title: "true"
tags: [b, a]
generated: {by: claude-code, at: 2026-07-29T14:30:00Z}
archived: true
beta: x
priority: "123"
---
`, string(Render(n)))

	back, err := Parse(Render(n))
	require.NoError(t, err)
	assert.Equal(t, "123", back.FM.ExtraValue("priority").Value)
	assert.Equal(t, "true", back.FM.Title)
}

func TestSourceExtraKeysPreserved(t *testing.T) {
	src := `---
type: Metric
sources:
  - id: ga4-schema
    resource: https://developers.google.com/analytics
    title: GA4 BigQuery Export schema
    author: team:ga4-docs
    usage_count: 5000
---

The table is sharded.[^ga4-schema]

[^ga4-schema]: GA4 BigQuery Export schema
`
	n, err := Parse([]byte(src))
	require.NoError(t, err)
	require.Len(t, n.FM.Sources, 1)
	assert.Equal(t, "ga4-schema", n.FM.Sources[0].ID)
	out := string(Render(n))
	assert.Contains(t, out, "    author: team:ga4-docs\n    usage_count: 5000\n")
	assert.Contains(t, out, "[^ga4-schema]: GA4 BigQuery Export schema\n", "non-numeric footnotes are not managed")
	assert.Contains(t, out, `id: "ga4-schema"`)
}

func TestNumericSourceIDReadAsString(t *testing.T) {
	n, err := Parse([]byte("---\ntype: note\nsources:\n  - id: 2\n    resource: x\n---\n"))
	require.NoError(t, err)
	assert.Equal(t, "2", n.FM.Sources[0].ID)
}
