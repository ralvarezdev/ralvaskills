package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ralvarezdev/ralvaskills/v2/internal/config"
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
	t.Chdir(projDir)

	var buf bytes.Buffer
	statusCmd.SetOut(&buf)
	t.Cleanup(func() { statusCmd.SetOut(nil) })

	if err := runStatus(statusCmd, statusOpts{global: true, output: outputText}); err != nil {
		t.Fatalf("runStatus with --global outside a project: unexpected error: %v", err)
	}
}

// TestRunStatusJSONOutput checks that -o json produces well-formed JSON with
// one section per scanned target directory, even when everything is empty —
// and that runStatus rejects an invalid --output value.
func TestRunStatusJSONOutput(t *testing.T) {
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

	projDir := t.TempDir()
	t.Chdir(projDir)

	var buf bytes.Buffer
	statusCmd.SetOut(&buf)
	t.Cleanup(func() { statusCmd.SetOut(nil) })

	if err := runStatus(statusCmd, statusOpts{global: true, output: outputJSON}); err != nil {
		t.Fatalf("runStatus with -o json: unexpected error: %v", err)
	}

	var sections []statusSectionEntry
	if err := json.Unmarshal(buf.Bytes(), &sections); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if len(sections) != 1 || sections[0].Dir != globalDir {
		t.Fatalf("unexpected sections: %+v", sections)
	}

	buf.Reset()
	if err := runStatus(statusCmd, statusOpts{global: true, output: outputFormat("bogus")}); err == nil {
		t.Fatal("expected error for invalid --output value, got nil")
	}
}
