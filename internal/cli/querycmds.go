package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/cndreisbach/ritari/internal/bundle"
	"github.com/cndreisbach/ritari/internal/lint"
	"github.com/cndreisbach/ritari/internal/note"
	"github.com/cndreisbach/ritari/internal/output"
	"github.com/cndreisbach/ritari/internal/search"
	"github.com/cndreisbach/ritari/internal/vocab"
)

// SearchCmd is `search`.
type SearchCmd struct {
	Query           []string `arg:"" optional:"" help:"Search terms (all must match)."`
	Type            string   `help:"Only notes of this type."`
	Category        string   `help:"Only notes with an observation in this category."`
	Tags            string   `help:"Only notes with all of these tags (comma-separated)."`
	Relation        string   `help:"Only notes with a relation of this type."`
	Meta            []string `sep:"none" placeholder:"KEY=VALUE" help:"Frontmatter filter; a|b for alternatives (repeatable)."`
	IncludeDaily    bool     `help:"Include daily notes."`
	IncludeArchived bool     `help:"Include archived notes."`
	Limit           int      `default:"20" help:"Maximum results; 0 or less for all."`
}

// Run executes `search`.
func (cmd *SearchCmd) Run(c *Ctx) error {
	b, err := c.Open()
	if err != nil {
		return err
	}
	hits, err := search.Run(b, search.Options{
		Query: strings.Join(cmd.Query, " "), Type: cmd.Type, Category: cmd.Category, Tags: splitTags(cmd.Tags),
		Relation: cmd.Relation, Meta: cmd.Meta, IncludeDaily: cmd.IncludeDaily, IncludeArchived: cmd.IncludeArchived,
		Limit: cmd.Limit,
	})
	if err != nil {
		return err
	}
	docs := []*output.Doc{output.New().Set("success", true).Set("count", len(hits))}
	for _, h := range hits {
		fm := &h.Concept.Note.FM
		docs = append(docs, output.New().Set("id", h.Concept.ID).Set("title", h.Concept.Title()).Set("type", fm.Type).
			Set("description", fm.Description).Set("score", h.Score).Set("matches", output.List(h.Matches)))
		c.Printf("%s  [%s]  %s\n", h.Concept.Title(), fm.Type, h.Concept.ID)
		if fm.Description != "" {
			c.Printf("    %s\n", fm.Description)
		}
		for i, m := range h.Matches {
			if i == 3 {
				c.Printf("    ... and %d more\n", len(h.Matches)-3)
				break
			}
			c.Printf("    - %s\n", m)
		}
	}
	if len(hits) == 0 {
		c.Printf("No results.\n")
	}
	return c.Emit(docs...)
}

// ContextCmd is `context`.
type ContextCmd struct {
	Ref   string `arg:"" help:"Starting note."`
	Depth int    `default:"1" help:"How many links to follow."`
}

// Run executes `context`.
func (cmd *ContextCmd) Run(c *Ctx) error {
	b, err := c.Open()
	if err != nil {
		return err
	}
	start, err := b.Resolve(cmd.Ref)
	if err != nil {
		return err
	}
	notes := b.Context(start, cmd.Depth)
	docs := []*output.Doc{output.New().Set("success", true).Set("start", start.ID).Set("count", len(notes))}
	for _, n := range notes {
		fm := &n.Note.FM
		docs = append(docs, output.New().Set("id", n.ID).Set("title", n.Title()).Set("type", fm.Type).
			Set("description", fm.Description).Set("links", output.List(note.Targets(n.Note.Body, n.ID))))
		line := fmt.Sprintf("- %s (%s)", n.Title(), n.ID)
		if fm.Description != "" {
			line += " - " + fm.Description
		}
		c.Printf("%s\n", line)
	}
	return c.Emit(docs...)
}

// RecentCmd is `recent`.
type RecentCmd struct {
	Limit int `default:"20" help:"Maximum notes; 0 or less for all."`
}

