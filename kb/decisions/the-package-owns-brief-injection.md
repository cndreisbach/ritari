---
type: decision
title: The package owns brief injection
description: The pi-ritari package injects rit brief; the generated .pi/extensions hook is the no-npm fallback, never both.
tags: [pi, extension, hooks]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-31
generated: {by: pi/~deepseek/deepseek-flash-latest, at: 2026-10-31T03:02:29Z}
state: accepted
---

# Observations
- [decision] the pi-ritari package starts each session (and re-briefs after a compaction) exactly as the generated hook does, so an npm install is a complete integration [^1]
- [decision] rit install hooks --agent pi stays the lightweight fallback for users without npm, and the two are documented as alternatives, not complements [^1]
- [constraint] both paths refuse to duplicate the brief if they are somehow loaded together; the package should detect the generated hook or brief only once per session [^1]
- [constraint] the hook bakes the project name at install time, but a globally installed package cannot, so the package resolves the project at runtime and refuses rather than falling back to the personal knowledge base [^1]
- [open-thread] how the package detects the generated hook — a shared custom message type, a settings check, or a note in the install output [^1]
- [outcome] the package injects rit brief before the first turn and after a compaction, and names the project it resolved [^1]
- [outcome] the generated hook was removed from rit install hooks, so there is no second writer and no duplicate-brief detection to build [^1]
- [constraint] the brief stays silent when rit fails or the project cannot be resolved, so a broken brief never blocks a session [^1]
- [outcome] closing the earlier open thread — no hook detection is needed because the package is the only path [^1]

# Relations
- affects [Pi extension](/ideas/pi-extension.md)
- affects [Session-start hooks](/decisions/session-start-hooks.md)
- affects [Build the pi-ritari package](/tasks/build-the-pi-ritari-package.md)

[^1]: Clinton, development session 2026-10-31
