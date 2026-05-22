package config

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/adrg/xdg"
)

//go:embed catalog.toml
var defaultCatalogData []byte

// SourceID identifies where a skill originates.
type SourceID string

const (
	SourceLocal    SourceID = "local"    // symlinked from the local ralvaskills repo
	SourceOfficial SourceID = "official" // fetched from anthropics/skills
)

// SkillRef is a reference to a named skill from a specific source.
type SkillRef struct {
	Name   string   `toml:"name"`
	Source SourceID `toml:"source"`
}

// Bundle is a named, ordered set of skill references.
type Bundle struct {
	Name        string     `toml:"name"`
	Description string     `toml:"description"`
	Skills      []SkillRef `toml:"skills"`
}

// catalogFile is the top-level shape of a catalog TOML file.
type catalogFile struct {
	Bundle []Bundle `toml:"bundle"`
}

// DefaultCatalogPath returns the canonical path of the user catalog
// (~/.config/rsk/catalog.toml).
func DefaultCatalogPath() string {
	return filepath.Join(xdg.ConfigHome, "rsk", "catalog.toml")
}

// LoadCatalog returns the merged bundle catalog. It always starts with the
// embedded defaults and then applies user overrides from userCatalogPath.
// Pass an empty string to use DefaultCatalogPath.
//
// User catalog errors (missing file excluded) are logged as warnings and the
// defaults are returned unchanged, so a misconfigured user catalog never
// prevents rsk from running.
func LoadCatalog(userCatalogPath string) []Bundle {
	defaults := mustParseEmbedded()

	if userCatalogPath == "" {
		userCatalogPath = DefaultCatalogPath()
	}

	raw, err := os.ReadFile(userCatalogPath)
	if os.IsNotExist(err) {
		return defaults
	}
	if err != nil {
		slog.Warn("user catalog unreadable; using defaults", "path", userCatalogPath, "err", err)
		return defaults
	}

	var user catalogFile
	if _, err = toml.Decode(string(raw), &user); err != nil {
		slog.Warn("user catalog parse error; using defaults", "path", userCatalogPath, "err", err)
		return defaults
	}

	return merge(defaults, user.Bundle)
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
	result := make([]Bundle, len(defaults))
	copy(result, defaults)
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

// FindBundle returns the bundle with the given name from bundles and whether it
// was found.
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
