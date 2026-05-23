// Package schema defines version constants for every rsk-managed file format
// and provides helpers for forward-compatibility validation and migration.
//
// Every persisted file (config.json, rsk.mod, rsk.lock, catalog.toml) carries
// a top-level "version" integer field. When rsk reads a file it calls Check to
// confirm the stored version is supported. When rsk writes a file it stamps the
// current version. Files that predate versioning carry version 0, which is
// treated as equivalent to version 1 (the initial stable format).
package schema

import (
	"errors"
	"fmt"
)

// Version is the schema version number embedded in persisted rsk files.
type Version int

const (
	// Config is the current schema version for ~/.config/rsk/config.json.
	Config Version = 1

	// Mod is the current schema version for .rsk/rsk.mod.
	Mod Version = 1

	// Lock is the current schema version for .rsk/rsk.lock.
	Lock Version = 1

	// Catalog is the current schema version for ~/.config/rsk/catalog.toml
	// and any user-override catalog.
	Catalog Version = 1
)

// ErrUnsupported is returned by Check when a file carries a version that this
// build of rsk cannot read.
var ErrUnsupported = errors.New("unsupported schema version")

// Check returns an error wrapping ErrUnsupported if stored is not compatible
// with current. Version 0 is accepted as legacy (pre-versioning) and treated
// as equivalent to version 1.
func Check(stored, current Version) error {
	if stored == 0 || stored == current {
		return nil
	}
	return fmt.Errorf("%w: file version %d, rsk supports %d — upgrade rsk or re-run 'rsk init'",
		ErrUnsupported, stored, current)
}
