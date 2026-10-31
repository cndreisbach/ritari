# Tasks

* [Store whole documents in Ritari](/tasks/store-whole-documents-in-ritari.md) - Make it easy for agents to keep plans, reports, and handoffs in Ritari in full: replace_body, hashes, and skill guidance.

# Archived

* [Check links to non-markdown files by existence](/tasks/check-links-to-non-markdown-files-by-existence.md) - rit lint reports links to existing images and other files as dangling, and never checks links into sources/.
* [Detect Pi as an agent](/tasks/detect-pi-as-an-agent.md) - Recognize the Pi coding agent in actor detection instead of falling back to the configured author.
* [Make reserved-name tests portable to case-insensitive filesystems](/tasks/make-reserved-name-tests-portable-to-case-insensitive-filesystems.md) - Two tests cannot build their fixtures on APFS, so go test ./... is red on default macOS.
* [Match reserved file names exactly](/tasks/match-reserved-file-names-exactly.md) - Recognize only lowercase index.md and log.md as reserved, and lint case variants on case-insensitive file systems.
* [Normalize the configured author identity](/tasks/normalize-the-configured-author-identity.md) - A bare author such as cndreisbach is stamped verbatim and looks like an agent name.
* [Stop flagging links to index files as dangling](/tasks/stop-flagging-links-to-index-files-as-dangling.md) - Links to index.md/log.md are valid OKF but rit's graph and lint treat them as dangling.
* [Test against the upstream OKF example bundles](/tasks/test-against-the-upstream-okf-example-bundles.md) - Vendor the four bundles from Google's OKF spec repo as test fixtures and check rit reads, lints, and searches them correctly.
* [Treat malformed frontmatter fields as warnings](/tasks/treat-malformed-frontmatter-fields-as-warnings.md) - A known field with the wrong shape should be ignored and reported, not make the whole note unparseable.
