---
type: task
title: Detect Pi as an agent
description: Recognize the Pi coding agent in actor detection instead of falling back to the configured author.
tags: [actor, agents, pi, provenance]
sources:
  - id: "1"
    resource: /daily/2026-10-18.md
    title: Ritari KB review, session 2026-10-18
generated: {by: pi/~deepseek/deepseek-flash-latest, at: 2026-10-18T19:58:13Z}
priority: medium
state: done
---

# Steps

- [x] Add pi to the agent table and RITARI_AGENT values in internal/actor
- [x] Detect AI_AGENT=pi and PI_CODING_AGENT, ordered after the existing signals
- [x] Carry the model in the actor as <agent>/<model>, with RITARI_AGENT_MODEL as an override
- [x] Cover signals, false positives, the model suffix and RITARI_AGENT=pi in actor tests
- [x] Add a testscript scenario and make the harness clear ambient agent markers
- [x] Update SPECIFICATION.md §9.4, §5.4 and §12, and the README's agent list

# Observations

- [problem] Writes made from Pi stamp the configured author instead of an agent identity, because actor.Agent only looked for rovo, claude-code, cursor and codex signals [^1]
- [problem] Pi sets AI_AGENT=pi and PI_CODING_AGENT=true, so detection has two reliable signals to use [^1]
- [context] Every note written in the 2026-10-18 session carried generated.by: cndreisbach, though pi and deepseek-flash-latest produced them [^1]
- [constraint] AI_AGENT is a generic marker other agents also set, so Pi is detected only when it names pi [^1]
- [decision] The actor carries the model as <agent>/<model>: PI_MODEL for Pi and RITARI_AGENT_MODEL as an override for any agent (see /decisions/agent-and-model-in-the-actor.md) [^1]
- [outcome] Pi is detected from PI_CODING_AGENT or AI_AGENT=pi and stamps pi/<model>; with no model signal it stamps pi. SPECIFICATION.md §9.4, §5.4 and scenario 16 plus the README describe it; unit and testscript coverage added [^1]
- [context] The acceptance harness now clears every agent and model marker before each scenario, so a developer's environment cannot change the stamped actor [^1]
- [context] Verified with the built binary: AI_AGENT=pi with PI_MODEL=~deepseek/deepseek-flash-latest stamps pi/~deepseek/deepseek-flash-latest, and an empty PI_MODEL stamps pi [^1]
- [context] SPECIFICATION.md §12 numbering was repaired in passing: author identity stays 15, agent and model is 16, and the two scenarios that collided move to 17 and 18 [^1]
- [state] done [^1]

[^1]: [Ritari KB review, session 2026-10-18](/daily/2026-10-18.md)
