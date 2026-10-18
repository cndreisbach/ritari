---
type: idea
title: rit fmt
description: Rewrite hand-edited notes into rit's canonical form, with a --check mode for CI.
tags: [roadmap, fmt, obsidian]
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
- [idea] Re-render notes as rit writes them: frontmatter key order, regenerated footnote definitions from sources, normalized blank lines [^1] [^2]
- [idea] --check reports notes that would change and exits non-zero, for CI and pre-commit [^1] [^2]
- [context] Notes edited in Obsidian drift from canonical form until rit next writes them, which then shows up as a noisy diff [^1] [^2]
- [context] Most of the code exists already in note.Render; W4G1/okf has the same command (okf fmt -w / --check) [^1] [^2]

# Relations
- part_of [Roadmap](/projects/roadmap.md)
- relates_to [Obsidian extension](/ideas/obsidian-extension.md)

[^1]: Clinton, development session 2026-10-18
[^2]: [W4G1/okf](https://github.com/W4G1/okf)
