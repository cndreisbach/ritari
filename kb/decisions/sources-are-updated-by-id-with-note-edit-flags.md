---
type: decision
title: Sources are updated by ID with note edit flags
description: note edit --source-id updates a source in place, using flags rather than an op, so --op stays body-only.
sources:
  - id: "1"
    resource: https://github.com/cndreisbach/ritari/issues/15
    title: GitHub issue 15
generated: {by: claude-code, at: 2026-11-06T16:21:47Z}
---

# Observations
- [decision] Use --source-id, --source-resource, and --source-title on note edit rather than a new op or command, because sources are frontmatter and --op stays body-only (Clinton agreed) [^1]
- [decision] Refuse unknown ids, non-numeric ids, blank resources, multi-line values, and a resource that duplicates another source in the note [^1]
- [decision] Keep a gone file's provenance as a descriptor rather than adding a --source-gone flag; a single token still renders as a link, which the guidance explains [^1]
- [decision] A changed source restamps generated, since provenance is content [^1]

# Relations
- relates_to [replace_body keeps structured sections](/decisions/replace-body-keeps-structured-sections.md)

[^1]: [GitHub issue 15](https://github.com/cndreisbach/ritari/issues/15)
