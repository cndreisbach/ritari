package note

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseObservation(t *testing.T) {
	o, ok := ParseObservation("  - [decision] Use JWT for   stateless auth #security #auth/jwt (per audit) [^1] [^2]")
	require.True(t, ok)
	assert.Equal(t, "decision", o.Category)
	assert.Equal(t, "Use JWT for stateless auth", o.Content)
	assert.Equal(t, []string{"security", "auth/jwt"}, o.Tags)
	assert.Equal(t, "per audit", o.Context)
	assert.Equal(t, []string{"1", "2"}, o.Refs)
	assert.Equal(t, "[decision] Use JWT for stateless auth", o.Summary())
	assert.Equal(t, "[decision] Use JWT for   stateless auth #security #auth/jwt (per audit) [^1] [^2]", o.Line())
}

func TestParseObservationRefsBeforeContext(t *testing.T) {
	o, ok := ParseObservation("* [fact] text [^3] (ctx)")
	require.True(t, ok)
	assert.Equal(t, "text", o.Content)
	assert.Equal(t, "ctx", o.Context)
	assert.Equal(t, []string{"3"}, o.Refs)
}

func TestParseObservationRejects(t *testing.T) {
	for _, l := range []string{
		"- [ ] todo", "- [x] done", "- [X] done", "- [] x",
		"- depends_on [A](/a.md)", "- [A](/a.md)", "plain [cat] text", "- [cat]text",
	} {
		_, ok := ParseObservation(l)
		assert.False(t, ok, l)
	}
}

func TestObservationsOnlyInObservationsSection(t *testing.T) {
	body := "- [WIP] Draft the intro\n\n# Plan\n* [Optional] Add a step\n\n## observations\n- [fact] one\n### Risks\n- [risk] two\n## Next\n- [Q3] not one\n# Observations\n- [fact] three\n"
	var got []string
	for _, o := range Observations(body) {
		got = append(got, o.Summary())
	}
	assert.Equal(t, []string{"[fact] one", "[risk] two", "[fact] three"}, got)
}

func TestMergeIgnoresProseBullets(t *testing.T) {
	n := &Note{Body: "- [fact] one\n"}
	res, err := n.Merge(Additions{Obs: []string{"fact: one"}})
	require.NoError(t, err)
	assert.Equal(t, 1, res.AddedObs)
	assert.Equal(t, "- [fact] one\n\n# Observations\n- [fact] one\n", n.Body)
}

func TestTagsExcludeNumbersAndHex(t *testing.T) {
	o, ok := ParseObservation("- [bug] Fixed #42 and color #ff00aa #3e8 #auth #cafe #fff")
	require.True(t, ok)
	assert.Equal(t, []string{"auth", "cafe", "fff"}, o.Tags)
	assert.Equal(t, "Fixed #42 and color #ff00aa #3e8", o.Content)
}

func TestSections(t *testing.T) {
	body := "intro\n# Observations\n- [a] b\n## Sub\nx\n# Relations\n- r\n"
	ls := lines(body)
	s, e, ok := findSection(ls, "observations")
	require.True(t, ok)
	assert.Equal(t, 1, s)
	assert.Equal(t, 5, e)
	s, e, ok = findSection(ls, "## sub")
	require.True(t, ok)
	assert.Equal(t, 3, s)
	assert.Equal(t, 5, e)
	_, _, ok = findSection(ls, "missing")
	assert.False(t, ok)
	assert.Equal(t, "intro\n", Narrative(body))
	assert.Equal(t, "Closed", headingText("Closed ##"))
	assert.Equal(t, "C#", headingText("C#"))
}

func TestLinks(t *testing.T) {
	body := "See [a](/decisions/a.md) and [b](../people/b.md) and [ext](https://x.y) and ![img](/i.png) and <https://auto.link>.\n" +
		"Again [a2](/decisions/a.md). [ref][r]\n\n[r]: /notes/r.md\n\n```\n[code](/code.md)\n```\n\n    [indented](/ind.md)\n"
	var hrefs []string
	for _, l := range Links(body) {
		hrefs = append(hrefs, l.Href)
	}
	assert.Equal(t, []string{"/decisions/a.md", "../people/b.md", "https://x.y", "/notes/r.md"}, hrefs)
	assert.Equal(t, []string{"decisions/a", "people/b", "notes/r"}, Targets(body, "projects/p"))
}

func TestResolveLink(t *testing.T) {
	cases := []struct {
		href, from, want string
		ok               bool
	}{
		{"/decisions/a.md", "x/y", "decisions/a", true},
		{"b.md", "notes/a", "notes/b", true},
		{"./b.md#sec", "notes/a", "notes/b", true},
		{"../../out.md", "notes/a", "", false},
		{"../top.md", "notes/a", "top", true},
		{"https://x", "a", "", false},
		{"mailto:me@x", "a", "", false},
		{"//cdn/x", "a", "", false},
		{"#frag", "a", "", false},
		{"/sources/call.txt", "a", "", false},
		{"/notes/A.MD", "a", "notes/A", true},
	}
	for _, c := range cases {
		got, ok := ResolveLink(c.href, c.from)
		assert.Equal(t, c.ok, ok, c.href)
		assert.Equal(t, c.want, got, c.href)
	}
}

func TestRelations(t *testing.T) {
	body := "# Relations\n- depends_on [User DB](/references/user-db.md)\n- [Plain](/notes/plain.md)\n" +
		"- two words [X](/x.md)\n- has [A](/a.md) trailing\n- [A](/a.md) [B](/b.md)\n" +
		"```\n- fake [F](/f.md)\n```\n"
	rels := Relations(body, "notes/n")
	require.Len(t, rels, 2)
	assert.Equal(t, Relation{Type: "depends_on", Text: "User DB", Href: "/references/user-db.md", Target: "references/user-db"}, rels[0])
	assert.Equal(t, "links_to", rels[1].Type)
	assert.Equal(t, "notes/plain", rels[1].Target)
}

