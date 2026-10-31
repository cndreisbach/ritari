---
type: task
title: Build the pi-ritari package
description: Ship the Pi extension as an in-repo npm package with typed tools, brief injection, and JSON CLI output.
tags: [pi, extension, npm]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-31
generated: {by: pi/~deepseek/deepseek-flash-latest, at: 2026-10-31T03:33:13Z}
archived: true
state: done
---

# Observations
- [state] open [^1]
- [goal] an npm-installable Pi package that registers typed rit tools and injects rit brief, versioned with the CLI [^1]
- [next-step] add rit --output json [^1]
- [next-step] create pi-ritari/ with package.json, extensions/ritari.ts, TypeBox tool schemas, tests, and a spec section [^1]
- [next-step] wire CI and the release workflow to publish pi-ritari when it changed since the previous tag [^1]
- [outcome] rit --output json shipped as newline-delimited flat JSON documents (spec §9.1, README, skill, unit and scenario 22 tests) [^1]
- [open-thread] project resolution for a globally installed package — refuse the registry default and require RITARI_PROJECT or a registered ritari.yaml, or accept the default and rely on the user [^1]
- [open-thread] tool set — start with rit_search, rit_read, rit_write, rit_edit, or a single rit tool with an action discriminator [^1]
- [open-thread] whether to commit the generated .pi/extensions/ritari.ts or gitignore it like bin/ [^1]
- [outcome] the package is built and tested locally; four tools, brief injection, project resolution, and vitest suites; CI and the release workflow are wired [^1]
- [outcome] rit install hooks no longer writes the Pi extension and this project's generated copy was removed; .pi/ is git-ignored [^1]
- [next-step] merge feat/pi-ritari, then try pi install ./pi-ritari in a real Pi session against this kb [^1]
- [next-step] add NPM_TOKEN before a release that touches pi-ritari [^1]
- [next-step] decide whether the end-of-run lint and open-thread report (build order step three) is worth adding to the package [^1]
- [outcome] pi install ./pi-ritari loads the package — pi list shows the local path and the session brief comes from the package (the generated hook and .pi/ are gone) [^1]
- [problem] the tools failed in the real session because rit on PATH is the released 3.0.0, which predates --output json; the brief still worked because it asks for text [^1]
- [fix] tools now report a version mismatch instead of a bare unknown-output-format error, and a local test needs the repo's bin/ first on PATH before starting Pi [^1]
- [next-step] after the release, pi-ritari and rit share the tag version, so this mismatch cannot happen from npm installs [^1]
- [outcome] all four tools ran in a real Pi session once this repo's bin/ was first on PATH — search, read, a guarded write, and an edit dry-run [^1]
- [outcome] the write guard held from the model's side too: a call with no observations was refused by rit rather than writing nothing silently [^1]
- [outcome] the package version is 0.0.0-placeholder until the release stamps the tag; the release workflow now uses trusted publishing instead of NPM_TOKEN [^1]
- [next-step] manual first publish to claim pi-ritari on npm, then trusted publisher setup, then merge [^1]
- [outcome] the name is reserved on npm and the trusted publisher is configured; the earlier NPM_TOKEN next-step is resolved and no secret is needed [^1]
- [outcome] the next release is v4.0.0, which will publish pi-ritari@4.0.0 and move the latest dist-tag off the stub [^1]
- [next-step] push feat/pi-ritari, open the PR, merge, and watch the release publish both the CLI and the package [^1]
- [next-step] after the release, upgrade the local rit, unpublish 0.0.0-stage, and re-run pi install npm:pi-ritari to test the published package [^1]
- [outcome] push and CI done — PR #8 (feat/pi-ritari) passes the Go tests, golangci-lint, govulncheck, npm ci, tsc, and vitest [^1]
- [next-step] merge PR #8 with a merge commit; the release job then publishes rit v4.0.0 and pi-ritari@4.0.0 [^1]
- [outcome] merged as PR #8 and released as v4.0.0; the release workflow published pi-ritari@4.0.0 with provenance in the same run and moved npm latest to it [^1]
- [outcome] verified the published artifact — 7 files, pi manifest, peer deps with *, no deprecation, latest = 4.0.0, next = 0.0.0-placeholder [^1]
- [next-step] on the local machine, brew upgrade ritari and replace the local package with npm:pi-ritari [^1]

# Relations
- implements [Pi extension](/ideas/pi-extension.md)
- part_of [Roadmap](/projects/roadmap.md)

[^1]: Clinton, development session 2026-10-31
