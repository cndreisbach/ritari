---
type: idea
title: Screen notes for hidden instructions
description: Detect content hidden from a human reader, and warn when hidden text reads like an instruction to a model.
tags: [roadmap, lint, security, prompt-injection]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-18
  - id: "2"
    resource: https://github.com/darshanNhb/okf-guard
    title: okf-guard
  - id: "3"
    resource: https://github.com/OffeneDatenmodellierung/Roteiro/blob/main/docs/OKF_BUNDLE.md
    title: Roteiro docs/OKF_BUNDLE.md
generated: {by: claude-code, at: 2026-10-18T17:51:51Z}
priority: high
state: proposed
---

# Observations
- [context] Ritari feeds notes back to agents, so a planted instruction is a persistent attack; a note with an HTML-comment directive, a zero-width space, and a right-to-left override passed rit lint and was returned by rit search [^1] [^2] [^3]
- [decision] Escalate only when content is both hidden and directive; visible instruction-like text is ignored, because notes legitimately contain instructions [^1] [^2] [^3]
- [decision] Findings are warnings (from lint and from note write, daily, and edit), never refusals [^1] [^2] [^3]
- [idea] Phase 1: a hidden-content lint rule using the markdown parser: HTML comments, raw HTML with hidden or display:none, link titles and reference definitions, invisible characters (zero-width, bidi controls, Unicode tag block, control characters), mixed-script words; also escape them in text output [^1] [^2] [^3]
- [idea] Phase 2: run a small pattern bank (okf-guard's categories: instruction override, fake system marker, fake role, jailbreak, data exfiltration) over hidden spans only, reported as hidden-instruction [^1] [^2] [^3]
- [idea] Phase 3: note read, search, and context mark notes that contain hidden content, so an agent is warned before trusting them [^1] [^2] [^3]
- [constraint] Allow legitimate invisible characters: the zero-width joiner in emoji sequences and the zero-width non-joiner in Persian and Indic scripts [^1] [^2] [^3]
- [context] okf-guard screens source documents (PDF, DOCX, HTML) but treats markdown as plain text, so it misses HTML comments in notes; the ritari-ingest skill could recommend running it on raw documents [^1] [^2] [^3]
- [constraint] Visible but false statements (a fake preference) cannot be caught deterministically; provenance and trust tiers are the defense there [^1] [^2] [^3]

# Relations
- part_of [Roadmap](/projects/roadmap.md)
- relates_to [Provenance and lifecycle lint rules](/ideas/provenance-and-lifecycle-lint-rules.md)

[^1]: Clinton, development session 2026-10-18
[^2]: [okf-guard](https://github.com/darshanNhb/okf-guard)
[^3]: [Roteiro docs/OKF_BUNDLE.md](https://github.com/OffeneDatenmodellierung/Roteiro/blob/main/docs/OKF_BUNDLE.md)
