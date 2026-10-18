---
type: task
title: Make reserved-name tests portable to case-insensitive filesystems
description: Two tests cannot build their fixtures on APFS, so go test ./... is red on default macOS.
tags: [tests, portability, lint]
sources:
  - id: "1"
    resource: /daily/2026-10-18.md
    title: "PR #2 CI investigation, 2026-10-18"
generated: {by: human:cndreisbach, at: 2026-10-18T19:36:28Z}
priority: medium
state: open
---

# Steps
- [ ] Add a case-sensitivity probe for t.TempDir() to the bundle and lint tests
- [ ] Branch the case-variant assertions on it, asserting the native collision where both names are one file
- [ ] Keep the hard-link simulation of a case-insensitive file system on case-sensitive hosts
- [ ] Run go test ./... on macOS and confirm CI stays green on Linux

# Observations
- [state] open [^1]
- [problem] TestReservedNamesMatchExactly (internal/bundle) and TestReservedCasing (internal/lint) fail on a case-insensitive file system, so go test ./... is red on a default macOS checkout while CI on Linux is green [^1]
- [problem] Both fixtures write notes/Index.md alongside notes/index.md and LOG.md alongside log.md; on APFS those pairs are one file, so the second write overwrites the first and os.Link then fails with either file exists or no such file or directory [^1]
- [constraint] The rules themselves are correct — ReservedCollision uses os.SameFile and is case-insensitive-safe; only the fixtures assume both names can coexist [^1]
- [constraint] The case-sensitive assertions are still worth keeping because they prove a case variant is an ordinary concept when it is a distinct file, so the fix must skip rather than delete that coverage [^1]
- [idea] Probe t.TempDir() for case sensitivity (create a file, stat its upper-cased name) and assert the native collision on an insensitive file system instead of hard-linking a variant that cannot exist [^1]
- [next-step] Add the probe helper, branch the two assertions on it, and confirm the suite passes on macOS and still covers the case-sensitive path on Linux [^1]

[^1]: [PR #2 CI investigation, 2026-10-18](/daily/2026-10-18.md)
