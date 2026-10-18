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
  - id: "3"
    resource: Development session 2026-10-18 (cloud)
generated: {by: claude-code, at: 2026-10-18T19:04:52Z}
archived: true
priority: medium
state: done
---

# Observations
- [state] open [^1] [^2]
- [context] rit has only been tested against bundles it wrote, so its reader agrees with its own writer by construction [^1] [^2]
- [context] Roteiro found four silent reader bugs this way, including flow-style generated: { by, at } blocks being dropped, which made human-reviewed notes read as unverified [^1] [^2]
- [next-step] Copy ga4, acme_retail, stackoverflow, and crypto_bitcoin from GoogleCloudPlatform/open-knowledge-format/bundles with license and commit recorded; assert concept counts, trust tiers, and zero lint errors [^1] [^2]
- [decision] Vendor only the markdown files, unmodified, under internal/lint/testdata/okf-upstream with the Apache 2.0 license and commit ad30107 recorded; the test lives in internal/lint/upstream_test.go [^3]
- [context] No reader bugs found: all 53 concepts parse, flow-style verified stamps give 8 human-reviewed notes in acme_retail, and lint raises no errors [^3]
- [context] Findings became ideas: Allow nested type directories, Rank notes that match every search term higher, and stale-index noise noted on OKF structure lint rules [^3]
- [state] done [^3]

# Relations
- part_of [Roadmap](/projects/roadmap.md)

[^1]: Clinton, development session 2026-10-18
[^2]: [Roteiro docs/OKF_BUNDLE.md](https://github.com/OffeneDatenmodellierung/Roteiro/blob/main/docs/OKF_BUNDLE.md)
[^3]: Development session 2026-10-18 (cloud)
