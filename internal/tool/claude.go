package tool

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/ralvarezdev/ralvaskills/internal"
	"github.com/ralvarezdev/ralvaskills/internal/fsperm"
	"github.com/ralvarezdev/ralvaskills/internal/fsx"
	"github.com/ralvarezdev/ralvaskills/internal/skill"
)

const (
	// ClaudeID is the tool identifier for Claude Code.
	ClaudeID ID = "claude-code"

	// ClaudeFolderName is the name of the Claude Code configuration directory
	// inside the user's home directory or inside the project.
	ClaudeFolderName = ".claude"

	// ClaudeFileName is the name of the generated file inside .rsk/ that lists
	// pinned skills for Claude Code to import.
	ClaudeFileName = "CLAUDE.md"

	// Settings.json configuration keys
	claudeSettingsFileName   = "settings.json"
	claudeSettingsPermKey    = "permissions"
	claudeSettingsAllowKey   = "allow"
	claudeSettingsDenyKey    = "deny"

	claudeImportLine    = "@" + internal.ProjectFolderName + "/" + ClaudeFileName
	claudeFileOpenFlags = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	rskTempPattern      = ".rsk-*.tmp"

	// claudeSkillPathReference is the format string for one skill import line in
	// .rsk/CLAUDE.md. Path is relative to .rsk/, so ../.claude/skills/<name>/SKILL.md
	// resolves to the project-local .claude/skills/ directory.
	// Uses "/" (not filepath.Separator) because CLAUDE.md imports use Unix-style paths.
	claudeSkillPathReference = "@../" + ClaudeFolderName + "/" + skill.SkillsFolderName + "/%s/" + skill.SkillFileName + "\n"
)

var (
	// ProjectClaudeFilePath is the path of CLAUDE.md relative to the project
	// root (.rsk/CLAUDE.md).
	ProjectClaudeFilePath = filepath.Join(internal.ProjectFolderName, ClaudeFileName)
)

func init() { Register(&ClaudeTool{}) }

// ClaudeTool implements Tool for Claude Code.
type ClaudeTool struct{}

// ID returns ClaudeID.
func (*ClaudeTool) ID() ID { return ClaudeID }

// SkillsDir returns the canonical path to the Claude Code global skills
// directory (~/.claude/skills).
func (*ClaudeTool) SkillsDir() string {
	return filepath.Join(xdg.Home, ClaudeFolderName, skill.SkillsFolderName)
}

// ProjectSkillsDir returns the project-local skills directory (.claude/skills/).
// Both Claude Code and OpenCode discover skills from this path.
func (*ClaudeTool) ProjectSkillsDir(projectRoot string) string {
	return filepath.Join(projectRoot, ClaudeFolderName, skill.SkillsFolderName)
}

// SyncPinned writes .rsk/CLAUDE.md atomically with one import line per pinned
// skill and appends the import marker to the project's CLAUDE.md (idempotent).
// projectDir is the project root (parent of .rsk/).
func (*ClaudeTool) SyncPinned(projectDir string, pinnedNames []string) error {
	rskDir := filepath.Join(projectDir, internal.ProjectFolderName)
	if err := writePinnedClaude(rskDir, pinnedNames); err != nil {
		return fmt.Errorf("sync claude-code pins: %w", err)
	}
	if err := appendClaudeImport(filepath.Join(projectDir, ClaudeFileName)); err != nil {
		return fmt.Errorf("sync claude-code import: %w", err)
	}
	return nil
}

// RemovePinned removes the import marker from the project CLAUDE.md.
// Returns nil if the file does not exist.
func (*ClaudeTool) RemovePinned(projectDir string) error {
	return removeClaudeImport(filepath.Join(projectDir, ClaudeFileName))
}

func writePinnedClaude(rskDir string, pinnedNames []string) error {
	path := filepath.Join(rskDir, ClaudeFileName)
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

// SettingsPath returns the path to the project's .claude/settings.json file.
func (*ClaudeTool) SettingsPath(projectDir string) string {
	return filepath.Join(projectDir, ClaudeFolderName, claudeSettingsFileName)
}

// ReadPermissions reads the tool permissions from .claude/settings.json.
// Returns empty slices if the file does not exist. Permissions is a map with
// "allow" and "deny" keys, each mapping to a []string of rules.
func (ct *ClaudeTool) ReadPermissions(projectDir string) (allow, deny []string, err error) {
	path := ct.SettingsPath(projectDir)
	cfg, err := readClaudeConfig(path)
	if err != nil {
		return nil, nil, err
	}

	perms, ok := cfg[claudeSettingsPermKey].(map[string]any)
	if !ok {
		return nil, nil, nil
	}

	allow = toStringSlice(perms[claudeSettingsAllowKey])
	deny = toStringSlice(perms[claudeSettingsDenyKey])
	return allow, deny, nil
}

// WritePermissions writes tool permissions to .claude/settings.json, preserving
// other keys in the file. Creates the file if it does not exist.
func (ct *ClaudeTool) WritePermissions(projectDir string, allow, deny []string) error {
	path := ct.SettingsPath(projectDir)
	cfg, err := readClaudeConfig(path)
	if err != nil {
		return err
	}

	perms := map[string]any{}
	if len(allow) > 0 {
		perms[claudeSettingsAllowKey] = allow
	}
	if len(deny) > 0 {
		perms[claudeSettingsDenyKey] = deny
	}

	if len(perms) == 0 {
		delete(cfg, claudeSettingsPermKey)
	} else {
		cfg[claudeSettingsPermKey] = perms
	}

	return writeClaudeConfig(path, cfg)
}

func readClaudeConfig(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return make(map[string]any), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	cfg := make(map[string]any)
	if err = json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return cfg, nil
}

func writeClaudeConfig(path string, cfg map[string]any) error {
	return fsx.WriteAtomic(path, rskTempPattern, func(w io.Writer) error {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(cfg)
	})
}

func toStringSlice(v any) []string {
	if v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
