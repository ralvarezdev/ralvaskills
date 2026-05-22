// Package config handles loading and saving the rsk configuration file.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/go-playground/validator/v10"
)

// Config holds rsk's runtime configuration stored at ~/.config/rsk/config.json.
type Config struct {
	// RepoPath is the local clone of ralvaskills used for filesystem-based skill
	// resolution. When set it takes precedence over RegistryURL.
	RepoPath string `json:"repo_path"`

	// RegistryURL is the base URL of the hosted skill registry
	// (e.g. https://skills.ralvarez.dev). Used when RepoPath is not set.
	RegistryURL string `json:"registry_url"`

	GlobalTargets      map[string]string `json:"global_targets"       validate:"required,min=1"`
	DefaultTargetScope string            `json:"default_target_scope" validate:"required"`
	OfficialCache      string            `json:"official_cache"       validate:"required"`
	VersionsCache      string            `json:"versions_cache"       validate:"required"`
}

// LocalMode reports whether the config points to a local repo clone rather than
// the hosted registry.
func (c Config) LocalMode() bool { return c.RepoPath != "" }

// RegistryCache returns the directory used to cache skill downloads from the
// hosted registry. Derived from OfficialCache so both caches sit under the
// same ~/.ralvaskills/cache/ root.
//
// OfficialCache is required by config validation, so the fallback branch is
// only reachable if RegistryCache is called on an unvalidated Config.
func (c Config) RegistryCache() string {
	if c.OfficialCache != "" {
		return filepath.Join(filepath.Dir(filepath.Clean(c.OfficialCache)), "registry")
	}
	return filepath.Join(xdg.Home, ".ralvaskills", "cache", "registry")
}

// DefaultPath returns the canonical path of config.json (~/.config/rsk/config.json).
func DefaultPath() string {
	return filepath.Join(xdg.ConfigHome, "rsk", "config.json")
}

// Exists reports whether the config file exists at the default path.
func Exists() bool {
	_, err := os.Stat(DefaultPath())
	return err == nil
}

// Load reads and validates config from the default path.
func Load() (Config, error) {
	return LoadFrom(DefaultPath())
}

// LoadFrom reads and validates config from path.
func LoadFrom(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("load config: %w", err)
	}
	var cfg Config
	if err = json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	v := validator.New()
	if err = v.Struct(cfg); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}
	if cfg.RepoPath == "" && cfg.RegistryURL == "" {
		return Config{}, fmt.Errorf("invalid config: one of repo_path or registry_url is required")
	}
	return cfg, nil
}

// Save marshals cfg as indented JSON to the default path, creating parent dirs.
func Save(cfg Config) error {
	return SaveTo(DefaultPath(), cfg)
}

// SaveTo marshals cfg as indented JSON to path, creating parent dirs.
func SaveTo(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	err = os.WriteFile(path, append(data, '\n'), 0o600)
	if err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}
	return nil
}
