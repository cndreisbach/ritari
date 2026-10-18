package lint

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cndreisbach/ritari/internal/bundle"
	"github.com/cndreisbach/ritari/internal/note"
	"github.com/cndreisbach/ritari/internal/search"
)

// TestUpstreamBundles reads the example bundles from Google's OKF repository
// (testdata/okf-upstream/README.md), which rit did not write.
func TestUpstreamBundles(t *testing.T) {
	cases := []struct {
		name     string
		concepts int
		trust    map[string]int
		query    string
		top      string
	}{
		{"acme_retail", 9, map[string]int{note.TrustHumanReviewed: 8, note.TrustUnverified: 1}, "gross margin", "metrics/gross-margin"},
		{"crypto_bitcoin", 9, map[string]int{note.TrustUnverified: 9}, "duplicate", "references/metrics/duplicate_transactions"},
		{"ga4", 9, map[string]int{note.TrustUnverified: 9}, "purchasers", "references/metrics/purchasers"},
		{"stackoverflow", 26, map[string]int{note.TrustUnverified: 26}, "accepted answer", "tables/posts_answers"},
	}
	// Their types aren't in rit's default vocabulary or directories, and their
	// hand-made indexes differ from the ones rit generates.
	expected := map[string]bool{UnknownType: true, TypeDirectoryMismatch: true, StaleIndex: true, Orphan: true}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b, err := bundle.Open(filepath.Join("testdata", "okf-upstream", c.name))
			require.NoError(t, err)
			assert.Empty(t, b.Problems)
			assert.Len(t, b.Concepts, c.concepts)

			trust := map[string]int{}
			for _, cn := range b.Concepts {
				fm := cn.Note.FM
				assert.Empty(t, fm.Malformed, cn.ID)
				assert.NotEmpty(t, fm.Type, cn.ID)
				assert.NotEmpty(t, fm.Title, cn.ID)
				if assert.NotNil(t, fm.Generated, cn.ID) {
					assert.NotEmpty(t, fm.Generated.By, cn.ID)
					assert.NotEmpty(t, fm.Generated.At, cn.ID)
				}
				for _, s := range fm.Sources {
					assert.NotEmpty(t, s.Resource, cn.ID)
				}
				trust[fm.Trust()]++
			}
			assert.Equal(t, c.trust, trust)

			rep, err := Run(b, Options{})
			require.NoError(t, err)
			assert.Zero(t, rep.Errors)
			for _, f := range rep.Findings {
				assert.True(t, expected[f.Rule], "unexpected %s on %s: %s", f.Rule, f.Path, f.Message)
			}

			hits, err := search.Run(b, search.Options{Query: c.query})
			require.NoError(t, err)
			require.NotEmpty(t, hits)
			assert.Equal(t, c.top, hits[0].Concept.ID)
		})
	}
}
