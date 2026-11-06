# Tasks: work that survives context loss

Read this before you decide whether work needs a task note, or create or update
one.

A task is a note of type `task`. It holds the steps, the current state, and —
most importantly — the context someone (you, after a compaction) needs to pick
the work back up cold.

## When to create one

Create a task for an ongoing effort with meaningful context to preserve, so that
you or someone else can resume it later. The test: if this stopped now, would
picking it up take more than a line in the daily note?

Count stages of the work, not steps of execution. Tool calls, delegating to an
agent, verification, background waits, and pending confirmations don't justify a
task on their own.

**Create a task:**

- Multi-stage implementation: migrating sessions to JWT, where inventorying
  readers, adding middleware, moving endpoints, and removing the store are each
  real work.
- An investigation that will continue across sessions, whose findings and dead
  ends must not be lost: tracing an intermittent token refresh failure.

**Don't create a task:**

- A routine review, check-in, or one-off request you complete. Reviewing a PR is
  one request even when it takes several tools and a review agent. Record the
  outcome in the daily note:
  `rit daily --note "Reviewed PR #1234: approved, with one comment on retry limits."`
- A one-off request with a brief unresolved follow-up. Keep the follow-up in the
  daily note:
  `rit daily --obs "open-thread: Re-check PR #1234 once CI is fixed"`
- A document you want to retain. It gets its own note and needs no task (see
  "Tasks and documents").

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

A retained document doesn't need a task to own it. When a task does exist, it
owns the current state, approvals, blockers, and the next step.

- A task's short plan can live in its body (a `# Plan` or `# Steps` section).
- A substantial report, brief, or proposal that is useful on its own gets its
  own note. If a task exists, link the note from it:
  `rit note write --title "Migrate sessions to JWT" --rel "documented_by: JWT migration proposal"`.
  See "Long documents" in `writing.md`.
- Don't create a task just to hold prose, and don't append every version of a
  proposal to a task that is already long.
- Keep versions distinguishable. Revise a draft in place with `replace_body`.
  Once a version has been reviewed or approved, keep it as it is and write the
  next one as a new note that `supersedes` it. Point any task at the current
  one, record the approval as an observation on the task, or on the document if
  there is no task
  (`--obs "decision: Approved JWT migration proposal v2 (Ana, 2026-07-29)"`),
  and archive versions nobody needs to find any more.

## Guidelines

- One task per unit of work.
- Concrete steps ("Add JWT middleware to api/auth"), not vague ones ("auth stuff").
- Once work qualifies, externalize early — create the task before you need it,
  not after context is gone.
- Link the work: the project it's part of, decisions it implements, people involved.
- Close finished tasks; stale open tasks poison searches.
- State lives in the task, not in daily notes.
