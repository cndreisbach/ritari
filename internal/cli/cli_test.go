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
		tty, want  bool
	}{
		{"", "", true, false},
		{"", "", false, true},
		{"", "text", false, false},
		{"yaml", "text", true, true},
		{"TEXT", "", false, false},
	} {
		got, err := mode(c.flag, IO{Getenv: env(c.envv), IsTerminal: c.tty})
		require.NoError(t, err)
		assert.Equal(t, c.want, got, "%+v", c)
	}
	_, err := mode("json", IO{Getenv: env("")})
	assert.EqualError(t, err, `unknown output format "json" (want text or yaml)`)
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
