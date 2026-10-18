// Package note parses and renders individual OKF concept files: frontmatter,
// observations, relations, links, sources, and body edits.
package note

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/cndreisbach/ritari/internal/yamlx"
)

// Stamp is a `{by, at}` record used by `generated` and `verified`.
type Stamp struct {
	By string
	At string
}

// Source is one entry of the `sources` frontmatter list (OKF §5.1).
type Source struct {
	ID       string
	Resource string
	Title    string
	// Extra holds the entry's other keys (author, usage_count, …) as
	// alternating key/value nodes, in their original order.
	Extra []*yaml.Node
}

// Extra is a frontmatter key the tool does not interpret. Raw is the exact
// YAML text of the key and its value, kept verbatim across rewrites.
type Extra struct {
	Key   string
	Raw   string
	Value *yaml.Node
}

// Frontmatter holds a concept's YAML metadata.
type Frontmatter struct {
	Type        string
	Title       string
	Description string
	Resource    string
	Tags        []string
	Sources     []Source
	Generated   *Stamp
	Verified    []Stamp
	Archived    bool
	Extra       []Extra
	// Malformed lists recognized fields whose values have the wrong shape
	// (§3.2). Each reads as absent and is kept verbatim in Extra.
	Malformed []Malformed
}

// Malformed is a recognized frontmatter field with the wrong shape.
type Malformed struct {
	Key    string
	Detail string
}

// Note is a parsed concept file.
type Note struct {
	FM Frontmatter
	// Body is everything after the frontmatter block and its separating
	// blank line, as it appears on disk.
	Body string
}

var knownKeys = map[string]bool{
	"type": true, "title": true, "description": true, "resource": true,
	"tags": true, "sources": true, "generated": true, "verified": true,
	"archived": true,
}

// Parse reads a concept file. A file without a frontmatter block parses
// successfully with empty frontmatter. An error means the file is unparseable.
func Parse(data []byte) (*Note, error) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	fmText, body, ok := split(text)
	n := &Note{Body: body}
	if !ok {
		n.Body = text
		return n, nil
	}
	fm, err := parseFrontmatter(fmText)
	if err != nil {
		return nil, err
	}
	n.FM = *fm
	return n, nil
}

// split separates the frontmatter text from the body. ok is false when the
// file has no complete frontmatter block.
func split(text string) (fm, body string, ok bool) {
	rest := strings.TrimLeft(text, " \t")
	first, after, found := strings.Cut(rest, "\n")
	if strings.TrimRight(first, " \t") != "---" {
		return "", "", false
	}
	if !found {
		return "", "", false
	}
	lines := strings.SplitAfter(after, "\n")
	for i, line := range lines {
		if strings.TrimRight(line, " \t\n") == "---" && strings.HasPrefix(line, "---") {
			fm = strings.Join(lines[:i], "")
			body = strings.Join(lines[i+1:], "")
			if b, cut := strings.CutPrefix(body, "\n"); cut {
				body = b
			}
			return fm, body, true
		}
	}
	return "", "", false
}

func parseFrontmatter(text string) (*Frontmatter, error) {
	fm := &Frontmatter{}
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(text), &doc); err != nil {
		return nil, err
	}
	if len(doc.Content) == 0 {
		return fm, nil
	}
	root := doc.Content[0]
	if root.Kind == yaml.ScalarNode && root.ShortTag() == "!!null" {
		return fm, nil
	}
	if root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("frontmatter is not a mapping")
	}
	lines := strings.SplitAfter(text, "\n")
	seen := map[string]bool{}
	for i := 0; i+1 < len(root.Content); i += 2 {
		k, v := root.Content[i], root.Content[i+1]
		key := k.Value
		if seen[key] {
			return nil, fmt.Errorf("duplicate key %q", key)
		}
		seen[key] = true
		var err error
		before := *fm
		switch key {
		case "type":
			fm.Type, err = scalarField("", v)
		case "title":
			fm.Title, err = scalarField("", v)
		case "description":
			fm.Description, err = scalarField("", v)
		case "resource":
			fm.Resource, err = scalarField("", v)
		case "tags":
			fm.Tags, err = tagsField(v)
		case "sources":
			fm.Sources, err = sourcesField(v)
		case "generated":
			fm.Generated, err = generatedField(v)
		case "verified":
			fm.Verified, err = verifiedField(v)
		case "archived":
			fm.Archived, err = archivedField(v)
		default:
			fm.Extra = append(fm.Extra, Extra{Key: key, Raw: rawChunk(root, i, lines), Value: v})
		}
		if err != nil {
			extra, malformed := fm.Extra, fm.Malformed
			*fm = before
			fm.Extra = append(extra, Extra{Key: key, Raw: rawChunk(root, i, lines), Value: v})
			fm.Malformed = append(malformed, Malformed{Key: key, Detail: err.Error()})
		}
	}
	return fm, nil
}

