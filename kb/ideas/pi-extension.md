---
type: idea
title: Pi extension
description: An extension that integrates Ritari into Pi.
tags: [roadmap, pi, integration]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-18
  - id: "2"
    resource: Clinton, development session 2026-10-31
  - id: "3"
    resource: Ritari suggestions email from Clinton's work agent (Pi user), 2026-10-31
generated: {by: claude-code, at: 2026-10-31T02:34:50Z}
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
- [outcome] Build-order step one (session-start injection) is covered by rit install hooks, which writes a project-scoped .pi/extensions/ hook that runs rit brief [^2]
- [context] Pi's native file-writing tool takes a Markdown string directly, while writing through Ritari needs a separate CLI step, so agents found saving documents as workspace files easier than keeping them in Ritari [^3]
- [idea] Thin document tools — write and edit — that take a note reference or creation metadata (title, type, description) plus the full Markdown body, call the rit CLI, and return the note's ID and path, whether it was created or updated, and its content hash for handoff or review [^3]
- [constraint] Reuse rit's search, validation and write behavior; never edit KB files directly or add a second persistence layer [^3]
- [constraint] Updating a document must preserve its metadata and structured sections and reject a stale expected revision; note edit --op replace_body with --expect-hash now provides both [^3]
- [constraint] Never silently write into an existing note whose title matches; keep a search or read step explicit so the agent picks the right existing note [^3]
- [decision] Leave the document tools aside for now; they are better authored inside Pi than in this repo (Clinton) [^2]

# Relations
- part_of [Roadmap](/projects/roadmap.md)
- relates_to [Store whole documents in Ritari](/tasks/store-whole-documents-in-ritari.md)

[^1]: Clinton, development session 2026-10-18
[^2]: Clinton, development session 2026-10-31
[^3]: Ritari suggestions email from Clinton's work agent (Pi user), 2026-10-31
