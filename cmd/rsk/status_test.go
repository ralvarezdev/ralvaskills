package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ralvarezdev/ralvaskills/internal/config"
)

// TestRunStatusGlobalDoesNotRequireProject guards against the regression where
// `rsk status --global` failed with "no rsk.mod found" even though --global
// explicitly means "skip the current project and look at the global install."
func TestRunStatusGlobalDoesNotRequireProject(t *testing.T) {
	home := t.TempDir()
	t.Setenv(config.EnvConfigHome, home)

	globalDir := filepath.Join(home, "global-claude")
	if err := os.MkdirAll(globalDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		RegistryURL:        "https://example.invalid",
		GlobalTargets:      map[string]string{"claude-code": globalDir},
		DefaultTargetScope: "all",
		OfficialCache:      filepath.Join(home, "official"),
		VersionsCache:      filepath.Join(home, "versions"),
	}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	// cwd has no .rsk/rsk.mod — --global must still work.
	projDir := t.TempDir()
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origWD) })
	if err = os.Chdir(projDir); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	statusCmd.SetOut(&buf)
	t.Cleanup(func() { statusCmd.SetOut(nil) })

	if err = runStatus(statusCmd, statusOpts{global: true}); err != nil {
		t.Fatalf("runStatus with --global outside a project: unexpected error: %v", err)
	}
}
