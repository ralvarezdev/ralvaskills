package config

import (
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/ralvarezdev/ralvaskills/cli/internal/skill"
)

const (
	// ConfigDirPermission is the permission bits mask for config directory creation.
	ConfigDirPermission = 0o750

	// ConfigFilePermission is the permission bits mask for config file creation.
	ConfigFilePermission = 0o600

	// ConfigFolderName is the name of the folder under ~/.config/ where rsk looks for config.json.
	ConfigFolderName = "rsk"

	// ConfigFileName is the name of the config file stored at ~/.config/rsk/config.json.
	ConfigFileName = "config.json"

	// CatalogFileName is the name of the catalog file stored at ~/.config/rsk/catalog.toml, which contains metadata about the local skill catalog.
	CatalogFileName = "catalog.toml"

	// RegistryFolderName is the name of the subdirectory under ~/.config/rsk/ where registry-related files are stored, such as cached tarballs and the registry index.
	RegistryFolderName = "registry"

	// DefaultRegistryURL is the default URL of the hosted skill registry.
	DefaultRegistryURL = "https://skills.ralvarez.dev"

	// GitDirPermission is the permission mode for the .git directory when initializing a new git repository.
	GitDirPermission = 0o750 // rwxr-x---

	// ClaudeFolderName is the name of the directory where Claude Code configuration file are stored in home.
	ClaudeFolderName = ".claude"

	// OpenCodeFolderName is the name of the directory where OpenCode configuration files are stored in home.
	OpenCodeFolderName = "opencode"
)

// DefaultConfigFolderPath returns the canonical path of the config folder (~/.config/rsk).
func DefaultConfigFolderPath() string {
	return filepath.Join(xdg.ConfigHome, ConfigFolderName)
}

// DefaultConfigFilePath returns the canonical path of config.json (~/.config/rsk/config.json).
func DefaultConfigFilePath() string {
	return filepath.Join(xdg.ConfigHome, ConfigFolderName, ConfigFileName)
}

// DefaultCatalogPath returns the canonical path of the user catalog
// (~/.config/rsk/catalog.toml).
func DefaultCatalogPath() string {
	return filepath.Join(xdg.ConfigHome, ConfigFolderName, CatalogFileName)
}

// DefaultRegistryCachePath returns the directory used to cache skill downloads
// from the hosted registry. Derived from officialCache so both caches sit under
// the same parent directory.
//
// officialCache is required by config validation, so the fallback branch is
// only reachable if called on an unvalidated Config.
func DefaultRegistryCachePath(officialCache string) string {
	if officialCache != "" {
		return filepath.Join(filepath.Dir(filepath.Clean(officialCache)), RegistryFolderName)
	}
	return filepath.Join(xdg.CacheHome, ConfigFolderName, RegistryFolderName)
}

// ClaudeSkillsDir returns the canonical path to the Claude Code skills directory, which is where Claude Code looks for skills to load. This is typically located at ~/.claude/skills.
func ClaudeSkillsDir() string {
	return filepath.Join(xdg.Home, ClaudeFolderName, skill.SkillsFolderName)
}

// OpenCodeSkillsDir returns the canonical path to the OpenCode skills directory, which is where OpenCode looks for skills to load. This is typically located at ~/.opencode/skills.
func OpenCodeSkillsDir() string {
	return filepath.Join(xdg.ConfigHome, OpenCodeFolderName, skill.SkillsFolderName)
}
