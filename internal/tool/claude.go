package tool

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/ralvarezdev/ralvaskills/internal/fsperm"
	"github.com/ralvarezdev/ralvaskills/internal/fsx"
	"github.com/ralvarezdev/ralvaskills/internal/skill"
)

const (
	// ClaudeID is the tool identifier for Claude Code.
	ClaudeID ID = "claude-code"

	claudeHomeDir       = ".claude"
	claudeImportLine    = "@.rsk/CLAUDE.md"
	claudeFileOpenFlags = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	claudeFileName      = "CLAUDE.md"
	rskFolderName       = ".rsk"
	rskTempPattern      = ".rsk-*.tmp"
)

// claudeSkillPathReference is the format string for one skill import line in
// .rsk/CLAUDE.md. Uses "/" (not filepath.Separator) because CLAUDE.md imports
// use Unix-style paths regardless of the host OS.
var claudeSkillPathReference = "@" + skill.SkillsFolderName + "/%s/" + skill.SkillFileName + "\n"

func init() { Register(&claudeTool{}) }

// claudeTool implements Tool for Claude Code.
type claudeTool struct{}

// ID returns ClaudeID.
func (*claudeTool) ID() ID { return ClaudeID }

// SkillsDir returns the canonical path to the Claude Code global skills
// directory (~/.claude/skills).
func (*claudeTool) SkillsDir() string {
	return filepath.Join(xdg.Home, claudeHomeDir, skill.SkillsFolderName)
}

// SyncPinned writes .rsk/CLAUDE.md atomically with one import line per pinned
// skill and appends the import marker to the project's CLAUDE.md (idempotent).
// projectDir is the project root (parent of .rsk/).
func (*claudeTool) SyncPinned(projectDir string, pinnedNames []string) error {
	rskDir := filepath.Join(projectDir, rskFolderName)
	if err := writePinnedClaude(rskDir, pinnedNames); err != nil {
		return fmt.Errorf("sync claude-code pins: %w", err)
	}
	if err := appendClaudeImport(filepath.Join(projectDir, claudeFileName)); err != nil {
		return fmt.Errorf("sync claude-code import: %w", err)
	}
	return nil
}

// RemovePinned removes the import marker from the project CLAUDE.md.
// Returns nil if the file does not exist.
func (*claudeTool) RemovePinned(projectDir string) error {
	return removeClaudeImport(filepath.Join(projectDir, claudeFileName))
}

func writePinnedClaude(rskDir string, pinnedNames []string) error {
	path := filepath.Join(rskDir, claudeFileName)
	return fsx.WriteAtomic(path, rskTempPattern, func(w io.Writer) error {
		for _, name := range pinnedNames {
			if _, err := fmt.Fprintf(w, claudeSkillPathReference, name); err != nil {
				return err
			}
		}
		return nil
	})
}

func appendClaudeImport(claudeMDPath string) error {
	data, err := os.ReadFile(claudeMDPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", claudeMDPath, err)
	}
	if strings.Contains(string(data), claudeImportLine) {
		return nil
	}

	f, err := os.OpenFile(claudeMDPath, claudeFileOpenFlags, fsperm.File)
	if err != nil {
		return fmt.Errorf("open %s: %w", claudeMDPath, err)
	}

	_, writeErr := fmt.Fprintln(f, claudeImportLine)
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

func removeClaudeImport(claudeMDPath string) error {
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
		if strings.TrimSpace(line) != claudeImportLine {
			kept = append(kept, line)
		}
	}

	payload := []byte(strings.Join(kept, "\n"))
	return fsx.WriteAtomic(claudeMDPath, rskTempPattern, func(w io.Writer) error {
		_, writeErr := w.Write(payload)
		return writeErr
	})
}
