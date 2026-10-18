package bundle

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T { return &v }

func seed(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	_, err := Write(root, testEnv(), jwtOpts())
	require.NoError(t, err)
	_, err = Write(root, testEnv(), WriteOpts{Title: "Login flow", Type: "project", Rels: []string{"uses: Auth uses JWT"}, Content: "See [jwt](../decisions/auth-uses-jwt.md#why)."})
	require.NoError(t, err)
	return root
}

func TestEdit(t *testing.T) {
	root := seed(t)
	_, err := Edit(root, testEnv(), "decisions/auth-uses-jwt", EditOpts{})
	assert.ErrorContains(t, err, "nothing to do")

	later := testEnv()
	later.Now = func() time.Time { return testNow.Add(time.Hour) }
	res, err := Edit(root, later, "Auth uses JWT", EditOpts{Title: ptr("Renamed"), Tags: ptr([]string{"x"}), Meta: [][2]string{{"state", "done"}}})
	require.NoError(t, err)
	assert.Equal(t, Result{Status: "updated", ID: "decisions/auth-uses-jwt", Path: filepath.Join(root, "decisions", "auth-uses-jwt.md")}, res)
	got := readFile(t, root, "decisions/auth-uses-jwt.md")
	assert.Contains(t, got, "title: Renamed\n")
	assert.Contains(t, got, "tags: [x]\n")
	assert.Contains(t, got, "at: 2026-07-29T12:00:00Z", "frontmatter-only edit keeps the stamp")

	_, err = Edit(root, later, "decisions/auth-uses-jwt", EditOpts{Op: "append", Content: "More prose."})
	require.NoError(t, err)
	got = readFile(t, root, "decisions/auth-uses-jwt.md")
	assert.Contains(t, got, "at: 2026-07-29T13:00:00Z")
	assert.Contains(t, got, "More prose.\n\n[^1]: ", "derived footnotes stay at the end")
}

func TestVerify(t *testing.T) {
	root := seed(t)
	res, err := Verify(root, testEnv(), "decisions/auth-uses-jwt", "")
	require.NoError(t, err)
	assert.Equal(t, "human-reviewed", res.Trust)
	assert.Equal(t, "human:jane@acme.com", res.By)
	res, err = Verify(root, testEnv(), "projects/login-flow", "process:nightly")
	require.NoError(t, err)
	assert.Equal(t, "machine-confirmed", res.Trust)
	assert.Contains(t, readFile(t, root, "projects/login-flow.md"), "verified:\n  - {by: process:nightly, at: 2026-07-29T12:00:00Z}\n")
}

func TestArchive(t *testing.T) {
	root := seed(t)
	_, err := SetArchived(root, testEnv(), "projects/login-flow", false)
	assert.EqualError(t, err, "projects/login-flow is not archived")
	res, err := SetArchived(root, testEnv(), "projects/login-flow", true)
	require.NoError(t, err)
	assert.Equal(t, "archived", res.Status)
	assert.Contains(t, readFile(t, root, "projects/login-flow.md"), "archived: true\n")
	assert.Contains(t, readFile(t, root, "projects/index.md"), "# Archived\n\n* [Login flow]")
	_, err = SetArchived(root, testEnv(), "projects/login-flow", true)
	assert.EqualError(t, err, "projects/login-flow is already archived")
	res, err = SetArchived(root, testEnv(), "projects/login-flow", false)
	require.NoError(t, err)
	assert.Equal(t, "restored", res.Status)
	assert.NotContains(t, readFile(t, root, "projects/login-flow.md"), "archived")
	assert.Contains(t, readFile(t, root, "log.md"), "* **Archive**: [Login flow](/projects/login-flow.md)\n* **Restore**: [Login flow](/projects/login-flow.md)\n")
}

