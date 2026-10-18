package lint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cndreisbach/ritari/internal/bundle"
)

func open(t *testing.T, files map[string]string) *bundle.Bundle {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
	}
	b, err := bundle.Open(root)
	require.NoError(t, err)
	_, err = b.SyncIndexes(false)
	require.NoError(t, err)
	b, err = bundle.Open(root)
	require.NoError(t, err)
	return b
}

func rules(fs []Finding, id string) []string {
	out := []string{}
	for _, f := range fs {
		if f.ID == id {
			out = append(out, f.Rule)
		}
	}
	return out
}

func TestCleanBundle(t *testing.T) {
	b := open(t, map[string]string{
		"decisions/a.md": "---\ntype: decision\ntitle: A\ndescription: d\n---\n\n[b](/notes/b.md)\n",
		"notes/b.md":     "---\ntype: note\ntitle: B\ndescription: d\n---\n",
	})
	rep, err := Run(b, Options{})
	require.NoError(t, err)
	assert.Empty(t, rep.Findings)
	assert.Equal(t, "", rep.Failure)
}

func TestRules(t *testing.T) {
	b := open(t, map[string]string{
		"decisions/a.md":      "---\ntype: decision\ntitle: A\ndescription: d\n---\n\n[b](/notes/b.md) [gone](/notes/gone.md)\n- [weird] x\n- [fact] y\n",
		"notes/b.md":          "---\ntype: note\ntitle: B\ndescription: d\n---\n",
		"notes/dup.md":        "---\ntype: note\ntitle: a\ndescription: d\n---\n\n[b](/notes/b.md)\n",
		"notes/untyped.md":    "---\ntitle: U\ndescription: d\n---\n\n[b](/notes/b.md)\n",
		"notes/bare.md":       "---\ntype: note\n---\n\n[b](/notes/b.md)\n",
		"notes/thing.md":      "---\ntype: gadget\ntitle: Thing\ndescription: d\n---\n\n[b](/notes/b.md)\n",
		"notes/lonely.md":     "---\ntype: note\ntitle: Lonely\ndescription: d\n---\n\n[me](/notes/lonely.md)\n",
		"notes/broken.md":     "---\ntype: [\n---\n",
		"people/ann.md":       "---\ntype: decision\ntitle: Ann\ndescription: d\n---\n\n[b](/notes/b.md)\n",
		"notes/old.md":        "---\ntype: decision\narchived: true\n---\n\n[gone](/x.md)\n",
		"notes/stale.md":      "---\ntype: note\ntitle: Stale\ndescription: d\ngenerated: {by: x, at: 2026-02-01T00:00:00Z}\nverified:\n  - {by: human:a, at: 2026-01-01T00:00:00Z}\n---\n\n[b](/notes/b.md)\n",
		"notes/fresh.md":      "---\ntype: note\ntitle: Fresh\ndescription: d\ngenerated: {by: x, at: 2026-01-01T00:00:00Z}\nverified: {by: human:a, at: 2026-02-01T00:00:00Z}\n---\n\n[b](/notes/b.md)\n",
		"daily/2026-01-01.md": "---\ntype: daily\ntitle: \"2026-01-01\"\ndescription: d\n---\n\n# Touched\n- [Orphan](/notes/orph.md) — created\n",
		"notes/orph.md":       "---\ntype: note\ntitle: Orphan\ndescription: d\n---\n",
		"notes/shape.md":      "---\ntype: note\ntitle: {a: b}\ndescription: d\nsources: foo\narchived: true\n---\n\n[b](/notes/b.md)\n",
		"notes/badtype.md":    "---\ntype: [note]\ntitle: Bad type\ndescription: d\n---\n\n[b](/notes/b.md)\n",
	})
	rep, err := Run(b, Options{Categories: []string{"fact"}})
	require.NoError(t, err)
	f := rep.Findings
	assert.Equal(t, []string{"dangling-link", "duplicate-title", "unknown-category"}, rules(f, "decisions/a"))
	assert.Equal(t, []string{"duplicate-title"}, rules(f, "notes/dup"))
	assert.Equal(t, []string{"missing-type"}, rules(f, "notes/untyped"))
	assert.Equal(t, []string{"missing-description", "missing-title"}, rules(f, "notes/bare"))
	assert.Equal(t, []string{"type-directory-mismatch", "unknown-type"}, rules(f, "notes/thing"))
	assert.Equal(t, []string{"orphan"}, rules(f, "notes/lonely"))
	assert.Equal(t, []string{"unparseable"}, rules(f, "notes/broken"))
	assert.Equal(t, []string{"type-directory-mismatch"}, rules(f, "people/ann"))
	assert.Equal(t, []string{"dangling-link"}, rules(f, "notes/old"), "archived notes skip hygiene warnings")
	assert.Equal(t, []string{"stale-verification"}, rules(f, "notes/stale"))
	assert.Equal(t, []string{}, rules(f, "notes/fresh"))
	assert.Equal(t, []string{"orphan"}, rules(f, "notes/orph"), "Touched links do not count")
	assert.Equal(t, []string{}, rules(f, "daily/2026-01-01"))
	assert.Equal(t, []string{"malformed-field", "malformed-field"}, rules(f, "notes/shape"), "archived notes still get malformed-field")
	assert.Equal(t, []string{"missing-type", "malformed-field"}, rules(f, "notes/badtype"), "a malformed type counts as missing")

	for _, x := range f {
		switch x.Rule {
		case "type-directory-mismatch":
			if x.ID == "people/ann" {
				assert.Equal(t, `type "decision" belongs in decisions/, but this note is in people/ (fix: rit note move people/ann decisions/ann)`, x.Message)
			}
		case "dangling-link":
			if x.ID == "decisions/a" {
				assert.Equal(t, "links to notes/gone, which does not exist", x.Message)
			}
		case "unparseable":
			assert.Equal(t, "error", x.Severity)
			assert.Contains(t, x.Message, "frontmatter could not be parsed: ")
		}
	}
	assert.Equal(t, 3, rep.Errors)
	assert.Equal(t, "3 conformance error(s)", rep.Failure)
	for i := 1; i < len(f); i++ {
		assert.LessOrEqual(t, f[i-1].Path, f[i].Path, "sorted by path")
	}
}

