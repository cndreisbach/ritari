---
type: idea
title: Skill routing
description: Collapse and re-scope the ritari skills so an agent reliably loads the right one.
tags: [skills, routing, agents]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-18
  - id: "2"
    resource: Clinton, development session 2026-10-19
  - id: "3"
    resource: skill-usage report, Clinton's transcripts, 2026-10-19
generated: {by: claude-code, at: 2026-10-19T02:51:37Z}
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
- [decision] Direction agreed; see Skill set after routing review [^2]
- [context] mise run skill-usage reports skill loads, rit calls, daily-note use and unanswered memory cues from local transcripts [^2]
- [context] First skill-usage run (7 Claude Code and 2 Pi sessions, mostly ritari development): one model-routed skill load in nine sessions, and the core ritari skill was never loaded [^3]
- [context] rit was called about 300 times regardless, including a Pi session with 100+ calls and no SKILL.md read, so AGENTS.md and user instructions drive rit use more than skill descriptions [^3]
- [context] The one load was ritari-notes and ritari-tasks, after rit was already in use — loaded at the moment of writing, the pattern references/ pointers tied to commands would rely on [^3]
- [context] No session read rit daily in its first 10 tool calls, so session-start continuity is not happening [^3]
- [open-thread] Rerun skill-usage with the parser fixes and --skip-project ritari to see usage outside ritari development [^3]

# Relations
- part_of [Roadmap](/projects/roadmap.md)
- decided_by [Skill set after routing review](/decisions/skill-set-after-routing-review.md)

[^1]: Clinton, development session 2026-10-18
[^2]: Clinton, development session 2026-10-19
[^3]: skill-usage report, Clinton's transcripts, 2026-10-19
