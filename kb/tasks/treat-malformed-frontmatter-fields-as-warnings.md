---
type: task
title: Treat malformed frontmatter fields as warnings
description: A known field with the wrong shape should be ignored and reported, not make the whole note unparseable.
tags: [lint, spec, frontmatter]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-18
  - id: "2"
    resource: Development session 2026-10-18 (cloud)
generated: {by: claude-code, at: 2026-10-18T19:01:21Z}
archived: true
priority: medium
state: done
---

# Observations
- [state] open [^1]
- [decision] Only invalid YAML, a non-mapping, or a repeated key makes a file unparseable; a malformed known field is read as absent, kept verbatim as an extra key, and reported as malformed-field (SPECIFICATION.md 3.2, 8) [^1]
- [context] Today title: {a: b} or sources: foo hides the note from every command, though OKF requires only parseable YAML and a type [^1]
- [next-step] Change the field parsers in internal/note/frontmatter.go to record malformed fields instead of erroring; add the lint rule; make a write that sets the field drop the malformed extra; update TestParseUnparseable [^1]
- [decision] Rendering keeps a malformed recognized key verbatim until its typed field is set; a malformed type suppresses the default type: note so the key isn't written twice [^2]
- [context] Every write restamps generated, so a malformed generated is replaced by the next write [^2]
- [state] done [^2]

# Relations
- part_of [Roadmap](/projects/roadmap.md)
- relates_to [Provenance and lifecycle lint rules](/ideas/provenance-and-lifecycle-lint-rules.md)

[^1]: Clinton, development session 2026-10-18
[^2]: Development session 2026-10-18 (cloud)
