---
type: decision
title: Skill set after routing review
description: Keep ritari, ritari-ingest and ritari-reflect; fold the rest into references; move session start/end out of routing.
tags: [skills, routing, agents]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-19
generated: {by: claude-code, at: 2026-10-19T02:48:56Z}
---

# Observations
- [decision] Ship three skills — ritari (recall, search, write; notes, tasks, daily and lint guidance as references/ files), ritari-ingest, and ritari-reflect [^1]
- [decision] ritari-ingest stays separate because users say things that match its cue (handing over material to capture) [^1]
- [decision] Session start and end are not a cue the model can route on; trigger daily-note reads and writes from AGENTS.md and agent hooks, not a skill description [^1]
- [decision] rit skills install warns about any installed ritari-* skill it does not ship, and removes them only with --cleanup [^1]
- [context] The current skills already work fairly well, so this is worth doing but not urgent [^1]
- [constraint] go:embed */SKILL.md and the single-file Install must grow to ship references/ directories [^1]
- [constraint] Spec §11 requires seven skills, so the change rewrites §11 [^1]
- [next-step] Run mise run skill-usage over local Pi and Claude Code transcripts before rewriting the skills [^1]

# Relations
- affects [Skill routing](/ideas/skill-routing.md)
- affects [Roadmap](/projects/roadmap.md)

[^1]: Clinton, development session 2026-10-19
