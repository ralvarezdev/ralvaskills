package manifest

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/ralvarezdev/ralvaskills/cli/internal/skill"
)

const LockFile = "rsk.lock"

// LockEntry records the resolved state of one installed skill.
type LockEntry struct {
	Name    string       `toml:"name"`
	Version string       `toml:"version"`
	Source  skill.Source `toml:"source"`
	Path    string       `toml:"path"` // resolved symlink target
}

// Lock is the in-memory representation of rsk.lock.
type Lock struct {
	Skills []LockEntry `toml:"skills"`
}

// LockPath returns the canonical path of rsk.lock inside rskDir.
func LockPath(rskDir string) string {
	return filepath.Join(rskDir, LockFile)
}

// ReadLock decodes rsk.lock from rskDir. Returns an empty Lock if the file does not exist.
func ReadLock(rskDir string) (Lock, error) {
	path := LockPath(rskDir)
	var l Lock
	if _, statErr := os.Stat(path); errors.Is(statErr, os.ErrNotExist) {
		return Lock{}, nil
	}
	if _, err := toml.DecodeFile(path, &l); err != nil {
		return Lock{}, fmt.Errorf("read %s: %w", path, err)
	}
	return l, nil
}

// WriteLock encodes l as TOML to rskDir/rsk.lock atomically, creating parent dirs as needed.
func WriteLock(rskDir string, l Lock) error {
	path := LockPath(rskDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	return writeFileAtomic(path, func(w io.Writer) error {
		return toml.NewEncoder(w).Encode(l)
	})
}

// UpsertLockEntry adds or replaces the entry matching entry.Name in l.
func UpsertLockEntry(l Lock, entry LockEntry) Lock {
	for i, e := range l.Skills {
		if e.Name == entry.Name {
			l.Skills[i] = entry
			return l
		}
	}
	l.Skills = append(l.Skills, entry)
	return l
}

// RemoveLockEntry removes the entry with the given name from l (no-op if absent).
func RemoveLockEntry(l Lock, name string) Lock {
	kept := l.Skills[:0]
	for _, e := range l.Skills {
		if e.Name != name {
			kept = append(kept, e)
		}
	}
	l.Skills = kept
	return l
}
