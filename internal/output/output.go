// Package output writes machine-mode results: streams of flat YAML documents
// whose values are scalars or inline lists (§9.1).
package output

import (
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/cndreisbach/ritari/internal/yamlx"
)

// Doc is one flat YAML mapping with ordered keys.
type Doc struct {
	keys []string
	vals []any
}

// List is a list value, omitted when empty.
type List []string

// ExplicitList is a list value written as `[]` when empty.
type ExplicitList []string

// New starts a document.
func New() *Doc { return &Doc{} }

// Set appends a key. Empty strings and empty Lists are omitted when written.
func (d *Doc) Set(key string, v any) *Doc {
	d.keys = append(d.keys, key)
	d.vals = append(d.vals, v)
	return d
}

// Write renders documents separated by `---`.
func Write(w io.Writer, docs ...*Doc) error {
	var b strings.Builder
	for i, d := range docs {
		if i > 0 {
			b.WriteString("---\n")
		}
		d.render(&b)
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func (d *Doc) render(b *strings.Builder) {
	for i, k := range d.keys {
		switch v := d.vals[i].(type) {
		case string:
			if v == "" {
				continue
			}
			b.WriteString(k + ":" + scalar(v) + "\n")
		case int:
			fmt.Fprintf(b, "%s: %d\n", k, v)
		case bool:
			fmt.Fprintf(b, "%s: %t\n", k, v)
		case List:
			if len(v) == 0 {
				continue
			}
			b.WriteString(k + ": " + flowList(v) + "\n")
		case ExplicitList:
			b.WriteString(k + ": " + flowList(v) + "\n")
		default:
			panic(fmt.Sprintf("output: unsupported value %T for %q", v, k))
		}
	}
}

func flowList(items []string) string {
	parts := make([]string, len(items))
	for i, s := range items {
		parts[i] = yamlx.FlowScalar(s)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// scalar renders a value including its leading separator: " plain",
// ` "quoted"`, or a literal block for multi-line text.
func scalar(s string) string {
	if strings.Contains(s, "\n") && literalOK(s) {
		return literal(s)
	}
	return " " + yamlx.Scalar(s)
}

func literalOK(s string) bool {
	for _, r := range s {
		if r != '\n' && r != '\t' && (unicode.IsControl(r) || !unicode.IsPrint(r) && r != ' ') {
			return false
		}
	}
	return strings.TrimSpace(s) != ""
}

// literal renders s as a `|` block scalar that reads back exactly.
func literal(s string) string {
	body := strings.TrimRight(s, "\n")
	trailing := len(s) - len(body)
	chomp := "-"
	switch {
	case trailing == 1:
		chomp = ""
	case trailing > 1:
		chomp = "+"
	}
	ind := ""
	if strings.HasPrefix(body, " ") || strings.HasPrefix(body, "\n") {
		ind = "2"
	}
	var b strings.Builder
	b.WriteString(" |" + ind + chomp + "\n")
	lines := strings.Split(body, "\n")
	if chomp == "+" {
		for i := 1; i < trailing; i++ {
			lines = append(lines, "")
		}
	}
	for i, l := range lines {
		if l != "" {
			b.WriteString("  " + l)
		}
		if i < len(lines)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
