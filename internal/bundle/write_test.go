package bundle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testNow = time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)

func testEnv() Env {
	return Env{Actor: "human:jane@acme.com", Now: func() time.Time { return testNow }}
}

func jwtOpts() WriteOpts {
	return WriteOpts{
		Title: "Auth uses JWT", Type: "decision", Description: "Stateless auth.",
		Tags:    []string{"auth", "security"},
		Obs:     []string{"decision: Use JWT for stateless auth #security"},
		Rels:    []string{"depends_on: User Database Schema"},
		Sources: []string{"PR #1234 https://github.com/acme/api/pull/1234"},
	}
}

const jwtFile = `---
type: decision
title: Auth uses JWT
description: Stateless auth.
tags: [auth, security]
sources:
  - id: "1"
    resource: https://github.com/acme/api/pull/1234
    title: "PR #1234"
generated: {by: human:jane@acme.com, at: 2026-07-29T12:00:00Z}
---

# Observations
- [decision] Use JWT for stateless auth #security [^1]

# Relations
- depends_on [User Database Schema](/notes/user-database-schema.md)

[^1]: [PR #1234](https://github.com/acme/api/pull/1234)
`

func TestWriteCreate(t *testing.T) {
	root := t.TempDir()
	res, err := Write(root, testEnv(), jwtOpts())
	require.NoError(t, err)
	assert.Equal(t, WriteResult{
		Status: "created", ID: "decisions/auth-uses-jwt", Path: filepath.Join(root, "decisions", "auth-uses-jwt.md"),
		AddedObs: 1, AddedRels: 1, AddedSources: 1,
	}, res)
	assert.Equal(t, jwtFile, readFile(t, root, "decisions/auth-uses-jwt.md"))
	assert.Equal(t, "# Update Log\n\n## 2026-07-29\n\n* **Create**: [Auth uses JWT](/decisions/auth-uses-jwt.md)\n", readFile(t, root, "log.md"))
	assert.Equal(t, "# Decisions\n\n* [Auth uses JWT](/decisions/auth-uses-jwt.md) - Stateless auth.\n", readFile(t, root, "decisions/index.md"))
	assert.Contains(t, readFile(t, root, "index.md"), "* [Decisions](/decisions/index.md) (`decision`) - A choice made, with its rationale.\n")
}

func TestWriteConverges(t *testing.T) {
	root := t.TempDir()
	_, err := Write(root, testEnv(), jwtOpts())
	require.NoError(t, err)
	later := testEnv()
	later.Now = func() time.Time { return testNow.Add(time.Hour) }
	res, err := Write(root, later, jwtOpts())
	require.NoError(t, err)
	assert.Equal(t, "updated", res.Status)
	assert.Equal(t, 0, res.AddedObs+res.AddedRels+res.AddedSources)
	assert.Equal(t, jwtFile, readFile(t, root, "decisions/auth-uses-jwt.md"), "unchanged, not restamped")
	assert.Equal(t, "# Update Log\n\n## 2026-07-29\n\n* **Create**: [Auth uses JWT](/decisions/auth-uses-jwt.md)\n* **Update**: [Auth uses JWT](/decisions/auth-uses-jwt.md)\n", readFile(t, root, "log.md"))

	o := jwtOpts()
	o.Sources = []string{"https://wiki/auth"}
	res, err = Write(root, later, o)
	require.NoError(t, err)
	assert.Equal(t, 1, res.AddedSources)
	got := readFile(t, root, "decisions/auth-uses-jwt.md")
	assert.Contains(t, got, "- [decision] Use JWT for stateless auth #security [^1] [^2]\n")
	assert.Contains(t, got, "generated: {by: human:jane@acme.com, at: 2026-07-29T13:00:00Z}")
	assert.Contains(t, got, "[^2]: [https://wiki/auth](https://wiki/auth)\n")
	assert.Equal(t, 1, strings.Count(readFile(t, root, "log.md"), "* **Update**"), "identical update bullets are deduplicated")
}

func TestWriteAdditiveErrorsAndForce(t *testing.T) {
	root := t.TempDir()
	_, err := Write(root, testEnv(), jwtOpts())
	require.NoError(t, err)

	_, err = Write(root, testEnv(), WriteOpts{Title: "Auth uses JWT", Content: "x"})
	assert.ErrorContains(t, err, `note "decisions/auth-uses-jwt" already exists; --content replaces the body, so use --force`)
	_, err = Write(root, testEnv(), WriteOpts{Title: "Auth uses JWT"})
	assert.ErrorContains(t, err, "already exists and nothing to add")

	res, err := Write(root, testEnv(), WriteOpts{Title: "Auth uses JWT", Content: "x", Force: true})
	require.NoError(t, err)
	assert.Equal(t, "replaced", res.Status)
	assert.Equal(t, "decisions/auth-uses-jwt", res.ID, "replaced in place even under another type")
	assert.Equal(t, "---\ntype: note\ntitle: Auth uses JWT\ngenerated: {by: human:jane@acme.com, at: 2026-07-29T12:00:00Z}\n---\n\nx\n",
		readFile(t, root, "decisions/auth-uses-jwt.md"))
}

