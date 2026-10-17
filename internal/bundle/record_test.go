package bundle

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppendLogNewestDayFirst(t *testing.T) {
	text := appendLog("", "2026-07-28", []Event{{Action: ActUpdate, ID: "a", Title: "A"}})
	assert.Equal(t, "# Update Log\n\n## 2026-07-28\n\n* **Update**: [A](/a.md)\n", text)
	text = appendLog(text, "2026-07-29", []Event{
		{Action: ActCreate, ID: "decisions/x", Title: "X [v2]"},
		{Action: ActMove, ID: "references/x", Title: "X", Detail: "was decisions/x"},
		{Action: ActDelete, ID: "notes/scratch", Title: "ignored"},
	})
	text = appendLog(text, "2026-07-29", []Event{{Action: ActCreate, ID: "decisions/x", Title: "X [v2]"}, {Action: ActArchive, ID: "b"}})
	assert.Equal(t, `# Update Log

## 2026-07-29

* **Create**: [X \[v2\]](/decisions/x.md)
* **Move**: [X](/references/x.md) (was decisions/x)
* **Delete**: notes/scratch
* **Archive**: [b](/b.md)

## 2026-07-28

* **Update**: [A](/a.md)
`, text)

	evs := LogEvents(text, "2026-07-29")
	assert.Equal(t, []Event{
		{Action: ActCreate, ID: "decisions/x", Title: "X [v2]"},
		{Action: ActMove, ID: "references/x", Title: "X", Detail: "was decisions/x"},
		{Action: ActDelete, ID: "notes/scratch"},
		{Action: ActArchive, ID: "b", Title: "b"},
	}, evs)
	assert.Equal(t, "decisions/x", movedFrom(evs[1].Detail))
	assert.Nil(t, LogEvents(text, "2026-01-01"))
}

func TestIndexes(t *testing.T) {
	b := openFiles(t, map[string]string{
		"decisions/b.md":      "---\ntype: decision\ntitle: beta\ndescription: Second.\n---\n",
		"decisions/a.md":      "---\ntype: decision\ntitle: Alpha\ndescription: \"First,\n  folded.\"\n---\n",
		"decisions/old.md":    "---\ntype: decision\ntitle: Old\narchived: true\n---\n",
		"daily/2026-07-29.md": "---\ntype: daily\ntitle: 2026-07-29\n---\n",
		"daily/2026-07-30.md": "---\ntype: daily\ntitle: 2026-07-30\n---\n",
		"daily/2025-12-31.md": "---\ntype: daily\ntitle: 2025-12-31\n---\n",
		"daily/misc.md":       "---\ntype: daily\ntitle: misc\n---\n",
		"custom/deep/x.md":    "---\ntype: thing\n---\n",
		"loose.md":            "---\ntype: note\ntitle: Loose Note\ndescription: its description\n---\n",
	})
	idx := b.Indexes()
	assert.Equal(t, `---
okf_version: "0.2"
---

# Directories

* [Deep](/custom/deep/index.md)
* [Daily](/daily/index.md) (`+"`daily`"+`) - A day's narrative: what came up, and what it touched.
* [Decisions](/decisions/index.md) (`+"`decision`"+`) - A choice made, with its rationale.

# Concepts

* [Loose Note](/loose.md) - its description
`, idx["index.md"])
	assert.Equal(t, `# Decisions

* [Alpha](/decisions/a.md) - First, folded.
* [beta](/decisions/b.md) - Second.

# Archived

* [Old](/decisions/old.md)
`, idx["decisions/index.md"])
	assert.Equal(t, `# 2026

* [2026-07-30](/daily/2026-07-30.md)
* [2026-07-29](/daily/2026-07-29.md)

# 2025

* [2025-12-31](/daily/2025-12-31.md)

# Undated

* [misc](/daily/misc.md)
`, idx["daily/index.md"])
	assert.Len(t, idx, 4)
}

func TestEmptyRootIndex(t *testing.T) {
	b := openFiles(t, map[string]string{})
	assert.Equal(t, "---\nokf_version: \"0.2\"\n---\n\n", b.Indexes()["index.md"])
}

func TestSyncIndexes(t *testing.T) {
	root := writeFiles(t, map[string]string{
		"notes/a.md":       "---\ntype: note\ntitle: A\n---\n",
		"gone/index.md":    "# Gone\n",
		".hidden/index.md": "keep\n",
	})
	b, err := Open(root)
	require.NoError(t, err)
	stale, err := b.SyncIndexes(true)
	require.NoError(t, err)
	assert.Equal(t, []string{"gone/index.md (removed)", "index.md", "notes/index.md"}, stale)
	_, err = os.Stat(filepath.Join(root, "index.md"))
	assert.True(t, os.IsNotExist(err), "check mode writes nothing")

	wrote, err := b.SyncIndexes(false)
	require.NoError(t, err)
	assert.Equal(t, stale, wrote)
	again, err := b.SyncIndexes(false)
	require.NoError(t, err)
	assert.Empty(t, again)
	assert.Equal(t, "keep\n", readFile(t, root, ".hidden/index.md"))
}

func TestLockContention(t *testing.T) {
	root := t.TempDir()
	old := LockTimeout
	LockTimeout = 200 * time.Millisecond
	defer func() { LockTimeout = old }()

	held := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error)
	go func() {
		done <- withLock(root, func() error {
			close(held)
			<-release
			return nil
		})
	}()
	<-held
	err := withLock(root, func() error { return nil })
	assert.ErrorContains(t, err, "bundle is locked by another process")
	close(release)
	require.NoError(t, <-done)
	assert.NoError(t, withLock(root, func() error { return nil }))
}

func TestWriteFileAtomic(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "a", "b.md")
	require.NoError(t, WriteFileAtomic(p, []byte("one")))
	require.NoError(t, WriteFileAtomic(p, []byte("two")))
	assert.Equal(t, "two", readFile(t, root, "a/b.md"))
	entries, err := os.ReadDir(filepath.Join(root, "a"))
	require.NoError(t, err)
	assert.Len(t, entries, 1, "no temp files left behind")
}
