package skill

import "fmt"

// Source identifies where an installed skill originates.
type Source int

const (
	SourceLocal    Source = iota // symlinked from the local ralvaskills repo
	SourceOfficial               // fetched from anthropics/skills
	SourceRegistry               // downloaded from the hosted registry
)

type sourceInfo struct {
	name  string
	label string
	short string
}

var sourceInfos = map[Source]sourceInfo{
	SourceLocal:    {name: "local", label: "[ralva]", short: "ralva"},
	SourceOfficial: {name: "official", label: "[anthr]", short: "anthr"},
	SourceRegistry: {name: "registry", label: "[reg]", short: "reg"},
}

// String returns the lowercase source identifier.
func (s Source) String() string {
	if info, ok := sourceInfos[s]; ok {
		return info.name
	}
	return "unknown"
}

// Label returns the short bracketed label shown in rsk status output.
func (s Source) Label() string {
	if info, ok := sourceInfos[s]; ok {
		return info.label
	}
	return "[?????]"
}

// CleanLabel returns the unstyled source label used in rsk list output.
func (s Source) CleanLabel() string {
	if info, ok := sourceInfos[s]; ok {
		return info.short
	}
	return "?????"
}

// MarshalText implements encoding.TextMarshaler so Source encodes as its
// string name in TOML, JSON, and other text-based formats.
func (s Source) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler so Source decodes from
// its string name in TOML, JSON, and other text-based formats.
func (s *Source) UnmarshalText(b []byte) error {
	for src, info := range sourceInfos {
		if string(b) == info.name {
			*s = src
			return nil
		}
	}
	return fmt.Errorf("unknown skill source %q", string(b))
}
