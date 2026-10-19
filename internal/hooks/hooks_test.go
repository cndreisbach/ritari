package hooks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandQuotesOddProjectNames(t *testing.T) {
	assert.Equal(t, "rit --project work --output text brief", Command("work"))
	assert.Equal(t, `rit --project 'my notes' --output text brief`, Command("my notes"))
	assert.Equal(t, `rit --project 'it'\''s' --output text brief`, Command("it's"))
}

func TestObjectRoundTripKeepsOrderAndValues(t *testing.T) {
	in := "{\n  \"z\": 1.50,\n  \"a\": [true, null, \"<b>&\"],\n  \"m\": {}\n}\n"
	o, err := parseObject([]byte(in))
	require.NoError(t, err)
	assert.Equal(t, "{\n  \"z\": 1.50,\n  \"a\": [\n    true,\n    null,\n    \"<b>&\"\n  ],\n  \"m\": {}\n}\n", string(encode(o)))

	_, err = parseObject([]byte("[1]"))
	assert.EqualError(t, err, "not a JSON object")
	_, err = parseObject([]byte("{} {}"))
	assert.Error(t, err)
	o, err = parseObject([]byte("  \n"))
	require.NoError(t, err)
	assert.Equal(t, "{}\n", string(encode(o)))
}

func TestClaudeCreatesThenLeavesAlone(t *testing.T) {
	root := t.TempDir()
	res, err := Claude(root, "work", false)
	require.NoError(t, err)
	assert.Equal(t, "created", res.Status)
	assert.Equal(t, filepath.Join(root, ".claude", "settings.json"), res.Path)
	data, err := os.ReadFile(res.Path)
	require.NoError(t, err)
	assert.Equal(t, `{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "rit --project work --output text brief"
          }
        ]
      }
    ]
  }
}
`, string(data))

	res, err = Claude(root, "work", false)
	require.NoError(t, err)
	assert.Equal(t, "unchanged", res.Status)

	res, err = Claude(root, "other", false)
	require.NoError(t, err)
	assert.Equal(t, "updated", res.Status)
	data, _ = os.ReadFile(res.Path)
	assert.Contains(t, string(data), `"rit --project other --output text brief"`)
	assert.NotContains(t, string(data), "--project work")

	res, err = Claude(root, "work", true)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(root, ".claude", "settings.local.json"), res.Path)
	assert.Equal(t, "created", res.Status)
}

func TestClaudeKeepsOtherSettingsAndHooks(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, ".claude", "settings.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
	require.NoError(t, os.WriteFile(p, []byte(`{"model": "opus", "hooks": {"SessionStart": [
		{"matcher": "startup", "hooks": [{"type": "command", "command": "echo hi"}]}]}, "env": {"A": "1"}}`), 0o644))
	res, err := Claude(root, "work", false)
	require.NoError(t, err)
	assert.Equal(t, "updated", res.Status)
	data, _ := os.ReadFile(p)
	assert.Equal(t, `{
  "model": "opus",
  "hooks": {
    "SessionStart": [
      {
        "matcher": "startup",
        "hooks": [
          {
            "type": "command",
            "command": "echo hi"
          }
        ]
      },
      {
        "hooks": [
          {
            "type": "command",
            "command": "rit --project work --output text brief"
          }
        ]
      }
    ]
  },
  "env": {
    "A": "1"
  }
}
`, string(data))
}

func TestClaudeRejectsMalformedSettings(t *testing.T) {
	for name, body := range map[string]string{
		"not json":       "{",
		"hooks not obj":  `{"hooks": []}`,
		"start not list": `{"hooks": {"SessionStart": {}}}`,
		"top not object": `"x"`,
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			p := filepath.Join(root, ".claude", "settings.json")
			require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
			require.NoError(t, os.WriteFile(p, []byte(body), 0o644))
			_, err := Claude(root, "work", false)
			require.Error(t, err)
			assert.Contains(t, err.Error(), p)
			data, _ := os.ReadFile(p)
			assert.Equal(t, body, string(data), "a file we cannot parse is left alone")
		})
	}
}

func TestPi(t *testing.T) {
	root := t.TempDir()
	res, err := Pi(root, "work", false)
	require.NoError(t, err)
	assert.Equal(t, "created", res.Status)
	assert.Equal(t, filepath.Join(root, ".pi", "extensions", "ritari.ts"), res.Path)
	data, _ := os.ReadFile(res.Path)
	assert.Contains(t, string(data), `const RIT_ARGS = ["--project", "work", "--output", "text", "brief"];`)

	res, err = Pi(root, "work", false)
	require.NoError(t, err)
	assert.Equal(t, "unchanged", res.Status)
	res, err = Pi(root, `we"ird`, false)
	require.NoError(t, err)
	assert.Equal(t, "updated", res.Status)
	data, _ = os.ReadFile(res.Path)
	assert.Contains(t, string(data), `"we\"ird"`)

	require.NoError(t, os.WriteFile(res.Path, []byte("// mine\n"), 0o644))
	_, err = Pi(root, "work", false)
	assert.EqualError(t, err, res.Path+" exists and was not written by rit (pass --force to replace it)")
	data, _ = os.ReadFile(res.Path)
	assert.Equal(t, "// mine\n", string(data))
	res, err = Pi(root, "work", true)
	require.NoError(t, err)
	assert.Equal(t, "updated", res.Status)
}
