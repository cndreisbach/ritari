package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

func TestSkillsHaveFrontmatter(t *testing.T) {
	assert.Equal(t, []string{"ritari", "ritari-ingest", "ritari-reflect"}, Names())
	for _, n := range Names() {
		data, err := Content(n)
		require.NoError(t, err)
		parts := strings.SplitN(string(data), "---\n", 3)
		require.Len(t, parts, 3, n)
		var fm struct{ Name, Description string }
		require.NoError(t, yaml.Unmarshal([]byte(parts[1]), &fm), n)
		assert.Equal(t, n, fm.Name)
		assert.NotEmpty(t, fm.Description, n)
		assert.NotContains(t, fm.Description, "\n", n)
		assert.LessOrEqual(t, len(fm.Description), 1024, n)
	}
}

func TestReferencesAreNamedInTheirSkill(t *testing.T) {
	got, err := Files("ritari")
	require.NoError(t, err)
	assert.Equal(t, []string{"SKILL.md", "references/daily.md", "references/lint.md", "references/tasks.md", "references/writing.md"}, got)
	skill, _ := Content("ritari")
	for _, f := range got[1:] {
		assert.Contains(t, string(skill), "`"+f+"`", "SKILL.md must say when to read %s", f)
	}
}

func TestInstall(t *testing.T) {
	dir := t.TempDir()
	installed, outdated, err := Install(dir, false)
	require.NoError(t, err)
	assert.Equal(t, Names(), installed)
	assert.Empty(t, outdated)
	assert.FileExists(t, filepath.Join(dir, "ritari", "references", "tasks.md"))

	installed, outdated, err = Install(dir, false)
	require.NoError(t, err)
	assert.Empty(t, installed)
	assert.Empty(t, outdated)

	// An edited file is kept and reported; a missing one is restored anyway.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ritari", "SKILL.md"), []byte("mine"), 0o644))
	require.NoError(t, os.Remove(filepath.Join(dir, "ritari", "references", "lint.md")))
	installed, outdated, err = Install(dir, false)
	require.NoError(t, err)
	assert.Equal(t, []string{"ritari"}, installed)
	assert.Equal(t, []string{"ritari"}, outdated)
	data, _ := os.ReadFile(filepath.Join(dir, "ritari", "SKILL.md"))
	assert.Equal(t, "mine", string(data))
	assert.FileExists(t, filepath.Join(dir, "ritari", "references", "lint.md"))

	installed, outdated, err = Install(dir, true)
	require.NoError(t, err)
	assert.Equal(t, []string{"ritari"}, installed)
	assert.Empty(t, outdated)
	data, _ = os.ReadFile(filepath.Join(dir, "ritari", "SKILL.md"))
	want, _ := Content("ritari")
	assert.Equal(t, string(want), string(data))
}

func TestRetired(t *testing.T) {
	dir := t.TempDir()
	got, err := Retired(filepath.Join(dir, "missing"))
	require.NoError(t, err)
	assert.Empty(t, got)

	_, _, err = Install(dir, false)
	require.NoError(t, err)
	for _, name := range []string{"ritari-old", "ritari-gone", "other-skill"} {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, name, "references"), 0o755))
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ritari-file"), []byte("not a skill"), 0o644))
	got, err = Retired(dir)
	require.NoError(t, err)
	assert.Equal(t, []string{"ritari-gone", "ritari-old"}, got)

	require.NoError(t, Remove(dir, got))
	assert.NoDirExists(t, filepath.Join(dir, "ritari-old"))
	assert.DirExists(t, filepath.Join(dir, "other-skill"))
	assert.DirExists(t, filepath.Join(dir, "ritari-ingest"))
	got, err = Retired(dir)
	require.NoError(t, err)
	assert.Empty(t, got)
}
