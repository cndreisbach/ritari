---
type: decision
title: pi-ritari lives in this repo
description: Ship the Pi extension as pi-ritari/ in the Ritari repo, named pi-ritari on npm, with its version coupled to the CLI release tag.
tags: [pi, extension, npm, packaging]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-31
generated: {by: pi/~deepseek/deepseek-flash-latest, at: 2026-10-31T03:14:31Z}
state: accepted
---

# Observations
- [decision] the package lives at pi-ritari/ in this repo and publishes to npm as pi-ritari, not a separate repo [^1]
- [decision] its version matches the CLI version from the same release tag (pi-ritari 3.1.0 pairs with rit 3.1.0), and it publishes only when pi-ritari/ changed since the previous tag [^1]
- [rationale] npm tarballs pack only the package directory, Go tooling ignores directories with no Go files, and one release keeps the wrapper and the CLI in lockstep [^1]
- [rationale] no path filtering is possible in svu, so independent npm versioning would bump on every Go-only commit [^1]
- [constraint] Pi packages are ordinary npm packages, so the manifest needs the pi-package keyword, a pi manifest, and host packages in peerDependencies with a * range [^1]
- [risk] node tooling enters CI and the release workflow; a spec section and tests are the price of keeping policy from drifting [^1]
- [outcome] pi-ritari/ is built — package.json with the pi-package keyword and pi manifest, extensions/ritari.ts, src/ (rit client, project resolution, tools, brief), and four vitest suites [^1]
- [outcome] CI runs npm ci, tsc --noEmit and vitest; the release workflow publishes pi-ritari@<tag> with provenance only when pi-ritari/ changed since the previous tag [^1]
- [next-step] add the NPM_TOKEN secret before the first release that touches pi-ritari [^1]
- [decision] publish through npm trusted publishing (OIDC) rather than an NPM_TOKEN secret; the release job already has id-token: write [^1]
- [constraint] trusted publishing needs the package to exist on npm first, so the name is claimed with a manual 0.0.0-placeholder published under the next dist-tag [^1]
- [constraint] 0.0.0-placeholder is deliberately below every release and outside latest, so nobody installs a package that needs the unreleased --output json [^1]
- [next-step] after the manual publish, configure the trusted publisher on npmjs.com for cndreisbach/ritari and release.yml [^1]

# Relations
- affects [Pi extension](/ideas/pi-extension.md)
- affects [Build the pi-ritari package](/tasks/build-the-pi-ritari-package.md)

[^1]: Clinton, development session 2026-10-31
