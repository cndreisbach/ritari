---
type: idea
title: OKF structure lint rules
description: Lint the structure OKF defines for documents and reserved files, following OpenKnowledge's OKF plugin rules.
tags: [roadmap, lint, obsidian]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-18
  - id: "2"
    resource: https://openknowledge.ai/docs/plugins/okf
    title: OpenKnowledge OKF plugin rules
generated: {by: claude-code, at: 2026-10-18T17:52:30Z}
priority: medium
state: proposed
---

# Observations
- [idea] no-wiki-links: flag `[[wiki-links]]` outside code; other OKF tools render them literally, and rit's graph does not see them as edges [^1] [^2]
- [idea] log-shape: log.md date headings must be real ISO YYYY-MM-DD dates in newest-first order [^1] [^2]
- [idea] index-shape and frontmatter-reserved-index: index.md lists links under section headings, and only the root index may carry frontmatter (okf_version alone, quoted) [^1] [^2]
- [idea] reserved-casing: flag Index.md and Log.md; OpenKnowledge matches reserved names case-sensitively [^1] [^2]
- [idea] frontmatter-computation: a document typed Attested Computation needs runtime [^1] [^2]
- [context] no-wiki-links matters most here: Clinton edits the knowledge base in Obsidian, which autocompletes `[[wiki-links]]` [^1] [^2]
- [context] stale-index already catches hand edits to generated indexes; these rules also cover bundles rit did not write [^1] [^2]
- [decision] Skip OpenKnowledge's project-no-mdx rule; Ritari will never use MDX [^1] [^2]
- [open-thread] rit matches index.md and log.md case-insensitively (SPECIFICATION.md 3.1), so it treats Index.md as reserved while OKF tools treat it as a concept missing a type; switch to exact matching? [^1] [^2]

# Relations
- part_of [Roadmap](/projects/roadmap.md)
- relates_to [Provenance and lifecycle lint rules](/ideas/provenance-and-lifecycle-lint-rules.md)

[^1]: Clinton, development session 2026-10-18
[^2]: [OpenKnowledge OKF plugin rules](https://openknowledge.ai/docs/plugins/okf)
