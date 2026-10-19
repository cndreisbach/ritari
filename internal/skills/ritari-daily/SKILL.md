---
name: ritari-daily
description: Session continuity with Ritari daily notes — read today's and recent days' notes when starting, and record the narrative and open threads before finishing. Use at the start and end of every working session.
---

# Daily notes: continuity between sessions

A daily note (`daily/YYYY-MM-DD`) is the authored record of a day: what came
up, what was learned, what is unresolved. It is how the next session picks up
where this one left off.

## When starting

```
rit brief                 # the latest two daily notes and open tasks (read-only)
rit brief --days 5        # more days
rit daily 2026-07-28      # one day's whole note
rit ls daily              # which days exist
rit recent --limit 10     # what else changed lately
```

If your context already holds `rit brief` output (a session hook installed with
`rit install hooks` runs it), you have this; don't run it again. Look for open
threads from recent days before starting new work. Use `brief` rather than bare
`rit daily` just to read: `rit daily` creates today's note if it is missing.

## Before finishing

Write the narrative and anything left open:

```
rit daily --note "Debugged the token refresh race; root cause was a stale cache." \
          --note "Ana prefers small PRs; split the migration." \
          --obs "open-thread: Decide whether refresh tokens rotate on every use"
```

- `--note` is repeatable: each value becomes its own paragraph.
- Calling `rit daily` again later the same day **appends** — earlier paragraphs
  stay.
- `--note -` reads a paragraph from stdin (only once per call).
- `--description` sets a one-line summary of the day.
- `--obs`, `--rel`, `--source` work as in `note write`.

## Structure

- Prose above the first heading **is** the narrative; new `--note` paragraphs
  go at its end.
- Sections you add under your own heading (`# Investigation log`) are left
  alone by `rit daily` — a stable place for longer material.
- **Never write `# Touched`.** It is generated from the day's change log (which
  notes were created, updated, moved, deleted). Refresh it with
  `rit daily --refresh`.

## What belongs where

- **Narrative** (`--note`): what happened, in a sentence or two each.
- **Observations** on the daily note: open threads and loose facts of the day.
- **Task state belongs to the task**, not the day: update the task note (see
  `ritari-tasks`) and let `# Touched` record that you did.
- Durable facts belong on typed notes (decisions, people, projects). Daily notes
  are excluded from search by default (`--include-daily` to include them), so a
  fact left only in a daily note is hard to find later.

## Promote and close

- When a fact will still matter next month, promote it to a typed note, citing
  the day: `--source "Daily note /daily/2026-07-29.md"` (see `ritari-reflect`).
- When an open thread is resolved, say so in today's narrative and record the
  outcome where it belongs (often a decision note).
