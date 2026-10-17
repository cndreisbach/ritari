---
name: ritari-lint
description: Turn `rit lint` findings into decisions — fix the mechanical ones, and ask the user about the rest with concrete options. Use when asked to check or clean up a knowledge base, or periodically during reflection.
---

# Lint: findings into decisions

`rit lint` finds problems; judgment about most of them belongs to the user.
Errors are OKF conformance failures (the bundle is broken for other tools);
warnings are hygiene opinions and never block anything.

## Workflow

1. **Run**: `rit lint` (add `--rule <name>` to focus, `--conformance` for errors only).
2. **Triage**: group findings by rule.
3. **Fix the mechanical ones yourself, and say so.**
4. **Ask about the rest**, with concrete options.
5. **Log** what was done: `rit daily --note "Lint pass: regenerated indexes, moved 2 notes, archived 1 orphan."`

## Mechanical fixes (no question needed)

- `stale-index` → `rit index`.
- `type-directory-mismatch` → run exactly the move the message names
  (`rit note move people/ann decisions/ann`); `move` rewrites inbound links.

Report each mechanical fix you made.

## Everything else: offer options

For each remaining finding, read the note first (`rit note read <id>`), then
present 2–4 concrete options with their consequences. Use the harness's
structured multiple-choice question tool if one is available; otherwise a
numbered list. Batch related findings into one question, and always offer
"skip the rest".

Per-rule guidance:

- **`orphan`** (no links in or out): link it to related notes (name which),
  archive it if it's stale, or leave it.
- **`duplicate-title`**: merge into one note (say which survives and what moves
  over), rename one with `note edit --title` (the ID doesn't change), or leave
  both if they're genuinely different.
- **`unknown-type` / `unknown-category`**: change the note's type (move it to
  the declared type's directory), add the type to `ritari.yaml` (start from
  `rit types --yaml`; a declared list *replaces* the defaults), or leave it.
- **`dangling-link`**: write the missing note (it was a forward reference),
  retarget the link to an existing note, or remove the link.
- **`stale-verification`** (content changed after verification): ask the user to
  re-verify (`rit note verify`) after reviewing the change, or leave it
  unverified-in-effect.
- **`missing-title` / `missing-description`**: propose a title or one-sentence
  description drawn from the note's content, for the user to accept or edit.
- **`missing-type` / `unparseable`** (errors): show the frontmatter problem and
  propose the corrected frontmatter; these break the bundle for other OKF tools,
  so recommend fixing them.

## Rules of thumb

- Never guess at a merge or a relation — propose, don't perform.
- Report every mechanical fix you made.
- An unresolved finding is fine; warnings are advisory.
- Read the note before proposing anything about it.
