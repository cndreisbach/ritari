---
type: decision
title: Touched names targets that no longer exist
description: A daily refresh keeps a touched entry whose target is gone as plain text with its historical action, rather than dropping it.
sources:
  - id: "1"
    resource: https://github.com/cndreisbach/ritari/issues/13
    title: GitHub issue 13
generated: {by: claude-code, at: 2026-11-06T15:49:44Z}
---

# Observations
- [decision] Keep the entry as plain text, not omit it, because the daily note is a record of what happened that day [^1]
- [decision] Use the existing named form with the concept ID, keeping the action from that day [^1]
- [constraint] Archived notes still exist, so they stay linked [^1]
- [risk] A note moved after that day loses its working link on refresh, because the projection reads only that date's log and does not follow later moves [^1]

# Relations
- relates_to [replace_body keeps structured sections](/decisions/replace-body-keeps-structured-sections.md)

[^1]: [GitHub issue 13](https://github.com/cndreisbach/ritari/issues/13)
