---
type: task
title: Check links to non-markdown files by existence
description: rit lint reports links to existing images and other files as dangling, and never checks links into sources/.
tags: [lint, graph, bug]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-18
  - id: "2"
    resource: https://github.com/OffeneDatenmodellierung/Roteiro/blob/main/docs/OKF_BUNDLE.md
    title: Roteiro docs/OKF_BUNDLE.md
  - id: "3"
    resource: Development session 2026-10-18 (cloud)
generated: {by: claude-code, at: 2026-10-18T19:02:57Z}
archived: true
priority: high
state: done
---

# Observations
- [state] open [^1] [^2]
- [context] In a scratch bundle, `[diagram](/img/diagram.png)` to an existing file was reported as dangling, while a link to a missing /sources/gone.txt was not reported at all [^1] [^2]
- [context] Roteiro hit the same bug (its issue #778) and fixed it by asking whether the bundle contains the target, not whether it is a concept [^1] [^2]
- [next-step] Generalize the reserved-file rule in SPECIFICATION.md 3.6: a link to any non-.md file (including under sources/) is not an edge and dangles only if the file is missing; reuse note.BundlePath [^1] [^2]
- [decision] Classify a file link by its path alone: under sources/, a reserved name, or a last segment whose extension isn't .md. Slugs never contain dots, so extension-less links still name concepts, and the note package stays free of file system access [^3]
- [state] done [^3]

# Relations
- relates_to [Stop flagging links to index files as dangling](/tasks/stop-flagging-links-to-index-files-as-dangling.md)
- part_of [Roadmap](/projects/roadmap.md)

[^1]: Clinton, development session 2026-10-18
[^2]: [Roteiro docs/OKF_BUNDLE.md](https://github.com/OffeneDatenmodellierung/Roteiro/blob/main/docs/OKF_BUNDLE.md)
[^3]: Development session 2026-10-18 (cloud)
