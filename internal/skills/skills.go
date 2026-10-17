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
