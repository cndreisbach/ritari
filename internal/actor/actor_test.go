package actor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestAgent(t *testing.T) {
	cases := []struct {
		env  map[string]string
		want string
	}{
		{map[string]string{"CLAUDECODE": "1"}, "claude-code"},
		{map[string]string{"CLAUDE_CODE_CHILD_SESSION": "x", "ROVODEV_CLI": "1"}, "rovo"},
		{map[string]string{"CURSOR_AGENT": "1"}, "cursor"},
		{map[string]string{"CODEX_CI": "1"}, "codex"},
		{map[string]string{"CODEX_CI": "0"}, ""},
		{map[string]string{"CODEX_THREAD_ID": " "}, ""},
		{map[string]string{"RITARI_AGENT": " Codex ", "CLAUDECODE": "1"}, "codex"},
		{map[string]string{"RITARI_AGENT": "human", "CLAUDECODE": "1"}, ""},
		{map[string]string{"RITARI_AGENT": "bogus", "CLAUDECODE": "1"}, "claude-code"},
		{nil, ""},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, Agent(env(c.env)), "%v", c.env)
	}
}

func TestDetect(t *testing.T) {
	none := func() string { return "" }
	assert.Equal(t, "claude-code", Detect(Lookup{Getenv: env(map[string]string{"CLAUDECODE": "1"}), Author: "human:me"}))
	assert.Equal(t, "human:me", Detect(Lookup{Getenv: env(nil), Author: " human:me ", GitEmail: func() string { return "x@y" }}))
	assert.Equal(t, "human:x@y", Detect(Lookup{Getenv: env(nil), GitEmail: func() string { return "x@y\n" }}))
	assert.Equal(t, "human:jane", Detect(Lookup{Getenv: env(nil), GitEmail: none, Username: func() string { return `CORP\jane` }}))
	assert.Equal(t, "human:unknown", Detect(Lookup{Getenv: env(nil), GitEmail: none, Username: none}))
}
