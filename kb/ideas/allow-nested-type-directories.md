---
type: idea
title: Allow nested type directories
description: type-directory-mismatch should accept a note in a subdirectory of its type's directory, as OKF bundles nest them.
sources:
  - id: "1"
    resource: Upstream OKF bundle test, development session 2026-10-18 (cloud)
generated: {by: claude-code, at: 2026-10-18T19:04:52Z}
priority: medium
state: proposed
---

# Observations
- [context] Google's ga4 and crypto_bitcoin bundles keep Reference notes in references/metrics/ and references/joins/; rit flags all 12 as type-directory-mismatch and suggests flattening them [^1]
- [idea] Treat <type-dir>/<any subpath> as matching; rit still creates notes at <type-dir>/<slug>.md [^1]

# Relations
- part_of [Roadmap](/projects/roadmap.md)
- relates_to [Test against the upstream OKF example bundles](/tasks/test-against-the-upstream-okf-example-bundles.md)

[^1]: Upstream OKF bundle test, development session 2026-10-18 (cloud)
