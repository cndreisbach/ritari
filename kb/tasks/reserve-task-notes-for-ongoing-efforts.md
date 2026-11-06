---
type: task
title: Reserve task notes for ongoing efforts
description: Change the skill guidance so task notes are for ongoing efforts that need resumption context, per GitHub issue 12.
sources:
  - id: "1"
    resource: https://github.com/cndreisbach/ritari/issues/12
    title: GitHub issue 12
generated: {by: claude-code, at: 2026-11-06T15:26:20Z}
priority: medium
state: open
---

# Steps

- [x] Update the ritari SKILL.md
- [x] Update references/tasks.md, daily.md, and writing.md
- [x] Update spec section 11
- [x] Run the checks
- [ ] Commit and open a PR

# Observations
- [state] open [^1]
- [context] The skills count tool calls, delegation, and approval checkpoints as steps, so routine one-off requests get task notes [^1]
- [next-step] Update SKILL.md, the tasks, daily, and writing references, and spec section 11 [^1]
- [next-step] Commit fix/task-note-threshold and open a PR that closes issue 12
- [open-thread] AGENTS.md, README.md, and ritari-ingest still use wording that predates the new threshold; left alone because the issue does not list them

# Relations
- implements [Task notes are for ongoing efforts](/decisions/task-notes-are-for-ongoing-efforts.md)

[^1]: [GitHub issue 12](https://github.com/cndreisbach/ritari/issues/12)
