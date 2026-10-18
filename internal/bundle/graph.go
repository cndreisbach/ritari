package bundle

import (
	"os"

	"github.com/cndreisbach/ritari/internal/note"
)

// Graph is the bundle's link graph.
type Graph struct {
	// Out maps a concept ID to the distinct existing concepts it links to.
	Out map[string][]string
	// In maps a concept ID to the concepts linking to it.
	In map[string][]string
	// Dangling maps a concept ID to the distinct in-bundle targets it links to
	// that do not exist: concept IDs, or reserved-file paths (ideas/index.md)
	// whose file is missing.
	Dangling map[string][]string
}

// Graph computes the link graph from the loaded concepts.
func (b *Bundle) Graph() *Graph {
	g := &Graph{Out: map[string][]string{}, In: map[string][]string{}, Dangling: map[string][]string{}}
	files := map[string]bool{}
	for _, id := range b.Files {
		files[id] = true
	}
	for _, c := range b.Concepts {
		for _, t := range note.Targets(c.Note.Body, c.ID) {
			switch {
			case b.byID[t] != nil:
				g.Out[c.ID] = append(g.Out[c.ID], t)
				g.In[t] = append(g.In[t], c.ID)
			case !files[t]: // an unparseable target exists but is not traversable
				g.Dangling[c.ID] = append(g.Dangling[c.ID], t)
			}
		}
		// Reserved-file links are not edges; they dangle only if the file is missing.
		seen := map[string]bool{}
		for _, l := range note.Links(c.Note.Body) {
			p, ok := note.ReservedLink(l.Href, c.ID)
			if !ok || seen[p] {
				continue
			}
			seen[p] = true
			if info, err := os.Stat(b.Abs(p)); err != nil || !info.Mode().IsRegular() {
				g.Dangling[c.ID] = append(g.Dangling[c.ID], p)
			}
		}
	}
	return g
}
