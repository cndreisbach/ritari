package bundle

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cndreisbach/ritari/internal/note"
)

// saveNote renders n and writes it to the concept's file if it changed.
func (b *Bundle) saveNote(rel string, n *note.Note) error {
	_, err := writeIfChanged(b.Abs(rel), note.Render(n))
	return err
}

// EditOpts are the inputs of `note edit`. Nil pointers mean "not given".
type EditOpts struct {
	Op          string
	Content     string
	Section     string
	Find        string
	Count       int
	Title       *string
	Description *string
	Resource    *string
	Tags        *[]string
	Meta        [][2]string
}

// Result is the common outcome of a single-note operation.
type Result struct {
	Status string
	ID     string
	Path   string // absolute
}

// Edit implements `note edit` (§9.7).
func Edit(root string, env Env, ref string, o EditOpts) (Result, error) {
	if o.Op == "" && o.Title == nil && o.Description == nil && o.Resource == nil && o.Tags == nil && len(o.Meta) == 0 {
		return Result{}, fmt.Errorf("nothing to do: pass --op for a body edit, or --title/--description/--tags/--resource/--meta to set frontmatter")
	}
	var res Result
	err := Update(root, env, func(b *Bundle) ([]Event, error) {
		c, err := b.Resolve(ref)
		if err != nil {
			return nil, err
		}
		n := c.Note
		if o.Op != "" {
			before := n.EditableBody()
			body, err := note.EditBody(before, o.Op, o.Content, o.Section, o.Find, o.Count)
			if err != nil {
				return nil, err
			}
			n.Body = body
			if body != before {
				n.Restamp(env.Actor, env.Stamp())
			}
		}
		if o.Title != nil {
			n.FM.Title = *o.Title
		}
		if o.Description != nil {
			n.FM.Description = *o.Description
		}
		if o.Resource != nil {
			n.FM.Resource = *o.Resource
		}
		if o.Tags != nil {
			n.FM.Tags = *o.Tags
		}
		for _, kv := range o.Meta {
			n.FM.SetExtra(kv[0], kv[1])
		}
		n.FillGenerated(env.Actor, env.Stamp())
		if err := b.saveNote(c.Path, n); err != nil {
			return nil, err
		}
		res = Result{Status: "updated", ID: c.ID, Path: b.Abs(c.Path)}
		return []Event{{Action: ActUpdate, ID: c.ID, Title: titleOf(n, c.ID)}}, nil
	})
	return res, err
}

// VerifyResult reports a verification.
type VerifyResult struct {
	Result
	By    string
	Trust string
}

// Verify implements `note verify` (§9.8).
func Verify(root string, env Env, ref, by string) (VerifyResult, error) {
	if by == "" {
		by = env.Actor
	}
	var res VerifyResult
	err := Update(root, env, func(b *Bundle) ([]Event, error) {
		c, err := b.Resolve(ref)
		if err != nil {
			return nil, err
		}
		n := c.Note
		n.FM.Verified = append(n.FM.Verified, note.Stamp{By: by, At: env.Stamp()})
		n.FillGenerated(env.Actor, env.Stamp())
		if err := b.saveNote(c.Path, n); err != nil {
			return nil, err
		}
		res = VerifyResult{Result: Result{Status: "verified", ID: c.ID, Path: b.Abs(c.Path)}, By: by, Trust: n.FM.Trust()}
		return []Event{{Action: ActUpdate, ID: c.ID, Title: titleOf(n, c.ID)}}, nil
	})
	return res, err
}

// SetArchived implements `note archive` and `note unarchive` (§9.10).
func SetArchived(root string, env Env, ref string, archived bool) (Result, error) {
	var res Result
	err := Update(root, env, func(b *Bundle) ([]Event, error) {
		c, err := b.Resolve(ref)
		if err != nil {
			return nil, err
		}
		n := c.Note
		switch {
		case archived && n.FM.Archived:
			return nil, fmt.Errorf("%s is already archived", c.ID)
		case !archived && !n.FM.Archived:
			return nil, fmt.Errorf("%s is not archived", c.ID)
		}
		n.FM.Archived = archived
		n.FillGenerated(env.Actor, env.Stamp())
		if err := b.saveNote(c.Path, n); err != nil {
			return nil, err
		}
		status, action := "archived", ActArchive
		if !archived {
			status, action = "restored", ActRestore
		}
		res = Result{Status: status, ID: c.ID, Path: b.Abs(c.Path)}
		return []Event{{Action: action, ID: c.ID, Title: titleOf(n, c.ID)}}, nil
	})
	return res, err
}

