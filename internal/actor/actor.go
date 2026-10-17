// Package actor determines who is writing: an agent, the configured author, or
// the human at the keyboard (§9.4).
package actor

import (
	"os/exec"
	"os/user"
	"strings"
)

// Lookup is the environment and system access Detect needs.
type Lookup struct {
	Getenv   func(string) string
	Author   string
	GitEmail func() string
	Username func() string
}

var agents = []string{"rovo", "claude-code", "cursor", "codex"}

// Agent reports the detected agent name, or "" for none.
func Agent(getenv func(string) string) string {
	set := func(k string) bool { return strings.TrimSpace(getenv(k)) != "" }
	if v := strings.ToLower(strings.TrimSpace(getenv("RITARI_AGENT"))); v != "" {
		switch v {
		case "unknown", "human":
			return ""
		default:
			for _, a := range agents {
				if v == a {
					return a
				}
			}
		}
	}
	switch {
	case set("ROVODEV_CLI") || set("ROVO_DEV_CLI_VERSION"):
		return "rovo"
	case set("CLAUDE_CODE_CHILD_SESSION") || set("CLAUDECODE"):
		return "claude-code"
	case set("CURSOR_AGENT"):
		return "cursor"
	case set("CODEX_THREAD_ID") || strings.TrimSpace(getenv("CODEX_CI")) == "1" || set("CODEX_SANDBOX"):
		return "codex"
	}
	return ""
}

// Detect returns the actor string to stamp into generated.by / verified[].by.
func Detect(l Lookup) string {
	if a := Agent(l.Getenv); a != "" {
		return a
	}
	if a := strings.TrimSpace(l.Author); a != "" {
		return a
	}
	if l.GitEmail != nil {
		if e := strings.TrimSpace(l.GitEmail()); e != "" {
			return "human:" + e
		}
	}
	if l.Username != nil {
		if u := strings.TrimSpace(l.Username()); u != "" {
			if i := strings.LastIndexByte(u, '\\'); i >= 0 {
				u = u[i+1:]
			}
			if u != "" {
				return "human:" + u
			}
		}
	}
	return "human:unknown"
}

// GitEmail reads `git config user.email`, or "" if unavailable.
func GitEmail() string {
	out, err := exec.Command("git", "config", "user.email").Output()
	if err != nil {
		return ""
	}
	return string(out)
}

// Username is the operating system's current user name, or "".
func Username() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return u.Username
}
