// Package hooks installs project-scoped session hooks that run `rit brief`,
// so an agent starts each session with the latest daily notes and open tasks
// (§9.22). Nothing here writes outside the project directory it is given.
package hooks

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cndreisbach/ritari/internal/bundle"
)

// Agents lists the agents hooks can be installed for.
var Agents = []string{"claude", "pi"}

// Result reports one agent's install.
type Result struct {
	Agent  string
	Path   string
	Status string // created, updated, unchanged
}

// Args is the rit command line a hook runs for project.
func Args(project string) []string {
	return []string{"--project", project, "--output", "text", "brief"}
}

// Command is Args as one shell command, for agents that take a string.
func Command(project string) string {
	words := []string{"rit"}
	for _, a := range Args(project) {
		words = append(words, shellQuote(a))
	}
	return strings.Join(words, " ")
}

var plainWord = regexp.MustCompile(`^[A-Za-z0-9._/:=-]+$`)

func shellQuote(s string) string {
	if plainWord.MatchString(s) {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// ownHook matches a hook command an earlier install wrote.
var ownHook = regexp.MustCompile(`^rit\s.*\bbrief$`)

// Claude adds the SessionStart hook to root/.claude/settings.json, or
// settings.local.json when local, keeping everything else in the file. A hook
// an earlier install wrote is updated in place rather than duplicated.
func Claude(root, project string, local bool) (Result, error) {
	name := "settings.json"
	if local {
		name = "settings.local.json"
	}
	p := filepath.Join(root, ".claude", name)
	res := Result{Agent: "claude", Path: p}
	data, err := os.ReadFile(p)
	exists := err == nil
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return res, err
	}
	settings, err := parseObject(data)
	if err != nil {
		return res, fmt.Errorf("%s: %w", p, err)
	}
	hooks, err := childObject(settings, "hooks")
	if err != nil {
		return res, fmt.Errorf("%s: %w", p, err)
	}
	var starts []any
	if v, ok := hooks.get("SessionStart"); ok {
		if starts, ok = v.([]any); !ok {
			return res, fmt.Errorf("%s: hooks.SessionStart is not a list", p)
		}
	}

	command := Command(project)
	found, changed := false, false
	for _, e := range starts {
		entry, ok := e.(*object)
		if !ok {
			continue
		}
		v, _ := entry.get("hooks")
		list, _ := v.([]any)
		for _, h := range list {
			hook, ok := h.(*object)
			if !ok {
				continue
			}
			c, _ := hook.get("command")
			if s, ok := c.(string); ok && ownHook.MatchString(s) {
				found = true
				if s != command {
					hook.set("command", command)
					changed = true
				}
			}
		}
	}
	if !found {
		hook := newObject()
		hook.set("type", "command")
		hook.set("command", command)
		entry := newObject()
		entry.set("hooks", []any{hook})
		hooks.set("SessionStart", append(starts, entry))
		changed = true
	}

	switch {
	case !exists:
		res.Status = "created"
	case changed:
		res.Status = "updated"
	default:
		res.Status = "unchanged"
		return res, nil
	}
	return res, bundle.WriteFileAtomic(p, encode(settings))
}

func childObject(parent *object, key string) (*object, error) {
	v, ok := parent.get(key)
	if !ok {
		child := newObject()
		parent.set(key, child)
		return child, nil
	}
	child, ok := v.(*object)
	if !ok {
		return nil, fmt.Errorf("%s is not an object", key)
	}
	return child, nil
}

// piHeader begins every extension rit writes; a file without it is someone
// else's and is replaced only with force.
const piHeader = "// Written by `rit install hooks`; run it again to update this file.\n"

const piBody = `// At the start of a session and after a compaction, adds rit brief's output
// (the latest daily notes and open tasks) to the agent's context.
import type { ExtensionAPI } from "@mariozechner/pi-coding-agent";

const RIT_ARGS = %s;

export default function (pi: ExtensionAPI) {
  let pending = true;
  pi.on("session_start", async () => {
    pending = true;
  });
  pi.on("session_compact", async () => {
    pending = true;
  });
  pi.on("before_agent_start", async () => {
    if (!pending) return;
    pending = false;
    try {
      const result = await pi.exec("rit", RIT_ARGS, { timeout: 10000 });
      if (result.code !== 0 || !result.stdout.trim()) return;
      return {
        message: { customType: "ritari-brief", content: result.stdout, display: false },
      };
    } catch {
      return;
    }
  });
}
`

// PiExtension is the extension source for project.
func PiExtension(project string) []byte {
	var args []string
	for _, a := range Args(project) {
		q, _ := json.Marshal(a)
		args = append(args, string(q))
	}
	return []byte(piHeader + fmt.Sprintf(piBody, "["+strings.Join(args, ", ")+"]"))
}

// Pi writes the extension to root/.pi/extensions/ritari.ts. An existing file
// rit did not write is left alone unless force.
func Pi(root, project string, force bool) (Result, error) {
	p := filepath.Join(root, ".pi", "extensions", "ritari.ts")
	res := Result{Agent: "pi", Path: p}
	want := PiExtension(project)
	data, err := os.ReadFile(p)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		res.Status = "created"
	case err != nil:
		return res, err
	case string(data) == string(want):
		res.Status = "unchanged"
		return res, nil
	case !strings.HasPrefix(string(data), piHeader) && !force:
		return res, fmt.Errorf("%s exists and was not written by rit (pass --force to replace it)", p)
	default:
		res.Status = "updated"
	}
	return res, bundle.WriteFileAtomic(p, want)
}
