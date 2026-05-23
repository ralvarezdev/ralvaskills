// Package fsx provides filesystem helpers used throughout rsk.
package fsx

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// WriteAtomic writes to a temporary file in the same directory as path, then
// renames the temporary file into place. The target is never partially written:
// readers always see either the previous file or the fully written new file.
//
// tempPattern is passed directly to os.CreateTemp; it must contain a "*"
// placeholder for the random suffix (e.g. ".rsk-*.tmp").
func WriteAtomic(path, tempPattern string, write func(io.Writer) error) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, tempPattern)
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	tmpPath := tmp.Name()

	if err = write(tmp); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err = tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close temp file: %w", err)
	}
	if err = os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename %s → %s: %w", tmpPath, path, err)
	}
	return nil
}
