package bundle

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cndreisbach/ritari/internal/vocab"
)

// IndexName is the generated directory listing.
const IndexName = "index.md"

// OKFVersion is the OKF version the root index declares.
const OKFVersion = "0.2"

// Indexes computes the desired content of every generated index.md, keyed by
// bundle-relative path.
func (b *Bundle) Indexes() map[string]string {
	byDir := map[string][]*Concept{}
	for _, c := range b.Concepts {
		byDir[c.Dir()] = append(byDir[c.Dir()], c)
	}
	out := map[string]string{}
	var dirs []string
	for d := range byDir {
		if d != "" {
			dirs = append(dirs, d)
		}
	}
	sort.Strings(dirs)
	out[IndexName] = b.rootIndex(dirs, byDir[""])
	for _, d := range dirs {
		out[d+"/"+IndexName] = b.dirIndex(d, byDir[d])
	}
	return out
}

func entry(c *Concept) string {
	s := "* [" + escapeLinkText(c.Title()) + "](/" + c.ID + ".md)"
	if d := strings.Join(strings.Fields(c.Note.FM.Description), " "); d != "" {
		s += " - " + d
	}
	return s
}

func sortByTitle(cs []*Concept) {
	sort.SliceStable(cs, func(i, j int) bool {
		ti, tj := strings.ToLower(cs[i].Title()), strings.ToLower(cs[j].Title())
		if ti != tj {
			return ti < tj
		}
		return cs[i].ID < cs[j].ID
	})
}

func section(heading string, items []string) string {
	return heading + "\n\n" + strings.Join(items, "\n") + "\n"
}

func (b *Bundle) rootIndex(dirs []string, loose []*Concept) string {
	var sections []string
	if len(dirs) > 0 {
		var items []string
		for _, d := range dirs {
			item := "* [" + escapeLinkText(vocab.Humanize(d)) + "](/" + d + "/" + IndexName + ")"
			if t, ok := b.Vocab.TypeForDir(d); ok {
				item += " (`" + t.Name + "`)"
				if t.Summary != "" {
					item += " - " + t.Summary
				}
			}
			items = append(items, item)
		}
		sections = append(sections, section("# Directories", items))
	}
	if len(loose) > 0 {
		cs := append([]*Concept(nil), loose...)
		sortByTitle(cs)
		var items []string
		for _, c := range cs {
			items = append(items, entry(c))
		}
		sections = append(sections, section("# Concepts", items))
	}
	head := "---\nokf_version: \"" + OKFVersion + "\"\n---\n"
	if len(sections) == 0 {
		return head + "\n"
	}
	return head + "\n" + strings.Join(sections, "\n")
}

func (b *Bundle) dirIndex(dir string, cs []*Concept) string {
	cs = append([]*Concept(nil), cs...)
	if dir == b.Vocab.DailyDir() {
		return dailyIndex(cs)
	}
	sortByTitle(cs)
	var live, archived []string
	for _, c := range cs {
		if c.Note.FM.Archived {
			archived = append(archived, entry(c))
		} else {
			live = append(live, entry(c))
		}
	}
	out := "# " + vocab.Humanize(dir) + "\n"
	if len(live) > 0 {
		out += "\n" + strings.Join(live, "\n") + "\n"
	}
	if len(archived) > 0 {
		out += "\n" + section("# Archived", archived)
	}
	return out
}

func dailyYear(c *Concept) string {
	for _, s := range []string{c.Title(), path.Base(c.ID)} {
		if len(s) >= 4 && strings.Trim(s[:4], "0123456789") == "" {
			return s[:4]
		}
	}
	return "Undated"
}

func dailyIndex(cs []*Concept) string {
	groups := map[string][]*Concept{}
	var years []string
	for _, c := range cs {
		y := dailyYear(c)
		if _, ok := groups[y]; !ok {
			years = append(years, y)
		}
		groups[y] = append(groups[y], c)
	}
	sort.Slice(years, func(i, j int) bool {
		if years[i] == "Undated" || years[j] == "Undated" {
			return years[j] == "Undated" && years[i] != "Undated"
		}
		return years[i] > years[j]
	})
	var sections []string
	for _, y := range years {
		g := groups[y]
		sort.Slice(g, func(i, j int) bool { return g[i].ID > g[j].ID })
		var items []string
		for _, c := range g {
			items = append(items, entry(c))
		}
		sections = append(sections, section("# "+y, items))
	}
	return strings.Join(sections, "\n")
}

// SyncIndexes writes the index files whose content differs and removes index
// files that are no longer generated. With check, it only reports. It returns
// the changed bundle-relative paths, sorted; removals end in " (removed)".
func (b *Bundle) SyncIndexes(check bool) ([]string, error) {
	want := b.Indexes()
	var changed []string
	for rel, content := range want {
		full := b.Abs(rel)
		old, err := os.ReadFile(full)
		if err == nil && string(old) == content {
			continue
		}
		changed = append(changed, rel)
		if !check {
			if err := WriteFileAtomic(full, []byte(content)); err != nil {
				return nil, err
			}
		}
	}
	existing, err := indexFiles(b.Root)
	if err != nil {
		return nil, err
	}
	for _, rel := range existing {
		if _, ok := want[rel]; ok {
			continue
		}
		changed = append(changed, rel+" (removed)")
		if !check {
			if err := os.Remove(b.Abs(rel)); err != nil {
				return nil, err
			}
		}
	}
	sort.Strings(changed)
	return changed, nil
}

// indexFiles lists every index.md in the bundle outside dot-directories.
func indexFiles(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if p != root && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == IndexName {
			rel, err := filepath.Rel(root, p)
			if err != nil {
				return err
			}
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	return out, err
}
