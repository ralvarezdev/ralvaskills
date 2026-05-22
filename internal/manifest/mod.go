// Package manifest handles rsk.mod and rsk.lock read/write for project-local installs.
package manifest

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/ralvarezdev/ralvaskills/internal"
)

// Mod is the in-memory representation of rsk.mod.
type Mod struct {
	Version string             `toml:"version"`
	Tools   []internal.ToolID  `toml:"tools"`
	Pinned  []string           `toml:"pinned"`
	Skills  map[string]string  `toml:"skills"`
}

// normalize fills nil/empty fields with safe defaults.
func (m *Mod) normalize() {
	if len(m.Tools) == 0 {
		m.Tools = []internal.ToolID{internal.ToolClaudeCode}
	}
	if m.Skills == nil {
		m.Skills = make(map[string]string)
	}
	if m.Pinned == nil {
		m.Pinned = []string{}
	}
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
	if err := os.MkdirAll(filepath.Dir(path), LocalDirPermission); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	return writeFileAtomic(path, func(w io.Writer) error {
		return toml.NewEncoder(w).Encode(m)
	})
}
