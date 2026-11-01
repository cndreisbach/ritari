---
type: task
title: Mark notes with hidden content in read, search, context, and brief
description: "Phase 3 of hidden-instruction screening: a hidden field on results that return note content, a stderr warning on raw reads, and hidden spans removed from brief."
sources:
  - id: "1"
    resource: Clinton, development session 2026-11-01
  - id: "2"
    resource: Review rounds 1-3 of Phase 3
generated: {by: claude-code, at: 2026-11-01T15:09:15Z}
archived: true
priority: high
state: done
---

# Steps

- [x] Span ranges and a Strip function in internal/screen
- [x] hidden field on search, context, brief; hidden and hidden_findings on note read
- [x] Text markers and warnings
- [x] stderr warning on note view and read
- [x] brief removes hidden spans
- [x] pi-ritari search and read-body
- [x] Spec, skills, acceptance tests
- [x] Three rounds of Sonnet review, findings fixed

# Observations
- [state] in-progress [^1]
- [next-step] Record hidden-span byte ranges in internal/screen so brief can remove them [^1]
- [state] done [^2]
- [context] Three review rounds moved brief's removals from regexes to parser positions: LinkReferenceDefinition nodes for definitions, and a destination/title scanner matched against parser titles for link titles [^2]
- [context] Left as known limits: goldmark's own parse is quadratic on huge runs of definitions or nested quotes; removing a whole-line comment between paragraph lines can join them in brief; lookalike letters are marked, not changed [^2]

# Relations
- implements [Screen notes for hidden instructions](/ideas/screen-notes-for-hidden-instructions.md)

[^1]: Clinton, development session 2026-11-01
[^2]: Review rounds 1-3 of Phase 3