// rawChunk returns the verbatim text of the i'th key of a block mapping, or a
// re-encoding when the mapping is not laid out one key per line.
func rawChunk(root *yaml.Node, i int, lines []string) string {
	k := root.Content[i]
	if root.Style&yaml.FlowStyle == 0 && k.Column == 1 {
		start := k.Line - 1
		end := len(lines)
		if i+2 < len(root.Content) {
			end = root.Content[i+2].Line - 1
		}
		if start >= 0 && start < end && end <= len(lines) {
			chunk := strings.Join(lines[start:end], "")
			return strings.TrimRight(chunk, " \t\n")
		}
	}
	return encodePair(k, root.Content[i+1])
}

func encodePair(k, v *yaml.Node) string {
	m := &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{k, v}}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	_ = enc.Encode(m)
	_ = enc.Close()
	return strings.TrimRight(buf.String(), "\n")
}

func isNull(v *yaml.Node) bool {
	return v.Kind == yaml.ScalarNode && v.ShortTag() == "!!null"
}

// scalarField reads a string field. path names a nested field in the error
// ("" for a top-level one).
func scalarField(path string, v *yaml.Node) (string, error) {
	if isNull(v) {
		return "", nil
	}
	if v.Kind != yaml.ScalarNode {
		if path == "" {
			return "", errors.New("not a string")
		}
		return "", fmt.Errorf("%s is not a string", path)
	}
	return v.Value, nil
}

func tagsField(v *yaml.Node) ([]string, error) {
	switch {
	case isNull(v):
		return nil, nil
	case v.Kind == yaml.ScalarNode:
		if strings.TrimSpace(v.Value) == "" {
			return nil, nil
		}
		return []string{v.Value}, nil
	case v.Kind == yaml.SequenceNode:
		var tags []string
		for _, e := range v.Content {
			if e.Kind != yaml.ScalarNode {
				return nil, errors.New("not a list of strings")
			}
			if !isNull(e) {
				tags = append(tags, e.Value)
			}
		}
		return tags, nil
	}
	return nil, errors.New("not a list of strings")
}

func stampField(v *yaml.Node) (Stamp, error) {
	if v.Kind != yaml.MappingNode {
		return Stamp{}, errors.New("not a {by, at} mapping")
	}
	var s Stamp
	for i := 0; i+1 < len(v.Content); i += 2 {
		var err error
		switch v.Content[i].Value {
		case "by":
			s.By, err = scalarField("by", v.Content[i+1])
		case "at":
			s.At, err = scalarField("at", v.Content[i+1])
		}
		if err != nil {
			return Stamp{}, err
		}
	}
	return s, nil
}

func generatedField(v *yaml.Node) (*Stamp, error) {
	if isNull(v) {
		return nil, nil
	}
	s, err := stampField(v)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func verifiedField(v *yaml.Node) ([]Stamp, error) {
	switch {
	case isNull(v):
		return nil, nil
	case v.Kind == yaml.MappingNode:
		s, err := stampField(v)
		if err != nil {
			return nil, err
		}
		return []Stamp{s}, nil
	case v.Kind == yaml.SequenceNode:
		var out []Stamp
		for _, e := range v.Content {
			s, err := stampField(e)
			if err != nil {
				return nil, err
			}
			out = append(out, s)
		}
		return out, nil
	}
	return nil, errors.New("not a list of {by, at} mappings")
}

func sourcesField(v *yaml.Node) ([]Source, error) {
	if isNull(v) {
		return nil, nil
	}
	if v.Kind != yaml.SequenceNode {
		return nil, errors.New("not a list of mappings")
	}
	var out []Source
	for _, e := range v.Content {
		if e.Kind != yaml.MappingNode {
			return nil, errors.New("not a list of mappings")
		}
		var s Source
		for i := 0; i+1 < len(e.Content); i += 2 {
			k, val := e.Content[i], e.Content[i+1]
			var err error
			switch k.Value {
			case "id":
				s.ID, err = scalarField("id", val)
			case "resource":
				s.Resource, err = scalarField("resource", val)
			case "title":
				s.Title, err = scalarField("title", val)
			default:
				s.Extra = append(s.Extra, k, val)
			}
			if err != nil {
				return nil, err
			}
		}
		out = append(out, s)
	}
	return out, nil
}

func archivedField(v *yaml.Node) (bool, error) {
	if isNull(v) {
		return false, nil
	}
	if v.Kind != yaml.ScalarNode {
		return false, errors.New("not a boolean")
	}
	switch strings.ToLower(strings.TrimSpace(v.Value)) {
	case "true", "yes":
		return true, nil
	}
	return false, nil
}

// ExtraValue returns an extra key's value node, or nil.
func (fm *Frontmatter) ExtraValue(key string) *yaml.Node {
	for _, e := range fm.Extra {
		if e.Key == key {
			return e.Value
		}
	}
	return nil
}

// SetExtra sets an extra key to a string value, replacing any existing value.
func (fm *Frontmatter) SetExtra(key, value string) {
	raw := yamlx.Scalar(key) + ": " + yamlx.Scalar(value)
	node := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}
	for i, e := range fm.Extra {
		if e.Key == key {
			fm.Extra[i] = Extra{Key: key, Raw: raw, Value: node}
			return
		}
	}
	fm.Extra = append(fm.Extra, Extra{Key: key, Raw: raw, Value: node})
}

