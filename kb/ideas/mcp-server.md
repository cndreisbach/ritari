---
type: idea
title: MCP server
description: Expose rit's operations as Model Context Protocol tools so agents can use a knowledge base without shelling out.
tags: [roadmap, mcp, integration]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-18
generated: {by: claude-code, at: 2026-10-18T14:08:54Z}
state: proposed
---

# Observations
- [idea] Serve search, note read/write/edit, context, daily, and lint as MCP tools [^1]
- [constraint] Keep to a local stdio transport; an HTTP listener would conflict with the no-network principle (SPECIFICATION.md §1) [^1]
- [open-thread] Ship as a rit subcommand (rit mcp) or a separate binary? [^1]

# Relations
- part_of [Roadmap](/projects/roadmap.md)

[^1]: Clinton, development session 2026-10-18
