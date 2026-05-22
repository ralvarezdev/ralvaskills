package manifest

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ralvarezdev/ralvaskills/internal/skill"
)

const (
	// ClaudeImportLine is the line that should be appended to CLAUDE.md to import pinned skills.
	ClaudeImportLine = "@.rsk/CLAUDE.md"

	// ClaudeFileOpenFlags are the flags used to open CLAUDE.md for appending.
	ClaudeFileOpenFlags = os.O_APPEND | os.O_CREATE | os.O_WRONLY
)

// ClaudeSkillPathReference is the format string for lines in CLAUDE.md referencing pinned skills.
// Uses "/" not filepath.Separator: CLAUDE.md imports use Unix-style paths regardless of host OS.
var ClaudeSkillPathReference = "@" + skill.SkillsFolderName + "/%s/" + skill.SkillFileName + "\n"

// WritePinned writes .rsk/CLAUDE.md atomically with one @skills/<name>/SKILL.md line per pinned entry.
// An empty pinnedEntries produces an empty file.
func WritePinned(rskDir string, pinnedEntries []string) error {
	path := filepath.Join(rskDir, ClaudeFileName)
	return writeFileAtomic(path, func(w io.Writer) error {
		for _, name := range pinnedEntries {
			if _, err := fmt.Fprintf(w, ClaudeSkillPathReference, name); err != nil {
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
	if strings.Contains(string(data), ClaudeImportLine) {
		return nil
	}

	f, err := os.OpenFile(claudeMDPath, ClaudeFileOpenFlags, LocalFilePermission)
	if err != nil {
		return fmt.Errorf("open %s: %w", claudeMDPath, err)
	}

	_, writeErr := fmt.Fprintln(f, ClaudeImportLine)
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
		if strings.TrimSpace(line) != ClaudeImportLine {
			kept = append(kept, line)
		}
	}

	payload := []byte(strings.Join(kept, "\n"))
	return writeFileAtomic(claudeMDPath, func(w io.Writer) error {
		_, err := w.Write(payload)
		return err
	})
}
