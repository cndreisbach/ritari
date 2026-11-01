---
type: task
title: Mark notes with hidden content in read, search, context, and brief
description: "Phase 3 of hidden-instruction screening: a hidden field on results that return note content, a stderr warning on raw reads, and hidden spans removed from brief."
sources:
  - id: "1"
    resource: Clinton, development session 2026-11-01
generated: {by: claude-code, at: 2026-11-01T14:45:56Z}
priority: high
state: in-progress
---

# Steps
- [ ] Span ranges and a Strip function in internal/screen
- [ ] hidden field on search, context, brief; hidden and hidden_findings on note read
- [ ] Text markers and warnings
- [ ] stderr warning on note view and read
- [ ] brief removes hidden spans
- [ ] pi-ritari search and read-body
- [ ] Spec, skills, acceptance tests

# Observations
- [state] in-progress [^1]
- [next-step] Record hidden-span byte ranges in internal/screen so brief can remove them [^1]

# Relations
- implements [Screen notes for hidden instructions](/ideas/screen-notes-for-hidden-instructions.md)

[^1]: Clinton, development session 2026-11-01
