---
type: idea
title: Observations only under the Observations section
description: "Parse observation lines only inside # Observations, so prose bullets like - [WIP] stay prose."
tags: [roadmap, observations, format]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-31
generated: {by: claude-code, at: 2026-10-31T02:25:11Z}
priority: high
state: done
---

# Observations
- [context] Spec §3.4 recognizes an observation on any bullet line of the form - [category] text anywhere in the body, including inside fenced code; only empty and x checkboxes are excluded [^1]
- [risk] Once full documents live in notes, ordinary bullets such as - [WIP] Draft, * [Optional] step or - [Q3] Revisit become observations and show up in search --category and note read [^1]
- [context] The syntax comes from Basic Memory, which Ritari replaced; OKF (§4.2) defines no observations, so narrowing the rule is Ritari's own choice and has no OKF conformance impact [^1]
- [decision] Lean toward recognizing observations only under an # Observations heading (any level, case-insensitive) (Clinton) [^1]
- [risk] Existing bundles may hold observations outside the section; they would stop counting, so lint or a migration should find them [^1]
- [open-thread] Decide whether relation lines get the same treatment, since a bullet holding only a link counts as a links_to relation anywhere in the body [^1]
- [outcome] Shipped on feat/replace-body as a breaking change; MergeBody also matches existing observations only inside the section
- [open-thread] No lint rule finds observation-shaped lines outside the section, since in documents those are usually prose; add one only if old bundles need migrating

# Relations
- part_of [Roadmap](/projects/roadmap.md)
- relates_to [Store whole documents in Ritari](/tasks/store-whole-documents-in-ritari.md)

[^1]: Clinton, development session 2026-10-31
