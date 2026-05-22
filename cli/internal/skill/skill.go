// Package skill defines core types and operations for discovered skills on disk.
package skill

import (
	"path/filepath"
	"strings"
)

// Source identifies where an installed skill originates.
type Source int

const (
	SourceLocal    Source = iota // symlinked from the local ralvaskills repo
	SourceOfficial               // fetched from anthropics/skills
)

// String returns the lowercase source identifier.
func (s Source) String() string {
	switch s {
	case SourceLocal:
		return "local"
	case SourceOfficial:
		return "official"
	default:
		return "unknown"
	}
}

// Label returns the short bracketed label shown in rsk status output.
func (s Source) Label() string {
	switch s {
	case SourceLocal:
		return "[ralva]"
	case SourceOfficial:
		return "[anthr]"
	default:
		return "[?????]"
	}
}

// Skill represents a skill discovered on disk.
type Skill struct {
	Name       string
	Version    string
	Path       string // absolute path to the skill folder
	Source     Source
	IsPersonal bool // true when path contains a "personal/" segment
}

// IsPersonalPath reports whether path contains a "personal" path segment.
func IsPersonalPath(path string) bool {
	clean := filepath.ToSlash(filepath.Clean(path))
	return strings.Contains(clean, "/personal/") ||
		strings.HasSuffix(clean, "/personal")
}
