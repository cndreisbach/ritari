---
name: ritari-reflect
description: Consolidate Ritari daily notes into durable memory — review recent days, promote lasting facts onto typed notes, and resolve open threads. Use when asked to reflect on, consolidate, or review recent work or memory, and when a scheduled reflection runs.
---

# Reflection: from daily notes to durable memory

Daily notes accumulate the raw record. Reflection turns what will still matter
into observations on typed notes, where search will find it. This builds on
the `ritari` skill; its references are in `../ritari/references/` from this
file.

## When to run

- At the end of a working session.
- On a schedule (daily or weekly).
- After a context compaction, while the material is still recent.

## Workflow

1. **Gather**
   - `rit recent --limit 30`
   - the last few days (`rit brief --days 7`, then `rit daily 2026-07-28` for a whole note, `rit ls daily`)
   - open threads: `rit search --category open-thread --include-daily`
   - active tasks: `rit search --type task --meta 'state=open|blocked'`
2. **Evaluate** each candidate: *will this still be true next month?* Durability,
   not importance — an urgent bug that's fixed by tomorrow doesn't belong; a
   quiet preference the user stated once does.
3. **Promote** durable facts as observations on **typed** notes (person,
   project, decision, preference — not another daily note):
   - search first, and add to the existing note (writes are additive);
   - cite the day: `--source "Daily note 2026-07-29 /daily/2026-07-29.md"`.
4. **Log** the reflection in today's note:
   `rit daily --note "Reflected on 07-27..07-29: promoted 4 facts to people/ana-ruiz and decisions/auth-uses-jwt."`

## Resolving open threads

When a thread from an earlier day has been settled, record the outcome where it
belongs (often a decision note), and mention it in today's narrative so the
thread visibly closes.

## Never rewrite daily notes

Daily notes are the raw record. Don't edit past days to tidy them or move facts
out — promote by *adding* to typed notes and leave the day as it was.

## Guidelines

- Distil: one fact per observation, in your own words.
- Be selective: most of a day doesn't need promoting.
- Prefer updating an existing note over creating a new one.
- Flag uncertainty: "(inferred)", "(unconfirmed)".
- Archive what is superseded (`rit note archive`) instead of leaving
  contradictions in search.
- Run `rit lint` occasionally and handle findings (read the `ritari` skill's
  `references/lint.md` first).
