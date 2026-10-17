package search

import (
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cndreisbach/ritari/internal/bundle"
)

func open(t *testing.T, files map[string]string) *bundle.Bundle {
	t.Helper()
	root := t.TempDir()
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	i := 0
	for _, rel := range slices.Sorted(maps.Keys(files)) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		require.NoError(t, os.WriteFile(p, []byte(files[rel]), 0o644))
		mt := base.Add(time.Duration(i) * time.Hour)
		require.NoError(t, os.Chtimes(p, mt, mt))
		i++
	}
	b, err := bundle.Open(root)
	require.NoError(t, err)
	return b
}

func ids(hits []Hit) []string {
	out := []string{}
	for _, h := range hits {
		out = append(out, h.Concept.ID)
	}
	return out
}

var corpus = map[string]string{
	"a/body.md":   "---\ntype: note\ntitle: Something\n---\n\nWe use jwt tokens here.\n",
	"b/title.md":  "---\ntype: decision\ntitle: JWT choice\ntags: [auth]\nstate: active\n---\n\n# Observations\n- [decision] Use JWT #security (audit)\n",
	"c/daily.md":  "---\ntype: daily\ntitle: 2026-07-29\n---\n\njwt jwt jwt\n",
	"d/old.md":    "---\ntype: note\ntitle: Old jwt\narchived: true\n---\n",
	"e/to.md":     "---\ntype: note\ntitle: Go to market\nstate: done\nowners: [Ann, bob]\n---\n",
	"f/broken.md": "---\ntype: [\n---\njwt\n",
	"g/rel.md":    "---\ntype: note\ntitle: Rel\nverified: [{by: human:x, at: 2026-01-01T00:00:00Z}]\n---\n\n- depends_on [B](/b/title.md)\n",
}

func TestTokenize(t *testing.T) {
	assert.Equal(t, []string{"auth", "uses", "jwt", "v2", "caf"}, Tokenize("Auth-uses JWT_v2 café"))
}

func TestRanking(t *testing.T) {
	b := open(t, corpus)
	hits, err := Run(b, Options{Query: "jwt"})
	require.NoError(t, err)
	assert.Equal(t, []string{"b/title", "a/body"}, ids(hits))
	assert.Equal(t, 10+3+1, hits[0].Score)
	assert.Equal(t, []string{"[decision] Use JWT"}, hits[0].Matches)
	assert.Equal(t, 1, hits[1].Score)

	hits, err = Run(b, Options{Query: "jwt", IncludeDaily: true, IncludeArchived: true})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"b/title", "a/body", "c/daily", "d/old"}, ids(hits))

	hits, err = Run(b, Options{Query: "nonexistent"})
	require.NoError(t, err)
	assert.Empty(t, hits)

	hits, err = Run(b, Options{Query: "jwt use"})
	require.NoError(t, err)
	assert.Equal(t, []string{"b/title", "a/body"}, ids(hits), "every term must match; use matches by prefix")
}

func TestPrefixRule(t *testing.T) {
	b := open(t, corpus)
	hits, err := Run(b, Options{Query: "tok"})
	require.NoError(t, err)
	assert.Equal(t, []string{"a/body"}, ids(hits))
	hits, err = Run(b, Options{Query: "to"})
	require.NoError(t, err)
	assert.Equal(t, []string{"e/to"}, ids(hits))
}

func TestFilters(t *testing.T) {
	b := open(t, corpus)
	run := func(o Options) []string {
		hits, err := Run(b, o)
		require.NoError(t, err)
		return ids(hits)
	}
	assert.Equal(t, []string{"b/title"}, run(Options{Type: "Decision"}))
	assert.Equal(t, []string{"b/title"}, run(Options{Category: "decision"}))
	assert.Equal(t, []string{"b/title"}, run(Options{Tags: []string{"AUTH"}}))
	assert.Equal(t, []string{"g/rel"}, run(Options{Relation: "depends_on"}))
	assert.Equal(t, []string{"e/to", "b/title"}, run(Options{Meta: []string{"state=active|done"}}), "no terms: newest first")
	assert.Equal(t, []string{"e/to"}, run(Options{Meta: []string{"owners=BOB"}}))
	assert.Equal(t, []string{"g/rel"}, run(Options{Meta: []string{"trust=human-reviewed"}}))
	assert.Equal(t, []string{"e/to"}, run(Options{Meta: []string{"state=done", "type=note"}}))
	assert.Equal(t, []string{"g/rel", "e/to"}, run(Options{Limit: 2}))
}

func TestMetaErrors(t *testing.T) {
	_, err := ParseMeta([]string{"state=active", "state=done"})
	assert.EqualError(t, err, "--meta state given twice (active and done); use state=a|b for alternatives")
	_, err = ParseMeta([]string{"nokey"})
	assert.EqualError(t, err, `--meta "nokey" is not key=value`)
	_, err = ParseMeta([]string{"=x"})
	assert.Error(t, err)
	_, err = ParseMeta([]string{"k= | "})
	assert.EqualError(t, err, `--meta "k= | " has no value`)
}
