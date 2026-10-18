---
type: idea
title: Pi extension
description: An extension that integrates Ritari into Pi.
tags: [roadmap, pi, integration]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-18
generated: {by: cndreisbach, at: 2026-10-18T19:19:34Z}
state: proposed
---

# Observations
- [idea] Integrate Ritari with Pi through its extension mechanism [^1]
- [constraint] Scope is the Pi coding agent; Pi already discovers ~/.agents/skills, the default target of rit skills install, so "use rit" and "read open threads first" need no extension (inferred) [^1]
- [idea] Narrow the extension to what a skill cannot do — a typed tool surface, deterministic session-start injection, and enforcement of write and move policy (proposed during design review, 2026-10-18) [^1]
- [constraint] rit under an agent is a CLI with edges (flat-YAML streams, --flag=value for values starting with a dash, --content - for multi-line bodies, RITARI_PROJECT resolution); a registered tool with a TypeBox schema removes that friction and is the one structural win [^1]
- [constraint] Pi discovers ~/.agents/skills, so rit skills install already targets Pi; an --agent pi flag pointing at ~/.pi/agent/skills would be cosmetic [^1]
- [risk] Auto-triggered memory writes cannot judge durability; a hook can only inject "consider writing a note", which duplicates the skills while adding latency and nudge fatigue [^1]
- [constraint] session_shutdown is cleanup-only and summoning a model there can hang quitting; agent_before_settle may request one continuation but must be guarded (tui mode, opt-in, KB touched) because it adds a full model turn per run [^1]
- [risk] rit daily --note appends, so retries, /reload, or a skill and extension both acting duplicate narrative; observation writes dedupe, narrative does not [^1]
- [risk] The extension must resolve the project explicitly (nearest ritari.yaml or the registry) and refuse rather than fall back to the user's personal knowledge base [^1]
- [risk] A TypeScript extension in a Go repo means node tooling, a second copy of product policy that can drift, and a test story testscript cannot cover; ship it as its own repo or Pi package, or in-repo with a spec section [^1]
- [constraint] Value asymmetry — MCP serves Pi, Claude Code, Codex and others, while a Pi extension serves only Pi and could later be a thin skin over rit mcp [^1]
- [idea] Build order — session_start injection, then typed tools, then a deterministic end-of-run lint and open-thread report, then an optional guarded agent_before_settle nudge; prefer a /ritari:wrap command over auto-write [^1]
- [idea] Measure first — log how often rit is invoked and whether daily notes get written, and fix skill routing before adding code (see /ideas/skill-routing.md) [^1]

# Relations
- part_of [Roadmap](/projects/roadmap.md)

[^1]: Clinton, development session 2026-10-18
