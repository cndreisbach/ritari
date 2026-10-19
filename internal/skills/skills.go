// Package skills embeds the agent skills that teach an agent to drive rit (§11).
package skills

import (
	"bytes"
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

//go:embed */SKILL.md */references/*.md
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

// Files lists a skill's files relative to its directory ('/'-delimited),
// sorted: SKILL.md and anything under references/.
func Files(name string) ([]string, error) {
	var out []string
	err := fs.WalkDir(files, name, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			out = append(out, strings.TrimPrefix(p, name+"/"))
		}
		return err
	})
	sort.Strings(out)
	return out, err
}

// Install writes each skill's files under <dir>/<name>/. Missing files are
// always written; a file that differs from this build's is overwritten only
// with force. It returns, sorted, the skills it wrote any file of, and the
// skills it left with a file that differs (empty when force).
func Install(dir string, force bool) (installed, outdated []string, err error) {
	installed, outdated = []string{}, []string{}
	for _, name := range Names() {
		rels, err := Files(name)
		if err != nil {
			return nil, nil, err
		}
		wrote, differs := false, false
		for _, rel := range rels {
			want, err := files.ReadFile(path.Join(name, rel))
			if err != nil {
				return nil, nil, err
			}
			dest := filepath.Join(dir, name, filepath.FromSlash(rel))
			have, err := os.ReadFile(dest)
			switch {
			case err == nil && bytes.Equal(have, want):
				continue
			case err == nil && !force:
				differs = true
				continue
			case err != nil && !errors.Is(err, fs.ErrNotExist):
				return nil, nil, err
			}
			if err := bundle.WriteFileAtomic(dest, want); err != nil {
				return nil, nil, err
			}
			wrote = true
		}
		if wrote {
			installed = append(installed, name)
		}
		if differs {
			outdated = append(outdated, name)
		}
	}
	return installed, outdated, nil
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