func TestOptions(t *testing.T) {
	b := open(t, map[string]string{
		"notes/bare.md":    "---\ntype: note\n---\n",
		"notes/untyped.md": "---\ntitle: U\n---\n",
	})
	rep, err := Run(b, Options{Conformance: true})
	require.NoError(t, err)
	assert.Equal(t, 1, rep.Errors)
	assert.Equal(t, 0, rep.Warnings)

	rep, err = Run(b, Options{Rules: []string{"missing-title"}})
	require.NoError(t, err)
	assert.Equal(t, 0, rep.Errors)
	assert.Equal(t, 1, rep.Warnings)
	assert.Equal(t, "", rep.Failure)

	rep, err = Run(b, Options{Rules: []string{"missing-title"}, Strict: true})
	require.NoError(t, err)
	assert.Equal(t, "1 warning(s) with --strict", rep.Failure)

	_, err = Run(b, Options{Rules: []string{"bogus"}})
	assert.ErrorContains(t, err, `unknown rule "bogus" (known: dangling-link, duplicate-title,`)
}

func TestStaleIndex(t *testing.T) {
	b := open(t, map[string]string{"notes/a.md": "---\ntype: note\ntitle: A\ndescription: d\n---\n"})
	require.NoError(t, os.WriteFile(filepath.Join(b.Root, "notes", "index.md"), []byte("hand edit\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(b.Root, "old"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(b.Root, "old", "index.md"), []byte("# Old\n"), 0o644))
	rep, err := Run(b, Options{Rules: []string{"stale-index"}})
	require.NoError(t, err)
	require.Len(t, rep.Findings, 2)
	assert.Equal(t, Finding{Severity: "warning", Rule: "stale-index", Path: "notes/index.md", Message: "generated listing is out of date (run: rit index)"}, rep.Findings[0])
	assert.Equal(t, "old/index.md", rep.Findings[1].Path)
}

func TestReservedCasing(t *testing.T) {
	b := open(t, map[string]string{
		"notes/a.md":     "---\ntype: note\ntitle: A\ndescription: d\n---\n",
		"notes/Index.md": "---\ntype: note\ntitle: Overview\ndescription: d\n---\n",
		"LOG.md":         "---\ntype: [\n---\n",
	})
	rep, err := Run(b, Options{Rules: []string{"reserved-casing"}})
	require.NoError(t, err)
	assert.Empty(t, rep.Findings, "distinct files on a case-sensitive file system do not collide")

	// A hard link stands in for a case-insensitive file system, where the two
	// names reach the same file.
	for _, p := range [][2]string{{"notes/Index.md", "notes/index.md"}, {"LOG.md", "log.md"}} {
		reserved := filepath.Join(b.Root, filepath.FromSlash(p[1]))
		if err := os.Remove(reserved); !os.IsNotExist(err) {
			require.NoError(t, err)
		}
		require.NoError(t, os.Link(filepath.Join(b.Root, filepath.FromSlash(p[0])), reserved))
	}
	b, err = bundle.Open(b.Root)
	require.NoError(t, err)
	rep, err = Run(b, Options{Rules: []string{"reserved-casing"}})
	require.NoError(t, err)
	assert.Equal(t, []Finding{
		{Severity: "warning", Rule: "reserved-casing", ID: "LOG", Path: "LOG.md", Message: "LOG.md collides with the reserved log.md on this case-insensitive file system; rename it"},
		{Severity: "warning", Rule: "reserved-casing", ID: "notes/Index", Path: "notes/Index.md", Message: "Index.md collides with the reserved index.md on this case-insensitive file system; rename it"},
	}, rep.Findings)
}
