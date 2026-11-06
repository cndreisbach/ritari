package note

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

func TestUpdateSource(t *testing.T) {
	str := func(s string) *string { return &s }
	extra := []*yaml.Node{{Kind: yaml.ScalarNode, Value: "author"}, {Kind: yaml.ScalarNode, Value: "Ana"}}
	fresh := func() []Source {
		return []Source{
			{ID: "1", Resource: "https://example.com/a", Title: "A", Extra: extra},
			{ID: "2", Resource: "Email from Ana"},
			{ID: "ext", Resource: "https://example.com/ext"},
		}
	}

	t.Run("resource keeps the id, title, and other keys", func(t *testing.T) {
		sources := fresh()
		changed, err := UpdateSource(sources, "n", "1", str(" /notes/a.md "), nil)
		require.NoError(t, err)
		assert.True(t, changed)
		assert.Equal(t, Source{ID: "1", Resource: "/notes/a.md", Title: "A", Extra: extra}, sources[0])
		assert.Equal(t, fresh()[1:], sources[1:], "other sources are untouched")
	})

	t.Run("title alone, and clearing it", func(t *testing.T) {
		sources := fresh()
		_, err := UpdateSource(sources, "n", "1", nil, str("B"))
		require.NoError(t, err)
		assert.Equal(t, "B", sources[0].Title)
		assert.Equal(t, "https://example.com/a", sources[0].Resource)
		_, err = UpdateSource(sources, "n", "1", nil, str(""))
		require.NoError(t, err)
		assert.Empty(t, sources[0].Title)
	})

	t.Run("the same values change nothing", func(t *testing.T) {
		sources := fresh()
		changed, err := UpdateSource(sources, "n", "1", str("https://example.com/a"), str("A"))
		require.NoError(t, err)
		assert.False(t, changed)
	})

	t.Run("a descriptor renders without a link", func(t *testing.T) {
		sources := fresh()
		_, err := UpdateSource(sources, "n", "1", str("Temporary file, since removed"), str(""))
		require.NoError(t, err)
		assert.Equal(t, "[^1]: Temporary file, since removed", FootnoteDef(sources[0]))
	})

	for _, tc := range []struct {
		name            string
		id              string
		resource, title *string
		want            string
	}{
		{"unknown id", "9", nil, str("x"), `note "n" has no source "9"; its source ids are 1, 2, ext`},
		{"non-numeric id", "ext", nil, str("x"), `source "ext" has a non-numeric id`},
		{"blank resource", "1", str("  "), nil, "a source needs a resource"},
		{"duplicate of a followable source", "2", str("https://example.com/a"), nil, `source "2" would duplicate source "1"`},
		{"duplicate of a descriptor", "1", str("EMAIL FROM ANA"), nil, `source "1" would duplicate source "2"`},
		{"multi-line resource", "1", str("two\nlines"), nil, "must each be one line"},
		{"multi-line title", "1", nil, str("two\nlines"), "must each be one line"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sources := fresh()
			changed, err := UpdateSource(sources, "n", tc.id, tc.resource, tc.title)
			assert.ErrorContains(t, err, tc.want)
			assert.False(t, changed)
			assert.Equal(t, fresh(), sources, "a refused update changes nothing")
		})
	}

	t.Run("a note with no sources", func(t *testing.T) {
		_, err := UpdateSource(nil, "n", "1", nil, str("x"))
		assert.EqualError(t, err, `note "n" has no sources`)
	})
}
