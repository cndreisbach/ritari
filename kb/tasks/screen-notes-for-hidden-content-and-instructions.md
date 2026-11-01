---
type: task
title: Screen notes for hidden content and instructions
description: "Phase 1 and 2 of hidden-instruction screening: hidden-content and hidden-instruction lint rules, write warnings, and escaping in text output."
sources:
  - id: "1"
    resource: Clinton, development session 2026-11-01
  - id: "2"
    resource: Branch claude/wonderful-feynman-mres83
  - id: "3"
    resource: "Adversarial review of PR #9"
generated: {by: claude-code, at: 2026-11-01T13:50:30Z}
priority: high
state: done
---

# Steps

- [x] Scanner for hidden spans and invisible characters (internal/screen)
- [x] Pattern bank over hidden spans
- [x] Lint rules hidden-content and hidden-instruction
- [x] Warnings on note write, note edit, and daily
- [x] Escape invisible characters in text output
- [x] Spec, skills, and acceptance tests

# Observations
- [state] in-progress [^1]
- [context] Phase 3 (marking notes in read, search, and context) is deferred to a later change [^1]
- [next-step] Add internal/screen with the scanner and pattern bank [^1]
- [state] done [^2]
- [context] Implemented as internal/screen, used by lint, the write paths, and Ctx.Printf; spec section 8.1 and acceptance scenario 23_hidden.txtar [^2]
- [context] The kb and the upstream OKF example bundles lint clean under both rules, so no false positives there [^2]
- [context] An adversarial review closed flag-wrapped tag smuggling, unquoted and other hiding CSS, multi-paragraph and nested hidden elements, script/style/template/noscript, attribute values, more invisible characters, and full frontmatter coverage; it also removed false positives for Thai ZWSP, Greek notation, attribute values containing hidden, and a leading BOM [^3]
- [context] Lookalike-letter folding and HTML entity decoding in the pattern bank were left out to avoid a golang.org/x/text dependency [^3]

# Relations
- implements [Screen notes for hidden instructions](/ideas/screen-notes-for-hidden-instructions.md)

[^1]: Clinton, development session 2026-11-01
[^2]: Branch claude/wonderful-feynman-mres83
[^3]: Adversarial review of PR #9
