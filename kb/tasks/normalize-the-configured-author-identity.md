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
generated: {by: human:cndreisbach, at: 2026-10-18T19:29:58Z}
priority: medium
state: done
---

# Steps

- [x] Decide between prefix-on-save, prefix-on-stamp, and reject-at-set
- [x] Implement the chosen behavior in internal/actor, with config author normalizing on save
- [x] Add unit tests for a bare value, an already-prefixed value, an empty value, and a dangling prefix
- [x] Add a testscript scenario covering rit config author then a note write
- [x] Update SPECIFICATION.md §9.4, §9.19 and §12, and the README

# Observations

- [problem] rit config author stored the value verbatim and Detect returned it unchanged, so a bare cndreisbach was stamped as the actor and could not be told apart from an agent name [^1]
- [problem] SPECIFICATION.md §9.19 said author values are conventionally human:<id>, and §9.4 step 3 passed the configured author through untouched, so nothing enforced the convention [^1]
- [context] The registry in this environment held author cndreisbach with no prefix, and every note written this session was attributed to it [^1]
- [constraint] Agent names are bare, so human: is the only marker that separates people from agents in generated.by [^1]
- [context] Correcting six existing notes required a line-level rewrite of generated.by, because no rit command can change an existing stamp: note edit --meta generated=… would add a duplicate key, and any content edit restamps the time too [^2]
- [decision] Normalize in both places — config author stores the prefixed value, and actor.Detect normalizes again, so hand-edited registries and existing bare configs stay safe [^1]
- [decision] Prefix rather than reject — bare values keep working, and the convention is enforced at the boundary instead of breaking existing configurations [^1]
- [constraint] The rule is "no colon means a person"; human: and process: pass through, and a dangling prefix such as human: counts as unset [^1]
- [outcome] actor.Normalize added and used by Detect; config author normalizes on save; spec §9.4, §9.19 and scenario 15 plus the README describe it; unit and testscript coverage added [^1]
- [state] done [^1]

[^1]: [Ritari KB review, session 2026-10-18](/daily/2026-10-18.md)
[^2]: Repair of existing notes, 2026-10-18
