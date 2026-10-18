---
type: idea
title: Skill routing
description: Collapse and re-scope the ritari skills so an agent reliably loads the right one.
tags: [skills, routing, agents]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-18
generated: {by: human:cndreisbach, at: 2026-10-18T19:19:44Z}
---

# Observations
- [idea] Collapse seven skills whose descriptions overlap into fewer, more distinct ones so the model routes reliably [^1]
- [problem] ritari ("whenever something learned should outlive this conversation"), ritari-notes ("whenever you are about to create or add to a note") and ritari-daily ("Use at the start and end of every working session") compete for the same trigger — remembering something [^1]
- [constraint] Pi and other Agent Skills implementations advertise only name, description and path at startup; the description alone decides whether the model loads the skill [^1]
- [constraint] Descriptions are capped at 1024 characters, but the real limit is competition — every advertised skill is a routing alternative for the same cue [^1]
- [idea] Proposed shape — one ritari skill covering recall, search and basic writing; ritari-daily for session continuity; ritari-reflect for consolidation; fold notes, tasks, ingest and lint guidance into references/ files the skill loads on demand [^1]
- [risk] Supporting files load only when the agent reads them, so a buried reference may never be read; each SKILL.md must say explicitly when to read which file [^1]
- [idea] Instrument routing before rewriting — record which skills are loaded per session and how often rit is called after being advertised [^1]
- [constraint] Shipped skills must stay portable (they install to ~/.agents/skills for any Agent Skills implementation), so no Pi-specific frontmatter such as disable-model-invocation [^1]
- [open-thread] Is the failure that agents load ritari and skip the rit calls, or that they never load it? Measure before rewriting [^1]

# Relations
- part_of [Roadmap](/projects/roadmap.md)

[^1]: Clinton, development session 2026-10-18
