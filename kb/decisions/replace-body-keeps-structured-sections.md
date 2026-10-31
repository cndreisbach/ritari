---
type: decision
title: replace_body keeps structured sections
description: note edit --op replace_body replaces only the document and keeps frontmatter, Observations, Relations, and derived footnotes.
tags: [notes, editing, documents]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-31
generated: {by: claude-code, at: 2026-10-31T02:25:11Z}
---

# Observations
- [decision] The document is the body minus Observations and Relations sections (any level) and derived footnote definitions; note view --body prints exactly that, so view, edit, replace round-trips [^1]
- [decision] Refuse content with a frontmatter block, an Observations or Relations heading, or a footnote definition that collides with a numbered source id; an exact copy of the derived definition is accepted and regenerated [^1]
- [decision] Refuse daily notes, whose narrative and # Touched belong to rit daily [^1]
- [decision] Relation lines outside the Relations section go with the old document, and the result carries a warning naming them, rather than hoisting them into Relations (a report's list of links is prose) [^1]
- [decision] The content hash is SHA-256 of the file bytes, so it matches sha256sum; --expect-hash is checked under the writer lock [^1]
- [rationale] note write --force replaced the whole note, dropping description, state, and observations, which made it unusable as a body save [^1]

# Relations
- affects [Observations only under the Observations section](/ideas/observations-only-under-the-observations-section.md)
- affects [Store whole documents in Ritari](/tasks/store-whole-documents-in-ritari.md)

[^1]: Clinton, development session 2026-10-31
