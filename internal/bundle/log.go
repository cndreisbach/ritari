package bundle

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// LogName is the generated change log at the bundle root.
const LogName = "log.md"

const logHeading = "# Update Log"

// Log actions.
const (
	ActCreate  = "Create"
	ActUpdate  = "Update"
	ActDelete  = "Delete"
	ActMove    = "Move"
	ActArchive = "Archive"
	ActRestore = "Restore"
)

// Event is one change-log entry.
type Event struct {
	Action string
	ID     string
	Title  string
	Detail string
}

// Bullet renders the event as a log line.
func (e Event) Bullet() string {
	subject := e.ID
	if e.Action != ActDelete {
		title := e.Title
		if title == "" {
			title = e.ID
		}
		subject = "[" + escapeLinkText(title) + "](/" + e.ID + ".md)"
	}
	s := "* **" + e.Action + "**: " + subject
	if e.Detail != "" {
		s += " (" + e.Detail + ")"
	}
	return s
}

func escapeLinkText(s string) string {
	return strings.NewReplacer(`[`, `\[`, `]`, `\]`).Replace(s)
}

func unescapeLinkText(s string) string {
	return strings.NewReplacer(`\[`, `[`, `\]`, `]`).Replace(s)
}

var (
	logDateRe   = regexp.MustCompile(`^## (\d{4}-\d{2}-\d{2})\s*$`)
	logBulletRe = regexp.MustCompile(`^\* \*\*([^*]+)\*\*: (.*)$`)
	logLinkRe   = regexp.MustCompile(`^\[(.*)\]\(/(.*)\.md\)(?: \((.*)\))?$`)
	logBareRe   = regexp.MustCompile(`^(\S+)(?: \((.*)\))?$`)
	movedFromRe = regexp.MustCompile(`^was (.+)$`)
)

// appendLog adds events to date's group of log text, creating the group
// directly after the document heading (so the file stays newest-first) and
// skipping events whose bullet is already in that group.
func appendLog(text, date string, events []Event) string {
	ls := strings.Split(strings.TrimRight(text, "\n"), "\n")
	if strings.TrimSpace(text) == "" {
		ls = []string{logHeading}
	}
	start, end := groupBounds(ls, date)
	if start < 0 {
		first := len(ls)
		for i, l := range ls {
			if logDateRe.MatchString(l) {
				first = i
				break
			}
		}
		head := trimBlankTail(ls[:first])
		if !slices.ContainsFunc(head, func(l string) bool { return strings.HasPrefix(l, "# ") }) {
			head = append([]string{logHeading}, head...)
		}
		group := []string{"", "## " + date}
		if rest := ls[first:]; len(rest) > 0 {
			group = append(group, "")
			group = append(group, rest...)
		}
		ls = slices.Concat(head, group)
		start, end = groupBounds(ls, date)
	}
	existing := map[string]bool{}
	for _, l := range ls[start:end] {
		existing[strings.TrimSpace(l)] = true
	}
	var add []string
	for _, e := range events {
		if b := e.Bullet(); !existing[b] {
			existing[b] = true
			add = append(add, b)
		}
	}
	if len(add) > 0 {
		at := end
		for at > start+1 && strings.TrimSpace(ls[at-1]) == "" {
			at--
		}
		if at == start+1 {
			add = append([]string{""}, add...)
		}
		ls = slices.Concat(ls[:at], add, ls[at:])
	}
	return strings.Join(trimBlankTail(ls), "\n") + "\n"
}

func trimBlankTail(ls []string) []string {
	for len(ls) > 0 && strings.TrimSpace(ls[len(ls)-1]) == "" {
		ls = ls[:len(ls)-1]
	}
	return ls
}

// groupBounds finds the `## date` heading line and the end of its group.
func groupBounds(ls []string, date string) (int, int) {
	for i, l := range ls {
		m := logDateRe.FindStringSubmatch(l)
		if m == nil || m[1] != date {
			continue
		}
		end := len(ls)
		for j := i + 1; j < len(ls); j++ {
			if strings.HasPrefix(ls[j], "#") {
				end = j
				break
			}
		}
		return i, end
	}
	return -1, -1
}

// LogEvents parses the entries recorded under date.
func LogEvents(text, date string) []Event {
	ls := strings.Split(text, "\n")
	start, end := groupBounds(ls, date)
	if start < 0 {
		return nil
	}
	var out []Event
	for _, l := range ls[start+1 : end] {
		m := logBulletRe.FindStringSubmatch(strings.TrimSpace(l))
		if m == nil {
			continue
		}
		e := Event{Action: m[1]}
		if lm := logLinkRe.FindStringSubmatch(m[2]); lm != nil {
			e.Title, e.ID, e.Detail = unescapeLinkText(lm[1]), lm[2], lm[3]
		} else if bm := logBareRe.FindStringSubmatch(m[2]); bm != nil {
			e.ID, e.Detail = bm[1], bm[2]
		}
		out = append(out, e)
	}
	return out
}

func readLog(root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, LogName))
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	return string(data), err
}

// movedFrom extracts the old ID from a Move event's detail.
func movedFrom(detail string) string {
	if m := movedFromRe.FindStringSubmatch(detail); m != nil {
		return m[1]
	}
	return ""
}