func TestParseSource(t *testing.T) {
	cases := map[string]Source{
		"[PR #1](https://x/1)": {Resource: "https://x/1", Title: "PR #1"},
		"<https://x/2>":        {Resource: "https://x/2"},
		"https://x/3":          {Resource: "https://x/3"},
		"PR #1234 https://github.com/acme/api/pull/1234": {Resource: "https://github.com/acme/api/pull/1234", Title: "PR #1234"},
		"Call transcript /sources/call.txt":              {Resource: "/sources/call.txt", Title: "Call transcript"},
		"Email from PJ, 2026-07-29":                      {Resource: "Email from PJ, 2026-07-29"},
		"mailto:a@b.c":                                   {Resource: "mailto:a@b.c"},
	}
	for in, want := range cases {
		got, ok := ParseSource(in)
		require.True(t, ok, in)
		assert.Equal(t, want, got, in)
	}
	_, ok := ParseSource("   ")
	assert.False(t, ok)
}

func TestFootnoteDef(t *testing.T) {
	assert.Equal(t, "[^1]: [PR \\[x\\]](https://x)", FootnoteDef(Source{ID: "1", Resource: "https://x", Title: "PR [x]"}))
	assert.Equal(t, "[^2]: [https://x](https://x)", FootnoteDef(Source{ID: "2", Resource: "https://x"}))
	assert.Equal(t, "[^3]: Standup (Slack thread, 2026-07-29)", FootnoteDef(Source{ID: "3", Resource: "Slack thread, 2026-07-29", Title: "Standup"}))
	assert.Equal(t, "[^4]: Email from PJ", FootnoteDef(Source{ID: "4", Resource: "Email from PJ"}))
}

func TestAddSources(t *testing.T) {
	existing := []Source{{ID: "1", Resource: "https://a"}, {ID: "ga4", Resource: "https://g"}, {ID: "2", Resource: "Email From PJ"}}
	body := "text [^5] here\n"
	out, ids, added := AddSources(existing, body, []Source{
		{Resource: "https://a"}, {Resource: "email from pj"}, {Resource: "https://new"}, {Resource: "https://new"},
	})
	assert.Equal(t, []string{"1", "2", "6"}, ids)
	assert.Equal(t, 1, added)
	assert.Len(t, out, 4)
	assert.Equal(t, Source{ID: "6", Resource: "https://new"}, out[3])
}

func TestFileLinks(t *testing.T) {
	for _, href := range []string{"/ideas/index.md", "../ideas/index.md", "/log.md", "index.md#top"} {
		_, ok := ResolveLink(href, "projects/roadmap")
		assert.False(t, ok, "%s is not a concept link", href)
	}
	p, ok := FileLink("../ideas/index.md#x", "projects/roadmap")
	require.True(t, ok)
	assert.Equal(t, "ideas/index.md", p)
	p, ok = FileLink("log.md", "a")
	require.True(t, ok)
	assert.Equal(t, "log.md", p)
	_, ok = FileLink("/ideas/mcp-server.md", "a")
	assert.False(t, ok)
	_, ok = FileLink("https://x/index.md", "a")
	assert.False(t, ok)
	for href, want := range map[string]string{
		"../img/diagram.PNG": "img/diagram.PNG",
		"/sources/call.txt":  "sources/call.txt",
		"/sources/notes.md":  "sources/notes.md",
		"/sources":           "sources",
	} {
		p, ok = FileLink(href, "projects/roadmap")
		require.True(t, ok, href)
		assert.Equal(t, want, p, href)
		_, ok = ResolveLink(href, "projects/roadmap")
		assert.False(t, ok, "%s is not a concept link", href)
	}
	for _, href := range []string{"/ideas/mcp-server.MD", "../decisions/auth-uses-jwt", "other"} {
		_, ok = FileLink(href, "projects/roadmap")
		assert.False(t, ok, "%s names a concept", href)
	}
	id, ok := ResolveLink("/ideas/Index.md", "a")
	require.True(t, ok, "a case variant is an ordinary concept")
	assert.Equal(t, "ideas/Index", id)
	_, ok = FileLink("../../index.md", "a/b")
	assert.False(t, ok, "escapes the bundle")

	assert.Equal(t, []string{"ideas/mcp-server"}, Targets("[i](/ideas/index.md) [m](/ideas/mcp-server.md) [l](../log.md)", "projects/roadmap"))

	rels := Relations("- see [Ideas](/ideas/index.md)\n", "projects/roadmap")
	require.Len(t, rels, 1)
	assert.Equal(t, "", rels[0].Target, "a reserved file is not a relation target")
}

func TestBundlePath(t *testing.T) {
	cases := []struct {
		href, from, want string
		ok               bool
	}{
		{"../sources/call.txt", "notes/a", "sources/call.txt", true},
		{"/decisions/a.md#why", "x", "decisions/a.md", true},
		{"other", "notes/a", "notes/other", true},
		{"../../x.md", "notes/a", "", false},
		{"https://x", "a", "", false},
		{"#frag", "a", "", false},
	}
	for _, c := range cases {
		got, ok := BundlePath(c.href, c.from)
		assert.Equal(t, c.ok, ok, c.href)
		assert.Equal(t, c.want, got, c.href)
	}
	assert.True(t, IsReservedName("index.md"))
	assert.True(t, IsReservedName("log.md"))
	assert.False(t, IsReservedName("Index.md"), "reserved names match exactly")
	assert.False(t, IsReservedName("LOG.MD"))
	assert.False(t, IsReservedName("indexes.md"))
}
