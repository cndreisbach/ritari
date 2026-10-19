// Package skills embeds the agent skills that teach an agent to drive rit (§11).
package skills

import (
	"embed"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cndreisbach/ritari/internal/bundle"
)

//go:embed */SKILL.md
var files embed.FS

// Names lists the embedded skills, sorted.
func Names() []string {
	entries, _ := fs.ReadDir(files, ".")
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}

// Content returns a skill's SKILL.md.
func Content(name string) ([]byte, error) {
	return files.ReadFile(path.Join(name, "SKILL.md"))
}

// Install writes each skill to <dir>/<name>/SKILL.md, leaving existing files
// alone unless force. It returns the names it wrote, sorted.
func Install(dir string, force bool) ([]string, error) {
	installed := []string{}
	for _, name := range Names() {
		dest := filepath.Join(dir, name, "SKILL.md")
		if !force {
			if _, err := os.Stat(dest); err == nil {
				continue
			} else if !errors.Is(err, fs.ErrNotExist) {
				return nil, err
			}
		}
		data, err := Content(name)
		if err != nil {
			return nil, err
		}
		if err := bundle.WriteFileAtomic(dest, data); err != nil {
			return nil, err
		}
		installed = append(installed, name)
	}
	return installed, nil
}

// Retired lists the skill directories in dir named ritari-* that are not
// embedded in this build, sorted: skills an earlier version installed and
// this one no longer ships.
func Retired(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return []string{}, nil
	} else if err != nil {
		return nil, err
	}
	shipped := map[string]bool{}
	for _, name := range Names() {
		shipped[name] = true
	}
	retired := []string{}
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "ritari-") && !shipped[e.Name()] {
			retired = append(retired, e.Name())
		}
	}
	return retired, nil
}

// Remove deletes the named skill directories from dir.
func Remove(dir string, names []string) error {
	for _, name := range names {
		if err := os.RemoveAll(filepath.Join(dir, name)); err != nil {
			return err
		}
	}
	return nil
}
