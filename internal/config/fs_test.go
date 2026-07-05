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
	if got, want2 := DefaultConfigFilePath(), filepath.Join(want, ConfigFileName); got != want2 {
		t.Errorf("DefaultConfigFilePath() = %q, want %q", got, want2)
	}
	if got, want2 := DefaultCatalogPath(), filepath.Join(want, CatalogFileName); got != want2 {
		t.Errorf("DefaultCatalogPath() = %q, want %q", got, want2)
	}
}

func TestDefaultConfigFolderPathNoEnvOverride(t *testing.T) {
	t.Setenv(EnvConfigHome, "")

	got := DefaultConfigFolderPath()
	if filepath.Base(got) != ConfigFolderName {
		t.Errorf("DefaultConfigFolderPath() = %q, want a path ending in %q", got, ConfigFolderName)
	}
}
