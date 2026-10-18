package lint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cndreisbach/ritari/internal/bundle"
)

// caseInsensitiveFS reports whether the file system holding t.TempDir() folds
// case, as APFS does by default. Fixtures that need two names differing only in
// case cannot exist there: the second write lands on the first file.
func caseInsensitiveFS(t *testing.T) bool {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "probe"), nil, 0o644))
	_, err := os.Stat(filepath.Join(dir, "PROBE"))
	return err == nil
}

// openRaw opens a bundle without generating indexes. Generating them would
// overwrite a concept named Index.md on a case-insensitive file system, which
// is exactly the fixture the reserved-casing rule needs.
func openRaw(t *testing.T, files map[string]string) *bundle.Bundle {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
	}
	b, err := bundle.Open(root)
	require.NoError(t, err)
	return b
}
