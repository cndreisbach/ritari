// Package registry manages the per-user project registry (§9.3).
package registry

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/cndreisbach/ritari/internal/bundle"
	"github.com/cndreisbach/ritari/internal/yamlx"
)

// Project is a named bundle.
type Project struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

// Registry is the parsed config file.
type Registry struct {
	Default  string    `yaml:"default"`
	Projects []Project `yaml:"projects"`
	Author   string    `yaml:"author"`

	path string
}

// Path is the registry location: RITARI_CONFIG, else
// $XDG_CONFIG_HOME/ritari/config.yaml, else ~/.config/ritari/config.yaml.
func Path(getenv func(string) string) (string, error) {
	if p := getenv("RITARI_CONFIG"); p != "" {
		return p, nil
	}
	if x := getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "ritari", "config.yaml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "ritari", "config.yaml"), nil
}

// Load reads the registry at path. A missing file is an empty registry.
func Load(path string) (*Registry, error) {
	r := &Registry{path: path}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return r, nil
	}
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, r); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return r, nil
}

// Save writes the registry atomically, creating parent directories.
func (r *Registry) Save() error {
	var b strings.Builder
	if r.Default != "" {
		b.WriteString("default: " + yamlx.Scalar(r.Default) + "\n")
	}
	if len(r.Projects) == 0 {
		b.WriteString("projects: []\n")
	} else {
		b.WriteString("projects:\n")
		for _, p := range r.Projects {
			b.WriteString("  - name: " + yamlx.Scalar(p.Name) + "\n    path: " + yamlx.Scalar(p.Path) + "\n")
		}
	}
	if r.Author != "" {
		b.WriteString("author: " + yamlx.Scalar(r.Author) + "\n")
	}
	return bundle.WriteFileAtomic(r.path, []byte(b.String()))
}

// FilePath is where the registry lives.
func (r *Registry) FilePath() string { return r.path }

// Find looks up a project by name.
func (r *Registry) Find(name string) (Project, bool) {
	for _, p := range r.Projects {
		if p.Name == name {
			return p, true
		}
	}
	return Project{}, false
}

func notFound(name string) error {
	return fmt.Errorf("project %q not found", name)
}

// ExpandPath expands a leading ~ and makes p absolute.
func ExpandPath(p string) (string, error) {
	if p == "~" || strings.HasPrefix(p, "~/") || strings.HasPrefix(p, `~\`) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		p = filepath.Join(home, p[1:])
	}
	return filepath.Abs(p)
}

// Add registers a project. It becomes the default if asked, or if there is none.
func (r *Registry) Add(name, path string, makeDefault bool) (Project, error) {
	if strings.TrimSpace(name) == "" {
		return Project{}, errors.New("project name is required")
	}
	if _, ok := r.Find(name); ok {
		return Project{}, fmt.Errorf("project %q already exists", name)
	}
	abs, err := ExpandPath(path)
	if err != nil {
		return Project{}, err
	}
	p := Project{Name: name, Path: abs}
	r.Projects = append(r.Projects, p)
	if makeDefault || r.Default == "" {
		r.Default = name
	}
	return p, nil
}

// Remove unregisters a project; it never touches files.
func (r *Registry) Remove(name string) error {
	for i, p := range r.Projects {
		if p.Name != name {
			continue
		}
		r.Projects = append(r.Projects[:i], r.Projects[i+1:]...)
		if r.Default == name {
			r.Default = ""
			if len(r.Projects) > 0 {
				r.Default = r.Projects[0].Name
			}
		}
		return nil
	}
	return notFound(name)
}

// SetDefault makes a registered project the default.
func (r *Registry) SetDefault(name string) error {
	if _, ok := r.Find(name); !ok {
		return notFound(name)
	}
	r.Default = name
	return nil
}

// Active resolves the active project: flag, else RITARI_PROJECT, else the default.
func (r *Registry) Active(flag string, getenv func(string) string) (Project, error) {
	name := flag
	if name == "" {
		name = getenv("RITARI_PROJECT")
	}
	if name == "" {
		name = r.Default
	}
	if name == "" {
		return Project{}, errors.New("no project specified and no default set (try: rit project init <name> <path>)")
	}
	p, ok := r.Find(name)
	if !ok {
		return Project{}, notFound(name)
	}
	return p, nil
}
