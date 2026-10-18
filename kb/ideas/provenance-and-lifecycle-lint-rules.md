---
type: idea
title: Provenance and lifecycle lint rules
description: Lint for broken provenance and OKF lifecycle values that hand edits can introduce.
tags: [roadmap, lint, provenance]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-18
  - id: "2"
    resource: https://github.com/W4G1/okf
    title: W4G1/okf
generated: {by: claude-code, at: 2026-10-18T17:39:45Z}
priority: medium
state: proposed
---

# Observations
- [idea] Flag footnote references with no matching source, footnotes never defined, sources never cited, and duplicate source ids [^1] [^2]
- [idea] Flag status outside draft, stable, deprecated, and links to deprecated notes [^1] [^2]
- [idea] Flag timestamps in the future and malformed stale_after values [^1] [^2]
- [context] okflint (S203-S210) and W4G1/okf (V9-V17, V29) both check these; rit's merge keeps provenance consistent, but Obsidian edits can break it [^1] [^2]

# Relations
- part_of [Roadmap](/projects/roadmap.md)

[^1]: Clinton, development session 2026-10-18
[^2]: [W4G1/okf](https://github.com/W4G1/okf)
