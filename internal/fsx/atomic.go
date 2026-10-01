// Package fsx provides filesystem helpers used throughout rsk.
package fsx

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ralvarezdev/ralvaskills/v3/internal/fsperm"
)

// WriteAtomic writes to a temporary file in the same directory as path, then
// renames the temporary file into place. The target is never partially written:
// readers always see either the previous file or the fully written new file.
//
// If path already exists, its permission bits are carried over to the new file.
// os.CreateTemp creates at 0600 and rename preserves the temp file's mode, so
// without this an overwrite would silently tighten a 0644 file to 0600. A
// nonexistent path gets newPerm.
//
// tempPattern is passed directly to os.CreateTemp; it must contain a "*"
// placeholder for the random suffix (e.g. ".rsk-*.tmp").
func WriteAtomic(path, tempPattern string, write func(io.Writer) error) error {
	return WriteAtomicPerm(path, tempPattern, fsperm.File, write)
}

// WriteAtomicPerm is WriteAtomic with an explicit mode for newly created files.
// Existing files keep their current mode regardless of newPerm.
func WriteAtomicPerm(path, tempPattern string, newPerm os.FileMode, write func(io.Writer) error) error {
	dir := filepath.Dir(path)

	// Stat (not Lstat) so a symlinked path adopts its target's mode; the
	// write-through guard for symlinks lives in the callers that own the
	// policy, not here.
	perm := newPerm.Perm()
	if fi, statErr := os.Stat(path); statErr == nil {
		perm = fi.Mode().Perm()
	}

	tmp, err := os.CreateTemp(dir, tempPattern)
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	tmpPath := tmp.Name()

	if err = write(tmp); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	// Chmod before Close so the mode is never briefly wrong, and so a
	// failure leaves nothing behind to clean up.
	if err = tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("chmod %s: %w", tmpPath, err)
	}
	if err = tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close temp file: %w", err)
	}
	if err = os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename %s → %s: %w", tmpPath, path, err)
	}
	return nil
}
