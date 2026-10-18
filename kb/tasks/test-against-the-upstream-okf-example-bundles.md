---
type: task
title: Test against the upstream OKF example bundles
description: Vendor the four bundles from Google's OKF spec repo as test fixtures and check rit reads, lints, and searches them correctly.
tags: [testing, interop]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-18
  - id: "2"
    resource: https://github.com/OffeneDatenmodellierung/Roteiro/blob/main/docs/OKF_BUNDLE.md
    title: Roteiro docs/OKF_BUNDLE.md
generated: {by: claude-code, at: 2026-10-18T17:45:27Z}
priority: medium
state: open
---

# Observations
- [state] open [^1] [^2]
- [context] rit has only been tested against bundles it wrote, so its reader agrees with its own writer by construction [^1] [^2]
- [context] Roteiro found four silent reader bugs this way, including flow-style generated: { by, at } blocks being dropped, which made human-reviewed notes read as unverified [^1] [^2]
- [next-step] Copy ga4, acme_retail, stackoverflow, and crypto_bitcoin from GoogleCloudPlatform/open-knowledge-format/bundles with license and commit recorded; assert concept counts, trust tiers, and zero lint errors [^1] [^2]

# Relations
- part_of [Roadmap](/projects/roadmap.md)

[^1]: Clinton, development session 2026-10-18
[^2]: [Roteiro docs/OKF_BUNDLE.md](https://github.com/OffeneDatenmodellierung/Roteiro/blob/main/docs/OKF_BUNDLE.md)
