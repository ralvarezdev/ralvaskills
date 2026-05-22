package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ralvarezdev/ralvaskills/internal/skill"
)

const (
	// InstructionsKey is the key in opencode.json where pinned instructions are stored.
	InstructionsKey = "instructions"
)

// SyncOpenCodeInstructions updates opencode.json in projectDir so that the
// .rsk/skills/<name>/SKILL.md entries match pinnedNames exactly.
// All non-rsk entries in the instructions array are left untouched.
// Creates opencode.json if it does not exist.
func SyncOpenCodeInstructions(projectDir string, pinnedNames []string) error {
	path := OpenCodeConfigFilePath(projectDir)

	cfg, err := readOpenCodeConfig(path)
	if err != nil {
		return err
	}

	kept := filterNonRskEntries(cfg[InstructionsKey])
	for _, name := range pinnedNames {
		kept = append(kept, LocalSkillsPrefix+name+"/"+skill.SkillFileName)
	}

	if len(kept) == 0 {
		delete(cfg, InstructionsKey)
	} else {
		cfg[InstructionsKey] = kept
	}

	return writeOpenCodeConfig(path, cfg)
}

// RemoveOpenCodeInstructions removes all .rsk/skills/ entries from opencode.json
// in projectDir. Returns nil if the file does not exist.
func RemoveOpenCodeInstructions(projectDir string) error {
	path := OpenCodeConfigFilePath(projectDir)

	cfg, err := readOpenCodeConfig(path)
	if err != nil {
		return err
	}

	kept := filterNonRskEntries(cfg[InstructionsKey])
	if len(kept) == 0 {
		delete(cfg, InstructionsKey)
	} else {
		cfg[InstructionsKey] = kept
	}

	return writeOpenCodeConfig(path, cfg)
}

// filterNonRskEntries returns all entries from raw that are not rsk-managed
// (.rsk/skills/ prefix). raw must be the value of cfg["instructions"].
func filterNonRskEntries(raw any) []any {
	existing, _ := raw.([]any)
	kept := make([]any, 0, len(existing))
	for _, e := range existing {
		if s, ok := e.(string); ok && strings.HasPrefix(s, LocalSkillsPrefix) {
			continue
		}
		kept = append(kept, e)
	}
	return kept
}

func readOpenCodeConfig(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return make(map[string]any), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	cfg := make(map[string]any)
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return cfg, nil
}

func writeOpenCodeConfig(path string, cfg map[string]any) error {
	return writeFileAtomic(path, func(w io.Writer) error {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(cfg)
	})
}
