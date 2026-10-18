---
type: idea
title: Rank notes that match every search term higher
description: search sums per-term scores, so a note that repeats one term can outrank a note whose title holds every term.
sources:
  - id: "1"
    resource: Upstream OKF bundle test, development session 2026-10-18 (cloud)
generated: {by: claude-code, at: 2026-10-18T19:04:52Z}
priority: low
state: proposed
---

# Observations
- [context] In crypto_bitcoin, 'duplicate transactions' ranks tables/transactions (37) above the note titled Duplicate Transactions Metric (36), as SPECIFICATION.md 9.11 specifies [^1]
- [idea] Add a bonus when every term matches, or when every term matches the title; it changes the 9.11 contract, so decide in the spec first [^1]

# Relations
- part_of [Roadmap](/projects/roadmap.md)
- relates_to [Test against the upstream OKF example bundles](/tasks/test-against-the-upstream-okf-example-bundles.md)

[^1]: Upstream OKF bundle test, development session 2026-10-18 (cloud)
