// Package yamlx renders YAML scalars so that they read back as exactly the
// string that was written.
package yamlx

import (
	"slices"
	"strconv"
	"strings"
	"unicode"

	"go.yaml.in/yaml/v3"
)

const (
	tagStr       = "!!str"
	TagTimestamp = "!!timestamp"
)

// Scalar renders s for use as a value in block context (`key: <s>`). It is
// written plain when that reads back as the same string, and double-quoted
// otherwise. Extra resolved tags (such as TagTimestamp) may be allowed plain.
func Scalar(s string, allow ...string) string {
	if plainOK(s, false, allow) {
		return s
	}
	return Quote(s)
}

// FlowScalar renders s for use inside a flow collection (`[a, b]`, `{k: v}`).
func FlowScalar(s string, allow ...string) string {
	if plainOK(s, true, allow) {
		return s
	}
	return Quote(s)
}

// Quote renders s as a YAML double-quoted scalar. Go's escape sequences are a
// subset of YAML's, so strconv's quoting is valid YAML.
func Quote(s string) string {
	return strconv.Quote(s)
}

func plainOK(s string, flow bool, allow []string) bool {
	if s == "" || s != strings.TrimSpace(s) {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) || !unicode.IsPrint(r) && r != ' ' {
			return false
		}
	}
	if strings.ContainsRune("-?:,[]{}#&*!|>'\"%@`", rune(s[0])) {
		return false
	}
	if strings.Contains(s, ": ") || strings.Contains(s, " #") || strings.HasSuffix(s, ":") {
		return false
	}
	if flow && strings.ContainsAny(s, ",[]{}") {
		return false
	}
	// YAML 1.1 readers (still common) resolve these as booleans.
	switch strings.ToLower(s) {
	case "y", "n", "yes", "no", "on", "off":
		return false
	}

	var doc yaml.Node
	src := "k: " + s
	if flow {
		src = "[" + s + "]"
	}
	if err := yaml.Unmarshal([]byte(src), &doc); err != nil || len(doc.Content) != 1 {
		return false
	}
	root := doc.Content[0]
	var n *yaml.Node
	switch {
	case !flow && root.Kind == yaml.MappingNode && len(root.Content) == 2:
		n = root.Content[1]
	case flow && root.Kind == yaml.SequenceNode && len(root.Content) == 1:
		n = root.Content[0]
	default:
		return false
	}
	if n.Kind != yaml.ScalarNode || n.Style != 0 || n.Value != s {
		return false
	}
	tag := n.ShortTag()
	return tag == tagStr || slices.Contains(allow, tag)
}
