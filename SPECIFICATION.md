# Ritari — Specification

This document specifies **Ritari**, a command-line memory tool for AI agents: a
single self-contained executable that stores long-term knowledge as plain
markdown files and lets an agent (or a person) write, search, link, and maintain
that knowledge across sessions. It is written so that an implementer can build
the tool from this text alone.

The specification describes observable behavior: file formats, command-line
interface, output formats, and algorithms where the exact result matters. It
deliberately says nothing about programming language, libraries, internal
structure, packaging, or distribution.

Requirement words (MUST, SHOULD, MAY) are used in their usual sense.

**Names.**

| Name | Meaning |
| --- | --- |
| `rit` | the executable's name |
| `ritari.yaml` | the per-bundle configuration file name (§4.5) |
| `RITARI_CONFIG`, `RITARI_PROJECT`, `RITARI_OUTPUT`, `RITARI_AGENT` | environment variables (§9) |
| `<config-dir>/ritari/config.yaml` | the project registry location (§9.3) |
| `ritari`, `ritari-notes`, `ritari-daily`, `ritari-tasks`, `ritari-ingest`, `ritari-reflect`, `ritari-lint` | the agent skills (§11) |

---

## 1. Purpose and principles

An AI agent normally forgets everything when a conversation ends. The tool gives
it a persistent memory: a **knowledge base** that survives sessions, can be
searched, linked, and added to as the agent works, and can be shared by any
agent that is taught to drive the tool. A set of **agent skills** (§11) shipped
inside the tool teaches an agent how to do so.