// Run executes `recent`.
func (cmd *RecentCmd) Run(c *Ctx) error {
	b, err := c.Open()
	if err != nil {
		return err
	}
	notes := b.Recent(cmd.Limit)
	docs := []*output.Doc{output.New().Set("success", true).Set("count", len(notes))}
	for _, n := range notes {
		fm := &n.Note.FM
		docs = append(docs, output.New().Set("id", n.ID).Set("title", n.Title()).Set("type", fm.Type).
			Set("description", fm.Description).Set("modified", n.ModTime.UTC().Format(time.RFC3339)))
		c.Printf("%s  %s  %s\n", n.ModTime.Local().Format("2006-01-02 15:04"), n.Title(), n.ID)
	}
	return c.Emit(docs...)
}

// LsCmd is `ls`.
type LsCmd struct {
	Dir string `arg:"" optional:"" help:"Only IDs under this directory."`
}

// Run executes `ls`.
func (cmd *LsCmd) Run(c *Ctx) error {
	b, err := c.Open()
	if err != nil {
		return err
	}
	prefix := strings.Trim(filepath.ToSlash(cmd.Dir), "/")
	ids := output.ExplicitList{}
	for _, id := range b.Files {
		if prefix == "" || strings.HasPrefix(id, prefix+"/") {
			ids = append(ids, id)
			c.Printf("%s\n", id)
		}
	}
	return c.Emit(output.New().Set("success", true).Set("count", len(ids)).Set("ids", ids))
}

// DailyCmd is `daily`.
type DailyCmd struct {
	Date        string   `arg:"" optional:"" help:"YYYY-MM-DD (default: today)."`
	Note        []string `sep:"none" placeholder:"PARAGRAPH" help:"Narrative paragraph, or - for stdin (repeatable)."`
	Description string   `help:"Set the day's description."`
	Obs         []string `sep:"none" placeholder:"CATEGORY: TEXT" help:"Observation to add (repeatable)."`
	Rel         []string `sep:"none" placeholder:"TYPE: TARGET" help:"Relation to add (repeatable)."`
	Source      []string `sep:"none" help:"Source for this write's observations (repeatable)."`
	Refresh     bool     `help:"Only re-derive the # Touched section; never creates the note."`
}

// Run executes `daily`.
func (cmd *DailyCmd) Run(c *Ctx) error {
	root, err := c.Root()
	if err != nil {
		return err
	}
	env, err := c.Env()
	if err != nil {
		return err
	}
	var notes []string
	stdin := false
	for _, n := range cmd.Note {
		if n == "-" {
			if stdin {
				return fmt.Errorf("--note - can be given only once; stdin is read to the end")
			}
			stdin = true
			if n, err = c.readStdinIf(n); err != nil {
				return err
			}
		}
		notes = append(notes, n)
	}
	res, err := bundle.Daily(root, env, bundle.DailyOpts{
		Date: cmd.Date, Notes: notes, Description: cmd.Description,
		Obs: cmd.Obs, Rels: cmd.Rel, Sources: cmd.Source, Refresh: cmd.Refresh,
	})
	if err != nil {
		return err
	}
	d := output.New().Set("success", true).Set("status", res.Status).Set("date", res.Date).Set("id", res.ID)
	if res.Status == "missing" {
		c.Printf("No daily note for %s.\n", res.Date)
		return c.Emit(d.Set("touched", 0))
	}
	d.Set("path", res.Path).Set("touched", res.Touched).Set("content", res.Content)
	c.Printf("%s %s (%d touched)\n\n%s", strings.ToUpper(res.Status[:1])+res.Status[1:], res.Path, res.Touched, res.Content)
	return c.Emit(d)
}

// IndexCmd is `index`.
type IndexCmd struct {
	Check bool `help:"Report stale index files without writing."`
}

