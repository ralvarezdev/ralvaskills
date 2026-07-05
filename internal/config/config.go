package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"

	"github.com/ralvarezdev/ralvaskills/internal/fsperm"
	"github.com/ralvarezdev/ralvaskills/internal/fsx"
	"github.com/ralvarezdev/ralvaskills/internal/schema"
)

const configTempPattern = ".rsk-cfg-*.tmp"

// Config holds rsk's runtime configuration stored at ~/.config/rsk/config.json.
type Config struct {
	// SchemaVersion identifies the file format version written by rsk.
	// Version 0 is accepted as legacy (pre-versioning) and treated as version 1.
	SchemaVersion schema.Version `json:"version,omitempty"`

	// RepoPath is the local clone of ralvaskills used for filesystem-based skill
	// resolution. When set it takes precedence over RegistryURL.
	RepoPath string `json:"repo_path"`

	// RegistryURL is the base URL of the hosted skill registry
	// (e.g. https://skills.ralvarez.dev). Used when RepoPath is not set.
	RegistryURL string `json:"registry_url"`

	// GlobalTargets maps tool IDs to their global skills directory paths.
	GlobalTargets map[string]string `json:"global_targets" validate:"required,min=1"`

	// DefaultTargetScope is the tool ID (or "all") selected when --for is omitted.
	DefaultTargetScope string `json:"default_target_scope" validate:"required"`

	// OfficialCache is the directory where the official anthropic/skills clone is stored.
	OfficialCache string `json:"official_cache" validate:"required"`

	// VersionsCache is the directory used to cache stack-version metadata.
	VersionsCache string `json:"versions_cache" validate:"required"`
}

// LocalMode reports whether the config points to a local repo clone rather than
// the hosted registry.
func (c Config) LocalMode() bool {
	return c.RepoPath != ""
}

// RegistryCache returns the directory used to cache skill downloads from the
// hosted registry. It is derived from OfficialCache so both caches share the
// same parent directory.
func (c Config) RegistryCache() string {
	return DefaultRegistryCachePath(c.OfficialCache)
}

// Exists reports whether the config file exists at the default path.
func Exists() bool {
	_, err := os.Stat(DefaultConfigFilePath())
	return err == nil
}

// Load reads and validates the config from the default path.
func Load() (Config, error) {
	return LoadFrom(DefaultConfigFilePath())
}

// LoadFrom reads and validates the config from path.
func LoadFrom(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("load config: %w", err)
	}
	var cfg Config
	if err = json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	if err = schema.Check(cfg.SchemaVersion, schema.Config); err != nil {
		return Config{}, fmt.Errorf("config %s: %w", path, err)
	}
	v := validator.New()
	if err = v.Struct(cfg); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}
	if cfg.RepoPath == "" && cfg.RegistryURL == "" {
		return Config{}, errors.New("invalid config: one of repo_path or registry_url is required")
	}
	return cfg, nil
}

// Save marshals cfg as indented JSON to the default path, creating parent dirs
// as needed.
func Save(cfg Config) error {
	return SaveTo(DefaultConfigFilePath(), cfg)
}

// SaveTo marshals cfg as indented JSON to path atomically, creating parent
// directories as needed. The file is never partially written.
func SaveTo(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), fsperm.Dir); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	cfg.SchemaVersion = schema.Config
	return fsx.WriteAtomic(path, configTempPattern, func(w io.Writer) error {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(cfg); err != nil {
			return fmt.Errorf("marshal config: %w", err)
		}
		return nil
	})
}
