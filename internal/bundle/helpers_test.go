package bundle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// writeFiles creates a bundle in a temp dir from a map of relative path to content.
func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		require.NoError(t, os.WriteFile(p, []byte(content), 0o644))
	}
	return root
}

func openFiles(t *testing.T, files map[string]string) *Bundle {
	t.Helper()
	b, err := Open(writeFiles(t, files))
	require.NoError(t, err)
	return b
}
