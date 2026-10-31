# pi-ritari

Ritari long-term memory for [Pi](https://pi.dev): four typed tools over the
`rit` CLI, plus the session-start brief.

```bash
pi install npm:pi-ritari
```

Requires `rit` on your `PATH` (`brew install cndreisbach/ritari/ritari`). The
package version matches the CLI version it ships with.

## What you get

| Tool | Does |
| --- | --- |
| `ritari_search` | Search by words, type, tag, observation category, or relation. Returns IDs, titles, and matching snippets. |
| `ritari_read` | Read a note (frontmatter, observations, relations, sources, hash) or just its markdown body. |
| `ritari_write` | Create a note, or add observations, relations, tags, and sources to an existing one. |
| `ritari_edit` | Replace a body (with the hash you last read), append to it, or change its description or state. |

At the start of a session, and again after a compaction, the extension adds the
output of `rit brief` — recent daily notes and open tasks — to the run. It is
read-only and silent if `rit` is missing.

There are no direct file tools: every write goes through `rit`, so frontmatter,
derived footnotes, indexes, and the log stay correct, and the model never needs
to build flag strings or parse YAML.

## Which knowledge base

Every call names its project explicitly and reports it back in the result:

1. `RITARI_PROJECT`, when set;
2. the nearest `ritari.yaml` at or above the working directory, matched against
   the registry by path;
3. the registry default.

If a directory declares a `ritari.yaml` that the registry does not know, the
call fails instead of falling back, so a note is never written to a different
knowledge base than the directory declares.

## Writing safely

`ritari_write` refuses a title that already exists unless you pass its `id`, so
a match is never a silent append. `ritari_edit` refuses to replace a body
without the `expect_hash` that `ritari_read` returned, so a stale edit fails
instead of overwriting someone else's change.

## Development

```bash
npm install
npm test
npm run typecheck
```
