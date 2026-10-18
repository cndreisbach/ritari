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
  - id: "3"
    resource: https://openknowledge.ai/docs/plugins/okf
    title: OpenKnowledge OKF plugin rules
generated: {by: claude-code, at: 2026-10-18T18:38:31Z}
priority: medium
state: proposed
---

# Observations
- [idea] Flag footnote references with no matching source, footnotes never defined, sources never cited, and duplicate source ids [^1] [^2]
- [idea] Flag status outside draft, stable, deprecated, and links to deprecated notes [^1] [^2]
- [idea] Flag timestamps in the future and malformed stale_after values [^1] [^2]
- [context] okflint (S203-S210) and W4G1/okf (V9-V17, V29) both check these; rit's merge keeps provenance consistent, but Obsidian edits can break it [^1] [^2]
- [idea] Every provenance timestamp (generated.at, verified[].at, stale_after, sources[].last_modified, usage_window bounds) must be an ISO 8601 datetime with an explicit UTC offset; flag bare dates and offset-less datetimes [^3]
- [idea] Check the shape of usage_window, and of title, description, resource, and tags (OpenKnowledge's frontmatter-recommended) [^3]
- [open-thread] rit makes a known field with the wrong shape (title: {a: b}) unparseable, an error that hides the note, but OKF only requires parseable YAML and a type; should wrong shapes be warnings instead? [^3]
- [decision] Resolved the shape open thread: malformed known fields become warnings (tasks/treat-malformed-frontmatter-fields-as-warnings) [^1]

# Relations
- part_of [Roadmap](/projects/roadmap.md)

[^1]: Clinton, development session 2026-10-18
[^2]: [W4G1/okf](https://github.com/W4G1/okf)
[^3]: [OpenKnowledge OKF plugin rules](https://openknowledge.ai/docs/plugins/okf)
