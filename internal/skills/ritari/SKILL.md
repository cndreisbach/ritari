---
name: ritari
description: Long-term memory across sessions with the `rit` CLI. Use when the user asks what we know, decided, or did before; asks you to remember, note, or record something; when you pick up earlier work or track a multi-step task; when you write the day's notes; or when asked to check or clean up the knowledge base.
---

# Ritari: long-term memory with `rit`

`rit` keeps a knowledge base (a **bundle**) as plain markdown files in the Open
Knowledge Format (OKF): one file per concept, YAML frontmatter on top. There is
no database and no index that can drift — every listing, search, and link graph
is computed from the files on demand. The files are the source of truth.

## Read the reference before the command

This file covers driving `rit`. The details of doing each job well are in
`references/`, next to this file. Each is short; read the whole file at the
moment named here, not before.

| Before you… | Read |
| --- | --- |
| run your first `rit note write` or `rit note edit` in a session | `references/writing.md`: types, descriptions, observations, relations, sources |
| create or update a note of type `task` | `references/tasks.md`: state, steps, writing for resumability |
| write to the daily note (`rit daily --note` / `--obs`), usually just before finishing | `references/daily.md`: what goes in the narrative, observations, and tasks |
| run `rit lint`, or check or clean up a knowledge base | `references/lint.md`: which findings to fix and which to ask about |

Two jobs have their own skills: **`ritari-ingest`** for turning a document,
transcript, email, or page into notes, and **`ritari-reflect`** for
consolidating recent daily notes into durable ones.

## A session

1. **Start**: `rit brief` prints the latest daily notes and open tasks. If your
   context already holds its output (a session hook installed with
   `rit install hooks` runs it), don't run it again. Pick up open threads and
   open tasks before starting something new.
2. **Work**: search before you write; record decisions and facts as they come
   up; keep a task note current for work that spans steps or sessions.
3. **Finish**: update the task's state, write the daily note's narrative and
   open threads, and run `rit lint` if you changed several notes.

## Output

Run without a terminal (as agents do), `rit` prints **flat YAML**: a single
document, or a stream of documents separated by `---` (a header with `success`
and `count`, then one document per record). Failures print `success: false` and
`error: …` and exit non-zero. Add `--output text` (or `-o text`) for human prose.
`rit note view` and `rit read` always print the note's raw markdown bytes.

## Values that start with `-`

A flag value beginning with `-` must use `--flag=value`. Markdown bullets hit
this constantly:

```
rit note edit tasks/ship-v2 --op append --content="- [ ] write release notes"
rit note edit tasks/ship-v2 --op find_replace --find="- [ ] tag release" --content="- [x] tag release"
```

`--content -` (and `--note -` on `daily`) reads standard input instead.

## Where to look

```
rit brief             # latest daily notes and open tasks
rit recent            # what changed lately
rit ls decisions      # concept IDs under a directory
```

The bundle root's `index.md` lists every directory with the type that lives
there; each directory's `index.md` lists its notes with one-line descriptions.

## Search before you write

Always look for an existing note before creating one — writing to an existing
title *adds* to it, and a near-duplicate under another title forks your memory.

```
rit search jwt auth                       # every term must match
rit search --type decision --tags auth    # filters alone: newest first
rit search token --category constraint    # notes with a [constraint] observation
rit search --relation depends_on
rit search --meta state=open --meta 'priority=high|urgent'
rit search jwt --include-daily --include-archived --limit 50
```

Matching is by whole word or by prefix (terms of 3+ letters: `tok` finds
`tokens`); there is no stemming, so try variants (`deploy`, `deployment`).
Title matches rank highest, then description, tags, observations, body. Daily
and archived notes are excluded unless asked for.

## Writing

```
rit note write --title "Auth uses JWT" --type decision \
  --description "Stateless auth via short-lived JWTs." --tags auth,security \
  --obs "decision: Use JWT for stateless auth #security" \
  --obs "constraint: Tokens expire after 15m" \
  --rel "depends_on: User Database Schema" \
  --source "PR #1234 https://github.com/acme/api/pull/1234"
```

The type decides the directory (`decisions/auth-uses-jwt`). Writing the same
title again **adds** observations, relations, and sources without duplicating
anything already there. Read `references/writing.md` before your first write of
the session.

Ritari holds whole documents, too: plans, briefs, research, drafts, handoffs.
Keep the full text in the note's body, not a summary that links to a copy in
the workspace. Search for an existing note first, and use a declared type
(`note` is fine when nothing more specific fits).

- `--content -` reads the body of a **new** note from standard input.
- Revise an existing body with `note edit`: `--op replace_body` for the whole
  document, or a section or find/replace op for part of it.
- `note write --force` replaces the **whole note**, dropping its description,
  metadata, observations, and relations unless you pass them again. It is not
  a way to save a body.
- Never edit note files with generic file tools; go through `rit`.

## Reading and traversal

A note can be addressed three equivalent ways: by **ID** (`decisions/auth-uses-jwt`),
by **path** (`decisions/auth-uses-jwt.md`), or by **title** (`"Auth uses JWT"`).

```
rit note read "Auth uses JWT"     # parsed: observations, relations, sources, trust
rit note view decisions/auth-uses-jwt   # raw markdown
rit context decisions/auth-uses-jwt --depth 2   # follow links outward
```

## Editing and moving

`note edit` **sets**; `note write` **merges**:

- `rit note write --tags api` adds `api` to the existing tags; `rit note edit --tags api`
  replaces the whole list with `api`.
- `note edit --title` changes the stored title but never moves the file —
  identity is the path.

```
rit note edit tasks/ship-v2 --op replace_section --section Steps --content="- [x] build"
rit note edit decisions/auth-uses-jwt --description "Updated summary."
```

Body ops: `append`, `prepend`, `find_replace` (`--find`, optional `--count`),
`replace_section`, `insert_before_section`, `insert_after_section` (`--section`),
and `replace_body`, which replaces the document but keeps the frontmatter and
the Observations and Relations sections. `rit note view --body` prints the
document; `--dry-run` previews an edit; `--expect-hash` (the `hash` from
`note read`) refuses the edit if the note changed since you read it.

**`rit note move <ref> <dest>` is the only correct way to relocate a note.** It
rewrites every link that pointed at the old path. Never rename files by hand.

Retire notes with `rit note archive` (still linkable, out of search). Use
`rit note delete` only for mistakes — it breaks inbound links.

## Projects

Each knowledge base is a registered project. The active one is `--project NAME`,
else `RITARI_PROJECT`, else the registry default.

```
rit project list
rit project init work ~/notes/work --default
rit project add team ~/src/team-kb
```

## Housekeeping

`index.md` and `log.md` are generated on every write — never edit them by hand.
`rit index --check` reports drift (for example after a hand edit of a
description); `rit index` regenerates. `rit lint` reports conformance errors and
hygiene warnings; read `references/lint.md` before acting on them.

## Principles

- Files are the source of truth; regenerate, don't hand-maintain, anything derived.
- Search before writing; add to what exists.
- One fact per observation.
- Link generously — forward references to notes that don't exist yet are fine.
