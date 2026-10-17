// Package vocab holds a bundle's type vocabulary: which types it recognizes and
// the directory each lives in.
package vocab

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

// FileName is the per-bundle vocabulary file.
const FileName = "ritari.yaml"

// Type is one entry in a vocabulary.
type Type struct {
	Name    string
	Dir     string
	Summary string
}

// Vocab is an ordered set of types.
type Vocab struct {
	Types []Type
}

var defaults = []Type{
	{"decision", "decisions", "A choice made, with its rationale."},
	{"note", "notes", "General knowledge that does not fit a more specific type."},
	{"task", "tasks", "Work in progress, with state."},
	{"person", "people", "A human and what you know about them."},
	{"project", "projects", "An ongoing effort."},
	{"meeting", "meetings", "A meeting and what came out of it."},
	{"reference", "references", "A pointer to an external resource."},
	{"preference", "preferences", "How the user wants things done."},
	{"daily", "daily", "A day's narrative: what came up, and what it touched."},
}

// Reserved type names, always present.
const (
	Note  = "note"
	Daily = "daily"
)

// Default returns the built-in vocabulary.
func Default() *Vocab {
	return &Vocab{Types: append([]Type(nil), defaults...)}
}

func defaultType(name string) Type {
	for _, t := range defaults {
		if t.Name == name {
			return t
		}
	}
	panic("no default type " + name)
}

// Load reads <root>/ritari.yaml. A missing file, or one with no types, means the
// default vocabulary. Any problem with the file is an error.
func Load(root string) (*Vocab, error) {
	data, err := os.ReadFile(filepath.Join(root, FileName))
	if errors.Is(err, fs.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", FileName, err)
	}
	v, err := parse(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", FileName, err)
	}
	return v, nil
}

func parse(data []byte) (*Vocab, error) {
	var raw struct {
		Types []struct {
			Name    string `yaml:"name"`
			Dir     string `yaml:"dir"`
			Summary string `yaml:"summary"`
		} `yaml:"types"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	if len(raw.Types) == 0 {
		return Default(), nil
	}
	v := &Vocab{}
	byName := map[string]bool{}
	byDir := map[string]string{}
	for i, rt := range raw.Types {
		name := strings.ToLower(strings.TrimSpace(rt.Name))
		if name == "" {
			return nil, fmt.Errorf("type %d has no name", i+1)
		}
		if byName[name] {
			return nil, fmt.Errorf("type %q is declared twice", name)
		}
		byName[name] = true
		dir := strings.TrimSpace(rt.Dir)
		if dir == "" {
			dir = Plural(Slug(name))
		}
		if dir != Slug(dir) {
			return nil, fmt.Errorf("type %q: dir %q must be a single lowercase path segment of [a-z0-9-]", name, dir)
		}
		if dir == "sources" {
			return nil, fmt.Errorf("type %q: dir %q is reserved for raw material", name, dir)
		}
		if other, ok := byDir[dir]; ok {
			return nil, fmt.Errorf("types %q and %q both use dir %q", other, name, dir)
		}
		byDir[dir] = name
		v.Types = append(v.Types, Type{Name: name, Dir: dir, Summary: strings.TrimSpace(rt.Summary)})
	}
	for _, r := range []string{Note, Daily} {
		if byName[r] {
			continue
		}
		t := defaultType(r)
		if other, ok := byDir[t.Dir]; ok {
			return nil, fmt.Errorf("types %q and %q both use dir %q", other, r, t.Dir)
		}
		byDir[t.Dir] = r
		v.Types = append(v.Types, t)
	}
	return v, nil
}

func norm(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// Lookup finds a declared type by name (case-insensitive, trimmed).
func (v *Vocab) Lookup(name string) (Type, bool) {
	n := norm(name)
	for _, t := range v.Types {
		if t.Name == n {
			return t, true
		}
	}
	return Type{}, false
}

// DirFor is the directory a type's notes live in. An undeclared type gets the
// pluralized slug of its name.
func (v *Vocab) DirFor(name string) string {
	if t, ok := v.Lookup(name); ok {
		return t.Dir
	}
	return Plural(Slug(norm(name)))
}

// TypeForDir finds the first declared type that lives in dir.
func (v *Vocab) TypeForDir(dir string) (Type, bool) {
	for _, t := range v.Types {
		if t.Dir == dir {
			return t, true
		}
	}
	return Type{}, false
}

// IDFor is the concept ID a new note of this type and title gets.
func (v *Vocab) IDFor(typ, title string) string {
	return v.DirFor(typ) + "/" + Slug(title)
}

// Names lists the declared type names in order.
func (v *Vocab) Names() []string {
	out := make([]string, len(v.Types))
	for i, t := range v.Types {
		out[i] = t.Name
	}
	return out
}

// DailyDir is the directory holding daily notes.
func (v *Vocab) DailyDir() string { return v.DirFor(Daily) }

// NoteDir is the directory of the default type.
func (v *Vocab) NoteDir() string { return v.DirFor(Note) }

// YAML renders the vocabulary as ritari.yaml content.
func (v *Vocab) YAML() string {
	var b strings.Builder
	b.WriteString("# The complete list of types this knowledge base recognizes.\n")
	b.WriteString("# It replaces the built-in types rather than adding to them.\n")
	b.WriteString("# `note` and `daily` are always present, even if omitted here.\n")
	b.WriteString("types:\n")
	for _, t := range v.Types {
		fmt.Fprintf(&b, "  - name: %s\n    dir: %s\n", t.Name, t.Dir)
		if t.Summary != "" {
			fmt.Fprintf(&b, "    summary: %s\n", quoteIfNeeded(t.Summary))
		}
	}
	return b.String()
}

func quoteIfNeeded(s string) string {
	var n yaml.Node
	if err := yaml.Unmarshal([]byte("k: "+s), &n); err == nil && len(n.Content) == 1 {
		if m := n.Content[0]; len(m.Content) == 2 && m.Content[1].Value == s && m.Content[1].ShortTag() == "!!str" {
			return s
		}
	}
	return fmt.Sprintf("%q", s)
}
