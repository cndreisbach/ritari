# Ritari: notes for agents

Ritari (`rit`) is a long-term memory CLI for AI agents. It stores a knowledge
base as an OKF v0.2 bundle (plain markdown files with YAML frontmatter) and
computes everything else from those files. Written in Go.

## Source of truth

- **`SPECIFICATION.md` is the contract.** File formats, CLI flags, output keys,
  error messages, and algorithms all come from it. If you change behavior,
  update the spec in the same change; if code and spec disagree, the spec wins
  unless the user says otherwise.
- The OKF format: <https://github.com/GoogleCloudPlatform/open-knowledge-format/blob/main/SPEC.md>.
- **Skills are part of the product.** Any change to how an agent should drive
  the CLI must update the matching `internal/skills/*/SKILL.md` in the same
  change (spec §11).

## Layout

| Path | What lives there |
| --- | --- |
| `cmd/rit` | `main` and the testscript acceptance suite (`testdata/scripts/*.txtar`) |
| `internal/note` | One concept file: frontmatter, observations, relations, links, sources/footnotes, merge, edit ops |
| `internal/vocab` | Type vocabulary (`ritari.yaml`), slug/plural/humanize |
| `internal/bundle` | Loading, resolving, link graph, writes, lock, atomic writes, `log.md`, `index.md`, daily notes, move/delete/archive |
| `internal/search` | Search scoring and filters |
| `internal/lint` | Lint rules |
| `internal/registry` | Per-user project registry (`~/.config/ritari/config.yaml`) |
| `internal/actor` | Actor detection for `generated.by` / `verified[].by` |
| `internal/output` | Machine-mode flat YAML document streams |
| `internal/yamlx` | YAML scalar quoting that round-trips exactly |
| `internal/cli` | kong command tree; text and machine output |
| `internal/skills` | Embedded agent skills, installed by `rit skills install` |

## Tooling

Tools are managed with mise (`mise.toml`): Go, golangci-lint, govulncheck,
`okf` (the third-party OKF linter used as a conformance check), and `svu`
(next-version calculation for releases).

```bash
mise install
mise run test          # go test ./...
golangci-lint run ./...
go build -o rit ./cmd/rit
```

Before finishing any change: `gofmt`, `go vet ./...`, `go test ./...`, and
`golangci-lint run ./...` must all be clean.

## CI and releases

- `.github/workflows/ci.yml` runs on every PR: `go test ./...`,
  `golangci-lint`, and `govulncheck` on Linux. It sets `RITARI_SKIP_OKF=1`, so
  the acceptance tests skip `okf lint` there (okf takes a long time to build).
  Run the full suite locally (with okf installed) before merging.
- `.github/workflows/release.yml` runs on every push to `main`: the CI job,
  then the release. The next version comes from the Conventional Commits since
  the last `v*` tag (via `svu`): a breaking change bumps major, `feat` bumps
  minor, `fix`/`perf` bump patch, and anything else releases nothing. The first
  release is `v1.0.0`. The job creates the GitHub release (notes from
  `.github/scripts/release-notes.sh`), then regenerates `Formula/ritari.rb`
  with `.github/scripts/update-formula.sh` and commits it to `main`.
- **Never edit `Formula/ritari.rb` by hand**; change the generator script.
- The repo is its own Homebrew tap. The formula builds from the tagged source
  and stamps the version with
  `-ldflags "-X github.com/cndreisbach/ritari/internal/cli.version=<v>"`:

  ```bash
  brew tap cndreisbach/ritari https://github.com/cndreisbach/ritari
  brew install ritari
  ```

- PRs are merged with merge commits, not squashed, so **every commit message**
  on a branch counts toward the version bump. A commit that isn't in
  Conventional Commits format is ignored.

## Conventions

- **Libraries:** kong for the CLI, `go.yaml.in/yaml/v3` (not `gopkg.in/yaml.v3`)
  for YAML, goldmark (with the footnote extension) for markdown, gofrs/flock for
  the writer lock, testify for tests, testscript for acceptance tests.
- **Tests:** every package has unit tests with testify. CLI behavior is tested
  end to end in `cmd/rit/testdata/scripts`; each scenario ends with
  `okflint <bundle>`, which runs `okf lint` (skipped if it isn't installed).
  Acceptance tests pin the clock with `RITARI_TEST_NOW` (RFC 3339) and the actor
  with `RITARI_AGENT`. `RITARI_TEST_NOW` is a test hook, not a user feature.
  testscript can't take multi-line quoted arguments; use `stdin <file>` with
  `--content -` instead.
- **Writes:** every write goes through `bundle.Update` (lock + record) and
  `WriteFileAtomic`/`writeIfChanged`. Never write files directly.
- **The tool never touches the network.** No HTTP clients.
- **Portability:** must work on macOS, Linux, and Windows. Concept IDs and
  bundle-relative paths always use `/`; convert with `filepath.ToSlash` /
  `FromSlash` at the filesystem boundary.
- **Provenance:** sources live in frontmatter `sources` with numeric ids; the
  `[^n]:` footnote definitions at the end of a body are derived from them on
  every render. Don't hand-maintain them.
- **Workflow state** goes in `state:`, never `status:` (OKF reserves `status`
  for lifecycle: draft, stable, deprecated).

## Commits

Use **Conventional Commits** (<https://www.conventionalcommits.org/>) for every
commit:

```
<type>(<optional scope>): <imperative summary, lowercase, no period>

<optional body: what and why, wrapped at ~72 columns>

<optional footers>
```

- Types: `feat`, `fix`, `docs`, `test`, `refactor`, `perf`, `build`, `ci`,
  `chore`, `style`, `revert`.
- Scopes are optional; use a package or area name when it helps: `note`,
  `bundle`, `vocab`, `search`, `lint`, `cli`, `skills`, `registry`, `spec`.
- Breaking changes: add `!` after the type/scope (`feat(cli)!: …`) and a
  `BREAKING CHANGE:` footer describing the break.
- One logical change per commit. Keep the subject under ~72 characters.

Examples:

```
feat(search): support --relation filter
fix(bundle): keep derived footnotes at the end after edits
docs(spec): record the hex-tag exclusion rule
test(cli): cover leading-dash flag values
```
