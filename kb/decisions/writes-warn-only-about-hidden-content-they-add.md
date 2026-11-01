---
type: decision
title: Writes warn only about hidden content they add
description: note write, note edit, and daily report only the screen findings a write introduces; lint reports everything.
sources:
  - id: "1"
    resource: Clinton, development session 2026-11-01
generated: {by: claude-code, at: 2026-11-01T13:50:30Z}
---

# Observations
- [rationale] Reporting every finding on every write made a note with one TODO comment, or a daily note listing a hostile title, warn forever with no way to silence it [^1]
- [decision] Dry runs report their findings too, so a write can be previewed [^1]

# Relations
- relates_to [Screen notes for hidden instructions](/ideas/screen-notes-for-hidden-instructions.md)

[^1]: Clinton, development session 2026-11-01
