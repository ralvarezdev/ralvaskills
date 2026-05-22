package manifest

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const importLine = "@.rsk/CLAUDE.md"

// WritePinned writes .rsk/CLAUDE.md atomically with one @skills/<name>/SKILL.md line per pinned entry.
// An empty pinnedEntries produces an empty file.
func WritePinned(rskDir string, pinnedEntries []string) error {
	path := filepath.Join(rskDir, "CLAUDE.md")
	return writeFileAtomic(path, func(w io.Writer) error {
		for _, name := range pinnedEntries {
			if _, err := fmt.Fprintf(w, "@skills/%s/SKILL.md\n", name); err != nil {
				return err
			}
		}
		return nil
	})
}

// AppendImport appends "@.rsk/CLAUDE.md" to claudeMDPath once (idempotent).
// Creates the file if it does not exist.
func AppendImport(claudeMDPath string) error {
	data, err := os.ReadFile(claudeMDPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", claudeMDPath, err)
	}
	if strings.Contains(string(data), importLine) {
		return nil
	}
	f, err := os.OpenFile(claudeMDPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open %s: %w", claudeMDPath, err)
	}
	_, writeErr := fmt.Fprintln(f, importLine)
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

// RemoveImport removes every occurrence of "@.rsk/CLAUDE.md" from claudeMDPath.
// Returns nil if the file does not exist.
func RemoveImport(claudeMDPath string) error {
	data, err := os.ReadFile(claudeMDPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read %s: %w", claudeMDPath, err)
	}
	lines := strings.Split(string(data), "\n")
	kept := lines[:0]
	for _, line := range lines {
		if strings.TrimSpace(line) != importLine {
			kept = append(kept, line)
		}
	}
	payload := []byte(strings.Join(kept, "\n"))
	return writeFileAtomic(claudeMDPath, func(w io.Writer) error {
		_, err := w.Write(payload)
		return err
	})
}
