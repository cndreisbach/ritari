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
Topic belongs in `tags`. Prefer a specific type when one fits; `note` is right
for anything else, including a long document. Length is never a reason to
invent a type.

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
- Only lines under an `# Observations` heading are observations. The same shape
  elsewhere (`- [WIP] Draft the intro` in a plan) is ordinary prose, and
  Markdown checkboxes (`- [ ] todo`, `- [x] done`) never are.

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

`--content` on an existing note is refused. `--force` gets past that by
replacing the **whole note**: the description, tags, `--meta` fields such as a
task's `state`, observations, and relations are all gone unless you pass them
again. To change the body, use `note edit` (next section).

## Long documents

Plans, briefs, research reports, drafts, and handoffs belong in Ritari in full.
The prose goes in the body as written: a report does not need to become dozens
of observations, and it does not need a new type. Add observations only for the
few facts worth finding with `search --category`.

Create one from standard input, a heredoc, or a staged file, so no shell
quoting is needed however long it is:

```
rit search "rollout plan"                 # is there already a note for this?
rit note write --type note --title "Rollout plan for v2" \
  --description "How v2 ships: phases, owners, and rollback." \
  --rel "part_of: API v2" --content - < /tmp/rollout-plan.md

rit note write --type note --title "Rollout plan for v2" \
  --description "How v2 ships: phases, owners, and rollback." --content - <<'EOF'
# Phases
...
EOF
```

Link it to the project or task it serves (`--rel "part_of: …"`, or
`rit note write --title "<Task>" --rel "documented_by: Rollout plan for v2"`).

To revise it:

1. `rit note read <ref>` for its `hash`, and `rit note view --body <ref>` for
   the document (the body without frontmatter, Observations, Relations, or
   source footnotes).
2. Edit that text, then replace it:
   `rit note edit <ref> --op replace_body --expect-hash <hash> --content - < draft.md`.
   Everything else in the note is kept. Add `--dry-run` to check first. If the
   hash no longer matches, someone changed the note: read it again and reapply
   your change. For a small change, `find_replace` or `replace_section` is
   enough.

Keep one authoritative copy:

- Retained prose lives in Ritari. Code, patches, fixtures, diagrams, build
  output, and a project's own README, API, or install docs stay with their
  source. A one-shot prompt or tool input can stay a temporary file.
- If a worker or review tool needs a file, export one
  (`rit note view --body <ref> > /tmp/plan.md`) and record the note's ID and
  the export's hash. Apply revisions through Ritari, then export again; never
  keep editing the export as a second copy.

## Before you write

1. **Search** — `rit search <terms>`, then by type/tag, for the note you're about
   to create.
2. **Add to the existing note** if there is one (same title, or `--title` of the
   match).
3. **Link** it to the people, projects, and decisions it touches.