// Run executes `index`.
func (cmd *IndexCmd) Run(c *Ctx) error {
	root, err := c.Root()
	if err != nil {
		return err
	}
	var changed []string
	if cmd.Check {
		b, err := bundle.Open(root)
		if err != nil {
			return err
		}
		if changed, err = b.SyncIndexes(true); err != nil {
			return err
		}
	} else {
		env, err := c.Env()
		if err != nil {
			return err
		}
		if changed, err = bundle.Reindex(root, env); err != nil {
			return err
		}
	}
	key, verb := "wrote", "Wrote"
	if cmd.Check {
		key, verb = "stale", "Stale"
	}
	if len(changed) == 0 {
		c.Printf("Indexes are current.\n")
	}
	for _, p := range changed {
		c.Printf("%s %s\n", verb, p)
	}
	if cmd.Check && len(changed) > 0 {
		msg := fmt.Sprintf("%d index file(s) are stale (run: rit index)", len(changed))
		if err := c.Emit(output.New().Set("success", false).Set("error", msg).Set("count", len(changed)).
			Set(key, output.ExplicitList(changed))); err != nil {
			return err
		}
		return reported{msg}
	}
	return c.Emit(output.New().Set("success", true).Set("count", len(changed)).Set(key, output.ExplicitList(changed)))
}

// LintCmd is `lint`.
type LintCmd struct {
	Conformance bool     `help:"Report errors only."`
	Rule        []string `sep:"none" help:"Only this rule (repeatable)."`
	Strict      bool     `help:"Fail on warnings too."`
	Categories  string   `help:"Allowed observation categories (comma-separated)."`
}

// Run executes `lint`.
func (cmd *LintCmd) Run(c *Ctx) error {
	b, err := c.Open()
	if err != nil {
		return err
	}
	rep, err := lint.Run(b, lint.Options{Conformance: cmd.Conformance, Rules: cmd.Rule, Strict: cmd.Strict, Categories: splitTags(cmd.Categories)})
	if err != nil {
		return err
	}
	docs := []*output.Doc{output.New().Set("success", rep.Failure == "").Set("errors", rep.Errors).
		Set("warnings", rep.Warnings).Set("error", rep.Failure)}
	last := ""
	for _, f := range rep.Findings {
		docs = append(docs, output.New().Set("severity", f.Severity).Set("rule", f.Rule).Set("id", f.ID).
			Set("path", f.Path).Set("message", f.Message))
		if f.Path != last {
			c.Printf("%s\n", f.Path)
			last = f.Path
		}
		c.Printf("  [%s] %s  %s\n", f.Severity, f.Rule, f.Message)
	}
	if len(rep.Findings) == 0 {
		c.Printf("No problems found.\n")
	} else {
		c.Printf("%d error(s), %d warning(s)\n", rep.Errors, rep.Warnings)
	}
	if err := c.Emit(docs...); err != nil {
		return err
	}
	if rep.Failure != "" {
		return reported{rep.Failure}
	}
	return nil
}

// TypesCmd is `types`.
type TypesCmd struct {
	YAML bool `name:"yaml" help:"Print the vocabulary as ritari.yaml content."`
}

// Run executes `types`.
func (cmd *TypesCmd) Run(c *Ctx) error {
	b, err := c.Open()
	if err != nil {
		return err
	}
	if cmd.YAML {
		_, err := fmt.Fprint(c.Stdout, b.Vocab.YAML())
		return err
	}
	docs := []*output.Doc{output.New().Set("success", true).Set("count", len(b.Vocab.Types)).
		Set("config", filepath.Join(b.Root, vocab.FileName))}
	for _, t := range b.Vocab.Types {
		docs = append(docs, output.New().Set("name", t.Name).Set("dir", t.Dir).Set("summary", t.Summary))
		c.Printf("%s  %s/  %s\n", t.Name, t.Dir, t.Summary)
	}
	return c.Emit(docs...)
}
