---
type: task
title: Normalize the configured author identity
description: A bare author such as cndreisbach is stamped verbatim and looks like an agent name.
tags: [actor, config, provenance]
sources:
  - id: "1"
    resource: /daily/2026-10-18.md
    title: Ritari KB review, session 2026-10-18
  - id: "2"
    resource: Repair of existing notes, 2026-10-18
generated: {by: human:cndreisbach, at: 2026-10-18T19:25:52Z}
priority: medium
state: open
---

# Steps
- [ ] Decide between prefix-on-save, prefix-on-stamp, and reject-at-set
- [ ] Implement the chosen behavior in internal/registry or internal/actor
- [ ] Add unit tests for a bare value, an already-prefixed value, and an empty value
- [ ] Add a testscript scenario covering rit config author then a note write
- [ ] Update SPECIFICATION.md §9.4 and §9.19

# Observations
- [state] open [^1]
- [problem] rit config author stores the value verbatim and Detect returns it unchanged, so a bare cndreisbach is stamped as the actor and cannot be told apart from an agent name [^1]
- [problem] SPECIFICATION.md §9.19 says author values are conventionally human:<id>, and §9.4 step 3 passes the configured author through untouched, so nothing enforces the convention [^1]
- [context] The registry in this environment holds author cndreisbach with no prefix, and every note written this session was attributed to it [^1]
- [expected] A configured author without a prefix is stamped as human:<value>, keeping existing bare configs working [^1]
- [alternative] Reject a value without human: at config author time, which is clearer but breaks existing bare configs and scripts [^1]
- [constraint] Agent names are bare, so human: is the only marker that separates people from agents in generated.by; anything that preserves that distinction is acceptable [^1]
- [open-thread] Prefix when saving, prefix when stamping, or reject at set time? Decide before implementing [^1]
- [next-step] Choose the approach, then change config author, actor.Detect or both, and add acceptance coverage [^1]
- [context] Correcting six existing notes required a line-level rewrite of generated.by, because no rit command can change an existing stamp: note edit --meta generated=… would add a duplicate key, and any content edit restamps the time too [^2]

[^1]: [Ritari KB review, session 2026-10-18](/daily/2026-10-18.md)
[^2]: Repair of existing notes, 2026-10-18
