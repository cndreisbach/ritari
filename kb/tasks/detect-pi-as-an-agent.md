---
type: task
title: Detect Pi as an agent
description: Recognize the Pi coding agent in actor detection instead of falling back to the configured author.
tags: [actor, agents, pi, provenance]
sources:
  - id: "1"
    resource: /daily/2026-10-18.md
    title: Ritari KB review, session 2026-10-18
generated: {by: cndreisbach, at: 2026-10-18T19:23:27Z}
priority: medium
state: open
---

# Steps
- [ ] Add pi to the agent table and RITARI_AGENT values in internal/actor
- [ ] Detect AI_AGENT=pi and PI_CODING_AGENT, ordered after the existing signals
- [ ] Cover both in actor tests, including RITARI_AGENT=pi and a false positive on AI_AGENT=<other>
- [ ] Update SPECIFICATION.md §9.4 and the README's agent list
- [ ] Run the testscript suite with RITARI_AGENT pinned so pinned-actor scenarios stay stable

# Observations
- [state] open [^1]
- [problem] Writes made from Pi stamp the configured author instead of an agent identity, because actor.Agent only looks for rovo, claude-code, cursor and codex signals [^1]
- [problem] Pi sets AI_AGENT=pi and PI_CODING_AGENT=true, so detection has two reliable signals to use [^1]
- [context] Every note written in the 2026-10-18 session carries generated.by: cndreisbach, though pi and deepseek-flash-latest produced them [^1]
- [expected] An agent named pi is returned when AI_AGENT=pi or PI_CODING_AGENT is set, and RITARI_AGENT accepts pi [^1]
- [expected] The actor is the bare name pi, matching the convention for other agents [^1]
- [expected] SPECIFICATION.md §9.4 lists the pi signals, the README list of recognized agents gains Pi, and actor tests cover both the env signal and RITARI_AGENT=pi [^1]
- [next-step] Add the signal and table cases in internal/actor, then update spec, README and tests [^1]

[^1]: [Ritari KB review, session 2026-10-18](/daily/2026-10-18.md)
