package bundle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMoveUpdatesSources(t *testing.T) {
	root := t.TempDir()
	env := testEnv()
	_, err := Write(root, env, WriteOpts{Title: "Sibling", Obs: []string{"a: b"}})
	require.NoError(t, err)
	// The moved note cites itself and a sibling by relative path.
	_, err = Write(root, env, WriteOpts{Title: "Plan", Obs: []string{"a: b"}, Sources: []string{"[Me](plan.md)", "[Sibling](sibling.md)"}})
	require.NoError(t, err)
	_, err = Write(root, env, WriteOpts{Title: "Cites", Obs: []string{"a: b"}, Sources: []string{"[Plan](../notes/plan.md)"}})
	require.NoError(t, err)

	// Another OKF tool's source: a non-numeric id, and no link in the body.
	external := `---
type: note
title: External
sources:
  - id: ext
    resource: /notes/plan.md
    author: Ana
---

Prose.
`
	require.NoError(t, os.WriteFile(filepath.Join(root, "notes", "external.md"), []byte(external), 0o644))

	res, err := Move(root, env, "notes/plan", "references/plan")
	require.NoError(t, err)
	assert.Equal(t, []string{"notes/cites", "notes/external"}, res.RewroteLinksIn)

	moved := readFile(t, root, "references/plan.md")
	assert.Contains(t, moved, "resource: /references/plan.md\n")
	assert.Contains(t, moved, "resource: /notes/sibling.md\n")
	assert.Contains(t, moved, "[^1]: [Me](/references/plan.md)\n")
	assert.Contains(t, moved, "[^2]: [Sibling](/notes/sibling.md)\n")

	cites := readFile(t, root, "notes/cites.md")
	assert.Contains(t, cites, "resource: /references/plan.md\n")
	assert.Contains(t, cites, "[^1]: [Plan](/references/plan.md)\n")

	got := readFile(t, root, "notes/external.md")
	assert.Contains(t, got, "resource: /references/plan.md\n")
	assert.Contains(t, got, "author: Ana\n", "other keys on the source are kept")
	assert.Contains(t, got, "Prose.\n")
}

func TestMoveLeavesOtherNotesAlone(t *testing.T) {
	root := t.TempDir()
	env := testEnv()
	_, err := Write(root, env, WriteOpts{Title: "Plan", Obs: []string{"a: b"}})
	require.NoError(t, err)
	_, err = Write(root, env, WriteOpts{Title: "Other", Obs: []string{"a: b"}, Sources: []string{"Plan B /notes/plan-b.md", "Plan by email"}})
	require.NoError(t, err)
	before := readFile(t, root, "notes/other.md")

	res, err := Move(root, env, "notes/plan", "references/plan")
	require.NoError(t, err)
	assert.Empty(t, res.RewroteLinksIn)
	assert.Equal(t, before, readFile(t, root, "notes/other.md"))
}
