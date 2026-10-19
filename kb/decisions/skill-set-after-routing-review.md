---
type: decision
title: Skill set after routing review
description: Keep ritari, ritari-ingest and ritari-reflect; fold the rest into references; move session start/end out of routing.
tags: [skills, routing, agents]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-19
generated: {by: claude-code, at: 2026-10-19T03:30:10Z}
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
- [outcome] rit skills install reports retired ritari-* skills and removes them with --cleanup (spec §9.20) [^1]
- [decision] Session hooks are project-scoped (a repo's .claude/settings.json or .pi/extensions/), never installed into user-level settings that fire in every project [^1]
- [context] In Pi, the agent wrote the daily note seven times but never read it, despite AGENTS.md asking it to at session start [^1]
- [decision] Agent integrations install under rit install: rit install skills now, rit install hooks next; rit skills install stays as a hidden alias [^1]
- [outcome] Consolidation drafted on feat/consolidate-skills: ritari with references/writing.md, tasks.md, daily.md and lint.md, plus ritari-ingest and ritari-reflect; spec §11 rewritten [^1]
- [decision] rit install skills installs per file and reports skills whose installed files differ as outdated; --force --cleanup is the upgrade path [^1]
- [context] Clinton chose to make the change without more routing data because the tool is a day old, and may not merge it [^1]

# Relations
- affects [Skill routing](/ideas/skill-routing.md)
- affects [Roadmap](/projects/roadmap.md)

[^1]: Clinton, development session 2026-10-19
