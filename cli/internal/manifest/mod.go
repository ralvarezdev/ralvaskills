// Package manifest handles rsk.mod and rsk.lock read/write for project-local installs.
package manifest

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// writeFileAtomic writes to a temp file in the same directory as path, then
// renames it into place. This ensures the target is never half-written: readers
// always see either the old file or the fully written new file.
func writeFileAtomic(path string, write func(io.Writer) error) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".rsk-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	tmpPath := tmp.Name()

	if err := write(tmp); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename %s → %s: %w", tmpPath, path, err)
	}
	return nil
}

// ToolID identifies a supported AI tool by its string name in rsk.mod.
type ToolID string

// MarshalText implements encoding.TextMarshaler.
func (t ToolID) MarshalText() ([]byte, error) { return []byte(t), nil }

// UnmarshalText implements encoding.TextUnmarshaler, rejecting unknown tool names.
func (t *ToolID) UnmarshalText(b []byte) error {
	switch ToolID(b) {
	case ToolClaudeCode, ToolOpenCode:
		*t = ToolID(b)
		return nil
	default:
		return fmt.Errorf("unknown tool %q: must be %q or %q", string(b), ToolClaudeCode, ToolOpenCode)
	}
}

// Tool name constants used in Mod.Tools and in cmd-layer dispatch.
const (
	ToolClaudeCode ToolID = "claude-code"
	ToolOpenCode   ToolID = "opencode"
)

const ModFile = "rsk.mod"

// Mod is the in-memory representation of rsk.mod.
type Mod struct {
	Version string            `toml:"version"`
	Tools   []ToolID          `toml:"tools"`
	Pinned  []string          `toml:"pinned"`
	Skills  map[string]string `toml:"skills"`
}

// normalize fills nil/empty fields with safe defaults.
func (m *Mod) normalize() {
	if len(m.Tools) == 0 {
		m.Tools = []ToolID{ToolClaudeCode}
	}
	if m.Skills == nil {
		m.Skills = make(map[string]string)
	}
	if m.Pinned == nil {
		m.Pinned = []string{}
	}
}

// ModPath returns the canonical path of rsk.mod inside rskDir.
func ModPath(rskDir string) string {
	return filepath.Join(rskDir, ModFile)
}

// ReadMod decodes rsk.mod from rskDir.
func ReadMod(rskDir string) (Mod, error) {
	path := ModPath(rskDir)
	var m Mod
	if _, err := toml.DecodeFile(path, &m); err != nil {
		return Mod{}, fmt.Errorf("read %s: %w", path, err)
	}
	m.normalize()
	return m, nil
}

// WriteMod encodes m as TOML to rskDir/rsk.mod atomically, creating parent dirs as needed.
func WriteMod(rskDir string, m Mod) error {
	m.normalize()
	path := ModPath(rskDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	return writeFileAtomic(path, func(w io.Writer) error {
		return toml.NewEncoder(w).Encode(m)
	})
}
