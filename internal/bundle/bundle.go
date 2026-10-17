// Package bundle loads an OKF bundle from disk and performs every operation on
// it: resolving notes, the link graph, writes, indexes, and the change log.
package bundle

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cndreisbach/ritari/internal/note"
	"github.com/cndreisbach/ritari/internal/vocab"
)

// Concept is one parsed concept file.
type Concept struct {
	ID      string // bundle-relative path minus .md, '/'-delimited
	Path    string // bundle-relative file path, '/'-delimited
	Note    *note.Note
	ModTime time.Time
}

// Title is the frontmatter title, or the last segment of the ID.
func (c *Concept) Title() string {
	if t := strings.TrimSpace(c.Note.FM.Title); t != "" {
		return t
	}
	return path.Base(c.ID)
}

// Dir is the concept's directory ("" for the root).
func (c *Concept) Dir() string {
	if d := path.Dir(c.ID); d != "." {
		return d
	}
	return ""
}

// IsDaily reports whether the concept is a daily note.
func (c *Concept) IsDaily() bool {
	return strings.EqualFold(strings.TrimSpace(c.Note.FM.Type), vocab.Daily)
}

// Problem is a concept file that could not be parsed.
type Problem struct {
	ID   string
	Path string
	Err  error
}

// Bundle is a loaded knowledge base.
type Bundle struct {
	Root     string
	Vocab    *vocab.Vocab
	Concepts []*Concept // sorted by ID
	Problems []Problem  // sorted by path
	// Files lists every concept file's ID, parseable or not, in directory order.
	Files []string

	byID map[string]*Concept
}

// Open loads the bundle at root. A bad vocabulary is fatal; a bad note is not.
func Open(root string) (*Bundle, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", abs)
	}
	v, err := vocab.Load(abs)
	if err != nil {
		return nil, err
	}
	b := &Bundle{Root: abs, Vocab: v, byID: map[string]*Concept{}}
	files, err := conceptFiles(abs)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		b.Files = append(b.Files, idOf(f))
	}
	b.load(files)
	return b, nil
}

// IsConceptFile reports whether a file name is a concept (not reserved).
func IsConceptFile(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".md") && lower != "index.md" && lower != "log.md"
}

func idOf(rel string) string {
	return rel[:len(rel)-3]
}

// conceptFiles walks root in lexical order, returning '/'-delimited relative
// paths of concept files. Dot-directories are skipped.
func conceptFiles(root string) ([]string, error) {
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
		if !d.Type().IsRegular() || !IsConceptFile(d.Name()) {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	return out, err
}

func (b *Bundle) load(files []string) {
	type result struct {
		c *Concept
		p *Problem
	}
	results := make([]result, len(files))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < runtime.GOMAXPROCS(0); w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				c, err := readConcept(b.Root, files[i])
				if err != nil {
					results[i] = result{p: &Problem{ID: idOf(files[i]), Path: files[i], Err: err}}
				} else {
					results[i] = result{c: c}
				}
			}
		}()
	}
	for i := range files {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	for _, r := range results {
		switch {
		case r.c != nil:
			b.Concepts = append(b.Concepts, r.c)
			b.byID[r.c.ID] = r.c
		case r.p != nil:
			b.Problems = append(b.Problems, *r.p)
		}
	}
	sort.Slice(b.Concepts, func(i, j int) bool { return b.Concepts[i].ID < b.Concepts[j].ID })
	sort.Slice(b.Problems, func(i, j int) bool { return b.Problems[i].Path < b.Problems[j].Path })
}

func readConcept(root, rel string) (*Concept, error) {
	full := filepath.Join(root, filepath.FromSlash(rel))
	data, err := os.ReadFile(full)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(full)
	if err != nil {
		return nil, err
	}
	n, err := note.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("frontmatter could not be parsed: %w", err)
	}
	return &Concept{ID: idOf(rel), Path: rel, Note: n, ModTime: info.ModTime()}, nil
}

// Get returns a loaded concept by exact ID.
func (b *Bundle) Get(id string) (*Concept, bool) {
	c, ok := b.byID[id]
	return c, ok
}

// Abs is the absolute file path of a bundle-relative path.
func (b *Bundle) Abs(rel string) string {
	return filepath.Join(b.Root, filepath.FromSlash(rel))
}

// ErrNoID is returned when a note reference is empty.
var ErrNoID = errors.New("no note id given")

// ErrNotFound is returned when a note reference matches nothing.
var ErrNotFound = errors.New("note not found")

// Resolve finds the concept a reference names: a path, an ID, or a title (§5.1).
func (b *Bundle) Resolve(ref string) (*Concept, error) {
	if strings.TrimSpace(ref) == "" {
		return nil, ErrNoID
	}
	if rel, ok := b.pathRef(ref); ok {
		if c, ok := b.byID[idOf(rel)]; ok {
			return c, nil
		}
		for _, p := range b.Problems {
			if p.Path == rel {
				return nil, fmt.Errorf("%s: %w", rel, p.Err)
			}
		}
	}
	if c, ok := b.byID[ref]; ok {
		return c, nil
	}
	if c := b.ByTitle(ref); c != nil {
		return c, nil
	}
	return nil, fmt.Errorf("%w: %s", ErrNotFound, ref)
}

// pathRef interprets ref as a file path, returning its bundle-relative path if
// it names an existing concept file inside the bundle.
func (b *Bundle) pathRef(ref string) (string, bool) {
	var candidates []string
	if !strings.HasSuffix(strings.ToLower(ref), ".md") {
		candidates = append(candidates, filepath.Join(b.Root, filepath.FromSlash(ref+".md")))
	}
	candidates = append(candidates, filepath.Join(b.Root, filepath.FromSlash(ref)))
	if filepath.IsAbs(ref) {
		candidates = append(candidates, filepath.Clean(ref))
	}
	for _, c := range candidates {
		info, err := os.Stat(c)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		rel, err := filepath.Rel(b.Root, c)
		if err != nil || strings.HasPrefix(rel, "..") {
			continue
		}
		rel = filepath.ToSlash(rel)
		if IsConceptFile(path.Base(rel)) {
			return rel, true
		}
	}
	return "", false
}

// ByTitle finds the concept whose title slugifies like s, preferring the
// smallest ID.
func (b *Bundle) ByTitle(s string) *Concept {
	want := vocab.Slug(s)
	if want == "" {
		return nil
	}
	for _, c := range b.Concepts {
		if vocab.Slug(c.Title()) == want {
			return c
		}
	}
	return nil
}