The knowledge base format is the **Open Knowledge Format (OKF), version 0.2**
(<https://github.com/GoogleCloudPlatform/open-knowledge-format/blob/main/SPEC.md>):
a knowledge base is a **bundle**, a directory of markdown files, one per
**concept**, each beginning with YAML frontmatter. The tool uses OKF's own
frontmatter fields where OKF defines one (`type`, `title`, `description`,
`resource`, `tags`, `sources`, `generated`, `verified`, `status`); its
extensions live only in the markdown body and in optional extra frontmatter
keys, which OKF permits. Any OKF-aware tool MUST be able to read a bundle
produced by this tool. There is no official OKF reference linter; a bundle
produced by this tool MUST satisfy OKF §11 conformance and MUST pass
`okflint validate` (<https://github.com/mattdav/okflint>) without errors.

The following principles are binding design decisions; changing one changes the
contract.

1. **Files are the only state.** There is no database, no index, no cache that
   can drift. Everything derived (directory listings, the link graph, search
   results) is computed from the files on demand and can be regenerated.
2. **Identity is the path.** A concept's identifier is its bundle-relative path
   minus the `.md` suffix. There is no separate permalink or UUID field.
3. **Malformed files are skipped, never fatal.** A file that cannot be parsed is
   left out of every command's view and is reported only by `lint`. One bad file
   must not lock anyone out of their knowledge base.
4. **Writes are additive and convergent.** Writing the same fact twice does not
   duplicate it; a repeat that arrives with a new source adds the source
   reference to the existing line.
5. **Generated files are never hand-edited.** `index.md` and `log.md` are
   maintained by the tool on every write.
6. **Output is a contract.** On a terminal the tool prints human text; otherwise
   it prints flat YAML, with no flag required, because agents run without a
   terminal.
7. **The tool never fetches anything.** It has no network client. Acquiring
   external material is the agent's job; the tool records what was learned and a
   pointer back to the source.
8. **Errors and warnings are different things.** Only OKF's hard requirements are
   errors. The tool's own opinions are warnings and never fail a command unless
   the user asks for strictness.
9. **Dates are local.** Day-grouping (daily notes, the change log) uses the
   user's local calendar date, not UTC.

---

## 2. Terminology

- **Bundle** — a directory tree containing concept files plus reserved files.
- **Concept** (also **note**) — one markdown file with frontmatter.
- **Concept ID** — bundle-relative path with forward slashes, minus the `.md`
  suffix: `decisions/auth-uses-jwt`. Always `/`-delimited on every platform.
- **Type** — the concept's category (`decision`, `task`, …). Determines the
  directory it lives in.
- **Vocabulary** — the set of types a bundle recognizes.
- **Observation** — one fact, written as a bulleted line with a bracketed
  category.
- **Relation** — a typed edge to another concept, written as a bulleted line
  containing a markdown link.
- **Source** — a material a concept derives from, recorded as an entry in the
  `sources` frontmatter list (OKF §5.1) with a numeric `id`, and mirrored in the
  body as a markdown footnote definition `[^id]: …` (§3.7).
- **Source reference** — a footnote reference `[^id]` attributing a claim
  (usually an observation) to a source.
- **Project** — a named bundle registered with the tool (§9.3).
- **Actor** — the identity string recorded as the author or verifier of content
  (§9.4), following the OKF actor convention (OKF §7).
- **Machine mode / text mode** — the two output modes (§9.1).

---

## 3. The bundle on disk

### 3.1 Layout

```
<bundle root>/
  index.md                 generated; root listing (declares the OKF version)
  log.md                   generated; change history
  ritari.yaml              optional; the bundle's type vocabulary (§4.5)
  sources/                 optional; raw material, never concepts (§3.8)
  <type-directory>/
    index.md               generated; one per directory holding concepts
    <slug>.md              one concept per file
  daily/
    YYYY-MM-DD.md          daily notes
```

A concept is any file whose name ends in `.md` (case-insensitive) other than the
reserved files `index.md` and `log.md`, in any directory. Reserved names match
**exactly**, lowercase, as in OKF §3.1: `Index.md` and `LOG.md` are ordinary
concepts (and, lacking a `type`, conformance errors).

On a case-insensitive file system (the default on macOS and Windows), a concept
named `Index.md` and the generated `index.md` are the same file, so writing the
index would overwrite the concept. `lint` therefore reports such files as
`reserved-casing` there (§8). The tool detects case-insensitivity for the
bundle's own directory rather than assuming it from the operating system.
Directories whose names begin with `.` are never traversed. Concepts may exist
in nested subdirectories or loose at the root; the tool reads them all but only
ever *creates* concepts at `<type-directory>/<slug>.md`.

### 3.2 Concept file format

A concept file is:

1. The frontmatter block: a line `---`, YAML, a line `---`.
2. A blank line.
3. The markdown body.

The file always ends with a newline. A `.md` file with no frontmatter block is
still readable (it has no type, which `lint` reports as an error — §8).

#### Frontmatter fields

| Key | Type | Notes |
| --- | --- | --- |
| `type` | string | **Required by OKF.** Non-empty. |
| `title` | string | Human-readable name. Falls back to the last path segment of the ID when absent. |
| `description` | string | One-sentence summary; this is what directory listings show. |
| `resource` | string | Canonical URI of the thing the note is *about* (not where a fact came from). |
| `tags` | list of strings | Topic. Free to change; type is not. A bare scalar (`tags: auth`) MUST be accepted on read as a one-element list. |
| `sources` | list of mappings | Provenance (OKF §5.1): the materials facts came from (§3.7). |
| `generated` | `{by, at}` | Who last meaningfully changed the content, and when (§5.4). |
| `verified` | list of `{by, at}` | Independent confirmations of the content (§5.5). |
| `archived` | boolean | `true` retires the note (§5.6). |
| `status` | string | OKF lifecycle (`draft`, `stable`, `deprecated`; absent means `stable`). The tool never sets it and treats it as an ordinary extra key; it is **not** workflow state. |
| `state` | string | Workflow state by convention (a task's `open`/`done`, a project's `active`). An ordinary extra key; the tool attaches no meaning to it. |
| any other key | any YAML | Preserved verbatim across rewrites. |

Reading rules:

- Unknown keys MUST be preserved when a note is rewritten.
- A legacy key `timestamp` (from OKF v0.1, superseded by `generated.at` in v0.2) is tolerated: where `generated.at` is
  absent, `timestamp` serves as the content timestamp for display, filtering,
  and staleness checks. It is not written by the tool; rewriting a note keeps
  an existing `timestamp` as an ordinary extra key.
- `generated.at` and `verified[].at` are read as plain strings even if the YAML
  resolves them to timestamps; they are RFC 3339 instants.
- `verified` MAY be a single `{by, at}` mapping instead of a list; it MUST be read
  as a one-element list (OKF §5.2). It is always written as a list.
- Each `sources` entry is a mapping. Its `id` is read as a string even if the YAML
  resolves it to a number. Keys of an entry other than `id`, `resource`, and
  `title` (OKF's `author`, `usage_count`, `last_modified`, …) are preserved.
- `archived` is true when the value is boolean `true` or a string equal,
  case-insensitively, to `true` or `yes`.
- Frontmatter that is not valid YAML, that is not a mapping, or that repeats a
  key makes the file **unparseable** (§10.1). Nothing else does: OKF requires only
  parseable YAML and a `type`.
- A recognized field whose value has the wrong shape (a mapping for `title`, a
  `sources` that is not a list of mappings, a `generated` that is not a mapping)
  is a **malformed field**. The tool reads the note as if that field were absent
  (a malformed `type` therefore counts as missing), keeps the value verbatim as an
  ordinary extra key when it rewrites the note, and `lint` reports it as
  `malformed-field` (§8). A write that sets the field replaces the malformed
  value.
- Leading spaces/tabs before the opening `---` are tolerated. If the content
  does not begin with a `---` line, the whole file is body and the frontmatter is
  empty. If an opening `---` has no closing `---` line, the whole file is body.
  The closing delimiter is a line consisting of `---` with optional trailing
  whitespace.

Writing rules (rendering): keys are emitted in this order — `type`, `title`,
`description`, `resource`, `tags`, `sources`, `generated`, `verified`, then all
remaining keys sorted alphabetically. Empty/absent fields are omitted (`type` is
always present; it defaults to `note`). `tags` is written as an inline list
(`[auth, security]`), empty and duplicate tags dropped. `sources` is a block list
of block mappings with keys `id`, `resource`, `title`, then any other keys of the
entry in their original order; `id` is always quoted (`id: "1"`). `generated` is
written inline as `{by: <actor>, at: <timestamp>}`; OKF requires `by`, so it is
always written (§5.4). `verified` is a block list of inline `{by, at}` maps. String values that would
otherwise be read as another YAML type (`true`, `123`) MUST be quoted so they
stay strings. Extra keys supplied from the command line are strings and MUST
read back as strings. The YAML indentation unit is two spaces.

Example:

```markdown
---
type: decision
title: Auth uses JWT
description: Stateless auth via short-lived JWTs, chosen over server sessions.
resource: https://github.com/acme/api/pull/1234
tags: [auth, security]
sources:
  - id: "1"
    resource: https://github.com/acme/api/pull/1234
    title: "PR #1234 — switch to JWT"
generated: {by: human:jane@acme.com, at: 2026-07-29T14:30:00Z}
verified:
  - {by: human:jane@acme.com, at: 2026-07-30T09:00:00Z}
state: shipped
---

Server-side sessions would have meant sticky routing, which we did not want.

# Observations
- [decision] Use JWT for stateless auth #security [^1]
- [constraint] Tokens expire after 15m (per security audit) [^1]

# Relations
- depends_on [User Database Schema](/references/user-database-schema.md)

[^1]: [PR #1234 — switch to JWT](https://github.com/acme/api/pull/1234)
```

### 3.3 Body structure

Everything above the first heading is free prose (the **narrative**). The
tool's conventional sections are top-level (`#`) headings: `# Observations`,
`# Relations`, and, for daily notes, `# Touched`. The footnote definitions that
mirror `sources` (§3.7) sit after everything else, under no heading. Heading
matching when *reading or editing* is case-insensitive and ignores heading level
(so a `## Observations` written by hand still counts); sections the tool
*creates* use a single `#`.

A **section** runs from its heading line to the next heading of the same or a
shallower level, or the end of the body. A heading is an ATX heading: one or
more `#` followed by a space. (Fenced code is not specially treated when locating
headings.)

### 3.4 Observations

```
- [category] content #tag #tag (context) [^1] [^2]
```

Recognized anywhere in the body, not only under `# Observations`, on any line
matching: optional leading whitespace, a `-` or `*` bullet, `[category]`, at
least one whitespace, then the rest.

Parsing the rest:

1. Trailing source references `[^label]` (label = letters, digits, `_`, `-`) are
   removed, in written order.
2. A trailing `(context)` (a parenthesized group at end of line) is removed and
   becomes the context. Then, if no references were found in step 1, trailing
   references are stripped again (so `text (ctx) [^1]` and `text [^1] (ctx)` both
   work).
3. Every `#word` (`word` = letters, digits, `_`, `-`, `/`) anywhere in the
   remaining text is a tag and is removed from the content — **except** a `word`
   made only of decimal digits (`#42`), or only of hexadecimal digits with at
   least one decimal digit among them (`#ff00aa`, `#3e8`). Those are issue
   numbers, colors, and hashes; they stay in the content. (`#cafe` and `#fff`
   are tags.)
4. The content is the remainder with whitespace runs collapsed to single spaces.

The category is trimmed. A line whose bracket content is empty/whitespace, or
exactly `x`/`X`, is a Markdown checkbox, **not** an observation.

### 3.5 Relations

```
- relation_type [Link text](/path/to/concept.md)
```

A relation is a list item (`-` or `*`) whose whole content is an optional single
whitespace-free token (the relation type) followed by one markdown link, and
nothing after the link. With no leading token the relation type is `links_to`.
A relation line only counts if the link is a *real* markdown link (so a line that
merely looks like one inside a fenced code block is not a relation).
Relation types are single tokens; underscores, not spaces.

A relation records: type, link text, the raw destination (`href`), and the
resolved concept ID if the destination is inside the bundle (empty otherwise).

### 3.6 Links, targets, and the graph

**Every** markdown link in a body is a graph edge, not only relation lines — in
OKF a link in prose connects two concepts. Links are extracted with a real
markdown parser (inline and reference-style links; images and `<autolinks>` are
ignored), so links inside fenced or indented code blocks are not edges. Links
with an empty destination, and repeated links with an identical destination
within one note, are ignored.

Link classification for a destination `href`:

- **External**: begins with `//`, or has a URI scheme (text before the first `:`,
  containing none of `/`, `?`, `#`) — `https:`, `mailto:`, etc. External links
  have no target.
- **Fragment-only** (`#section`): neither external nor in-bundle; no target.
- **In-bundle**: anything else. Strip any `#fragment`. A destination starting
  with `/` is relative to the bundle root; otherwise it is relative to the
  directory of the linking note. Normalize (`.` and `..` segments). If the result
  escapes the bundle root, it has no target. If it lies under `sources/` (raw
  material, §3.8), it has no target. If its last segment is a reserved file name
  (`index.md` or `log.md`, matched as in §3.1), it is a **reserved-file link**:
  it names that file, not a concept, and has no target. Otherwise the target
  concept ID is the normalized path minus a trailing `.md`.

The **graph** has a directed edge from each concept to each distinct in-bundle
target. Edges to targets that do not exist are **dangling** and are tracked
separately (they appear in `lint`, not in traversal). The tool MUST tolerate
dangling links everywhere — forward references are cheap and intended.

A reserved-file link (a link to a directory listing or a change log, which OKF
§8–9 define) is never an edge: it doesn't count for traversal, `context`,
`links` in `note read`, or `orphan`. It is dangling only if no file exists at
its normalized path, for example a link to the `index.md` of a directory that
no longer holds any concepts. A dangling reserved-file link is reported by its
bundle-relative path including the file name (`ideas/index.md`).

### 3.7 Sources and source references

Provenance follows OKF v0.2 §5.1: sources live in the `sources` frontmatter list,
and claims point at them with markdown footnote references keyed by the entry's
`id`. Because ordinary markdown readers (Obsidian, GitHub) only render a footnote
reference that has a definition in the body, every source the tool numbers is
**also** written as a footnote definition at the end of the body:

```markdown
---
type: meeting
title: H1 planning
sources:
  - id: "1"
    resource: https://github.com/acme/api/pull/1234
    title: "PR #1234 — switch to JWT"
  - id: "2"
    resource: Email from PJ Perger, 2026-07-29, subject: H1 commitments
---

# Observations
- [decision] Ship JWT auth in H1 [^1] [^2]

[^1]: [PR #1234 — switch to JWT](https://github.com/acme/api/pull/1234)
[^2]: Email from PJ Perger, 2026-07-29, subject: H1 commitments
```

**Entries.** A source the tool writes has `id` (a positive decimal integer,
written as a quoted string), `resource` (required by OKF), and an optional
`title`. A `resource` is **followable** when it contains no whitespace (a URL, or
a bundle path such as `/sources/call.txt`); otherwise it is a **descriptor**
(OKF's term for a source a consumer cannot follow, such as an email or a
conversation).

**Input forms** (`--source` on the command line):

| Form | Title | Resource |
| --- | --- | --- |
| `[Title](URI)` | Title | URI |
| `<URI>` | — | URI |
| `URI` (one token: has a scheme followed by `//`, starts with `mailto:`, or starts with `/`) | — | URI |
| `Some title URI` (the last whitespace-delimited token is URI-looking as above, with text before it) | Some title | URI |
| `Some text` (anything else) | — | the text, trimmed (a descriptor) |

Blank values are ignored.

**Footnote definitions.** For every entry whose `id` is a positive decimal
integer, the body ends with one line `[^id]: <rendered>`, in ascending numeric
`id` order, after all other content and separated from it by one blank line.
Rendering: `[Title](resource)` when the resource is followable and there is a
title; `[resource](resource)` when it is followable with no title (an ordinary
link, not the `<URI>` autolink form, because OKF consumers resolve autolink
destinations relative to the citing note); `Title (resource)` for a descriptor
with a title; the bare descriptor otherwise. Square brackets in link text are
backslash-escaped.

These definitions are **derived**: `sources` is the truth. Whenever the tool
rewrites a note it removes every footnote-definition line in the body whose
label is the `id` of a numbered entry, wherever it sits, and appends fresh ones
as above. Body operations (merge, `note edit`, `# Touched`) act on the body with
those lines removed. Footnote definitions with any other label, and entries
with a non-numeric `id` (written by other OKF tools, which author their own
footnote prose), are left untouched.

**Allocation.** A new source gets `id` = 1 + the largest positive integer among
the note's existing `sources` ids **and** the numeric footnote labels already
used anywhere in its body (so a hand-written `[^3]` is never reused), or 1 if
there are none. Two sources are the *same* if their resources are equal —
exactly for followable resources, case-insensitively after trimming for
descriptors. Citing the same source again reuses its `id` (the entry is left as
it is) rather than adding an entry.

**Legacy citations.** OKF v0.1 bodies may carry a `# Citations` list with `[n]`
references. The tool treats that list as ordinary body text: it never parses it
as provenance, never writes one, and never rewrites it.

### 3.8 Raw material: `sources/`

Raw material (transcripts, exports, PDFs) lives in `sources/` at the bundle
root. It is **not kept by default**: a source with a URL already lives somewhere
durable. Nothing in `sources/` is a concept, so files there MUST NOT be `.md` and
MUST NOT be placed in hidden (dot-) directories, because other OKF tools descend
into hidden directories and would treat a stray `.md` there as a conformance
error. Notes point at material in `sources/` with a bundle-relative path (OKF
§6.2): `resource: /sources/<file>`, or `--source "Call transcript /sources/<file>"`.
Links into `sources/` are not graph edges (§3.6).

Creating a project (§9.3) scaffolds `sources/` with a plain-text `README.txt`
explaining these two rules. The README is deliberately not markdown.

---

## 4. Types and directories

### 4.1 Type determines directory

A note's directory is derived from its type; there is no way to choose a
directory when writing. Type must therefore be the most stable property of a
note. Topic belongs in `tags`.

**Slug**: lowercase and trim the title, replace every maximal run of characters
outside `[a-z0-9]` with a single `-`, trim leading/trailing `-`. (Non-ASCII
characters are therefore dropped; a title with no ASCII alphanumerics has an
empty slug and cannot be used as a note title.)

A new note's ID is `<directory-for-type>/<slug-of-title>`.

### 4.2 The default vocabulary

| Type | Directory | Summary |
| --- | --- | --- |
| `decision` | `decisions` | A choice made, with its rationale. |
| `note` | `notes` | General knowledge that does not fit a more specific type. |
| `task` | `tasks` | Work in progress, with state. |
| `person` | `people` | A human and what you know about them. |
| `project` | `projects` | An ongoing effort. |
| `meeting` | `meetings` | A meeting and what came out of it. |
| `reference` | `references` | A pointer to an external resource. |
| `preference` | `preferences` | How the user wants things done. |
| `daily` | `daily` | A day's narrative: what came up, and what it touched. |

`note` is the default type (used when none is given or when frontmatter has no
type). `daily` is the type of daily notes (§7). The default observation category
(used when an observation has no `category:` prefix) is also `note`; these are
two separate concepts that happen to share a name.

### 4.3 Unknown types

Type lookup is case-insensitive and trims whitespace. A type outside the
vocabulary is **not an error** (OKF requires consumers to tolerate unknown
types). Its directory is the pluralized slug of the type name:

- ends in `s`, `x`, `ch`, or `sh` → append `es`
- ends in `y` preceded by a non-vowel (`aeiou`) → replace `y` with `ies`
- otherwise → append `s`

(An empty slug stays empty.) A write with an undeclared type succeeds and returns
a warning (§10.4).

### 4.4 Directory ↔ type reverse lookups

The tool can ask which type lives in a directory and what its summary is. When
two types claim the same directory, the first declared wins for the reverse
lookup.

### 4.5 Per-bundle vocabulary: `ritari.yaml`

A bundle may declare its own types in `ritari.yaml` at its root:

```yaml
types:
  - name: incident
    dir: incidents
    summary: An outage and its timeline.
```

- A non-empty `types` list **replaces** the default vocabulary (it does not
  extend it), so a bundle can be narrow. An absent file, or a file with an empty
  list, means the default vocabulary.
- `note` and `daily` are **reserved** and always present: if the list omits
  either, it is appended with its default directory and summary. A list may give
  either a different directory or summary.
- Names are lowercased and trimmed. A type with `dir` omitted gets its directory
  derived as in §4.3. The first declaration of a name wins.
- Validation (any failure is a hard error from every command that opens the
  bundle, with the message prefixed by the config file name): a type has no
  name; a type name is declared twice; a `dir` is not equal to its own slug
  (it must be a single lowercase path segment of `[a-z0-9-]`); two types want the
  same `dir`; a `dir` equals `sources`.
- A malformed or invalid `ritari.yaml` is fatal, unlike a malformed *note*. Reason:
  one bad note costs that note; a bad vocabulary decides where everything belongs.
- The file is read once per command invocation. It is not a concept (not
  markdown).
- The file lives in the bundle (not in the user-level config) so the vocabulary
  travels with a cloned bundle and an agent can be pointed at it.

Nothing breaks when a vocabulary changes: notes filed under a type that is no
longer declared stay where they are and stay addressable; `lint` warns about
them (§8).

---

## 5. Note semantics

### 5.1 Resolving a note reference

Commands that take a note argument accept any of: the **concept ID**, the
**file path** (bundle-relative, or absolute; with or without `.md`), or the
**title**. Resolution order:

1. Treat the argument as a path: try `<root>/<arg>.md` (adding `.md` unless it
   already ends in it, case-insensitively), then `<root>/<arg>` as given, then — if
   the argument is absolute — the absolute path itself. The first that is an
   existing regular file wins and is read directly from disk.
2. Exact match on a loaded concept ID.
3. Title match: the concept whose title's slug equals the slug of the argument.
   When several match, the one with the lexicographically smallest ID wins.

An empty argument is an error (`no note id given`); no match is an error
`note not found: <arg>`. If the matched file is unparseable, the command fails
with the parse error.

### 5.2 Composing observation and relation lines

**Observation input** (`--obs`) forms; a leading `- ` or `* ` is stripped:

- `category: text #tag (context)` — preferred. The category is the text before
  the first `:` **if and only if** that text is a single token with no whitespace
  (so `Use JWT: fast` is not split, but `http://x` would be — callers should
  prefer an explicit category).
- `[category] text #tag (context)` — already bracketed.
- `text` — category defaults to `note`.

The canonical output line is `- [category] content` followed by source references
` [^id]` for each ref. Tags and context are left exactly as typed inside the
content text. Refs already written at the end of the input are kept and unioned
(de-duplicated, order preserved) with the refs allocated for this write.

**Relation input** (`--rel`), leading bullet stripped:

- `type: Target` — preferred. The type is the text before the first `:` under the
  same single-token rule.
- `Target` — untyped; type `links_to`.

`Target` is one of:

1. A markdown link `[text](href)` — destination taken apart (leading `/` and
   trailing `.md` stripped to get the concept ID), text kept.
2. A path — contains `/` or ends in `.md`. The ID is the path with any leading `/`
   stripped, `.md` removed, normalized. Link text is the resolved note's title if
   such a note exists, otherwise the last path segment.
3. A title (anything else). Resolved per §5.1; on success the link goes to the
   found note's ID with its title as text. If nothing resolves, it is a **forward
   reference**: the link goes to `<default-type-directory>/<slug-of-target>`
   (`notes/<slug>` in the default vocabulary) with the target as typed as the
   text. It dangles until the target is written; that is intended.

The canonical relation line is `- type [text](/<id>.md)` (bundle-absolute link).

### 5.3 Body composition and additive merge

**Fresh composition** (new note): body = prose (trailing newlines trimmed, then
one newline) if any; then the sections, each emitted only if it has lines, in the
order Observations, Relations; each section is `# Name` followed by its lines,
separated from preceding content by one blank line; then the footnote
definitions (§3.7). Sources for the `--source` values are numbered from 1 (or
after any numeric footnote labels already in the supplied prose); the resulting
refs are attached to **every** observation in the same write (facts captured
together came from the same source).

**Additive merge** (existing note) — must converge, not duplicate:

- Work on the body without its derived footnote definitions (§3.7).
- Allocate sources against the note's existing ones (§3.7); the refs apply to every
  observation in this write.
- For each new observation line: if an existing body line has the same text
  ignoring trailing source references, **do not add a line**; instead union the refs
  into that existing line (rewrite it only if the ref set grew, preserving
  indentation). Otherwise, if the exact line (trimmed) is not already present, add
  it. Duplicates within one write are added once.
- For each new relation line: add it unless the identical line (trimmed) already
  exists.
- Append new observation lines at the end of the existing `Observations` section
  (before that section's trailing blank lines), likewise Relations; create a
  missing section at the end of the body as `# Name` preceded by a blank line.
- Re-append the footnote definitions for the (possibly extended) `sources`.
- Report counts of observations and relations actually added (a ref-only
  merge counts as zero added).

### 5.4 `generated`

`generated: {by, at}` records who produced the **current content** and when it
last meaningfully changed.

- It is stamped automatically on creation, replacement, and on any write/edit that
  changes the note's content (body) — `by` = the actor (§9.4), `at` = current time,
  UTC, RFC 3339, second precision.
- A write or edit that touches only frontmatter (`description`, `tags`,
  `resource`, `title`, `meta`, archive/restore, verify) leaves the existing stamp
  untouched.
- Whenever a note is rendered and its `generated.at` is blank, the current time is
  filled in; whenever its `generated.by` is blank, the current actor is filled in
  (so every written note has a complete `generated` field — OKF requires `by`).
- `by` convention (OKF §7): `<producer>/<version>` for tools, `human:<id>` for
  people, `process:<id>` for automated processes. This tool uses a bare producer
  name for agents (no version is available); OKF requires the `human:` prefix only
  for people, which the tool always uses (§9.4).

### 5.5 `verified` and trust tiers

`verified` is a list of independent confirmations, appended only by the explicit
`note verify` command — never automatically, because verifying is a deliberate
act. Entries are never removed by the tool. Verifying does **not** change
`generated`.

A **trust tier** is derived from `verified`:

| Tier | Condition |
| --- | --- |
| `unverified` | no entries |
| `human-reviewed` | at least one entry whose `by` starts with `human:` (OKF §5.3) |
| `machine-confirmed` | entries exist, none human |

Tiers are advisory (not access control) but filterable (`search --meta
trust=human-reviewed`). The tier is never stored; it is computed.

### 5.6 Archiving

`archived: true` in frontmatter retires a note: it stays in the bundle, stays
linkable and addressable, but drops out of search by default, is listed under an
`# Archived` heading in its directory index, and is exempt from hygiene warnings
(§8). It is deliberately **not** a workflow `state` value (a task can be `done`
without being retired, or archived while its `state` records how it ended), and
it is independent of OKF's lifecycle `status`, which the tool leaves to the
author. Unarchiving removes the key entirely.

Deletion removes the file and breaks inbound links; archiving is the normal way
to retire something and `delete` is for mistakes.

### 5.7 Body edit operations

`note edit` applies one operation to the body:

| Op | Behavior |
| --- | --- |
| `append` | ensure body ends in a newline, then add `content` and a newline |
| `prepend` | `content`, newline, then the existing body |
| `find_replace` | replace **all** occurrences of `find` (required, non-empty) with `content`; if `count` > 0 and the number of occurrences differs, error `expected N occurrences of "…", found M`; if none found, error saying the text was not found and that only the body (not frontmatter) is searched, pointing to `note write` for frontmatter fields |
| `replace_section` | keep the heading line, replace the section's contents with a blank line, `content`, a blank line |
| `insert_before_section` | insert `content` and a blank line immediately before the heading |
| `insert_after_section` | insert `content` and a blank line immediately after the section ends |

Section ops require a section name (heading text, leading `#`s and spaces
ignored, matched case-insensitively at any level); a missing name is an error
(`<op> requires a section heading`) and an unmatched one is
`section not found: "<name>"`. An unknown op is `unknown edit op: "<op>"`.

---

## 6. Reserved files

### 6.1 `index.md` (generated)

Every directory that directly holds at least one concept has an `index.md`, and
the bundle root always has one. They are regenerated, never hand-edited, and
progressively disclose the bundle: an agent reads the root, picks a directory,
and reads descriptions rather than documents.

**Root `index.md`:**

```markdown
---
okf_version: "0.2"
---

# Directories

* [Decisions](/decisions/index.md) (`decision`) - A choice made, with its rationale.
* [People](/people/index.md) (`person`) - A human and what you know about them.

# Concepts

* [Loose Note](/loose-note.md) - its description
```

- Frontmatter is exactly `okf_version: "0.2"`.
- `# Directories` lists every concept-bearing directory (excluding the root),
  sorted lexicographically by path. Each line is
  `* [<Humanized>](/<dir>/index.md)`, then, if the bundle's vocabulary maps that
  directory to a type, ` (\`<type>\`)`, then, if that type has a summary,
  ` - <summary>`. The type is *named* here because this listing is the first thing
  an agent reads and it is how it learns that `incidents/` means `--type incident`.
- **Humanize**: take the last path segment, split on `-`/`_`, capitalize the first
  letter of each word, join with spaces.
- `# Concepts` appears only if concepts sit directly at the root; entries sorted
  by title.
- Sections are separated by a blank line. A bundle with nothing in it produces
  only the frontmatter block followed by a blank line.

**Directory `index.md`:**

```markdown
# Decisions

* [Auth uses JWT](/decisions/auth-uses-jwt.md) - Stateless auth via short-lived JWTs.

# Archived

* [Old approach](/decisions/old-approach.md) - description
```

- Heading is the humanized directory name; entries sorted by title (ties by ID);
  each entry `* [Title](/<id>.md)` plus ` - <description>` when the description is
  non-blank. Archived notes are listed under `# Archived` (preceded by a blank
  line), omitted if none.
- The **daily directory** (the directory the bundle's vocabulary assigns to type
  `daily`) differs: no top heading; one `# <YYYY>` heading per year, newest year
  first (year taken from the first four digits of the title, else of the last ID
  segment, else the bucket `Undated`), days within a year newest-first by ID, groups
  separated by a blank line. Archived daily notes are not separated.

**Synchronization.** After computing the desired content of every index, the tool
writes only the files whose bytes differ (keeping mtimes and VCS diffs quiet), and
**removes** any `index.md` in the bundle (outside dot-directories) that is no longer
generated (e.g. a directory's last concept was moved away). It returns the
bundle-relative paths it changed, sorted, with removals suffixed ` (removed)`.
A check mode computes the same list without writing.

### 6.2 `log.md` (generated)

An append-only, newest-day-first change history:

```markdown
# Update Log

## 2026-07-29

* **Create**: [Auth uses JWT](/decisions/auth-uses-jwt.md)
* **Move**: [Auth uses JWT](/references/auth-uses-jwt.md) (was decisions/auth-uses-jwt)
* **Delete**: notes/scratch

## 2026-07-28

* **Update**: [Title](/path.md)
```

- Heading `# Update Log`. One `## YYYY-MM-DD` group per local date; a new day's
  group is inserted directly after the document heading (so the file stays
  newest-first); within a day, entries are appended chronologically at the end of
  the day's group. Groups are separated by blank lines.
- Entry: `* **<Action>**: <subject>` plus ` (<detail>)` if there is a detail.
  Subject is a bundle-absolute link `[Title](/<id>.md)` (title falls back to the ID),
  except for `Delete`, whose subject is the bare concept ID (it would not resolve).
- Actions: `Create`, `Update`, `Delete`, `Move` (detail `was <old id>`, subject is the
  new ID), `Archive`, `Restore`.
- An event is **skipped** if an identical rendered bullet is already present in
  today's group, so repeated no-op writes do not pile up lines.
- Reading: the entries of a given date are recoverable by parsing these bullets
  (`* **Action**: subject (detail)`), where the subject is a link or, for deletions,
  a bare ID.

### 6.3 Recording

Every command that changes a concept (write, edit, verify, move, archive,
unarchive, delete, daily write) ends by **recording**: re-read the bundle from disk,
append the change event(s) to `log.md` (subject to the dedupe rule), and
synchronize indexes. Regenerating a derived section (a daily note's `# Touched`) is
**not** recorded, since logging it would feed itself.

---

## 7. Daily notes

A daily note is the authored record of a day: what came up, what was learned,
what is unresolved. There is one per local calendar date at
`<daily-directory>/YYYY-MM-DD` (the directory comes from the vocabulary, `daily`
by default), type `daily`, title equal to the date.

A date MUST be a valid ISO-8601 calendar date `YYYY-MM-DD` (zero-padded); anything
else is an error (`daily notes are named by date: "<x>" is not YYYY-MM-DD`). With no
date given, the current local date is used.

**Writing** (`daily` command, §10.7): if the note does not exist it is created with
type `daily`, title = date, the optional description, and a body composed from the
narrative prose and additions (§5.3). If it exists, additions are merged (§5.3),
narrative paragraphs are appended **at the end of the narrative** (the text before the
first heading — so anything under a heading is untouched and a user-added section is
a stable home), and a supplied description replaces the old one. An existing note's
`generated` stamp is **not** restamped, and a newly created daily note carries
`generated: {by: <actor>, at: <now>}`. The write is recorded (Create or Update with the
date as title), then `# Touched` is refreshed.

**`# Touched`** is a *projection* of that day's `log.md` group, never authored. It is
rebuilt wholesale (replacing the section, created at the end of the body if missing,
removed entirely if empty):

Entries are derived from the log events of the date, excluding the daily note's own
events and events with no ID, collapsed to one entry per concept in order of first
appearance:

- The action reported is the **first non-delete** action recorded for the concept that
  day (a note created then updated reads as created).
- If the concept's *latest* relevant event is a delete (not followed by a create or
  update), the entry is the deletion. A create/update after a delete "resurrects" it
  and the first action after the resurrection is reported.
- A `Move` event retargets the entry already recorded for the old ID (taken from the
  event's `was <old id>` detail) to the new ID/title, keeping its original action
  (created-then-moved reads as created, at the new path). A move with no earlier entry
  becomes its own `moved` entry.

Lines: `- [Title](/<id>.md) — <action>` (action past tense: `created`, `updated`,
`moved`; other actions are the lower-cased action name, e.g. `archive`, `restore`),
and for deletions `- <id> — deleted` (named, not linked). Note the action trails the
link so the line is a real link (a graph edge) but **not** a typed relation.

Refreshing rewrites the note only if the section text changed, and does so without
recording and without restamping `generated`. Refreshing a date with no daily note is
a no-op that reports `missing`.

---

## 8. Lint

`lint` reports problems; it never writes. Findings are **errors** (OKF's hard
requirements; a bundle with any is non-conformant) or **warnings** (this tool's
opinions). Only errors fail the command by default, because OKF forbids rejecting a
bundle over a missing optional field.

| Rule | Severity | Fires when |
| --- | --- | --- |
| `unparseable` | error | A concept file cannot be parsed (invalid frontmatter). Message: `frontmatter could not be parsed: <detail>`. |
| `missing-type` | error | Frontmatter `type` is absent or blank. |
| `unknown-type` | warning | `type` is non-blank but not in the vocabulary. Message names the vocabulary. |
| `missing-title` | warning | Frontmatter `title` blank. |
| `missing-description` | warning | `description` blank (the note appears bare in listings). |
| `type-directory-mismatch` | warning | The note's directory differs from the directory its type implies. Message includes the fix: the move command with the proposed destination `<expected-dir>/<current-basename>`. |
| `duplicate-title` | warning | Two or more notes share a title (case-insensitive, trimmed; title falls back to the last path segment). One finding per involved note naming the others. |
| `dangling-link` | warning | A note links to an in-bundle target that does not exist, or a reserved-file link points at a missing file (§3.6) (one finding per distinct (note, target) pair). Message: `links to <target>, which does not exist`. |
| `stale-index` | warning | A generated listing is out of date or should be removed; the path is the index file. Message points at the regenerate command. |
| `unknown-category` | warning | Only when an allowed-category list is supplied: an observation category outside it (case-insensitive). Off by default; categories are open-ended. |
| `stale-verification` | warning | The note has `verified` entries and `generated.at` is later than the newest well-formed `verified[].at`. Never fires for unverified notes. |
| `orphan` | warning | Not archived, not a daily note, has no outbound edge to an existing concept, and no inbound edge from a non-daily note (a daily note's `# Touched` links do not count). |
| `malformed-field` | warning | A recognized frontmatter field has the wrong shape and is being ignored (§3.2). One finding per field. Message: `<field> is malformed (<detail>) and is ignored`. |
| `reserved-casing` | warning | Only on a case-insensitive file system: a file whose name equals `index.md` or `log.md` ignoring case, but not exactly (§3.1). The path is that file. Message: `<name> collides with the reserved <reserved> on this case-insensitive file system; rename it`. |

**Archived notes** are exempt from every hygiene warning: they still receive
`missing-type`/`unknown-type` (conformance applies), `malformed-field`,
`reserved-casing`, `dangling-link` and graph-level checks, but skip `missing-title`, `missing-description`, `type-directory-mismatch`,
`duplicate-title`, `unknown-category`, `stale-verification`, and `orphan`.

Findings are sorted by path, then errors before warnings, then rule name. A finding
has `severity`, `rule`, `id` (empty for bundle-level findings such as `stale-index`),
`path` (bundle-relative file path), and `message`.

Options: `--conformance` keeps errors only; `--rule` (repeatable) keeps only the
named rules and rejects an unknown rule name with `unknown rule "<x>" (known: <sorted
list>)`; `--strict` makes warnings fail the command; `--categories a,b,c` supplies
the allowed-category list. Counts in the result reflect filtering. The command fails
(exit status 1, `success: false`) when errors > 0 (`N conformance error(s)`) or, with
`--strict`, when warnings > 0 (`N warning(s) with --strict`).

---

## 9. Command-line interface

### 9.1 Global behavior

```
rit [--output text|yaml] [--project NAME] <command> ...
```

- `--output`/`-o`: `text` or `yaml`. If unset, the `RITARI_OUTPUT` environment variable
  is consulted; if still unset, **auto**: machine (YAML) mode unless standard output
  is a terminal. An unrecognized value prints `<tool>: unknown output format "<x>"
  (want text or yaml)` to stderr and exits with status 2.
- `--project`/`-p`: project name; overrides the default and `RITARI_PROJECT`.
- A string flag value that begins with `-` MUST be given in the `--flag=value` form;
  the space-separated form is a usage error (this is the standard behavior of the
  flag parser and matters because bodies are often markdown bullets).
- Usage errors (unknown command, missing argument, invalid flag value) exit with
  status 2 with the problem on stderr.
- `-` as the value of `--content` (and `--note` on `daily`) means "read standard
  input to the end".

**The output contract (machine mode).** Every result is one or more YAML
**documents**, each a **flat mapping**: values are scalars or inline (flow) lists;
nothing nests.

- Lists of records are a **document stream** separated by `---`: first a header
  document carrying `success: true` and `count`, then one document per record.
- A field whose value is an empty string is **omitted**; an empty list is omitted
  unless the spec below says it is emitted as an explicit `[]` ("none" as an answer).
- Key order is as listed in this specification.
- On failure: a document `success: false` + `error: <message>` on standard output
  (machine mode only), the message on standard error as `<tool>: <message>`, and exit
  status 1. A command that has already emitted its own failure document (`lint`,
  `index --check`) emits no second one; it prints the message to stderr and exits 1.
- In text mode nothing is written as YAML; commands print human prose (formats below
  are informative unless stated).
- A command emits its result document unconditionally; human prose is printed only in
  text mode.
- `note view` and `read` (and `types --yaml`) write raw bytes verbatim in **both**
  modes.

### 9.2 Command summary

| Command | Purpose |
| --- | --- |
| `note write` | create a note, or add to an existing one |
| `note read` | parsed view of a note |
| `note view` / `read` | raw markdown (two names, one operation) |
| `note edit` | edit body (granular ops) and/or set frontmatter fields |
| `note verify` | record an independent confirmation |
| `note move` | relocate, rewriting inbound links |
| `note archive` / `note unarchive` | retire / restore |
| `note delete` | remove, reporting broken links |
| `search` | text + structured filters |
| `context` | follow links outward from a note |
| `daily` | open or write a day's note |
| `recent` | recently modified notes |
| `ls` | list concept IDs |
| `index` | regenerate or check `index.md` files |
| `lint` | conformance errors and hygiene warnings |
| `types` | list the vocabulary (or emit it as config) |
| `project list/init/add/remove/default` | manage the project registry |
| `config author` | get/set the configured actor identity |
| `skills install` | install the agent skills |
| `version` | print the version |

All commands except `project`, `config`, `skills`, and `version` operate on the
**active project's bundle**: resolved as `--project`, else `RITARI_PROJECT`, else the
registry default (an empty `RITARI_PROJECT` counts as unset). If none can be
determined: `no project specified and no default set (try: rit project init <name>
<path>)`; an unknown name: `project "<x>" not found`. There is no cross-project
addressing; concept IDs are bundle-relative.

### 9.3 The project registry

A **project** is a name plus an absolute directory path. The registry is a YAML file:

```yaml
default: work
projects:
  - name: work
    path: /home/me/notes/work
author: human:me@example.com
```

The registry is per user and per machine (where *your* bundles live and who *you*
are); it is never part of a bundle. Contrast `ritari.yaml` (§4.5), which belongs
to a bundle and travels with it.

Location: `RITARI_CONFIG` if set; else `$XDG_CONFIG_HOME/ritari/config.yaml`; else
`~/.config/ritari/config.yaml` (on every platform). Keys in the order shown,
two-space indent, trailing newline; parent directories are created on save, and
the file is replaced atomically (§10.7); `default` and `author` omitted when empty. A
missing file is an empty registry. An unparseable file is an error
(`parsing <path>: …`).

Operations:

- **add**: name required (`project name is required`); path has a leading `~`
  expanded and is made absolute; a duplicate name is an error (`project "<x>" already
  exists`); the project becomes the default if requested **or** if no default exists.
- **remove**: unknown name is `project "<x>" not found`; if it was the default, the
  default becomes the first remaining project, or empty. Never touches files.
- **set default**: unknown name is `project "<x>" not found`.

### 9.4 Actor identity

The actor string stamped into `generated.by` and `verified[].by`:

1. **Agent detection.** `RITARI_AGENT` (case-insensitive, trimmed) may name one of
   `unknown`, `human`, `claude-code`, `codex`, `cursor`, `rovo`; a recognized value is
   taken as the detected identity immediately (`unknown`/`human` count as *not an
   agent*). An unrecognized value is ignored. Otherwise the first matching signal in
   this order wins (a variable counts as set if non-blank after trimming):
   - `rovo`: `ROVODEV_CLI` or `ROVO_DEV_CLI_VERSION`
   - `claude-code`: `CLAUDE_CODE_CHILD_SESSION` or `CLAUDECODE`
   - `cursor`: `CURSOR_AGENT`
   - `codex`: `CODEX_THREAD_ID`, or `CODEX_CI` equal to `1`, or `CODEX_SANDBOX`
2. If an agent was detected, the actor is its bare name (`claude-code`, …).
3. Otherwise the registry's configured `author` if non-empty.
4. Otherwise `human:<email>` from the VCS user email (`git config user.email`) if
   obtainable.
5. Otherwise `human:<username>` from the operating system's current user name
   (on Windows, without any `DOMAIN\` prefix).
6. Otherwise `human:unknown`.

### 9.5 `note write`

```
rit note write --title T [--type TYPE] [--description D] [--tags a,b]
    [--resource URI] [--meta k=v]... [--content BODY|-]
    [--obs "category: text #tag (ctx)"]... [--rel "type: Target"]...
    [--source "source"]... [--force] [--dry-run] [--allow-empty]
```

- `--title` is required. `--type` defaults to `note`. `--tags` is comma-separated.
  `--obs`, `--rel`, `--source` are repeatable and are **not** split on commas.
  `--meta` is repeatable `key=value`.
- The new note's ID is `IDFor(type, title)` (§4.1). If the title has an empty slug:
  error `note requires a title that slugifies to something`.
- **Existing-note detection:** look up that ID (§5.1); if not found, look up the title
  (so re-capturing under a different type still finds it instead of forking).
- **Existing note and no `--force` → additive mode.** Otherwise → fresh mode.

**Additive mode.**
- If `--content` is non-blank: error `note "<id>" already exists; --content replaces the
  body, so use --force (only --obs/--rel/--source are added by default)`.
- If there are no observations/relations/sources and no description, resource, meta,
  or tags: error `note "<id>" already exists and nothing to add (provide
  --obs/--rel/--source/--description/--tags/--meta, or --force to replace)`.
- Merge additions into the body (§5.3). `generated` is restamped (actor, now) **only
  if the body changed**. `--description` and `--resource` **replace**; `--tags` is
  **unioned** with existing tags; `--meta` replaces per key (other keys untouched).
  The note's type, title, and verification list are preserved.
- Status `updated`; `added_observations`/`added_relations` are the counts actually added
  (§5.3); `added_sources` is the number of `--source` values that created a new
  `sources` entry (a re-cited source counts zero).

**Fresh mode.**
- Body = `--content` (or stdin), then additions composed per §5.3. If the resulting
  body is blank and `--allow-empty` is not set: error `refusing to write an empty note
  (no --content/--obs/--rel); use --allow-empty to override`.
- Frontmatter: type, title, description, resource, tags, `generated` (actor, now), extra
  from `--meta`. No `verified`.
- Writes to `<dir>/<slug>.md` (creating directories), or — when replacing an existing
  note via `--force` — over the existing file in place. Status `created` or `replaced`.
  Counts are the numbers of `--obs`, `--rel`, `--source` values supplied.
- If the type is outside the bundle's vocabulary, the result carries a `warning`:
  `type "<t>" is not one this knowledge base declares (<comma-separated type names>); it
  will live in <dir>/`. The write still succeeds.

**`--dry-run`:** performs everything except writing and recording, and reports status
`dry-run` with the full rendered file.

Results (machine): `success`, `status` (`created`|`updated`|`replaced`|`dry-run`), `id`,
`path` (absolute; omitted on dry-run), `added_observations`, `added_relations`,
`added_sources`, `warning` (omitted when empty); a dry run replaces `path`/`warning`
with `rendered` (the whole file as a string). Text: `warning: <msg>` (if any) then
`Wrote <path>` / `Updated <path> (+N obs, +M rel)` / `Replaced <path>`; dry run prints
`--- dry-run (<id>) ---` then the rendered file.

### 9.6 `note read`, `note view`, `read`

`note view <ref>` and `read <ref>` print the file's exact bytes in both modes.

`note read <ref>` machine result (one document; empty values/lists omitted): `success`,
`id`, `type`, `title`, `description`, `tags` (list), `trust`, `resource`, `path`,
`generated_by`, `generated_at` (only if `generated` exists), `observations` (list of
canonical lines without the leading `- `, e.g. `[decision] Use JWT #security [^1]`),
`relations` (list of `type -> target` where target is the resolved ID or, for external
links, the raw href), `links` (distinct in-bundle target IDs), `sources` (list of
the rendered footnote definitions, `[^id]: …`, for every entry with an `id`, in
`sources` order). Text mode prints the title, an `id: … type: …` line, the
description, `trust: …`, then Observations / Relations / Sources blocks.

### 9.7 `note edit`

```
rit note edit <ref> [--op OP] [--content C|-] [--section S] [--find F] [--count N]
    [--title T] [--description D] [--tags a,b] [--resource URI] [--meta k=v]...
```

- Error `nothing to do: pass --op for a body edit, or --title/--description/--tags/
  --resource/--meta to set frontmatter` if neither an op nor a frontmatter flag is given.
- Body ops are §5.7. Frontmatter flags **set outright** (not merge): `--tags` *replaces*
  the list; `--meta` replaces per key; `--title` changes the stored title but does **not**
  move the file or change the ID (identity is the path). An op and frontmatter flags may
  be combined in one call/write.
- `generated` is restamped only if the body actually changed.
- Always recorded as an `Update`.
- Result: `success`, `status: updated`, `id`, `path`, `op` (if given). Text: `Edited <path>`.

(The two commands are complementary by design: `note write` adds and merges, matched
by title; `note edit` sets, addressed by ID — one cannot merge its way to renaming.)

### 9.8 `note verify`

```
rit note verify <ref> [--by ACTOR]
```

Appends `{by, at: now UTC RFC 3339}` to `verified` (`by` defaults to the detected
actor). Does not change `generated`. Recorded as `Update`. Result: `success`,
`status: verified`, `id`, `path`, `by`, `trust` (the new tier). Text: `Verified <id> by
<by> (trust: <tier>)`.

### 9.9 `note move`

```
rit note move <ref> <dest>
```

`dest` is a path relative to the bundle root; `.md` is appended if missing and a leading
`/` is stripped. Errors: `note <id> is already there` (same ID), `destination already
exists: <id>`. Effects, in order:

1. Collect the notes that link to the moved note (from the graph, before the move).
2. Move the file, creating directories.
3. Rewrite the **moved note's own** relative in-bundle links to bundle-absolute form
   (its old relative paths would otherwise break), including links to targets that
   do not exist.
4. In each inbound note, rewrite every link whose target was the old ID to the new
   bundle-absolute path (`](/<newid>.md)`). Replacement is keyed on the exact
   destination text as written so unrelated links are untouched.
5. Record a `Move` event (subject new ID, detail `was <old id>`).

Result: `success`, `status: moved`, `from`, `to`, `rewrote_links_in` (sorted IDs;
explicit `[]` if none). Text: `Moved A -> B`, then
`Rewrote links in N note(s): …` when any.

`move` is the **only** correct way to relocate a note.

### 9.10 `note archive`, `note unarchive`, `note delete`

- `archive`/`unarchive` take a ref. Archiving an archived note errors `<id> is already
  archived`; unarchiving a live one errors `<id> is not archived`. Sets/removes
  `archived`; does not restamp `generated`. Recorded as `Archive` / `Restore`. Result:
  `success`, `status` (`archived`|`restored`), `id`, `path`. Text: `Archived <id>` /
  `Restored <id>`.
- `delete` resolves the note, computes the concepts that currently link to it, removes the
  file, and records a `Delete`. Result: `success`, `status: deleted`, `id`, `path`,
  `broke_links_in` (IDs; explicit `[]` if none). Text additionally prints
  `Broke N inbound link(s): …` and a hint that archiving would have kept them working.

### 9.11 `search`

```
rit search [QUERY] [--type T] [--category C] [--tags a,b] [--relation R]
    [--meta key=value]... [--include-daily] [--include-archived] [--limit N]
```

No index: the whole bundle is read and scored on every query.

**Tokenization.** Lowercase, then split on every character that is not an ASCII letter
or digit (non-ASCII letters act as separators). The query is tokenized into **terms**.

**Filters** (applied first; all must pass; matching is case-insensitive and exact, never
fuzzy):
- Daily notes (type `daily`) are excluded unless `--include-daily`.
- Archived notes are excluded unless `--include-archived`.
- `--meta key=value` (repeatable): filters on **any** frontmatter field, well-known
  ones included. Values for a key are alternatives separated by `|` (`priority=high|low`).
  Different keys are ANDed. The **same key given twice** is an error (`--meta <k> given
  twice (<a> and <b>); use <k>=a|b for alternatives`); a spec without `=` or with an empty
  key is `--meta "<s>" is not key=value`; a spec with no non-blank value is `--meta "<s>"
  has no value`. An absent field never matches. Field lookup by key (case-insensitive for
  these): `type`, `title`, `description`, `resource`, `tags` (all tags), `trust` (derived
  tier), `generated.by`, `timestamp` (= `generated.at`, or the legacy `timestamp`);
  otherwise the exact extra-frontmatter key — a list-valued field yields all its
  elements, other scalars their string form.
- `--type`: note type equals (case-insensitive).
- `--category`: at least one observation has this category.
- `--relation`: at least one relation has this type.
- `--tags a,b`: **all** listed tags present.

**Scoring** (only when there are terms). Every term must match somewhere in the note
(AND); otherwise the note is excluded. A term matches a token **exactly**, or **by
prefix** when the term has at least 3 characters (so `token` finds `tokens`, but `to`
does not match everything). There is no stemming. For each term the contribution is
the sum over matching tokens, weighted by field:

| Field | Weight |
| --- | --- |
| title | 10 |
| description | 6 |
| frontmatter tags | 5 |
| each observation (category + content + its tags + its context) | 3 |
| full body text | 1 |

A term inside an observation therefore scores both the observation weight and the body
weight (intended: a deliberately recorded fact is a stronger signal than prose). The
note's score is the sum over terms. `matches` is the list of observations in which any
term matched.

**Ordering and limit.** With terms: score descending, ties by ID ascending. With no
terms (filters only): most recently modified file first, `score` 0. `--limit` defaults to
20; a non-positive limit means unlimited.

Machine output: header `success: true`, `count`; then per hit `id`, `title`, `type`,
`description`, `score`, `matches` (list of `[category] content` strings; **all** matched
observations). Text: per hit `Title  [type]  id`, an indented description, up to 3
matched observations as `- [category] content`, then `... and N more`; `No results.` if
empty.

### 9.12 `context`

```
rit context <ref> [--depth N]
```

Breadth-first traversal outward along the link graph from the starting note (depth
default 1), over **all** in-bundle links to existing concepts (not only typed relations).
Output order: start note, then reachable notes in BFS order, each at most once.
Machine: header `success`, `start` (ID), `count`; then per note `id`, `title`, `type`,
`description`, `links` (distinct in-bundle targets). Text: `- Title (id) - description`.

### 9.13 `recent`, `ls`

- `recent [--limit N]` (default 20): concepts by **file modification time**, newest first
  (so out-of-band edits count; it does not use `log.md`). Header `success`, `count`; per
  note `id`, `title`, `type`, `description`, `modified` (UTC RFC 3339). Text: `YYYY-MM-DD
  HH:MM  title  id`.
- `ls [DIR]`: concept IDs in directory order (every `.md` file as described in §3.1,
  whether or not it parses), optionally only those under `DIR/`. Machine: one document
  `success`, `count`, `ids` (list; explicit `[]` if none). Text: one ID per line.

### 9.14 `daily`

```
rit daily [DATE] [--note PARA|-]... [--description D] [--obs ...]... [--rel ...]...
    [--source ...]... [--refresh]
```

Semantics in §7. `--note` is repeatable; each non-blank value (trimmed) becomes its own
paragraph, joined by blank lines; `--note -` reads stdin and may be given only once
(`--note - can be given only once; stdin is read to the end`). `--refresh` only
re-derives `# Touched`, never creates the note, and ignores the other flags.

Result: `success`, `status` (`created`|`updated`|`refreshed`|`missing`), `date`, `id`,
`path` (omitted for `missing`), `touched` (entry count), and — machine mode, except
`missing` — `content` (the entire note, so a resuming agent needs no second call).
Text: `<Status> <path> (N touched)`, a blank line, then the note verbatim; `missing`
prints `No daily note for <date>.` and still exits successfully (`--refresh` is an
explicit "do not create", so "nothing to refresh" is an honest answer, not a failure).

### 9.15 `index`

```
rit index [--check]
```

Regenerates (or, with `--check`, only reports) the `index.md` files (§6.1). Result:
`success`, `count`, and a list named `wrote` (or `stale` with `--check`), explicit `[]`
when nothing changed. Text: `Indexes are current.` or `Wrote <path>` / `Stale <path>` lines.
With `--check`, drift is a **failure**: emit `success: false`, `error: <N> index file(s)
are stale (run: rit index)`, `count`, `stale`, then exit 1.

### 9.16 `lint`

See §8. Machine output: header `success` (false when the run fails), `errors`,
`warnings`, `error` (the failure reason, only when failing); then one document per
finding: `severity`, `rule`, `id`, `path`, `message`. Text: findings grouped under their
file path as `  [<severity>] <rule>  <message>`, then `N error(s), M warning(s)`; `No
problems found.` if none.

### 9.17 `types`

`types` lists the active vocabulary in declaration order: header `success`, `count`,
`config` (absolute path of the bundle's `ritari.yaml`, whether or not it exists); then per
type `name`, `dir`, `summary`. Text: `name  dir/  summary`.

`types --yaml` writes, raw in both modes, the vocabulary as `ritari.yaml` content: three
comment lines explaining that the list is the whole set, that it replaces the built-in
types rather than adding to them, and that `note` and `daily` are always present;
followed by `types:` and a block list of `name`/`dir`/`summary`. It exists because a
declared vocabulary replaces rather than extends the default, so the sane way to write one
is to start from the current set and edit it.

### 9.18 `project`

- `project list`: header `success`, `count`, `default`; per project `name`, `path`,
  `is_default`. Empty registry: `success`, `count: 0`, `projects: []`, text `No projects.
  Create one: memory project init <name> <path>`. Text lines: `* name  path` (the default
  marked `*`).
- `project init <name> <path> [--default]`: **creates** a new bundle and registers it.
  Fails if the name exists, or if the directory already holds concepts (`<path> already
  holds notes; use \`rit project add <name> <path>\` to register it`; an empty or
  nonexistent directory is fine). Creates the directory, `sources/` with `README.txt`
  (if `sources/` is absent), and the root `index.md`. Writes **no** `ritari.yaml`: an absent
  file means the default vocabulary, and writing the defaults out would pin them. The
  registry is saved only after the bundle exists. Result: `success`, `status: created`,
  `name`, `path`.
- `project add <name> <path> [--default]`: registers an **existing** directory without
  modifying it (a cloned bundle brings its own indexes and config). Errors if the path
  does not exist (`<path> does not exist; use \`rit project init …\` to create it`) or is
  not a directory. Result: `status: added`.
- `project remove <name>`: `status: removed`. `project default <name>`: `status: default`.
  Both report `name`.

### 9.19 `config author`

`config author [VALUE]`: with no value prints the configured author (machine: `success`,
`author`, omitted when unset); with a value stores it in the registry (`success`,
`status: set`, `author`). Values are conventionally `human:<id>`.

### 9.20 `skills install`

```
rit skills install [--agent agents|claude] [--dir PATH] [--force]
```

Writes each embedded skill (§11) to `<dir>/<skill-name>/SKILL.md`. Default target is
`~/.agents/skills` (agent `agents`); agent `claude` is `~/.claude/skills`. `--agent` and
`--dir` are mutually exclusive (`--agent and --dir are alternatives; pass one`); an unknown
agent is `unknown agent "<x>" (known: agents, claude); use --dir for anything else`. Existing
`SKILL.md` files are left alone unless `--force`. Result: `success`, `dir`, `count`,
`installed` (sorted skill names; explicit `[]` if none). Text: `Installed skills to <dir>:
a, b` or `All skills already present in <dir> (use --force to overwrite).`

### 9.21 `version`

Prints `<tool> <version>` (text) / `success`, `version` (machine). The version string is
whatever the build identified itself as (a release version, or a short VCS revision with a
`-dirty` suffix when built from a modified tree, or `dev`).

---

## 10. Behavioral details worth getting right

### 10.1 Loading and tolerance

Loading a bundle reads every concept file. A file that cannot be parsed is **skipped** and
recorded as a problem keyed by its bundle-relative path; the load succeeds. Only `lint`
(and an attempt to address that exact file) surfaces it. Parsing an unfamiliar-but-valid
file (no frontmatter, no type, extra keys) always succeeds.

### 10.2 Reading never mutates

Only commands documented as writers modify files. Reading, searching, linting, `index
--check`, and `ls` leave the bundle byte-for-byte unchanged.

### 10.3 Idempotence

Repeating a `note write` with the same observations/relations/sources produces no further
change to the note (counts of 0), and a repeated fact with a new source gains only the new
source reference. Repeating a write whose log event is identical to one already logged today adds
no log line. Index synchronization writes only what changed.

### 10.4 Warnings do not block

An undeclared type, dangling links, and all lint warnings are advisory. Writing such content
succeeds.

### 10.5 No fetching

No command opens a network connection or reads anything outside the bundle, the registry, and
(for the actor) the VCS user email.

### 10.6 Performance expectations

Because there is no index, a query reads the whole bundle. A bundle of a few hundred notes
SHOULD answer a search in well under a second including process start; several thousand notes
SHOULD stay within a few hundred milliseconds. Loading SHOULD be parallel.

### 10.7 Concurrency and atomic writes

Several agents may drive one bundle at once, so writers are serialized:

- Every command that writes to a bundle holds an exclusive advisory lock on
  `<bundle root>/.ritari.lock` (created if absent; never deleted; a dot-file, so never
  traversed) from before it first reads the bundle until it has finished recording.
  Read-only commands take no lock.
- A writer that cannot obtain the lock within 10 seconds fails with `bundle is locked by
  another process (<lock path>)`.
- Every file the tool writes (notes, `index.md`, `log.md`, the registry, installed
  skills) is written atomically: to a temporary file in the same directory, then renamed
  over the target, so a reader never sees a partial file.

### 10.8 Portability

The tool MUST work on macOS, Linux, and Windows. Concept IDs, links, and every path
printed in machine output that is bundle-relative use `/`; absolute paths use the
platform's native form. File names are compared case-sensitively when resolving IDs
and when recognizing reserved files (even on case-insensitive file systems), except
where this specification says otherwise. On a case-insensitive file system, `lint`
reports names that differ from a reserved name only in case (§3.1).
Files are written with `\n` line endings on every platform; `\r\n` is accepted on read.

---

## 11. Agent skills

The tool carries, inside the executable, a fixed set of **agent skills**: markdown
documents that teach an agent to drive the tool, installed by `skills install` (§9.20).
Each skill document begins with YAML frontmatter containing `name` (the skill name, which is
also its install directory name) and `description` (one line stating what it is for and when
to use it), followed by guidance. The skills are part of the product: **any behavior change
that alters how an agent should drive the CLI requires the matching skill to be updated in
the same change.**

The implementer MUST ship these seven skills (`ritari` is the core skill's name and the
others are prefixed by it). Content below states what each must
teach; wording is the implementer's.

1. **`ritari`** (core). The bundle model (markdown files, no database). The output contract
   (§9.1) including that `note view`/`read` return raw bytes and that `--output text` gives
   human output. The leading-`-` rule: values starting with `-` need `--flag=value`
   (with examples for `--content` and `--find`). Where to start (`recent`, `daily`, `ls`).
   **Search before you write**, with the full set of search filters and the note that
   matching is by whole word or prefix with no stemming. Writing (including that writing to an
   existing title *adds*). Reading and traversal (`note read`, `note view`, `context`; the three
   equivalent ways to address a note). Editing and moving, and that **`move` is the only correct
   way to relocate a note**; the difference between `note edit`'s *set* semantics and `note
   write`'s *merge* semantics for frontmatter flags; archive rather than delete. Projects
   and their resolution order. Housekeeping (`index`, `index --check`; never hand-edit
   generated files). Principles: files are the source of truth; search before writing; one fact
   per observation; link generously.
2. **`ritari-notes`** (writing well). The note format and an annotated example; identity is the path;
   how to pick a type (a table of the default types and when to use each, preferring a specific
   type over `note`); that a bundle may declare its own vocabulary and how to discover it (the
   root index names the type beside each directory; `types`; the write-time warning); always
   write a description; observations (one fact per line, reusable categories, tags, context,
   checkboxes are not observations); relations as links (single-token types with underscores,
   title or ID targets, forward references are fine, prose links are edges too); recording
   sources (`--source`, `--resource`); trust signals (`generated` is automatic, `verified` is
   deliberate and not to be called reflexively); archiving; writing again adds; **`--tags` merges
   but `--description`/`--resource` replace and `--meta` replaces per key**, with a warning about
   batch loops that would clobber existing descriptions; a before-you-write checklist (search,
   add to existing, link).
3. **`ritari-daily`** (session continuity). Read today's note and recent ones when starting; write the
   narrative and open threads before finishing; `--note` is repeatable (one paragraph each) and
   repeat calls append; sections the agent adds under its own heading are left alone while prose
   above the first heading *is* the narrative; never write `# Touched` (generated; refresh with
   `--refresh`); what belongs where (narrative vs observations vs task state — task state belongs
   to the task, not the day); daily notes are excluded from search by default; promote durable
   facts out; close open threads.
4. **`ritari-tasks`** (multi-step work that survives context loss). A task is a note of type `task`;
   when to create one; creating with steps, queryable state in **both** frontmatter (`state`,
   never `status`, which is OKF lifecycle) and an observation, and resumption context; finding open tasks; updating progress (tick steps with a
   section replace, record learnings with sources, move a pointer); completing with an outcome;
   writing for resumability (context beats steps; reusable categories such as `state`, `context`,
   `decision`, `blocked`, `outcome`, `next-step`); guidelines (one task per unit of work, concrete
   steps, externalize early, link the work, close finished tasks, state lives in the task).
5. **`ritari-ingest`** (external material into notes). The tool never fetches: the agent acquires
   material with its own tools; provenance is a pointer, not a copy (`--resource`, `--source`,
   short summary, never paste the source; record an "as of" observation for changing sources);
   whether a source is a concept or merely a vehicle (meetings and documents get notes, emails and
   chat messages usually do not — their facts go on the notes they belong to, cited to the
   message); workflow (acquire, classify, extract, search, plan, write); search several ways before
   proposing; present **one plan and take one approval**; write the source note and update the
   notes it touches, each cited to the same source; action items become tasks; a source with no
   URL is a plain-text descriptor (`--source "Email from …, 2026-07-29"`); raw material kept in
   `sources/` is cited by bundle path (`/sources/<file>`); `sources/` rules (never `.md`, never hidden directories);
   hedge inferences; do not re-ingest.
6. **`ritari-reflect`** (consolidation). Daily notes accumulate; reflection turns them into durable
   memory. When to run (end of session, on a schedule, after compaction). Workflow: gather
   (`recent`, daily notes, open threads, active tasks), evaluate (will this still be true next
   month? durability not importance), promote (observations on **typed** notes, cited with a link
   to the day, search first, additive), log (record the reflection in today's note). Resolving
   open threads. **Never rewrite daily notes** (they are the raw record). Guidelines: distil,
   be selective, prefer updating over creating, flag uncertainty, archive what is superseded,
   run lint occasionally.
7. **`ritari-lint`** (findings into decisions). Lint finds problems; judgment belongs to the user.
   Workflow: run, triage, fix mechanical ones (and say so), ask about the rest, log. Mechanical
   fixes: `stale-index` (regenerate) and `type-directory-mismatch` (the move the message names).
   For every other finding, present 2–4 concrete options with consequences using the harness's
   structured multiple-choice tool if available, else a numbered list — with per-rule guidance
   for `orphan`, `duplicate-title`, `unknown-type`/`unknown-category`, `dangling-link`,
   `stale-verification`, `missing-title`/`missing-description`, `missing-type`/`unparseable`.
   Batch related questions and offer "skip the rest". Never guess at a merge or a relation;
   report mechanical fixes; an unresolved finding is fine; read the note before proposing.

---

## 12. Acceptance scenarios

A conforming implementation SHOULD behave as follows. (Run non-interactively, so output is
machine mode.)

1. **Create a project.** `project init work <dir> --default` on an empty directory → `success:
   true`, `status: created`; `<dir>/index.md` exists containing `okf_version: "0.2"`;
   `<dir>/sources/README.txt` exists; no `ritari.yaml` exists; `project list` shows `work` as
   default. Running it again with the same name fails; running it on a directory with notes
   fails and advises `project add`.
2. **Write and read back.** `note write --title "Auth uses JWT" --type decision --description
   "…" --tags auth,security --obs "decision: Use JWT for stateless auth #security" --rel
   "depends_on: User Database Schema" --source "PR #1234 https://github.com/acme/api/pull/1234"`
   → `status: created`, `id: decisions/auth-uses-jwt`, `added_observations: 1`,
   `added_relations: 1`, `added_sources: 1`. The file has the shape shown in §3.2: frontmatter
   `sources` holds one entry `id: "1"`, `resource: https://github.com/acme/api/pull/1234`,
   `title: "PR #1234"`; the observation carries ` [^1]`; the relation is `- depends_on [User
   Database Schema](/notes/user-database-schema.md)` — a forward reference; the body ends with
   `[^1]: [PR #1234](https://github.com/acme/api/pull/1234)`; `generated.by` is set. `log.md` has a `Create` entry for today;
   `decisions/index.md` and the root index list it. `lint` reports a `dangling-link` warning
   and exits 0; `lint --strict` exits 1.
3. **Convergence.** Repeating the same command reports `status: updated` with
   `added_observations: 0`, `added_relations: 0`, and the file is unchanged. Repeating with the
   same `--obs` and a *different* `--source` leaves one observation line whose refs are
   ` [^1] [^2]`, two `sources` entries, and two footnote definitions.
4. **Additive vs. replace.** `note write --title "Auth uses JWT" --content "x"` (no `--force`)
   fails; with `--force` it replaces the body and reports `replaced`.
5. **Title fallback.** Writing an existing title with a different `--type` adds to the existing
   note rather than creating a second file.
6. **Move rewrites links.** After `note move decisions/auth-uses-jwt references/auth-uses-jwt`,
   every note that linked to the old path now links to the new one, `rewrote_links_in` lists
   them, and the old directory's index is removed if it is now empty.
7. **Archive.** `note archive <id>` → excluded from `search`, included with `--include-archived`,
   listed under `# Archived` in its directory index, file contains `archived: true`, `lint` raises
   no hygiene warnings for it; `unarchive` removes the key.
8. **Malformed file.** A concept with broken YAML frontmatter is invisible to `search`/`recent`
   (but still named by `ls`), does not fail those commands, and `lint` reports an `unparseable`
   **error** (exit 1).
9. **Custom vocabulary.** A `ritari.yaml` declaring one type `incident` (dir `incidents`) → `types`
   lists `incident`, `note`, `daily`; `note write --type meeting …` succeeds with a `warning`
   naming the declared types and `meetings/`; `lint` flags it as `unknown-type`. A `ritari.yaml`
   with two types sharing a directory makes every bundle command fail with a message naming the
   conflict.
10. **Daily.** `daily --note "A" --note "B" --obs "open-thread: x"` creates `daily/<today>` with
    two paragraphs and an observation; a second call with `--note "C"` appends a third paragraph
    above the first heading; after creating and then deleting some other note the same day,
    `daily --refresh` lists the deleted one as `- <id> — deleted` and the created-then-updated
    one as `created`. A date like `2026-7-1` is rejected.
11. **Search.** With two notes, one with `jwt` in its title and one only in body text, a query
    `jwt` ranks the title match first; a query for a term absent from either returns
    `count: 0`; the 3-character prefix rule makes `tok` match `tokens` while `to` matches only
    the exact token `to`; `--meta state=active --meta state=done` is an error while
    `--meta 'state=active|done'` is accepted.
12. **Trust.** `note verify <id> --by human:a@b.c` → `trust: human-reviewed`; a later content
    write makes `lint` raise `stale-verification` (when `generated.at` is later than the
    verification time).
13. **Index drift.** After editing a note's description by hand, `index --check` exits 1 listing
    the stale index; `index` rewrites only that file; `index --check` then exits 0.
14. **Output failures.** Any failing command run in machine mode prints `success: false` and an
    `error` on stdout, a message on stderr, and exits non-zero.
15. **OKF conformance.** After each scenario above, `okflint validate` (§1) reports zero errors
    for the bundle, and it reports an error for a bundle with a malformed file.
16. **Tags.** `--obs "bug: Fixed #42 and color #ff00aa #auth"` yields the tag `auth` only; `#42`
    and `#ff00aa` stay in the content.
