package skill

import "fmt"

// Source identifies where an installed skill originates.
type Source string

const (
	// SourceLocal is a skill symlinked from the local ralvaskills repo clone.
	SourceLocal Source = "local"

	// SourceOfficial is a skill fetched from the official anthropic/skills cache.
	SourceOfficial Source = "official"

	// SourceRegistry is a skill downloaded from the hosted registry.
	SourceRegistry Source = "registry"
)

// ParseSource converts a string to its Source value. Returns an error for
// unrecognized values.
func ParseSource(s string) (Source, error) {
	switch Source(s) {
	case SourceLocal, SourceOfficial, SourceRegistry:
		return Source(s), nil
	default:
		return "", fmt.Errorf("unknown skill source %q", s)
	}
}

// String returns the string representation of the source, identical to its
// wire encoding.
func (s Source) String() string { return string(s) }

// Label returns the short display label used in table output.
func (s Source) Label() string {
	switch s {
	case SourceLocal:
		return "ralva"
	case SourceOfficial:
		return "anthr"
	case SourceRegistry:
		return "reg  "
	default:
		return "?????"
	}
}

// MarshalText implements encoding.TextMarshaler so Source encodes as its
// string value in TOML, JSON, and other text-based formats.
func (s Source) MarshalText() ([]byte, error) { return []byte(s), nil }

// UnmarshalText implements encoding.TextUnmarshaler so Source decodes from
// its string value in TOML, JSON, and other text-based formats.
func (s *Source) UnmarshalText(b []byte) error {
	src, err := ParseSource(string(b))
	if err != nil {
		return err
	}
	*s = src
	return nil
}
