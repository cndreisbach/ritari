package bundle

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/cndreisbach/ritari/internal/note"
	"github.com/cndreisbach/ritari/internal/vocab"
)

var relLinkInputRe = regexp.MustCompile(`^\[(.*)\]\((.*)\)$`)

// ComposeRelation turns a --rel value into a canonical relation line (§5.2).
func (b *Bundle) ComposeRelation(input string) (string, error) {
	s := strings.TrimSpace(input)
	if r, ok := strings.CutPrefix(s, "- "); ok {
		s = r
	} else if r, ok := strings.CutPrefix(s, "* "); ok {
		s = r
	}
	s = strings.TrimSpace(s)
	typ, target := "links_to", s
	if before, after, ok := strings.Cut(s, ":"); ok && before != "" && !strings.ContainsAny(before, " \t[") {
		typ, target = before, strings.TrimSpace(after)
	}
	if target == "" {
		return "", fmt.Errorf("relation %q has no target", input)
	}
	text, href, err := b.relationTarget(target)
	if err != nil {
		return "", err
	}
	return "- " + typ + " [" + escapeLinkText(text) + "](" + href + ")", nil
}

func (b *Bundle) relationTarget(target string) (text, href string, err error) {
	if m := relLinkInputRe.FindStringSubmatch(target); m != nil {
		text, dest := unescapeLinkText(m[1]), m[2]
		if note.Classify(dest) != note.InBundle {
			return text, dest, nil
		}
		return text, "/" + normalizeID(dest) + ".md", nil
	}
	if strings.Contains(target, "/") || strings.HasSuffix(strings.ToLower(target), ".md") {
		id := normalizeID(target)
		text := path.Base(id)
		if c, ok := b.Get(id); ok {
			text = c.Title()
		}
		return text, "/" + id + ".md", nil
	}
	if c, err := b.Resolve(target); err == nil {
		return c.Title(), "/" + c.ID + ".md", nil
	}
	slug := vocab.Slug(target)
	if slug == "" {
		return "", "", fmt.Errorf("relation target %q has no usable title", target)
	}
	return target, "/" + b.Vocab.NoteDir() + "/" + slug + ".md", nil
}

// normalizeID strips a leading "/" and a trailing ".md" and cleans the path.
func normalizeID(p string) string {
	p = strings.TrimPrefix(path.Clean("/"+strings.TrimPrefix(p, "/")), "/")
	if strings.HasSuffix(strings.ToLower(p), ".md") {
		p = p[:len(p)-3]
	}
	return p
}

// WriteOpts are the inputs of `note write`.
type WriteOpts struct {
	Title       string
	Type        string
	Description string
	Resource    string
	Tags        []string
	Meta        [][2]string
	Content     string
	Obs         []string
	Rels        []string
	Sources     []string
	Force       bool
	DryRun      bool
	AllowEmpty  bool
	ExpectHash  string
}

// WriteResult reports what `note write` did.
type WriteResult struct {
	Status       string // created, updated, replaced, dry-run
	ID           string
	Path         string // absolute
	AddedObs     int
	AddedRels    int
	AddedSources int
	Warning      string
	Hash         string // of the file as written; omitted on dry-run
	Rendered     string
}

// Write implements `note write` (§9.5).
func Write(root string, env Env, o WriteOpts) (WriteResult, error) {
	var res WriteResult
	err := Update(root, env, func(b *Bundle) ([]Event, error) {
		var ev []Event
		var err error
		res, ev, err = b.write(env, o)
		if o.DryRun {
			return nil, err
		}
		return ev, err
	})
	return res, err
}

func (b *Bundle) parseAdditions(o WriteOpts) (note.Additions, error) {
	var add note.Additions
	add.Obs = o.Obs
	for _, r := range o.Rels {
		line, err := b.ComposeRelation(r)
		if err != nil {
			return add, err
		}
		add.Rels = append(add.Rels, line)
	}
	for _, s := range o.Sources {
		if src, ok := note.ParseSource(s); ok {
			add.Sources = append(add.Sources, src)
		}
	}
	return add, nil
}

func (b *Bundle) findExisting(id, title string) (*Concept, error) {
	c, err := b.Resolve(id)
	if errors.Is(err, ErrNotFound) {
		return b.ByTitle(title), nil
	}
	return c, err
}

