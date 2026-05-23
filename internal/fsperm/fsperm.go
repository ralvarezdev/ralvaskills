// Package fsperm defines the filesystem permission constants used throughout rsk.
// All three values are intentionally centralized here so that adding or changing
// a permission only ever requires editing one file.
package fsperm

const (
	// Dir is the permission mode applied to every directory rsk creates
	// (rwxr-x---). The group bits allow system-level tooling to read the
	// tree; others have no access.
	Dir = 0o750

	// File is the permission mode applied to every data file rsk writes
	// (rw-r--r--). rsk files contain no secrets, so world-readable is fine
	// and consistent with standard config/lock-file conventions.
	File = 0o644

	// Mask is a bitwise AND mask applied to permission bits read from tar
	// archive headers before writing extracted files. It prevents tarball
	// entries from granting execute or setuid/setgid bits that were never
	// intended.
	Mask = 0o777
)
