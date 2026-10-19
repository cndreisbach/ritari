# Ritari

Ritari (`rit`) gives AI agents a long-term memory. An agent forgets everything
when a conversation ends; Ritari keeps what it learned in a knowledge base that
survives between sessions and can be searched, linked, and added to as the agent
works. People can use it too.

The knowledge base is a directory of plain markdown files in the
[Open Knowledge Format (OKF) v0.2](https://github.com/GoogleCloudPlatform/open-knowledge-format/blob/main/SPEC.md):
one file per note, with YAML frontmatter on top. There is no database. Search
results, the link graph, and directory listings are computed from the files
every time, so you can read, edit, diff, and commit the knowledge base like any
other text.

## Install

With Homebrew (the formula builds from source, so it pulls in Go):

```bash
brew tap cndreisbach/ritari https://github.com/cndreisbach/ritari
brew install ritari
```

With Go 1.26.9 or later:

```bash
go install github.com/cndreisbach/ritari/cmd/rit@latest
```

## Quick start

Create a knowledge base and make it the default:

```bash
rit project init work ~/notes/work --default
```

Write a note. The type decides its directory and the title its file name, so
this creates `decisions/auth-uses-jwt.md`:

```bash
rit note write --title "Auth uses JWT" --type decision \
  --description "Stateless auth via short-lived JWTs." \
  --tags auth,security \
  --obs "decision: Use JWT for stateless auth #security" \
  --rel "depends_on: User Database Schema" \
  --source "PR #1234 https://github.com/acme/api/pull/1234"
```

Find it and read it:

```bash
rit search jwt
rit note read "Auth uses JWT"
```

Writing to the same title again adds to the note instead of duplicating it.
The result looks like this:

```markdown
---
type: decision
title: Auth uses JWT
description: Stateless auth via short-lived JWTs.
tags: [auth, security]
sources:
  - id: "1"
    resource: https://github.com/acme/api/pull/1234
    title: "PR #1234"
generated: {by: human:you@example.com, at: 2026-07-29T14:30:00Z}
---

# Observations
- [decision] Use JWT for stateless auth #security [^1]

# Relations
- depends_on [User Database Schema](/notes/user-database-schema.md)

[^1]: [PR #1234](https://github.com/acme/api/pull/1234)
```

## Using it with an agent

Ritari ships with agent skills that teach an agent how to use it: searching
before writing, keeping daily notes for session continuity, tracking multi-step
tasks, ingesting documents, consolidating what it learned, and cleaning up lint
findings. Install them for your agent:

```bash
rit install skills                 # ~/.agents/skills
rit install skills --agent claude  # ~/.claude/skills
```

Run it again after upgrading. If an earlier version installed a `ritari-*`
skill that this one no longer ships, `rit install skills` warns about it, and
`rit install skills --cleanup` removes it.

When `rit` runs without a terminal, as it does under an agent, it prints flat
YAML that is easy to parse. In a terminal it prints readable text. Use
`--output text` or `--output yaml` to choose. Writes are stamped with who made
them: `rit` recognizes Claude Code, Codex, Cursor, Rovo Dev, and Pi, and otherwise
uses your configured author (`rit config author human:you@example.com`) or your
git email. An agent that names its model is stamped as `<agent>/<model>` — Pi's
`PI_MODEL` gives `pi/deepseek-flash-latest` — and `RITARI_AGENT_MODEL` overrides
the model for any agent. A bare author such as `rit config author you@example.com`
is stored as `human:you@example.com`, so a person is never mistaken for an agent.

## Commands

| Command | What it does |
| --- | --- |
| `rit note write` | Create a note, or add observations, relations, and sources to one |
| `rit note read` / `rit note view` | Show a note parsed, or as raw markdown |
| `rit note edit` | Edit the body (append, find/replace, replace a section…) or set frontmatter |
| `rit note move` | Move a note, rewriting every link to it |
| `rit note archive` / `unarchive` | Retire a note while keeping it linkable |
| `rit note verify` | Record that someone checked a note |
| `rit note delete` | Delete a note, reporting the links it breaks |
| `rit search` | Search by text, type, tags, observation category, relation, or any frontmatter field |
| `rit context` | Follow links outward from a note |
| `rit daily` | Read or write today's (or any day's) note |
| `rit recent` / `rit ls` | Recently changed notes; all note IDs |
| `rit lint` | Check OKF conformance and note hygiene |
| `rit index` | Regenerate the `index.md` listings |
| `rit types` | Show the note types this knowledge base uses |
| `rit project` | Manage registered knowledge bases |
| `rit install skills` | Install the agent skills |

Run `rit <command> --help` for flags. A flag value that starts with `-` must
be written as `--flag=value`, for example `--content="- [ ] write tests"`.

## Knowledge base layout

```
work/
  index.md            generated listing of directories
  log.md              generated change history
  ritari.yaml         optional: this knowledge base's note types
  decisions/          one directory per type
    index.md
    auth-uses-jwt.md
  daily/
    2026-07-29.md
  sources/            optional raw material (never .md files)
```

The default types are `decision`, `note`, `task`, `person`, `project`,
`meeting`, `reference`, `preference`, and `daily`. To use your own, list them in
`ritari.yaml`; `rit types --yaml` prints the current set as a starting point.

## Development

[SPECIFICATION.md](SPECIFICATION.md) defines Ritari's behavior: file formats,
commands, output, and error messages. [AGENTS.md](AGENTS.md) covers the code
layout, conventions, and release process.

Tools are managed with [mise](https://mise.jdx.dev):

```bash
mise install
mise run build                         # builds bin/rit
mise run test
mise exec -- golangci-lint run ./...
```

Run commands through mise (`mise run <task>` or `mise exec -- <command>`) so
the pinned tools and environment from `mise.toml` are in place.

Commits follow [Conventional Commits](https://www.conventionalcommits.org/).
Merging to `main` runs the tests, then publishes a release whose version comes
from those commit messages and updates the Homebrew formula.

## License

[MIT](LICENSE)
