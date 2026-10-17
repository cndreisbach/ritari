package registry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestPath(t *testing.T) {
	p, err := Path(env(map[string]string{"RITARI_CONFIG": "/x/c.yaml", "XDG_CONFIG_HOME": "/xdg"}))
	require.NoError(t, err)
	assert.Equal(t, "/x/c.yaml", p)
	p, err = Path(env(map[string]string{"XDG_CONFIG_HOME": "/xdg"}))
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("/xdg", "ritari", "config.yaml"), p)
	p, err = Path(env(nil))
	require.NoError(t, err)
	assert.True(t, filepath.IsAbs(p))
	assert.True(t, strings.HasSuffix(p, filepath.Join(".config", "ritari", "config.yaml")), p)
}

func TestLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "config.yaml")
	r, err := Load(path)
	require.NoError(t, err)
	assert.Empty(t, r.Projects)

	_, err = r.Add("", "/x", false)
	assert.EqualError(t, err, "project name is required")
	work, err := r.Add("work", "/w", false)
	require.NoError(t, err)
	assert.Equal(t, "work", r.Default, "first project becomes the default")
	_, err = r.Add("work", "/w2", false)
	assert.EqualError(t, err, `project "work" already exists`)
	_, err = r.Add("home", "/h", true)
	require.NoError(t, err)
	assert.Equal(t, "home", r.Default)
	r.Author = "human:me@example.com"
	require.NoError(t, r.Save())

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "default: home\nprojects:\n  - name: work\n    path: "+work.Path+"\n  - name: home\n    path: "+filepath.Join(filepath.VolumeName(work.Path)+string(filepath.Separator), "h")+"\nauthor: human:me@example.com\n", string(data))

	back, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, r.Projects, back.Projects)
	assert.Equal(t, "home", back.Default)

	p, err := back.Active("", env(nil))
	require.NoError(t, err)
	assert.Equal(t, "home", p.Name)
	p, err = back.Active("", env(map[string]string{"RITARI_PROJECT": "work"}))
	require.NoError(t, err)
	assert.Equal(t, "work", p.Name)
	p, err = back.Active("home", env(map[string]string{"RITARI_PROJECT": "work"}))
	require.NoError(t, err)
	assert.Equal(t, "home", p.Name)
	_, err = back.Active("nope", env(nil))
	assert.EqualError(t, err, `project "nope" not found`)

	require.NoError(t, back.Remove("home"))
	assert.Equal(t, "work", back.Default)
	assert.EqualError(t, back.Remove("home"), `project "home" not found`)
	assert.EqualError(t, back.SetDefault("home"), `project "home" not found`)
	require.NoError(t, back.Remove("work"))
	assert.Equal(t, "", back.Default)
	_, err = back.Active("", env(nil))
	assert.EqualError(t, err, "no project specified and no default set (try: rit project init <name> <path>)")
}

func TestLoadUnparseable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("projects: [oops\n"), 0o644))
	_, err := Load(path)
	assert.ErrorContains(t, err, "parsing "+path+": ")
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	p, err := ExpandPath("~/notes")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, "notes"), p)
}
