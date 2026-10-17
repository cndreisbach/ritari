package yamlx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

func TestScalarPlain(t *testing.T) {
	for _, s := range []string{
		"decision",
		"Auth uses JWT",
		"https://github.com/acme/api/pull/1234",
		"human:jane@acme.com",
		"Stateless auth via short-lived JWTs, chosen over server sessions.",
		"Tokens — expire",
	} {
		assert.Equal(t, s, Scalar(s), s)
	}
}

func TestScalarQuoted(t *testing.T) {
	for _, s := range []string{
		"", " lead", "trail ", "true", "yes", "null", "~", "123", "1.5",
		"PR #1234", "a: b", "ends:", "-dash", "[x]", "{x}", "*ref", "&anchor",
		"'single'", "\"double\"", "multi\nline", "tab\there", "2026-07-29",
	} {
		got := Scalar(s)
		assert.NotEqual(t, s, got, "expected quoting for %q", s)
		assertRoundTrip(t, s, "k: "+got)
	}
}

func TestScalarAllowTimestamp(t *testing.T) {
	ts := "2026-07-29T14:30:00Z"
	assert.Equal(t, `"`+ts+`"`, Scalar(ts))
	assert.Equal(t, ts, Scalar(ts, TagTimestamp))
	assert.Equal(t, ts, FlowScalar(ts, TagTimestamp))
}

func TestFlowScalar(t *testing.T) {
	assert.Equal(t, "auth", FlowScalar("auth"))
	assert.Equal(t, "human:jane@acme.com", FlowScalar("human:jane@acme.com"))
	for _, s := range []string{"a,b", "a]", "a}", "true"} {
		got := FlowScalar(s)
		assert.NotEqual(t, s, got, s)
		var v []string
		require.NoError(t, yaml.Unmarshal([]byte("["+got+"]"), &v))
		assert.Equal(t, []string{s}, v)
	}
}

func assertRoundTrip(t *testing.T, want, src string) {
	t.Helper()
	var m map[string]string
	require.NoError(t, yaml.Unmarshal([]byte(src), &m), src)
	assert.Equal(t, want, m["k"], src)
}
