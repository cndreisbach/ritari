package output

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

func TestWrite(t *testing.T) {
	var b strings.Builder
	require.NoError(t, Write(&b,
		New().Set("success", true).Set("count", 2),
		New().Set("id", "decisions/a").Set("empty", "").Set("tags", List{"a", "b,c"}).Set("none", List{}).Set("ids", ExplicitList{}).Set("score", 0),
	))
	assert.Equal(t, "success: true\ncount: 2\n---\nid: decisions/a\ntags: [a, \"b,c\"]\nids: []\nscore: 0\n", b.String())
}

func TestMultilineRoundTrip(t *testing.T) {
	for _, s := range []string{
		"---\ntype: note\n---\n\nbody\n",
		"no trailing newline\nsecond",
		"two trailing\n\n",
		"  leading spaces\nnext\n",
		"\nleading blank\n",
		"tab\there\n  indented\n\n\nend\n",
		"line with trailing space \n",
	} {
		var b strings.Builder
		require.NoError(t, Write(&b, New().Set("content", s).Set("after", "x")))
		var m map[string]string
		require.NoError(t, yaml.Unmarshal([]byte(b.String()), &m), b.String())
		assert.Equal(t, s, m["content"], b.String())
		assert.Equal(t, "x", m["after"])
	}
}

func TestScalarQuoting(t *testing.T) {
	var b strings.Builder
	require.NoError(t, Write(&b, New().Set("error", `note "x" already exists: use --force`).Set("v", "true")))
	var m map[string]string
	require.NoError(t, yaml.Unmarshal([]byte(b.String()), &m))
	assert.Equal(t, `note "x" already exists: use --force`, m["error"])
	assert.Equal(t, "true", m["v"])
}
