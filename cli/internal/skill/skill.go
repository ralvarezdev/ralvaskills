// Package skill defines core types and operations for discovered skills on disk.
package skill

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

// Source identifies where an installed skill originates.
type Source int

const (
	SourceLocal    Source = iota // symlinked from the local ralvaskills repo
	SourceOfficial               // fetched from anthropics/skills
	SourceRegistry               // downloaded from the hosted registry
)

// String returns the lowercase source identifier.
func (s Source) String() string {
	switch s {
	case SourceLocal:
		return "local"
	case SourceOfficial:
		return "official"
	case SourceRegistry:
		return "registry"
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
	case SourceRegistry:
		return "[reg]"
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

// MarshalText implements encoding.TextMarshaler so Source encodes as its
// string name in TOML, JSON, and other text-based formats.
func (s Source) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler so Source decodes from
// its string name in TOML, JSON, and other text-based formats.
func (s *Source) UnmarshalText(b []byte) error {
	switch string(b) {
	case "local":
		*s = SourceLocal
	case "official":
		*s = SourceOfficial
	case "registry":
		*s = SourceRegistry
	default:
		return fmt.Errorf("unknown skill source %q", string(b))
	}
	return nil
}

// IsPersonalPath reports whether path contains a "personal" path segment.
func IsPersonalPath(path string) bool {
	clean := filepath.ToSlash(filepath.Clean(path))
	return slices.Contains(strings.Split(clean, "/"), "personal")
}
