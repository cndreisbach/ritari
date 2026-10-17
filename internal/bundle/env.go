package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Env supplies the identity and clock for writes.
type Env struct {
	Actor string
	Now   func() time.Time
}

func (e Env) now() time.Time {
	if e.Now == nil {
		return time.Now()
	}
	return e.Now()
}

// Stamp is the current instant as an RFC 3339 UTC timestamp, second precision.
func (e Env) Stamp() string {
	return e.now().UTC().Truncate(time.Second).Format(time.RFC3339)
}

// Today is the current local calendar date, YYYY-MM-DD.
func (e Env) Today() string {
	return e.now().Local().Format(time.DateOnly)
}

// Update runs fn against a freshly loaded bundle while holding the writer lock,
// then records the events fn returns (§6.3).
func Update(root string, env Env, fn func(b *Bundle) ([]Event, error)) error {
	abs, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	if info, err := os.Stat(abs); err != nil {
		return err
	} else if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", abs)
	}
	return withLock(abs, func() error {
		b, err := Open(abs)
		if err != nil {
			return err
		}
		events, err := fn(b)
		if err != nil {
			return err
		}
		if len(events) == 0 {
			return nil
		}
		return Record(b.Root, env, events)
	})
}

// Record re-reads the bundle, appends events to log.md, and synchronizes indexes.
func Record(root string, env Env, events []Event) error {
	b, err := Open(root)
	if err != nil {
		return err
	}
	text, err := readLog(b.Root)
	if err != nil {
		return err
	}
	if _, err := writeIfChanged(b.Abs(LogName), []byte(appendLog(text, env.Today(), events))); err != nil {
		return err
	}
	_, err = b.SyncIndexes(false)
	return err
}

// Reindex regenerates index files under the writer lock.
func Reindex(root string, env Env) ([]string, error) {
	var changed []string
	err := Update(root, env, func(b *Bundle) ([]Event, error) {
		var err error
		changed, err = b.SyncIndexes(false)
		return nil, err
	})
	return changed, err
}
