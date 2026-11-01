# Decisions

* [Agent and model in the actor](/decisions/agent-and-model-in-the-actor.md) - Stamp an agent as <agent>/<model> when the model is known; humans never carry a model.
* [Archived notes are screened like live ones](/decisions/archived-notes-are-screened-like-live-ones.md) - Archived notes are screened like live ones.
* [Escape invisible characters in text output](/decisions/escape-invisible-characters-in-text-output.md) - Invisible characters are escaped in text-mode output only; note view and read stay raw.
* [JSON output](/decisions/json-output.md) - Add rit --output json as newline-delimited JSON documents mirroring the flat YAML stream.
* [Link titles are scanned, not flagged](/decisions/link-titles-are-scanned-not-flagged.md) - Link titles and unused reference definitions are scanned for instructions but are not hidden-content on their own.
* [Pi extension scope](/decisions/pi-extension-scope.md) - Build only the parts of a Pi extension a skill cannot do; no auto-triggered memory writes.
* [pi-ritari lives in this repo](/decisions/pi-ritari-lives-in-this-repo.md) - Ship the Pi extension as pi-ritari/ in the Ritari repo, named pi-ritari on npm, with its version coupled to the CLI release tag.
* [replace_body keeps structured sections](/decisions/replace-body-keeps-structured-sections.md) - note edit --op replace_body replaces only the document and keeps frontmatter, Observations, Relations, and derived footnotes.
* [Session-start hooks](/decisions/session-start-hooks.md) - rit install hooks writes project-scoped hooks that run the read-only rit brief at session start.
* [Skill set after routing review](/decisions/skill-set-after-routing-review.md) - Keep ritari, ritari-ingest and ritari-reflect; fold the rest into references; move session start/end out of routing.
* [The package owns brief injection](/decisions/the-package-owns-brief-injection.md) - The pi-ritari package injects rit brief; the generated .pi/extensions hook is the no-npm fallback, never both.
* [Two hidden-content lint rules](/decisions/two-hidden-content-lint-rules.md) - hidden-content warns on hidden content; hidden-instruction warns when hidden text matches the pattern bank. Both are warnings.
* [Writes warn only about hidden content they add](/decisions/writes-warn-only-about-hidden-content-they-add.md) - note write, note edit, and daily report only the screen findings a write introduces; lint reports everything.
