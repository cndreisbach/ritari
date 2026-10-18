---
type: task
title: Match reserved file names exactly
description: Recognize only lowercase index.md and log.md as reserved, and lint case variants on case-insensitive file systems.
tags: [lint, spec, portability]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-18
generated: {by: claude-code, at: 2026-10-18T18:38:31Z}
priority: medium
state: open
---

# Observations
- [state] open [^1]
- [decision] Reserved names match exactly (OKF 3.1); Index.md is an ordinary concept (SPECIFICATION.md 3.1, 10.8) [^1]
- [decision] On a case-insensitive file system, lint reports reserved-casing for names equal to index.md or log.md ignoring case, because writing the generated index would overwrite them [^1]
- [context] Clinton raised this as a macOS issue; Windows file systems are case-insensitive by default too, so detect it per bundle directory instead of by operating system [^1]
- [next-step] Make note.IsReservedName exact; add a case-insensitivity probe for the bundle root (os.SameFile on a case-swapped path) and the reserved-casing rule; update tests that expect INDEX.md to be reserved [^1]

# Relations
- part_of [Roadmap](/projects/roadmap.md)
- relates_to [OKF structure lint rules](/ideas/okf-structure-lint-rules.md)

[^1]: Clinton, development session 2026-10-18
