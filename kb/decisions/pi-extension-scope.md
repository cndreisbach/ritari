---
type: decision
title: Pi extension scope
description: Build only the parts of a Pi extension a skill cannot do; no auto-triggered memory writes.
tags: [pi, extension, skills, agents]
sources:
  - id: "1"
    resource: /daily/2026-10-18.md
    title: Clinton, development session 2026-10-18
generated: {by: human:cndreisbach, at: 2026-10-18T19:18:34Z}
---

# Observations
- [decision] A Pi extension earns its place only for what a skill cannot do — a typed tool surface, deterministic session-start context injection, and enforcement of write/move policy [^1]
- [decision] Do not auto-trigger memory writes on tool or turn events; the judgment that matters is what is durable, and a hook can only nag [^1]
- [decision] Never summon a model from session_shutdown; deterministic housekeeping there, and the judgment-heavy review behind an explicit /ritari:wrap command or a guarded agent_before_settle continuation [^1]
- [decision] Measure skill routing and how often rit is actually invoked before writing extension code [^1]
- [rationale] Pi already discovers ~/.agents/skills (the default rit skills install target) and loads AGENTS.md, so "use rit" and "read open threads first" are already solved at zero cost [^1]
- [constraint] MCP remains the wider integration bet because it serves every agent, not just Pi; the extension can later be a thin skin over rit mcp [^1]

# Relations
- affects [Pi extension](/ideas/pi-extension.md)
- affects [Skill routing](/ideas/skill-routing.md)
- affects [MCP server](/ideas/mcp-server.md)

[^1]: [Clinton, development session 2026-10-18](/daily/2026-10-18.md)
