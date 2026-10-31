---
type: task
title: Build the pi-ritari package
description: Ship the Pi extension as an in-repo npm package with typed tools, brief injection, and JSON CLI output.
tags: [pi, extension, npm]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-31
generated: {by: pi/~deepseek/deepseek-flash-latest, at: 2026-10-31T02:54:35Z}
state: open
---

# Observations
- [state] open [^1]
- [goal] an npm-installable Pi package that registers typed rit tools and injects rit brief, versioned with the CLI [^1]
- [next-step] add rit --output json [^1]
- [next-step] create pi-ritari/ with package.json, extensions/ritari.ts, TypeBox tool schemas, tests, and a spec section [^1]
- [next-step] wire CI and the release workflow to publish pi-ritari when it changed since the previous tag [^1]

# Relations
- implements [Pi extension](/ideas/pi-extension.md)
- part_of [Roadmap](/projects/roadmap.md)

[^1]: Clinton, development session 2026-10-31
