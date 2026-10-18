---
type: decision
title: Agent and model in the actor
description: Stamp an agent as <agent>/<model> when the model is known; humans never carry a model.
tags: [actor, agents, provenance, okf]
sources:
  - id: "1"
    resource: /daily/2026-10-18.md
    title: Branch feat/detect-pi-and-model, 2026-10-18
generated: {by: pi/~deepseek/deepseek-flash-latest, at: 2026-10-18T19:57:35Z}
---

# Observations
- [decision] An agent's actor carries its model as <agent>/<model> (pi/deepseek-flash-latest), matching OKF's producer/version convention, instead of a separate generated.model field [^1]
- [rationale] OKF §7 already defines <producer>/<version> for tools and okflint warns about bare agent names (S206); the model says which producer actually ran [^1]
- [rationale] A separate generated.model key would need a spec extension and would be dropped by other OKF readers, while the actor string is the portable place for it [^1]
- [constraint] RITARI_AGENT_MODEL overrides the model for any agent so tests and CI can pin it; otherwise only an agent that names its own model contributes one (Pi sets PI_MODEL) [^1]
- [constraint] A human identity never carries a model — the author and detected-human paths are unchanged [^1]
- [constraint] When no model is known the actor stays the bare agent name, so existing stamps and every other agent are unaffected [^1]

# Relations
- affects [Detect Pi as an agent](/tasks/detect-pi-as-an-agent.md)

[^1]: [Branch feat/detect-pi-and-model, 2026-10-18](/daily/2026-10-18.md)