// ContentTime is `generated.at`, falling back to the legacy `timestamp` key.
func (fm *Frontmatter) ContentTime() string {
	if fm.Generated != nil && fm.Generated.At != "" {
		return fm.Generated.At
	}
	if v := fm.ExtraValue("timestamp"); v != nil && v.Kind == yaml.ScalarNode {
		return v.Value
	}
	return ""
}

// renderFrontmatter renders the frontmatter block, including both `---` lines.
func renderFrontmatter(fm *Frontmatter) string {
	var b strings.Builder
	b.WriteString("---\n")
	typ := fm.Type
	if strings.TrimSpace(typ) == "" {
		typ = "note"
	}
	if fm.isSet("type") || !fm.isMalformed("type") {
		b.WriteString("type: " + yamlx.Scalar(typ) + "\n")
	}
	for _, kv := range [][2]string{{"title", fm.Title}, {"description", fm.Description}, {"resource", fm.Resource}} {
		if strings.TrimSpace(kv[1]) != "" {
			b.WriteString(kv[0] + ": " + yamlx.Scalar(kv[1]) + "\n")
		}
	}
	if tags := cleanTags(fm.Tags); len(tags) > 0 {
		parts := make([]string, len(tags))
		for i, t := range tags {
			parts[i] = yamlx.FlowScalar(t)
		}
		b.WriteString("tags: [" + strings.Join(parts, ", ") + "]\n")
	}
	if len(fm.Sources) > 0 {
		b.WriteString("sources:\n")
		for _, s := range fm.Sources {
			b.WriteString(renderSource(s))
		}
	}
	if fm.Generated != nil {
		b.WriteString("generated: " + renderStamp(*fm.Generated) + "\n")
	}
	if len(fm.Verified) > 0 {
		b.WriteString("verified:\n")
		for _, v := range fm.Verified {
			b.WriteString("  - " + renderStamp(v) + "\n")
		}
	}
	extra := append([]Extra(nil), fm.Extra...)
	if fm.Archived {
		extra = append(extra, Extra{Key: "archived", Raw: "archived: true"})
	}
	sort.SliceStable(extra, func(i, j int) bool { return extra[i].Key < extra[j].Key })
	for _, e := range extra {
		// A recognized key appears in Extra only when malformed (or set
		// with --meta); it gives way once the field itself is set.
		if knownKeys[e.Key] && e.Value != nil && fm.isSet(e.Key) {
			continue
		}
		b.WriteString(e.Raw + "\n")
	}
	b.WriteString("---\n")
	return b.String()
}

// isSet reports whether a recognized field has a value that rendering writes.
func (fm *Frontmatter) isSet(key string) bool {
	switch key {
	case "type":
		return strings.TrimSpace(fm.Type) != ""
	case "title":
		return strings.TrimSpace(fm.Title) != ""
	case "description":
		return strings.TrimSpace(fm.Description) != ""
	case "resource":
		return strings.TrimSpace(fm.Resource) != ""
	case "tags":
		return len(cleanTags(fm.Tags)) > 0
	case "sources":
		return len(fm.Sources) > 0
	case "generated":
		return fm.Generated != nil
	case "verified":
		return len(fm.Verified) > 0
	case "archived":
		return fm.Archived
	}
	return false
}

func (fm *Frontmatter) isMalformed(key string) bool {
	for _, m := range fm.Malformed {
		if m.Key == key {
			return true
		}
	}
	return false
}

func renderStamp(s Stamp) string {
	var parts []string
	if s.By != "" {
		parts = append(parts, "by: "+yamlx.FlowScalar(s.By))
	}
	if s.At != "" {
		parts = append(parts, "at: "+yamlx.FlowScalar(s.At, yamlx.TagTimestamp))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func renderSource(s Source) string {
	var lines []string
	if s.ID != "" {
		lines = append(lines, "id: "+yamlx.Quote(s.ID))
	}
	if s.Resource != "" {
		lines = append(lines, "resource: "+yamlx.Scalar(s.Resource))
	}
	if s.Title != "" {
		lines = append(lines, "title: "+yamlx.Scalar(s.Title))
	}
	for i := 0; i+1 < len(s.Extra); i += 2 {
		lines = append(lines, strings.Split(encodePair(s.Extra[i], s.Extra[i+1]), "\n")...)
	}
	if len(lines) == 0 {
		return "  - {}\n"
	}
	var b strings.Builder
	for i, l := range lines {
		if i == 0 {
			b.WriteString("  - " + l + "\n")
		} else {
			b.WriteString("    " + l + "\n")
		}
	}
	return b.String()
}

func cleanTags(tags []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	return out
}

// Trust tiers (§5.5).
const (
	TrustUnverified       = "unverified"
	TrustMachineConfirmed = "machine-confirmed"
	TrustHumanReviewed    = "human-reviewed"
)

// Trust derives the trust tier from `verified`.
func (fm *Frontmatter) Trust() string {
	if len(fm.Verified) == 0 {
		return TrustUnverified
	}
	for _, v := range fm.Verified {
		if strings.HasPrefix(v.By, "human:") {
			return TrustHumanReviewed
		}
	}
	return TrustMachineConfirmed
}
