---
type: decision
title: Session-start hooks
description: rit install hooks writes project-scoped hooks that run the read-only rit brief at session start.
tags: [hooks, agents, daily]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-19
  - id: "2"
    resource: Clinton, development session 2026-10-31
generated: {by: pi/~deepseek/deepseek-flash-latest, at: 2026-10-31T03:02:26Z}
---

# Observations
- [decision] rit brief prints the latest daily notes and open tasks without writing; bare rit daily creates today's note, so it can't be what a hook runs [^1]
- [decision] rit install hooks --agent claude,pi writes only under the project directory: .claude/settings.json (settings.local.json with --local) and .pi/extensions/ritari.ts [^1]
- [decision] The hook names its knowledge base at install time (rit --project NAME --output text brief), so it doesn't depend on RITARI_PROJECT when it runs [^1]
- [decision] The Claude Code hook has no matcher, so it also runs on resume, clear and compaction; the Pi extension briefs before the first turn and again after a compaction [^1]
- [decision] The Claude Code settings merge keeps every other key in file order and updates an earlier rit hook in place [^1]
- [constraint] rit must be on the agent's PATH; the hook does not go through mise [^1]
- [context] The generated Pi extension type-checks under tsc --strict against @mariozechner/pi-coding-agent, and its event logic was exercised with a fake API; it has not run inside a real Pi session [^1]
- [outcome] the generated Pi extension is superseded by the pi-ritari package; rit install hooks --agent pi was removed and --agent pi now points at the package [^2]
- [decision] rit install hooks installs only for Claude Code; the Pi integration is npm install pi-ritari [^2]
- [outcome] this project's .pi/extensions/ritari.ts was deleted and .pi/ is git-ignored like bin/ [^2]

# Relations
- affects [Skill routing](/ideas/skill-routing.md)
- affects [Pi extension](/ideas/pi-extension.md)
- affects [Skill set after routing review](/decisions/skill-set-after-routing-review.md)

[^1]: Clinton, development session 2026-10-19
[^2]: Clinton, development session 2026-10-31