func (b *Bundle) write(env Env, o WriteOpts) (WriteResult, []Event, error) {
	typ := strings.TrimSpace(o.Type)
	if typ == "" {
		typ = vocab.Note
	}
	if vocab.Slug(o.Title) == "" {
		return WriteResult{}, nil, fmt.Errorf("note requires a title that slugifies to something")
	}
	id := b.Vocab.IDFor(typ, o.Title)
	add, err := b.parseAdditions(o)
	if err != nil {
		return WriteResult{}, nil, err
	}
	existing, err := b.findExisting(id, o.Title)
	if err != nil {
		return WriteResult{}, nil, err
	}
	if o.ExpectHash != "" {
		if existing == nil {
			return WriteResult{}, nil, fmt.Errorf("--expect-hash given, but note %q does not exist", id)
		}
		if err := b.checkHash(existing, o.ExpectHash); err != nil {
			return WriteResult{}, nil, err
		}
	}
	if existing != nil && !o.Force {
		return b.writeAdditive(env, existing, o, add)
	}
	return b.writeFresh(env, existing, id, typ, o, add)
}

func (b *Bundle) writeAdditive(env Env, c *Concept, o WriteOpts, add note.Additions) (WriteResult, []Event, error) {
	if strings.TrimSpace(o.Content) != "" {
		return WriteResult{}, nil, fmt.Errorf("note %q already exists; replace its body with rit note edit %s --op replace_body --content - (--force replaces the whole note, dropping its metadata, observations, and relations; by default only --obs/--rel/--source are added)", c.ID, c.ID)
	}
	if len(o.Obs)+len(o.Rels)+len(add.Sources)+len(o.Meta)+len(o.Tags) == 0 && o.Description == "" && o.Resource == "" {
		return WriteResult{}, nil, fmt.Errorf("note %q already exists and nothing to add (provide --obs/--rel/--source/--description/--tags/--meta, or --force to replace)", c.ID)
	}
	n := c.Note
	r, err := n.Merge(add)
	if err != nil {
		return WriteResult{}, nil, err
	}
	if r.BodyChanged {
		n.Restamp(env.Actor, env.Stamp())
	}
	if o.Description != "" {
		n.FM.Description = o.Description
	}
	if o.Resource != "" {
		n.FM.Resource = o.Resource
	}
	n.FM.Tags = append(n.FM.Tags, o.Tags...)
	for _, kv := range o.Meta {
		n.FM.SetExtra(kv[0], kv[1])
	}
	res := WriteResult{
		Status: "updated", ID: c.ID, Path: b.Abs(c.Path),
		AddedObs: r.AddedObs, AddedRels: r.AddedRels, AddedSources: r.AddedSources,
	}
	return b.finish(env, res, c.Path, n, o.DryRun, Event{Action: ActUpdate, ID: c.ID})
}

func (b *Bundle) writeFresh(env Env, existing *Concept, id, typ string, o WriteOpts, add note.Additions) (WriteResult, []Event, error) {
	n := &note.Note{FM: note.Frontmatter{
		Type: typ, Title: o.Title, Description: o.Description, Resource: o.Resource, Tags: o.Tags,
	}}
	for _, kv := range o.Meta {
		n.FM.SetExtra(kv[0], kv[1])
	}
	r, err := n.Compose(o.Content, add)
	if err != nil {
		return WriteResult{}, nil, err
	}
	if strings.TrimSpace(n.Body) == "" && !o.AllowEmpty {
		return WriteResult{}, nil, fmt.Errorf("refusing to write an empty note (no --content/--obs/--rel); use --allow-empty to override")
	}
	n.Restamp(env.Actor, env.Stamp())
	rel, status, action := id+".md", "created", ActCreate
	if existing != nil {
		rel, status, action, id = existing.Path, "replaced", ActUpdate, existing.ID
	}
	res := WriteResult{
		Status: status, ID: id, Path: b.Abs(rel),
		AddedObs: r.AddedObs, AddedRels: r.AddedRels, AddedSources: r.AddedSources,
	}
	if _, ok := b.Vocab.Lookup(typ); !ok {
		res.Warning = fmt.Sprintf("type %q is not one this knowledge base declares (%s); it will live in %s/",
			typ, strings.Join(b.Vocab.Names(), ", "), b.Vocab.DirFor(typ))
	}
	return b.finish(env, res, rel, n, o.DryRun, Event{Action: action, ID: id})
}

// finish renders n, writes it (unless dry-run), and returns the event to record.
func (b *Bundle) finish(env Env, res WriteResult, rel string, n *note.Note, dryRun bool, ev Event) (WriteResult, []Event, error) {
	n.FillGenerated(env.Actor, env.Stamp())
	data := note.Render(n)
	res.Warning = withScreen(res.Warning, data)
	if dryRun {
		res.Status, res.Path, res.Warning, res.Rendered = "dry-run", "", "", string(data)
		return res, nil, nil
	}
	if _, err := writeIfChanged(b.Abs(rel), data); err != nil {
		return WriteResult{}, nil, err
	}
	res.Hash = Hash(data)
	ev.Title = titleOf(n, ev.ID)
	return res, []Event{ev}, nil
}

func titleOf(n *note.Note, id string) string {
	if t := strings.TrimSpace(n.FM.Title); t != "" {
		return t
	}
	return path.Base(id)
}
