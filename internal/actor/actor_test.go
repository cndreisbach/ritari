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
		{map[string]string{"AI_AGENT": "pi"}, "pi"},
		{map[string]string{"AI_AGENT": " PI "}, "pi"},
		{map[string]string{"PI_CODING_AGENT": "true"}, "pi"},
		{map[string]string{"AI_AGENT": "claude"}, ""},
		{map[string]string{"RITARI_AGENT": "pi", "CLAUDECODE": "1"}, "pi"},
		{map[string]string{"AI_AGENT": "pi", "CLAUDECODE": "1"}, "claude-code"},
		{nil, ""},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, Agent(env(c.env)), "%v", c.env)
	}
}

func TestModel(t *testing.T) {
	cases := []struct {
		env   map[string]string
		agent string
		want  string
	}{
		{map[string]string{"PI_MODEL": "deepseek-flash"}, "pi", "deepseek-flash"},
		{map[string]string{"PI_MODEL": " deepseek-flash "}, "pi", "deepseek-flash"},
		{map[string]string{"PI_MODEL": "deepseek-flash"}, "claude-code", ""},
		{nil, "pi", ""},
		{map[string]string{"RITARI_AGENT_MODEL": "pinned", "PI_MODEL": "ignored"}, "pi", "pinned"},
		{map[string]string{"RITARI_AGENT_MODEL": "pinned"}, "claude-code", "pinned"},
		{map[string]string{"RITARI_AGENT_MODEL": "  "}, "pi", ""},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, Model(env(c.env), c.agent), "%v agent=%s", c.env, c.agent)
	}
}

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"":                  "",
		"   ":               "",
		"cndreisbach":       "human:cndreisbach",
		" me@example.com ":  "human:me@example.com",
		"human:cndreisbach": "human:cndreisbach",
		"process:nightly":   "process:nightly",
		"human:":            "",
		"process:":          "",
		"CORP\\jane":        "human:CORP\\jane",
	}
	for in, want := range cases {
		assert.Equal(t, want, Normalize(in), "%q", in)
	}
}

func TestDetect(t *testing.T) {
	none := func() string { return "" }
	assert.Equal(t, "claude-code", Detect(Lookup{Getenv: env(map[string]string{"CLAUDECODE": "1"}), Author: "human:me"}))
	assert.Equal(t, "pi/deepseek-flash", Detect(Lookup{Getenv: env(map[string]string{"AI_AGENT": "pi", "PI_MODEL": "deepseek-flash"})}))
	assert.Equal(t, "pi", Detect(Lookup{Getenv: env(map[string]string{"RITARI_AGENT": "pi"})}), "no model signal leaves the bare name")
	assert.Equal(t, "claude-code/sonnet-4", Detect(Lookup{Getenv: env(map[string]string{"RITARI_AGENT": "claude-code", "RITARI_AGENT_MODEL": "sonnet-4"})}))
	assert.Equal(t, "human:me", Detect(Lookup{Getenv: env(map[string]string{"RITARI_AGENT_MODEL": "ignored"}), Author: "me"}), "a human never carries a model")
	assert.Equal(t, "human:me", Detect(Lookup{Getenv: env(nil), Author: " human:me ", GitEmail: func() string { return "x@y" }}))
	assert.Equal(t, "human:me", Detect(Lookup{Getenv: env(nil), Author: "me", GitEmail: func() string { return "x@y" }}), "a bare author is a human id")
	assert.Equal(t, "human:me", Detect(Lookup{Getenv: env(nil), Author: "me", Username: none}), "a bare author beats the OS user name")
	assert.Equal(t, "human:x@y", Detect(Lookup{Getenv: env(nil), GitEmail: func() string { return "x@y\n" }}))
	assert.Equal(t, "human:jane", Detect(Lookup{Getenv: env(nil), GitEmail: none, Username: func() string { return `CORP\jane` }}))
	assert.Equal(t, "human:unknown", Detect(Lookup{Getenv: env(nil), GitEmail: none, Username: none}))
}
