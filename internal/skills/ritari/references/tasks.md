# Tasks: work that survives context loss

Read this before you create or update a task note.

A task is a note of type `task`. It holds the steps, the current state, and —
most importantly — the context someone (you, after a compaction) needs to pick
the work back up cold.

## When to create one

- The work has several steps or will outlast this session.
- You might lose context partway (long investigations, big refactors).
- Someone else might resume it.

Single-step work doesn't need a task.

## Create

State is queryable in **both** places: the `state` frontmatter key (for
`search --meta`) and a `[state]` observation (readable in the body). Use
`state`, never `status` — `status` is OKF's lifecycle field (draft, stable,
deprecated).

```
rit note write --title "Migrate sessions to JWT" --type task \
  --description "Replace server sessions with JWT auth across the API." \
  --meta state=open --meta priority=high \
  --obs "state: open" \
  --obs "context: Sessions need sticky routing, which blocks autoscaling" \
  --obs "next-step: Inventory endpoints that read the session" \
  --rel "part_of: API v2" --rel "implements: Auth uses JWT" \
  --content="# Steps
- [ ] Inventory session readers
- [ ] Add JWT middleware
- [ ] Migrate endpoints
- [ ] Remove session store"
```

## Find open tasks

```
rit search --type task --meta state=open
rit search --type task --meta 'state=open|blocked'
rit search migrate --type task
```

## Update progress

- Tick steps with a section replace:

  ```
  rit note edit tasks/migrate-sessions-to-jwt --op replace_section --section Steps \
    --content="- [x] Inventory session readers
  - [ ] Add JWT middleware
  - [ ] Migrate endpoints
  - [ ] Remove session store"
  ```

- Record what you learned, with sources:
  `rit note write --title "Migrate sessions to JWT" --obs "context: 14 endpoints read the session" --source "grep of api/ at a1b2c3d"`
- Move the pointer: add a new `next-step` observation as you go.
- When the state changes, update both places:

  ```
  rit note edit tasks/migrate-sessions-to-jwt --meta state=blocked
  rit note write --title "Migrate sessions to JWT" --obs "blocked: Waiting on key rotation design"
  ```

## Complete

```
rit note edit tasks/migrate-sessions-to-jwt --meta state=done
rit note write --title "Migrate sessions to JWT" --obs "state: done" \
  --obs "outcome: All endpoints on JWT; session store removed in PR #1290"
```

Archive it later if it no longer needs to show up in searches.

## Write for resumability

- **Context beats steps.** A list of unchecked boxes is useless without *why*
  and *what you already know*. Record dead ends too.
- Reuse categories so they're searchable: `state`, `context`, `decision`,
  `blocked`, `outcome`, `next-step`.
- Write as if the reader has never seen this conversation — because after a
  compaction, they haven't.

## Tasks and documents

The task owns the current state, approvals, blockers, and the next step.

- A short plan can live in the task's body (a `# Plan` or `# Steps` section).
- A substantial report, brief, or proposal that is useful on its own gets its
  own note, linked from the task:
  `rit note write --title "Migrate sessions to JWT" --rel "documented_by: JWT migration proposal"`.
  See "Long documents" in `writing.md`.
- Don't create a task just to hold prose, and don't append every version of a
  proposal to a task that is already long.
- Keep versions distinguishable. Revise a draft in place with `replace_body`.
  Once a version has been reviewed or approved, keep it as it is and write the
  next one as a new note that `supersedes` it. Point the task at the current
  one, record the approval as an observation
  (`--obs "decision: Approved JWT migration proposal v2 (Ana, 2026-07-29)"`),
  and archive versions nobody needs to find any more.

## Guidelines

- One task per unit of work.
- Concrete steps ("Add JWT middleware to api/auth"), not vague ones ("auth stuff").
- Externalize early — create the task before you need it, not after context is gone.
- Link the work: the project it's part of, decisions it implements, people involved.
- Close finished tasks; stale open tasks poison searches.
- State lives in the task, not in daily notes.
