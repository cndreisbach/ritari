package vocab

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlug(t *testing.T) {
	assert.Equal(t, "auth-uses-jwt", Slug("  Auth uses JWT! "))
	assert.Equal(t, "caf", Slug("Café"))
	assert.Equal(t, "", Slug("日本"))
	assert.Equal(t, "a-b", Slug("--a__b--"))
}

func TestPlural(t *testing.T) {
	for in, want := range map[string]string{
		"incident": "incidents", "class": "classes", "box": "boxes", "match": "matches",
		"wish": "wishes", "story": "stories", "day": "days", "": "", "y": "ys",
	} {
		assert.Equal(t, want, Plural(in), in)
	}
}

func TestHumanize(t *testing.T) {
	assert.Equal(t, "Decisions", Humanize("decisions"))
	assert.Equal(t, "Api Endpoints", Humanize("ref/api-endpoints"))
	assert.Equal(t, "Foo Bar", Humanize("foo_bar"))
}

func TestDefault(t *testing.T) {
	v := Default()
	assert.Equal(t, "people", v.DirFor("Person "))
	assert.Equal(t, "incidents", v.DirFor("incident"))
	assert.Equal(t, "decisions/auth-uses-jwt", v.IDFor("decision", "Auth uses JWT"))
	assert.Equal(t, "daily", v.DailyDir())
	assert.Equal(t, "notes", v.NoteDir())
	typ, ok := v.TypeForDir("people")
	require.True(t, ok)
	assert.Equal(t, "person", typ.Name)
}

func load(t *testing.T, content string) (*Vocab, error) {
	t.Helper()
	dir := t.TempDir()
	if content != "" {
		require.NoError(t, os.WriteFile(filepath.Join(dir, FileName), []byte(content), 0o644))
	}
	return Load(dir)
}

func TestLoadMissingIsDefault(t *testing.T) {
	v, err := load(t, "")
	require.NoError(t, err)
	assert.Equal(t, Default(), v)
	v, err = load(t, "types: []\n")
	require.NoError(t, err)
	assert.Equal(t, Default(), v)
}

func TestLoadReplacesAndAppendsReserved(t *testing.T) {
	v, err := load(t, "types:\n  - name: Incident\n    dir: incidents\n    summary: An outage.\n  - name: runbook\n")
	require.NoError(t, err)
	assert.Equal(t, []string{"incident", "runbook", "note", "daily"}, v.Names())
	assert.Equal(t, "runbooks", v.DirFor("runbook"))
	assert.Equal(t, "meetings", v.DirFor("meeting"))
	_, ok := v.Lookup("meeting")
	assert.False(t, ok)
}

func TestLoadErrors(t *testing.T) {
	for content, msg := range map[string]string{
		"types: [{dir: x}]\n":                             "type 1 has no name",
		"types: [{name: a}, {name: A}]\n":                 `type "a" is declared twice`,
		"types: [{name: a, dir: Bad}]\n":                  `dir "Bad" must be a single lowercase path segment`,
		"types: [{name: a, dir: x/y}]\n":                  `dir "x/y"`,
		"types: [{name: a, dir: x}, {name: b, dir: x}]\n": `types "a" and "b" both use dir "x"`,
		"types: [{name: a, dir: sources}]\n":              `dir "sources" is reserved`,
		"types: [{name: a, dir: notes}]\n":                `types "a" and "note" both use dir "notes"`,
		"types: [unclosed\n":                              "ritari.yaml: ",
	} {
		_, err := load(t, content)
		require.Error(t, err, content)
		assert.Contains(t, err.Error(), msg, content)
		assert.Contains(t, err.Error(), "ritari.yaml: ", content)
	}
}

func TestYAMLRoundTrip(t *testing.T) {
	v := Default()
	back, err := parse([]byte(v.YAML()))
	require.NoError(t, err)
	assert.Equal(t, v, back)
	assert.Contains(t, v.YAML(), "summary: \"A day's narrative: what came up, and what it touched.\"")
}
