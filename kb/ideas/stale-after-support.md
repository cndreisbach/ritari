---
type: idea
title: stale_after support
description: Let notes declare when they expire (OKF stale_after) and surface expired ones.
tags: [roadmap, lifecycle]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-18
  - id: "2"
    resource: https://github.com/W4G1/okf
    title: W4G1/okf
generated: {by: claude-code, at: 2026-10-18T17:39:45Z}
priority: low
state: proposed
---

# Observations
- [idea] --stale-after on note write, a lint rule for notes past their date, and a search filter for stale notes [^1] [^2]
- [context] Low priority: Clinton doesn't expect to use it, but it's reasonable for facts that change (prices, versions) [^1] [^2]
- [context] W4G1/okf only checks staleness against a date passed in (--today) so results stay reproducible [^1] [^2]

# Relations
- part_of [Roadmap](/projects/roadmap.md)

[^1]: Clinton, development session 2026-10-18
[^2]: [W4G1/okf](https://github.com/W4G1/okf)
