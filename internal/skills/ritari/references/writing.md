# Writing good notes

Read this before your first `rit note write` or `rit note edit` in a session.

## The format

```markdown
---
type: decision                      # decides the directory: decisions/
title: Auth uses JWT                # becomes the file name: auth-uses-jwt.md
description: Stateless auth via short-lived JWTs, chosen over server sessions.
tags: [auth, security]              # topic; free to change
sources:                            # where facts came from (OKF provenance)
  - id: "1"
    resource: https://github.com/acme/api/pull/1234
    title: "PR #1234 — switch to JWT"
generated: {by: claude-code, at: 2026-07-29T14:30:00Z}   # automatic
---

Server-side sessions would have meant sticky routing.   ← narrative prose

# Observations
- [decision] Use JWT for stateless auth #security [^1]  ← one fact; [^1] cites source 1
- [constraint] Tokens expire after 15m (per security audit) [^1]

# Relations
- depends_on [User Database Schema](/references/user-database-schema.md)

[^1]: [PR #1234 — switch to JWT](https://github.com/acme/api/pull/1234)  ← generated from sources
```

**Identity is the path.** `decisions/auth-uses-jwt` is the note's ID; there is
no other identifier. The type picks the directory and the title picks the file,
so choose both deliberately.

## Pick a type

Type is the most stable property of a note — it decides where the note lives.
Topic belongs in `tags`. Prefer a specific type over `note`.

| Type | Use for |
| --- | --- |
| `decision` | A choice that was made, with its rationale. |
| `task` | Multi-step work in progress, with state (see `tasks.md`). |
| `person` | A human: role, preferences, what you know about them. |
| `project` | An ongoing effort and its context. |
| `meeting` | A meeting and what came out of it. |
| `reference` | A pointer to an external resource (doc, repo, dashboard). |
| `preference` | How the user wants things done. |
| `note` | General knowledge that fits nothing more specific. |
| `daily` | A day's narrative — written with `rit daily`, not `note write`. |

A bundle may declare its own vocabulary in `ritari.yaml`. Discover it before
writing: the root `index.md` names the type beside each directory
(`* [Incidents](/incidents/index.md) (`incident`)`), and `rit types` lists them.
If you use a type the bundle doesn't declare, the write still succeeds but
returns a `warning` naming the declared types — take the hint.

## Always write a description

One sentence. It is what directory listings and search results show; a note
without one appears bare and is hard to choose from a list.

## Observations

`--obs "category: text #tag (context)"` — one fact per line.

- Reuse categories (`decision`, `constraint`, `fact`, `risk`, `preference`,
  `open-thread`…) so `rit search --category` stays useful.
- `#tags` inside an observation are searchable; `#42` and `#ff00aa` are kept as
  text (issue numbers, colors), not tags.
- A trailing `(context)` records where or why.
- Markdown checkboxes (`- [ ] todo`, `- [x] done`) are **not** observations.

## Relations are links

`--rel "type: Target"` writes `- type [Target](/path.md)`.

- Types are single tokens with underscores: `depends_on`, `part_of`, `works_with`.
- The target may be a title (`Jane Doe`), an ID (`people/jane`), or a link.
- Forward references are fine: a target that doesn't exist yet links to
  `notes/<slug>` and starts working when that note is written.
- Every markdown link in prose is a graph edge too — linking in sentences counts.

## Recording sources

- `--source` says where *this write's* facts came from. Its number is attached
  to every observation in the same write. Forms: `"Title https://url"`,
  `"[Title](url)"`, `"https://url"`, `"Call transcript /sources/call.txt"`, or
  plain text for something with no URL (`"Email from Ana, 2026-07-29"`).
- `--resource` is the canonical URI of the thing the note is *about*, not where
  a fact came from.
- Never edit the `[^n]:` footnote lines at the bottom — they are regenerated
  from `sources` on every write.

## Trust signals

- `generated` (who last changed the content, and when) is stamped automatically.
- `verified` is deliberate: `rit note verify <ref>` records that someone checked
  the content against its sources. Don't call it reflexively after writing —
  verifying your own fresh write adds nothing.

## Archiving

When something is superseded, `rit note archive <ref>`: it stays linkable but
leaves search and hygiene checks. Delete only mistakes.

## Writing again adds — but not everything merges

Repeating `note write` with an existing title adds observations, relations, and
sources without duplicating. Frontmatter flags differ:

- `--tags` **merges** with existing tags.
- `--description` and `--resource` **replace**.
- `--meta key=value` **replaces that key** (other keys are untouched).

Beware batch loops: running `note write --description "…"` over many existing
notes clobbers every description you had. Leave `--description` off when you
only mean to add facts.

## Before you write

1. **Search** — `rit search <terms>`, then by type/tag, for the note you're about
   to create.
2. **Add to the existing note** if there is one (same title, or `--title` of the
   match).
3. **Link** it to the people, projects, and decisions it touches.
