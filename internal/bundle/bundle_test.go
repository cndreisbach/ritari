package bundle

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenSkipsReservedHiddenAndBroken(t *testing.T) {
	b := openFiles(t, map[string]string{
		"index.md":           "---\nokf_version: \"0.2\"\n---\n",
		"log.md":             "# Update Log\n",
		"decisions/index.md": "# Decisions\n",
		"decisions/a.md":     "---\ntype: decision\ntitle: Auth uses JWT\n---\n\nSee [b](/notes/b.md).\n",
		"notes/b.MD":         "---\ntype: note\n---\n",
		"notes/broken.md":    "---\ntype: [oops\n---\n",
		".git/x.md":          "---\ntype: note\n---\n",
		"loose.md":           "no frontmatter\n",
		"sources/README.txt": "raw\n",
	})
	var ids []string
	for _, c := range b.Concepts {
		ids = append(ids, c.ID)
	}
	assert.Equal(t, []string{"decisions/a", "loose", "notes/b"}, ids)
	assert.Equal(t, []string{"decisions/a", "loose", "notes/b", "notes/broken"}, b.Files)
	require.Len(t, b.Problems, 1)
	assert.Equal(t, "notes/broken.md", b.Problems[0].Path)
	assert.Contains(t, b.Problems[0].Err.Error(), "frontmatter could not be parsed")
}

func TestOpenBadVocabIsFatal(t *testing.T) {
	_, err := Open(writeFiles(t, map[string]string{"ritari.yaml": "types: [{name: a, dir: x}, {name: b, dir: x}]\n"}))
	assert.ErrorContains(t, err, `ritari.yaml: types "a" and "b" both use dir "x"`)
}

func TestResolve(t *testing.T) {
	b := openFiles(t, map[string]string{
		"decisions/a.md":  "---\ntype: decision\ntitle: Auth uses JWT\n---\n",
		"z/dup.md":        "---\ntitle: Same Title\n---\n",
		"m/dup.md":        "---\ntitle: same title!\n---\n",
		"notes/broken.md": "---\ntype: [oops\n---\n",
	})
	for _, ref := range []string{"decisions/a", "decisions/a.md", "Auth uses JWT", "auth-uses-jwt", filepath.Join(b.Root, "decisions", "a.md")} {
		c, err := b.Resolve(ref)
		require.NoError(t, err, ref)
		assert.Equal(t, "decisions/a", c.ID, ref)
	}
	c, err := b.Resolve("Same Title")
	require.NoError(t, err)
	assert.Equal(t, "m/dup", c.ID)

	_, err = b.Resolve("")
	assert.ErrorIs(t, err, ErrNoID)
	_, err = b.Resolve("nothing")
	assert.EqualError(t, err, "note not found: nothing")
	_, err = b.Resolve("notes/broken")
	assert.ErrorContains(t, err, "frontmatter could not be parsed")
}

func TestGraph(t *testing.T) {
	b := openFiles(t, map[string]string{
		"a.md":      "---\ntype: note\n---\n[b](/b.md) [b again](b.md) [missing](/nope.md) [broken](/broken.md) [ext](https://x)\n",
		"b.md":      "---\ntype: note\n---\n[a](./a.md)\n",
		"broken.md": "---\ntype: [\n---\n",
	})
	g := b.Graph()
	assert.Equal(t, []string{"b"}, g.Out["a"])
	assert.Equal(t, []string{"a"}, g.In["b"])
	assert.Equal(t, []string{"nope"}, g.Dangling["a"])
	assert.Equal(t, []string{"b"}, g.In["a"])
}
