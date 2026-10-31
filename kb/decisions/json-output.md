---
type: decision
title: JSON output
description: Add rit --output json as newline-delimited JSON documents mirroring the flat YAML stream.
tags: [cli, output, json, pi]
sources:
  - id: "1"
    resource: Clinton, development session 2026-10-31
generated: {by: pi/~deepseek/deepseek-flash-latest, at: 2026-10-31T02:54:39Z}
state: accepted
---

# Observations
- [decision] machine mode gains json alongside text and yaml, exposed as --output json and RITARI_OUTPUT=json [^1]
- [decision] json output is one flat JSON object per line (NDJSON), the same documents, fields, order and omission rules as the YAML stream, not a JSON array [^1]
- [rationale] the extension must not carry a YAML parser, and NDJSON keeps the document-stream contract while letting a single-document result parse directly [^1]
- [constraint] raw-byte commands (note view, read, types --yaml) ignore the format in both modes, so json changes nothing there [^1]
- [constraint] the unknown-format error and its spec text must list text, yaml and json [^1]

# Relations
- affects [Build the pi-ritari package](/tasks/build-the-pi-ritari-package.md)
- affects [Pi extension](/ideas/pi-extension.md)

[^1]: Clinton, development session 2026-10-31
