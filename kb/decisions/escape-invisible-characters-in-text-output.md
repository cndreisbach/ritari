---
type: decision
title: Escape invisible characters in text output
description: Invisible characters are escaped in text-mode output only; note view and read stay raw.
sources:
  - id: "1"
    resource: Clinton, development session 2026-11-01
generated: {by: claude-code, at: 2026-11-01T02:57:05Z}
---

# Observations
- [rationale] Spec 9.1 promises raw bytes from view and read in both modes [^1]

# Relations
- relates_to [Screen notes for hidden instructions](/ideas/screen-notes-for-hidden-instructions.md)

[^1]: Clinton, development session 2026-11-01