func TestDelete(t *testing.T) {
	root := seed(t)
	res, err := Delete(root, testEnv(), "decisions/auth-uses-jwt")
	require.NoError(t, err)
	assert.Equal(t, []string{"projects/login-flow"}, res.BrokeLinksIn)
	_, err = os.Stat(filepath.Join(root, "decisions"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(root, "decisions", "index.md"))
	assert.True(t, os.IsNotExist(err), "empty directory's index is removed")
	assert.Contains(t, readFile(t, root, "log.md"), "* **Delete**: decisions/auth-uses-jwt\n")
}

func TestMove(t *testing.T) {
	root := seed(t)
	// Give the moved note a relative link and a self-link.
	_, err := Edit(root, testEnv(), "decisions/auth-uses-jwt", EditOpts{Op: "append", Content: "Back to [login](../projects/login-flow.md) and [me](auth-uses-jwt.md). See [all](index.md) and [raw](../sources/call.txt)."})
	require.NoError(t, err)

	_, err = Move(root, testEnv(), "decisions/auth-uses-jwt", "decisions/auth-uses-jwt")
	assert.EqualError(t, err, "note decisions/auth-uses-jwt is already there")
	_, err = Move(root, testEnv(), "decisions/auth-uses-jwt", "/projects/login-flow.md")
	assert.EqualError(t, err, "destination already exists: projects/login-flow")

	res, err := Move(root, testEnv(), "decisions/auth-uses-jwt", "/references/auth-uses-jwt")
	require.NoError(t, err)
	assert.Equal(t, MoveResult{From: "decisions/auth-uses-jwt", To: "references/auth-uses-jwt", RewroteLinksIn: []string{"projects/login-flow"}}, res)

	login := readFile(t, root, "projects/login-flow.md")
	assert.Contains(t, login, "See [jwt](/references/auth-uses-jwt.md#why).")
	assert.Contains(t, login, "- uses [Auth uses JWT](/references/auth-uses-jwt.md)")
	moved := readFile(t, root, "references/auth-uses-jwt.md")
	assert.Contains(t, moved, "Back to [login](/projects/login-flow.md) and [me](/references/auth-uses-jwt.md).")
	assert.Contains(t, moved, "See [all](/decisions/index.md) and [raw](/sources/call.txt).", "reserved-file and sources links keep pointing where they did")
	assert.Contains(t, moved, "- depends_on [User Database Schema](/notes/user-database-schema.md)")
	_, err = os.Stat(filepath.Join(root, "decisions", "index.md"))
	assert.True(t, os.IsNotExist(err))
	assert.Contains(t, readFile(t, root, "log.md"), "* **Move**: [Auth uses JWT](/references/auth-uses-jwt.md) (was decisions/auth-uses-jwt)\n")
}

func TestDaily(t *testing.T) {
	root := t.TempDir()
	env := testEnv()
	res, err := Daily(root, env, DailyOpts{Notes: []string{"A", " ", "B"}, Obs: []string{"open-thread: x"}})
	require.NoError(t, err)
	assert.Equal(t, "created", res.Status)
	assert.Equal(t, "daily/2026-07-29", res.ID)
	assert.Equal(t, `---
type: daily
title: "2026-07-29"
generated: {by: human:jane@acme.com, at: 2026-07-29T12:00:00Z}
---

A

B

# Observations
- [open-thread] x
`, res.Content)

	later := env
	later.Now = func() time.Time { return testNow.Add(time.Hour) }
	res, err = Daily(root, later, DailyOpts{Notes: []string{"C"}, Description: "A day."})
	require.NoError(t, err)
	assert.Equal(t, "updated", res.Status)
	assert.Contains(t, res.Content, "description: A day.\n")
	assert.Contains(t, res.Content, "at: 2026-07-29T12:00:00Z", "daily updates are not restamped")
	assert.Contains(t, res.Content, "A\n\nB\n\nC\n\n# Observations\n")

	// Touch other notes, then refresh.
	_, err = Write(root, env, WriteOpts{Title: "Kept", Obs: []string{"a: b"}})
	require.NoError(t, err)
	_, err = Write(root, env, WriteOpts{Title: "Kept", Obs: []string{"a: c"}})
	require.NoError(t, err)
	_, err = Write(root, env, WriteOpts{Title: "Scratch", Obs: []string{"a: b"}})
	require.NoError(t, err)
	_, err = Delete(root, env, "notes/scratch")
	require.NoError(t, err)
	_, err = Write(root, env, WriteOpts{Title: "Mover", Obs: []string{"a: b"}})
	require.NoError(t, err)
	_, err = Move(root, env, "notes/mover", "references/mover")
	require.NoError(t, err)

	res, err = Daily(root, env, DailyOpts{Refresh: true, Notes: []string{"ignored"}})
	require.NoError(t, err)
	assert.Equal(t, "refreshed", res.Status)
	assert.Equal(t, 3, res.Touched)
	assert.Contains(t, res.Content, `# Touched
- [Kept](/notes/kept.md) — created
- notes/scratch — deleted
- [Mover](/references/mover.md) — created
`)
	assert.NotContains(t, res.Content, "ignored")

	res, err = Daily(root, env, DailyOpts{Date: "2026-01-01", Refresh: true})
	require.NoError(t, err)
	assert.Equal(t, DailyResult{Status: "missing", Date: "2026-01-01", ID: "daily/2026-01-01"}, res)

	_, err = Daily(root, env, DailyOpts{Date: "2026-7-1"})
	assert.EqualError(t, err, `daily notes are named by date: "2026-7-1" is not YYYY-MM-DD`)
	_, err = Daily(root, env, DailyOpts{Date: "2026-02-30"})
	assert.Error(t, err)
}

func TestTouchedResurrection(t *testing.T) {
	evs := []Event{
		{Action: ActCreate, ID: "a", Title: "A"},
		{Action: ActDelete, ID: "a"},
		{Action: ActUpdate, ID: "a", Title: "A"},
		{Action: ActMove, ID: "b", Title: "B", Detail: "was z"},
		{Action: ActUpdate, ID: "daily/x"},
		{Action: ActArchive, ID: "c", Title: "C"},
	}
	got := touchedEntries(evs, "daily/x")
	require.Len(t, got, 3)
	assert.Equal(t, touchedEntry{id: "a", title: "A", action: ActUpdate}, *got[0])
	assert.Equal(t, touchedEntry{id: "b", title: "B", action: ActMove}, *got[1])
	assert.Equal(t, "archive", pastTense(got[2].action))
}
