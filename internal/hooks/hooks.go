// Package hooks installs project-scoped session hooks that run `rit brief`,
// so an agent starts each session with the latest daily notes and open tasks
// (§9.22). Nothing here writes outside the project directory it is given.
package hooks

import (
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
var Agents = []string{"claude"}

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
