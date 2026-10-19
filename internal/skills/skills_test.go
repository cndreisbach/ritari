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
	assert.Equal(t, []string{"ritari", "ritari-daily", "ritari-ingest", "ritari-lint", "ritari-notes", "ritari-reflect", "ritari-tasks"}, Names())
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
	}
}

func TestInstall(t *testing.T) {
	dir := t.TempDir()
	got, err := Install(dir, false)
	require.NoError(t, err)
	assert.Equal(t, Names(), got)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ritari", "SKILL.md"), []byte("mine"), 0o644))
	got, err = Install(dir, false)
	require.NoError(t, err)
	assert.Empty(t, got)
	data, _ := os.ReadFile(filepath.Join(dir, "ritari", "SKILL.md"))
	assert.Equal(t, "mine", string(data))
	got, err = Install(dir, true)
	require.NoError(t, err)
	assert.Len(t, got, 7)
}

func TestRetired(t *testing.T) {
	dir := t.TempDir()
	got, err := Retired(filepath.Join(dir, "missing"))
	require.NoError(t, err)
	assert.Empty(t, got)

	_, err = Install(dir, false)
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
	assert.DirExists(t, filepath.Join(dir, "ritari-notes"))
	got, err = Retired(dir)
	require.NoError(t, err)
	assert.Empty(t, got)
}
