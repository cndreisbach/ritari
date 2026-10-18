---
type: task
title: Stop flagging links to index files as dangling
description: Links to index.md/log.md are valid OKF but rit's graph and lint treat them as dangling.
tags: [lint, graph, spec]
sources:
  - id: "1"
    resource: rit lint on kb/, 2026-10-18
generated: {by: claude-code, at: 2026-10-18T14:10:04Z}
state: open
---

# Observations
- [state] open [^1]
- [context] A prose link to /ideas/index.md in projects/roadmap raised a dangling-link warning, because reserved files are not concepts and the spec (§3.6) only resolves links to concepts [^1]
- [next-step] Decide in the spec how links to reserved files count (not an edge, not dangling), then fix note.ResolveLink or Graph and add tests [^1]

[^1]: rit lint on kb/, 2026-10-18
