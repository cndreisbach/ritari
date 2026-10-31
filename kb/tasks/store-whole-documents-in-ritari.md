---
type: task
title: Store whole documents in Ritari
description: "Make it easy for agents to keep plans, reports, and handoffs in Ritari in full: replace_body, hashes, and skill guidance."
tags: [documents, editing, skills]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-31
generated: {by: claude-code, at: 2026-10-31T02:25:04Z}
priority: high
state: done
---

# Observations
- [state] done [^1]
- [context] Feedback from a Pi user: agents saved substantial documents as workspace Markdown and put only a summary and link in Ritari; note write --force dry-run dropped a task's description, state, and observations [^1]
- [outcome] note edit --op replace_body, note view --body, hash in read/write/edit results, --expect-hash on write and edit, note edit --dry-run [^1]
- [outcome] ritari skill, writing.md (long documents, one authoritative copy) and tasks.md (tasks and documents) updated; spec §5.7, §9.5–9.7, §11 [^1]
- [decision] Leave the proposed Pi document tools to Pi itself for now (Clinton) [^1]
- [next-step] Review and merge feat/replace-body; the observation scoping commit is breaking and releases a major version [^1]

# Relations
- implements [Observations only under the Observations section](/ideas/observations-only-under-the-observations-section.md)
- part_of [Roadmap](/projects/roadmap.md)

[^1]: Clinton, development session 2026-10-31