func TestWriteTitleFallbackAcrossTypes(t *testing.T) {
	root := t.TempDir()
	_, err := Write(root, testEnv(), jwtOpts())
	require.NoError(t, err)
	res, err := Write(root, testEnv(), WriteOpts{Title: "Auth Uses JWT", Type: "reference", Obs: []string{"fact: x"}})
	require.NoError(t, err)
	assert.Equal(t, "updated", res.Status)
	assert.Equal(t, "decisions/auth-uses-jwt", res.ID)
	_, err = os.Stat(filepath.Join(root, "references"))
	assert.True(t, os.IsNotExist(err))
}

func TestWriteFrontmatterMerge(t *testing.T) {
	root := t.TempDir()
	_, err := Write(root, testEnv(), jwtOpts())
	require.NoError(t, err)
	_, err = Write(root, testEnv(), WriteOpts{Title: "Auth uses JWT", Tags: []string{"api", "auth"}, Description: "New.", Meta: [][2]string{{"state", "shipped"}}})
	require.NoError(t, err)
	got := readFile(t, root, "decisions/auth-uses-jwt.md")
	assert.Contains(t, got, "description: New.\ntags: [auth, security, api]\n")
	assert.Contains(t, got, "state: shipped\n")
}

func TestWriteValidation(t *testing.T) {
	root := t.TempDir()
	_, err := Write(root, testEnv(), WriteOpts{Title: "!!!"})
	assert.EqualError(t, err, "note requires a title that slugifies to something")
	_, err = Write(root, testEnv(), WriteOpts{Title: "Empty"})
	assert.ErrorContains(t, err, "refusing to write an empty note")
	res, err := Write(root, testEnv(), WriteOpts{Title: "Empty", AllowEmpty: true})
	require.NoError(t, err)
	assert.Equal(t, "created", res.Status)
}

func TestWriteUnknownTypeWarns(t *testing.T) {
	root := writeFiles(t, map[string]string{"ritari.yaml": "types:\n  - name: incident\n    dir: incidents\n"})
	res, err := Write(root, testEnv(), WriteOpts{Title: "Standup", Type: "meeting", Obs: []string{"x: y"}})
	require.NoError(t, err)
	assert.Equal(t, `type "meeting" is not one this knowledge base declares (incident, note, daily); it will live in meetings/`, res.Warning)
	assert.Equal(t, "meetings/standup", res.ID)
}

func TestWriteDryRun(t *testing.T) {
	root := t.TempDir()
	o := jwtOpts()
	o.DryRun = true
	res, err := Write(root, testEnv(), o)
	require.NoError(t, err)
	assert.Equal(t, "dry-run", res.Status)
	assert.Equal(t, jwtFile, res.Rendered)
	entries, err := os.ReadDir(root)
	require.NoError(t, err)
	for _, e := range entries {
		assert.Equal(t, LockFile, e.Name(), "dry run writes nothing but the lock file")
	}
}

func TestComposeRelation(t *testing.T) {
	b := openFiles(t, map[string]string{"people/jane.md": "---\ntype: person\ntitle: Jane Doe\n---\n"})
	cases := map[string]string{
		"knows: Jane Doe":              "- knows [Jane Doe](/people/jane.md)",
		"knows: people/jane":           "- knows [Jane Doe](/people/jane.md)",
		"knows: /people/jane.md":       "- knows [Jane Doe](/people/jane.md)",
		"knows: people/bob":            "- knows [bob](/people/bob.md)",
		"Future Thing":                 "- links_to [Future Thing](/notes/future-thing.md)",
		"cites: [Spec](/refs/spec.md)": "- cites [Spec](/refs/spec.md)",
		"[Ext](https://x.y)":           "- links_to [Ext](https://x.y)",
		"- part_of: Big [Plan]":        "- part_of [Big \\[Plan\\]](/notes/big-plan.md)",
	}
	for in, want := range cases {
		got, err := b.ComposeRelation(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, got, in)
	}
	_, err := b.ComposeRelation("x: !!!")
	assert.Error(t, err)
}
