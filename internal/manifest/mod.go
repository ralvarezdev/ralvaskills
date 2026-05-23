// Package manifest handles rsk.mod and rsk.lock read/write for project-local installs.
package manifest

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/ralvarezdev/ralvaskills/internal/fsperm"
	"github.com/ralvarezdev/ralvaskills/internal/fsx"
	"github.com/ralvarezdev/ralvaskills/internal/schema"
	"github.com/ralvarezdev/ralvaskills/internal/tool"
)

// Mod is the in-memory representation of rsk.mod.
type Mod struct {
	// SchemaVersion identifies the file format version. Version 0 is accepted
	// as legacy (pre-versioning) and treated as version 1.
	SchemaVersion schema.Version    `toml:"version,omitempty"`
	Tools         []tool.ID         `toml:"tools"`
	Pinned        []string          `toml:"pinned"`
	Skills        map[string]string `toml:"skills"`
}

// normalize fills nil/empty fields with safe zero-value defaults.
func (m *Mod) normalize() {
	if len(m.Tools) == 0 {
		m.Tools = []tool.ID{tool.ClaudeID}
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
	if err := schema.Check(m.SchemaVersion, schema.Mod); err != nil {
		return Mod{}, fmt.Errorf("rsk.mod: %w", err)
	}
	m.normalize()
	return m, nil
}

// WriteMod encodes m as TOML to rskDir/rsk.mod atomically, creating parent
// directories as needed.
func WriteMod(rskDir string, m Mod) error {
	m.normalize()
	m.SchemaVersion = schema.Mod
	path := ModPath(rskDir)
	if err := os.MkdirAll(filepath.Dir(path), fsperm.Dir); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	return fsx.WriteAtomic(path, TempFilePattern, func(w io.Writer) error {
		return toml.NewEncoder(w).Encode(m)
	})
}
