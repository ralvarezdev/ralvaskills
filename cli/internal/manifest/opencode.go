package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	openCodeConfigFile = "opencode.json"
	rskSkillsPrefix    = ".rsk/skills/"
)

// SyncOpenCodeInstructions updates opencode.json in projectDir so that the
// .rsk/skills/<name>/SKILL.md entries match pinnedNames exactly.
// All non-rsk entries in the instructions array are left untouched.
// Creates opencode.json if it does not exist.
func SyncOpenCodeInstructions(projectDir string, pinnedNames []string) error {
	path := filepath.Join(projectDir, openCodeConfigFile)

	cfg, err := readOpenCodeConfig(path)
	if err != nil {
		return err
	}

	kept := filterNonRskEntries(cfg["instructions"])
	for _, name := range pinnedNames {
		kept = append(kept, rskSkillsPrefix+name+"/SKILL.md")
	}

	if len(kept) == 0 {
		delete(cfg, "instructions")
	} else {
		cfg["instructions"] = kept
	}

	return writeOpenCodeConfig(path, cfg)
}

// RemoveOpenCodeInstructions removes all .rsk/skills/ entries from opencode.json
// in projectDir. Returns nil if the file does not exist.
func RemoveOpenCodeInstructions(projectDir string) error {
	path := filepath.Join(projectDir, openCodeConfigFile)

	cfg, err := readOpenCodeConfig(path)
	if err != nil {
		return err
	}

	kept := filterNonRskEntries(cfg["instructions"])
	if len(kept) == 0 {
		delete(cfg, "instructions")
	} else {
		cfg["instructions"] = kept
	}

	return writeOpenCodeConfig(path, cfg)
}

// filterNonRskEntries returns all entries from raw that are not rsk-managed
// (.rsk/skills/ prefix). raw must be the value of cfg["instructions"].
func filterNonRskEntries(raw any) []any {
	existing, _ := raw.([]any)
	kept := make([]any, 0, len(existing))
	for _, e := range existing {
		if s, ok := e.(string); ok && strings.HasPrefix(s, rskSkillsPrefix) {
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
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", path, err)
	}
	payload := append(data, '\n')
	return writeFileAtomic(path, func(w io.Writer) error {
		_, err := w.Write(payload)
		return err
	})
}