// DeleteResult reports a deletion.
type DeleteResult struct {
	Result
	BrokeLinksIn []string
}

// Delete implements `note delete` (§9.10).
func Delete(root string, env Env, ref string) (DeleteResult, error) {
	var res DeleteResult
	err := Update(root, env, func(b *Bundle) ([]Event, error) {
		c, err := b.Resolve(ref)
		if err != nil {
			return nil, err
		}
		inbound := others(b.Graph().In[c.ID], c.ID)
		if err := os.Remove(b.Abs(c.Path)); err != nil {
			return nil, err
		}
		res = DeleteResult{Result: Result{Status: "deleted", ID: c.ID, Path: b.Abs(c.Path)}, BrokeLinksIn: inbound}
		return []Event{{Action: ActDelete, ID: c.ID}}, nil
	})
	return res, err
}

// others returns ids sorted, without self.
func others(ids []string, self string) []string {
	out := []string{}
	for _, id := range ids {
		if id != self {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

// MoveResult reports a move.
type MoveResult struct {
	From           string
	To             string
	RewroteLinksIn []string
}

// Move implements `note move` (§9.9).
func Move(root string, env Env, ref, dest string) (MoveResult, error) {
	var res MoveResult
	err := Update(root, env, func(b *Bundle) ([]Event, error) {
		c, err := b.Resolve(ref)
		if err != nil {
			return nil, err
		}
		newRel := strings.TrimPrefix(path.Clean("/"+filepath.ToSlash(dest)), "/")
		if !strings.HasSuffix(strings.ToLower(newRel), ".md") {
			newRel += ".md"
		}
		newID := idOf(newRel)
		if newID == c.ID {
			return nil, fmt.Errorf("note %s is already there", c.ID)
		}
		if _, err := os.Stat(b.Abs(newRel)); err == nil {
			return nil, fmt.Errorf("destination already exists: %s", newID)
		} else if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		inbound := b.Graph().In[c.ID]

		if err := os.MkdirAll(filepath.Dir(b.Abs(newRel)), 0o755); err != nil {
			return nil, err
		}
		if err := os.Rename(b.Abs(c.Path), b.Abs(newRel)); err != nil {
			return nil, err
		}
		// The moved note's own relative links would break: make them absolute,
		// and point any self-links at the new location.
		if err := rewriteLinks(b.Abs(newRel), func(href string) (string, bool) {
			target, ok := note.ResolveLink(href, c.ID)
			switch {
			case ok && target == c.ID:
				return "/" + newID + ".md" + fragment(href), true
			case ok && !strings.HasPrefix(href, "/"):
				return "/" + target + ".md" + fragment(href), true
			}
			return "", false
		}); err != nil {
			return nil, err
		}
		rewrote := []string{}
		for _, id := range inbound {
			if id == c.ID {
				continue
			}
			src, ok := b.Get(id)
			if !ok {
				continue
			}
			if err := rewriteLinks(b.Abs(src.Path), func(href string) (string, bool) {
				if target, ok := note.ResolveLink(href, id); ok && target == c.ID {
					return "/" + newID + ".md" + fragment(href), true
				}
				return "", false
			}); err != nil {
				return nil, err
			}
			rewrote = append(rewrote, id)
		}
		sort.Strings(rewrote)
		res = MoveResult{From: c.ID, To: newID, RewroteLinksIn: rewrote}
		return []Event{{Action: ActMove, ID: newID, Title: c.Title(), Detail: "was " + c.ID}}, nil
	})
	return res, err
}

func fragment(href string) string {
	if i := strings.IndexByte(href, '#'); i >= 0 {
		return href[i:]
	}
	return ""
}

// rewriteLinks replaces link destinations in a concept file. fn maps a
// destination as written to its replacement. Replacement is keyed on the exact
// destination text, so unrelated links are untouched.
func rewriteLinks(file string, fn func(href string) (string, bool)) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	n, err := note.Parse(data)
	if err != nil {
		return err
	}
	text := string(data)
	for _, l := range note.Links(n.Body) {
		repl, ok := fn(l.Href)
		if !ok || repl == l.Href {
			continue
		}
		text = strings.ReplaceAll(text, "]("+l.Href+")", "]("+repl+")")
		text = strings.ReplaceAll(text, "]: "+l.Href+"\n", "]: "+repl+"\n")
	}
	_, err = writeIfChanged(file, []byte(text))
	return err
}
