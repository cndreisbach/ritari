package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMode(t *testing.T) {
	env := func(v string) func(string) string { return func(string) string { return v } }
	for _, c := range []struct {
		flag, envv string
		tty        bool
		want       string
	}{
		{"", "", true, "text"},
		{"", "", false, "yaml"},
		{"", "text", false, "text"},
		{"yaml", "text", true, "yaml"},
		{"TEXT", "", false, "text"},
		{"json", "text", true, "json"},
		{"", "JSON", true, "json"},
	} {
		got, err := mode(c.flag, IO{Getenv: env(c.envv), IsTerminal: c.tty})
		require.NoError(t, err)
		assert.Equal(t, c.want, got, "%+v", c)
	}
	_, err := mode("toml", IO{Getenv: env("")})
	assert.EqualError(t, err, `unknown output format "toml" (want text, yaml, or json)`)
}

func TestParseMeta(t *testing.T) {
	got, err := parseMeta([]string{"state=done", "url=a=b"})
	require.NoError(t, err)
	assert.Equal(t, [][2]string{{"state", "done"}, {"url", "a=b"}}, got)
	_, err = parseMeta([]string{"nokey"})
	assert.EqualError(t, err, `--meta "nokey" is not key=value`)
}

func TestSplitTags(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, splitTags(" a, ,b,"))
	assert.Nil(t, splitTags(""))
}

func TestUsageErrorsExit2(t *testing.T) {
	var out, errb strings.Builder
	code := Main([]string{"nonsense"}, IO{Stdin: strings.NewReader(""), Stdout: &out, Stderr: &errb, Getenv: func(string) string { return "" }})
	assert.Equal(t, 2, code)
	assert.Contains(t, errb.String(), "rit: ")
}

func TestHelpExits0(t *testing.T) {
	var out, errb strings.Builder
	code := Main([]string{"--help"}, IO{Stdin: strings.NewReader(""), Stdout: &out, Stderr: &errb, Getenv: func(string) string { return "" }})
	assert.Equal(t, 0, code)
	assert.Contains(t, out.String(), "Usage: rit")
}

func TestVersionFromLinker(t *testing.T) {
	old := version
	defer func() { version = old }()
	version = "1.2.3"
	assert.Equal(t, "1.2.3", Version())
}
