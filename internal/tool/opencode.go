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
	"github.com/ralvarezdev/ralvaskills/internal/fsx"
	"github.com/ralvarezdev/ralvaskills/internal/skill"
)

const (
	// OpenCodeID is the tool identifier for OpenCode.
	OpenCodeID ID = "opencode"

	openCodeFolderName      = "opencode"
	openCodeFileName        = "opencode.json"
	openCodeInstructionsKey = "instructions"
	openCodeSkillsPrefix    = ".rsk/" + skill.SkillsFolderName + "/"
)

func init() { Register(&openCodeTool{}) }

// openCodeTool implements Tool for OpenCode.
type openCodeTool struct{}

// ID returns OpenCodeID.
func (*openCodeTool) ID() ID { return OpenCodeID }

// SkillsDir returns the canonical path to the OpenCode global skills directory
// (~/.config/opencode/skills).
func (*openCodeTool) SkillsDir() string {
	return filepath.Join(xdg.ConfigHome, openCodeFolderName, skill.SkillsFolderName)
}

// SyncPinned updates opencode.json in projectDir so that the
// .rsk/skills/<name>/SKILL.md entries match pinnedNames exactly. Non-rsk
// entries are preserved. Creates the file if it does not exist.
func (*openCodeTool) SyncPinned(projectDir string, pinnedNames []string) error {
	return updateOCInstructions(filepath.Join(projectDir, openCodeFileName), pinnedNames)
}

// RemovePinned removes all rsk-managed .rsk/skills/ entries from opencode.json
// in projectDir. Returns nil if the file does not exist.
func (*openCodeTool) RemovePinned(projectDir string) error {
	return updateOCInstructions(filepath.Join(projectDir, openCodeFileName), nil)
}

// updateOCInstructions reads opencode.json, replaces the rsk-managed
// .rsk/skills/ entries with pinnedNames, and writes the file back atomically.
// Pass a nil or empty pinnedNames to remove all rsk entries.
func updateOCInstructions(path string, pinnedNames []string) error {
	cfg, err := readOCConfig(path)
	if err != nil {
		return err
	}

	kept := filterOCEntries(cfg[openCodeInstructionsKey])
	for _, name := range pinnedNames {
		kept = append(kept, openCodeSkillsPrefix+name+"/"+skill.SkillFileName)
	}

	if len(kept) == 0 {
		delete(cfg, openCodeInstructionsKey)
	} else {
		cfg[openCodeInstructionsKey] = kept
	}

	return writeOCConfig(path, cfg)
}

// filterOCEntries returns the subset of raw that are not rsk-managed
// (.rsk/skills/ prefix). raw must be cfg[openCodeInstructionsKey].
func filterOCEntries(raw any) []any {
	existing, ok := raw.([]any)
	if !ok {
		return nil
	}
	kept := make([]any, 0, len(existing))
	for _, e := range existing {
		if s, isStr := e.(string); isStr && strings.HasPrefix(s, openCodeSkillsPrefix) {
			continue
		}
		kept = append(kept, e)
	}
	return kept
}

func readOCConfig(path string) (map[string]any, error) {
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

func writeOCConfig(path string, cfg map[string]any) error {
	return fsx.WriteAtomic(path, rskTempPattern, func(w io.Writer) error {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(cfg)
	})
}
