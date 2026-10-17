package bundle

import (
	"sort"
)

// Context walks the link graph breadth-first from start, up to depth hops,
// returning the start concept and every concept reached, each once.
func (b *Bundle) Context(start *Concept, depth int) []*Concept {
	g := b.Graph()
	seen := map[string]bool{start.ID: true}
	out := []*Concept{start}
	frontier := []string{start.ID}
	for d := 0; d < depth && len(frontier) > 0; d++ {
		var next []string
		for _, id := range frontier {
			for _, t := range g.Out[id] {
				if seen[t] {
					continue
				}
				seen[t] = true
				next = append(next, t)
				out = append(out, b.byID[t])
			}
		}
		frontier = next
	}
	return out
}

// Recent lists concepts by file modification time, newest first. A
// non-positive limit means unlimited.
func (b *Bundle) Recent(limit int) []*Concept {
	cs := append([]*Concept(nil), b.Concepts...)
	sort.SliceStable(cs, func(i, j int) bool {
		if !cs[i].ModTime.Equal(cs[j].ModTime) {
			return cs[i].ModTime.After(cs[j].ModTime)
		}
		return cs[i].ID < cs[j].ID
	})
	if limit > 0 && len(cs) > limit {
		cs = cs[:limit]
	}
	return cs
}
