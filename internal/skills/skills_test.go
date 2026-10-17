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
