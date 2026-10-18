---
type: idea
title: Note merge and delete with redirect
description: Merge two notes into one, and let delete point inbound links at a replacement.
tags: [roadmap, refactor, lint]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-18
  - id: "2"
    resource: https://github.com/W4G1/okf
    title: W4G1/okf
generated: {by: claude-code, at: 2026-10-18T17:39:45Z}
priority: high
state: proposed
---

# Observations
- [idea] rit note merge <from> <into>: fold observations, relations, sources, and verified events into one note and rewrite links to the merged-away note [^1] [^2]
- [idea] rit note delete --redirect-to <ref> rewrites inbound links to the replacement; --unlink turns them into plain text [^1] [^2]
- [context] lint flags duplicate-title and the ritari-lint skill suggests merging, but no command does it; agents create near-duplicates often [^1] [^2]
- [open-thread] How should narrative prose combine: append it, or refuse when both notes have prose and leave that to the agent? [^1] [^2]

# Relations
- part_of [Roadmap](/projects/roadmap.md)

[^1]: Clinton, development session 2026-10-18
[^2]: [W4G1/okf](https://github.com/W4G1/okf)
