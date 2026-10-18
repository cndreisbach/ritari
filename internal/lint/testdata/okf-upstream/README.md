# Upstream OKF example bundles

The four example bundles from Google's Open Knowledge Format repository,
used to check that rit reads, lints, and searches bundles it did not write
(`../../upstream_test.go`).

- Source: <https://github.com/GoogleCloudPlatform/open-knowledge-format/tree/ad30107c31c06aec8a7d5636e0d1058118604e6f/bundles>
- Commit: `ad30107c31c06aec8a7d5636e0d1058118604e6f` (2026-08-21)
- License: Apache License 2.0 (`LICENSE.md`, copied from that commit)

Only the markdown files are copied, unmodified; each bundle's `viz.html`
and `acme_retail/attesters/sql_equality.py` are left out.

To update, replace the bundle directories with the markdown files from a newer
commit, update the commit above, and adjust the expectations in the test.
