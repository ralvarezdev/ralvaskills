package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultConfigFolderPathEnvOverride(t *testing.T) {
	want := filepath.Join(t.TempDir(), "rsk-sandbox")
	t.Setenv(EnvConfigHome, want)

	if got := DefaultConfigFolderPath(); got != want {
		t.Errorf("DefaultConfigFolderPath() = %q, want %q", got, want)
	}
	if got, want := DefaultConfigFilePath(), filepath.Join(want, ConfigFileName); got != want {
		t.Errorf("DefaultConfigFilePath() = %q, want %q", got, want)
	}
	if got, want := DefaultCatalogPath(), filepath.Join(want, CatalogFileName); got != want {
		t.Errorf("DefaultCatalogPath() = %q, want %q", got, want)
	}
}

func TestDefaultConfigFolderPathNoEnvOverride(t *testing.T) {
	t.Setenv(EnvConfigHome, "")

	got := DefaultConfigFolderPath()
	if filepath.Base(got) != ConfigFolderName {
		t.Errorf("DefaultConfigFolderPath() = %q, want a path ending in %q", got, ConfigFolderName)
	}
}
