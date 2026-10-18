---
type: task
title: Stop flagging links to index files as dangling
description: Links to index.md/log.md are valid OKF but rit's graph and lint treat them as dangling.
tags: [lint, graph, spec]
sources:
  - id: "1"
    resource: rit lint on kb/, 2026-10-18
  - id: "2"
    resource: Development session 2026-10-18
generated: {by: claude-code, at: 2026-10-18T16:11:46Z}
state: done
---

# Steps

- [x] Write the rule into SPECIFICATION.md §3.6
- [x] Separate reserved-file links from concept links in internal/note
- [x] Check reserved links for existence in the bundle graph
- [x] Unit and acceptance tests
- [x] Close the task

# Observations
- [state] open [^1]
- [context] A prose link to /ideas/index.md in projects/roadmap raised a dangling-link warning, because reserved files are not concepts and the spec (§3.6) only resolves links to concepts [^1]
- [next-step] Decide in the spec how links to reserved files count (not an edge, not dangling), then fix note.ResolveLink or Graph and add tests [^1]
- [state] in-progress [^2]
- [decision] A link whose last segment is a reserved file name (index.md, log.md) names that file, not a concept: it is not a graph edge, and it is dangling only if the file does not exist [^2]
- [context] okflint and OKPH both accept links to index.md as valid; okf (abcubed3) flags them like rit did [^2]
- [state] done [^2]
- [outcome] note.ReservedLink/BundlePath split reserved-file links from concept links; the graph reports them dangling only when the file is missing (SPECIFICATION.md §3.6, scenario 17_reserved_links) [^2]
- [context] note move now rewrites every relative in-bundle link to an absolute path, so relative links to index.md and sources/ files also survive a move [^2]

# Relations
- affects [Roadmap](/projects/roadmap.md)

[^1]: rit lint on kb/, 2026-10-18
[^2]: Development session 2026-10-18
