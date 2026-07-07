package config

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"

	"github.com/ralvarezdev/ralvaskills/internal/schema"
	"github.com/ralvarezdev/ralvaskills/internal/skill"
)

//go:embed catalog.toml
var defaultCatalogData []byte

type (
	// SkillRef is a reference to a named skill from a specific source.
	SkillRef struct {
		Name   string       `toml:"name"`
		Source skill.Source `toml:"source"`
	}

	// Bundle is a named, ordered set of skill references.
	Bundle struct {
		Name        string     `toml:"name"`
		Description string     `toml:"description"`
		Skills      []SkillRef `toml:"skills"`
	}

	// catalogFile is the top-level shape of a catalog TOML file.
	catalogFile struct {
		// SchemaVersion identifies the file format version. Version 0 is
		// accepted as legacy (pre-versioning) and treated as version 1.
		SchemaVersion schema.Version `toml:"version,omitempty"`
		Bundle        []Bundle       `toml:"bundle"`
	}
)

// LoadCatalog returns the merged bundle catalog. It always starts with the
// embedded defaults and then applies user overrides from userCatalogPath.
// Pass an empty string to use DefaultCatalogPath.
//
// When the user catalog exists but cannot be read or parsed, the embedded
// defaults are returned together with a non-nil error. Callers should surface
// this error as a warning — rsk continues to work, but user customisations are
// not applied.
func LoadCatalog(userCatalogPath string) ([]Bundle, error) {
	defaults := mustParseEmbedded()

	if userCatalogPath == "" {
		userCatalogPath = DefaultCatalogPath()
	}

	raw, err := os.ReadFile(userCatalogPath)
	if os.IsNotExist(err) {
		return defaults, nil
	}
	if err != nil {
		return defaults, fmt.Errorf("user catalog unreadable (%s): %w", userCatalogPath, err)
	}

	var user catalogFile
	if _, err = toml.Decode(string(raw), &user); err != nil {
		return defaults, fmt.Errorf("user catalog parse error (%s): %w", userCatalogPath, err)
	}
	if err = schema.Check(user.SchemaVersion, schema.Catalog); err != nil {
		return defaults, fmt.Errorf("user catalog %s: %w", userCatalogPath, err)
	}

	return merge(defaults, user.Bundle), nil
}

// mustParseEmbedded parses the embedded catalog.toml and panics on failure.
// A parse error here is a build-time bug, not a runtime condition.
func mustParseEmbedded() []Bundle {
	var cat catalogFile
	if _, err := toml.Decode(string(defaultCatalogData), &cat); err != nil {
		panic(fmt.Sprintf("bundles: embedded catalog.toml is malformed: %v", err))
	}
	return cat.Bundle
}

// merge applies user bundles on top of defaults. A user bundle whose name
// matches an existing default replaces it entirely; bundles with new names
// are appended in the order they appear in the user catalog.
func merge(defaults, user []Bundle) []Bundle {
	index := make(map[string]int, len(defaults))
	result := make([]Bundle, 0, len(defaults)+len(user))
	result = append(result, defaults...)
	for i, b := range result {
		index[b.Name] = i
	}
	for _, b := range user {
		if i, ok := index[b.Name]; ok {
			result[i] = b
		} else {
			result = append(result, b)
		}
	}
	return result
}

// FindBundle returns the bundle with the given name, and whether it was found.
func FindBundle(bundles []Bundle, name string) (Bundle, bool) {
	for _, b := range bundles {
		if b.Name == name {
			return b, true
		}
	}
	return Bundle{}, false
}

// BundleNames returns all bundle names in catalog order.
func BundleNames(bundles []Bundle) []string {
	names := make([]string, len(bundles))
	for i, b := range bundles {
		names[i] = b.Name
	}
	return names
}
