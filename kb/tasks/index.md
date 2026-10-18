# Tasks

* [Check links to non-markdown files by existence](/tasks/check-links-to-non-markdown-files-by-existence.md) - rit lint reports links to existing images and other files as dangling, and never checks links into sources/.
* [Match reserved file names exactly](/tasks/match-reserved-file-names-exactly.md) - Recognize only lowercase index.md and log.md as reserved, and lint case variants on case-insensitive file systems.
* [Stop flagging links to index files as dangling](/tasks/stop-flagging-links-to-index-files-as-dangling.md) - Links to index.md/log.md are valid OKF but rit's graph and lint treat them as dangling.
* [Test against the upstream OKF example bundles](/tasks/test-against-the-upstream-okf-example-bundles.md) - Vendor the four bundles from Google's OKF spec repo as test fixtures and check rit reads, lints, and searches them correctly.
* [Treat malformed frontmatter fields as warnings](/tasks/treat-malformed-frontmatter-fields-as-warnings.md) - A known field with the wrong shape should be ignored and reported, not make the whole note unparseable.
