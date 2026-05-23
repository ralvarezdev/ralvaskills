package manifest

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/ralvarezdev/ralvaskills/internal/fsperm"
	"github.com/ralvarezdev/ralvaskills/internal/fsx"
	"github.com/ralvarezdev/ralvaskills/internal/schema"
	"github.com/ralvarezdev/ralvaskills/internal/skill"
)

type (
	// LockEntry records the resolved state of one installed skill.
	LockEntry struct {
		Name    string       `toml:"name"`
		Version string       `toml:"version"`
		Source  skill.Source `toml:"source"`
		Path    string       `toml:"path"`
	}

	// Lock is the in-memory representation of rsk.lock.
	Lock struct {
		// SchemaVersion identifies the file format version. Version 0 is
		// accepted as legacy (pre-versioning) and treated as version 1.
		SchemaVersion schema.Version `toml:"version,omitempty"`
		Skills        []LockEntry    `toml:"skills"`
	}
)

// ReadLock decodes rsk.lock from rskDir. Returns an empty Lock if the file
// does not exist.
func ReadLock(rskDir string) (Lock, error) {
	path := LockPath(rskDir)
	if _, statErr := os.Stat(path); errors.Is(statErr, os.ErrNotExist) {
		return Lock{}, nil
	}

	var l Lock
	if _, err := toml.DecodeFile(path, &l); err != nil {
		return Lock{}, fmt.Errorf("read %s: %w", path, err)
	}
	if err := schema.Check(l.SchemaVersion, schema.Lock); err != nil {
		return Lock{}, fmt.Errorf("rsk.lock: %w", err)
	}
	return l, nil
}

// WriteLock encodes l as TOML to rskDir/rsk.lock atomically, creating parent
// directories as needed.
func WriteLock(rskDir string, l Lock) error {
	l.SchemaVersion = schema.Lock
	path := LockPath(rskDir)
	if err := os.MkdirAll(filepath.Dir(path), fsperm.Dir); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	return fsx.WriteAtomic(path, TempFilePattern, func(w io.Writer) error {
		return toml.NewEncoder(w).Encode(l)
	})
}

// UpsertLockEntry adds entry to l if no entry with the same name exists,
// or replaces the existing entry in place.
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

// RemoveLockEntry removes the entry with the given name from l. No-op if
// the name is not present.
func RemoveLockEntry(l Lock, name string) Lock {
	kept := make([]LockEntry, 0, len(l.Skills))
	for _, e := range l.Skills {
		if e.Name != name {
			kept = append(kept, e)
		}
	}
	l.Skills = kept
	return l
}
