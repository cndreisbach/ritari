package bundle

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func conceptIDs(cs []*Concept) []string {
	var out []string
	for _, c := range cs {
		out = append(out, c.ID)
	}
	return out
}

func TestContext(t *testing.T) {
	b := openFiles(t, map[string]string{
		"a.md": "---\ntype: note\n---\n[b](/b.md) [c](/c.md) [x](/missing.md)\n",
		"b.md": "---\ntype: note\n---\n[d](/d.md) [a](/a.md)\n",
		"c.md": "---\ntype: note\n---\n",
		"d.md": "---\ntype: note\n---\n[e](/e.md)\n",
		"e.md": "---\ntype: note\n---\n",
	})
	a, _ := b.Get("a")
	assert.Equal(t, []string{"a", "b", "c"}, conceptIDs(b.Context(a, 1)))
	assert.Equal(t, []string{"a", "b", "c", "d"}, conceptIDs(b.Context(a, 2)))
	assert.Equal(t, []string{"a"}, conceptIDs(b.Context(a, 0)))
}

func TestRecent(t *testing.T) {
	root := writeFiles(t, map[string]string{
		"old.md": "---\ntype: note\n---\n",
		"new.md": "---\ntype: note\n---\n",
		"mid.md": "---\ntype: note\n---\n",
	})
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, n := range []string{"old.md", "mid.md", "new.md"} {
		mt := base.Add(time.Duration(i) * time.Hour)
		require.NoError(t, os.Chtimes(filepath.Join(root, n), mt, mt))
	}
	b, err := Open(root)
	require.NoError(t, err)
	assert.Equal(t, []string{"new", "mid", "old"}, conceptIDs(b.Recent(0)))
	assert.Equal(t, []string{"new", "mid"}, conceptIDs(b.Recent(2)))
}
