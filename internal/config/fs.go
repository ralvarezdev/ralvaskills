// Package config handles loading and saving the rsk configuration file.
package config

import (
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/ralvarezdev/ralvaskills/internal/skill"
)

const (
	// ConfigFolderName is the name of the folder under ~/.config/ where rsk
	// stores its configuration.
	ConfigFolderName = "rsk"

	// ConfigFileName is the filename of the main rsk configuration file.
	ConfigFileName = "config.json"

	// CatalogFileName is the filename of the user-overridable skill catalog.
	CatalogFileName = "catalog.toml"

	// RegistryFolderName is the subdirectory under the config folder used to
	// cache tarballs and the registry index.
	RegistryFolderName = "registry"

	// DefaultRegistryURL is the base URL of the hosted skill registry used when
	// no local repo clone is configured.
	DefaultRegistryURL = "https://skills.ralvarez.dev"

	// ClaudeFolderName is the name of the Claude Code configuration directory
	// inside the user's home directory.
	ClaudeFolderName = ".claude"

	// OpenCodeFolderName is the name of the OpenCode configuration directory
	// inside the XDG config home.
	OpenCodeFolderName = "opencode"
)

// DefaultConfigFolderPath returns the canonical path of the rsk config folder
// (~/.config/rsk).
func DefaultConfigFolderPath() string {
	return filepath.Join(xdg.ConfigHome, ConfigFolderName)
}

// DefaultConfigFilePath returns the canonical path of the rsk config file
// (~/.config/rsk/config.json).
func DefaultConfigFilePath() string {
	return filepath.Join(xdg.ConfigHome, ConfigFolderName, ConfigFileName)
}

// DefaultCatalogPath returns the canonical path of the user skill catalog
// (~/.config/rsk/catalog.toml).
func DefaultCatalogPath() string {
	return filepath.Join(xdg.ConfigHome, ConfigFolderName, CatalogFileName)
}

// DefaultRegistryCachePath returns the directory used to cache skill downloads
// from the hosted registry. The cache is placed next to officialCache so both
// caches share the same parent directory.
//
// officialCache is required by config validation. The fallback branch is only
// reachable when called on an unvalidated Config.
func DefaultRegistryCachePath(officialCache string) string {
	if officialCache != "" {
		return filepath.Join(filepath.Dir(filepath.Clean(officialCache)), RegistryFolderName)
	}
	return filepath.Join(xdg.CacheHome, ConfigFolderName, RegistryFolderName)
}

// ClaudeSkillsDir returns the canonical path to the Claude Code global skills
// directory (~/.claude/skills).
func ClaudeSkillsDir() string {
	return filepath.Join(xdg.Home, ClaudeFolderName, skill.SkillsFolderName)
}

// OpenCodeSkillsDir returns the canonical path to the OpenCode global skills
// directory (~/.config/opencode/skills).
func OpenCodeSkillsDir() string {
	return filepath.Join(xdg.ConfigHome, OpenCodeFolderName, skill.SkillsFolderName)
}
