package bundle

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
)

// LockFile is the advisory lock writers hold, at the bundle root.
const LockFile = ".ritari.lock"

// LockTimeout is how long a writer waits for another writer to finish.
var LockTimeout = 10 * time.Second

// WriteFileAtomic writes data to path via a temporary file and a rename, so a
// reader never sees a partial file. Parent directories are created.
func WriteFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	mode := fs.FileMode(0o644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer func() { _ = os.Remove(name) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(name, mode); err != nil {
		return err
	}
	return os.Rename(name, path)
}

// writeIfChanged writes data unless the file already holds exactly those
// bytes. It reports whether it wrote.
func writeIfChanged(path string, data []byte) (bool, error) {
	if old, err := os.ReadFile(path); err == nil && bytes.Equal(old, data) {
		return false, nil
	}
	return true, WriteFileAtomic(path, data)
}

// withLock runs fn while holding the bundle's writer lock.
func withLock(root string, fn func() error) error {
	lockPath := filepath.Join(root, LockFile)
	fl := flock.New(lockPath)
	ctx, cancel := context.WithTimeout(context.Background(), LockTimeout)
	defer cancel()
	ok, err := fl.TryLockContext(ctx, 50*time.Millisecond)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("locking bundle: %w", err)
	}
	if !ok {
		return fmt.Errorf("bundle is locked by another process (%s)", lockPath)
	}
	defer func() { _ = fl.Unlock() }()
	return fn()
}

func readAll(path string) (string, error) {
	data, err := os.ReadFile(path)
	return string(data), err
}
