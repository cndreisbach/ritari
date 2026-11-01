package bundle

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/cndreisbach/ritari/internal/note"
	"github.com/cndreisbach/ritari/internal/vocab"
)

var dateRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// ValidDate checks a daily note date.
func ValidDate(d string) error {
	if dateRe.MatchString(d) {
		if t, err := time.Parse(time.DateOnly, d); err == nil && t.Format(time.DateOnly) == d {
			return nil
		}
	}
	return fmt.Errorf("daily notes are named by date: %q is not YYYY-MM-DD", d)
}

// DailyOpts are the inputs of `daily`.
type DailyOpts struct {
	Date        string
	Notes       []string
	Description string
	Obs         []string
	Rels        []string
	Sources     []string
	Refresh     bool
}

// DailyResult reports what `daily` did.
type DailyResult struct {
	Status  string // created, updated, refreshed, missing
	Date    string
	ID      string
	Path    string
	Touched int
	Content string
	Warning string // screen findings for the note as written (§8.1); not for --refresh
}

// Daily implements the `daily` command (§7, §9.14).
func Daily(root string, env Env, o DailyOpts) (DailyResult, error) {
	date := o.Date
	if date == "" {
		date = env.Today()
	}
	if err := ValidDate(date); err != nil {
		return DailyResult{}, err
	}
	var res DailyResult
	err := Update(root, env, func(b *Bundle) ([]Event, error) {
		id := b.Vocab.DailyDir() + "/" + date
		res = DailyResult{Date: date, ID: id, Path: b.Abs(id + ".md")}
		if o.Refresh {
			if _, ok := b.Get(id); !ok {
				res.Status, res.Path = "missing", ""
				return nil, nil
			}
			res.Status = "refreshed"
		} else {
			ev, err := b.writeDaily(env, id, date, o)
			if err != nil {
				return nil, err
			}
			res.Status = "updated"
			if ev.Action == ActCreate {
				res.Status = "created"
			}
			if err := Record(b.Root, env, []Event{ev}); err != nil {
				return nil, err
			}
		}
		n, err := RefreshTouched(b.Root, id, date)
		if err != nil {
			return nil, err
		}
		res.Touched = n
		data, err := readAll(b.Abs(id + ".md"))
		res.Content = data
		if !o.Refresh {
			res.Warning = withScreen("", []byte(data))
		}
		return nil, err
	})
	return res, err
}

func (b *Bundle) writeDaily(env Env, id, date string, o DailyOpts) (Event, error) {
	add, err := b.parseAdditions(WriteOpts{Obs: o.Obs, Rels: o.Rels, Sources: o.Sources})
	if err != nil {
		return Event{}, err
	}
	var paras []string
	for _, p := range o.Notes {
		if p = strings.TrimSpace(p); p != "" {
			paras = append(paras, p)
		}
	}
	c, exists := b.Get(id)
	var n *note.Note
	action := ActUpdate
	if !exists {
		n = &note.Note{FM: note.Frontmatter{Type: vocab.Daily, Title: date, Description: o.Description}}
		if _, err := n.Compose(strings.Join(paras, "\n\n"), add); err != nil {
			return Event{}, err
		}
		n.Restamp(env.Actor, env.Stamp())
		action = ActCreate
	} else {
		n = c.Note
		if _, err := n.Merge(add); err != nil {
			return Event{}, err
		}
		n.Body = note.AppendNarrative(n.EditableBody(), paras)
		if o.Description != "" {
			n.FM.Description = o.Description
		}
		n.FillGenerated(env.Actor, env.Stamp())
	}
	if err := b.saveNote(id+".md", n); err != nil {
		return Event{}, err
	}
	return Event{Action: action, ID: id, Title: date}, nil
}

type touchedEntry struct {
	id, title, action string
	deleted           bool
}

// touchedEntries projects a day's log events onto one entry per concept (§7).
func touchedEntries(events []Event, dailyID string) []*touchedEntry {
	var order []*touchedEntry
	byID := map[string]*touchedEntry{}
	for _, e := range events {
		if e.ID == "" || e.ID == dailyID {
			continue
		}
		if e.Action == ActMove {
			if old := movedFrom(e.Detail); old != "" && byID[old] != nil {
				t := byID[old]
				delete(byID, old)
				t.id, t.title = e.ID, e.Title
				byID[e.ID] = t
				continue
			}
		}
		t := byID[e.ID]
		if t == nil {
			t = &touchedEntry{id: e.ID, title: e.Title}
			byID[e.ID] = t
			order = append(order, t)
		}
		if e.Action == ActDelete {
			t.deleted = true
			continue
		}
		if e.Title != "" {
			t.title = e.Title
		}
		if t.deleted || t.action == "" {
			t.deleted = false
			t.action = e.Action
		}
	}
	return order
}

func pastTense(action string) string {
	switch action {
	case ActCreate:
		return "created"
	case ActUpdate:
		return "updated"
	case ActMove:
		return "moved"
	}
	return strings.ToLower(action)
}

// RefreshTouched rebuilds the # Touched section of a daily note from the log,
// without recording or restamping. It returns the number of entries.
func RefreshTouched(root, id, date string) (int, error) {
	b, err := Open(root)
	if err != nil {
		return 0, err
	}
	c, ok := b.Get(id)
	if !ok {
		return 0, nil
	}
	text, err := readLog(b.Root)
	if err != nil {
		return 0, err
	}
	var lines []string
	for _, t := range touchedEntries(LogEvents(text, date), id) {
		if t.deleted {
			lines = append(lines, "- "+t.id+" — deleted")
			continue
		}
		title := t.title
		if title == "" {
			title = t.id
		}
		lines = append(lines, "- ["+escapeLinkText(title)+"](/"+t.id+".md) — "+pastTense(t.action))
	}
	n := c.Note
	n.Body = note.SetSection(n.EditableBody(), "Touched", lines)
	return len(lines), b.saveNote(c.Path, n)
}
