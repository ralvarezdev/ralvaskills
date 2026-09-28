package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ralvarezdev/termkit"

	"github.com/ralvarezdev/ralvaskills/v2/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/v2/internal/config"
	"github.com/ralvarezdev/ralvaskills/v2/internal/ui"
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

// TestRunStatusCapture guards the picker's captured/scrollable path for
// `rsk status --global`: under a ui.WithCapture-wrapped context, the section
// table must land in the capture and nothing must be written to the
// command's own writer.
func TestRunStatusCapture(t *testing.T) {
	home := t.TempDir()
	t.Setenv(config.EnvConfigHome, home)

	skillSrc := filepath.Join(home, "skill-src", "demo-skill")
	if err := os.MkdirAll(skillSrc, 0o755); err != nil {
		t.Fatal(err)
	}
	skillMD := "---\nname: demo-skill\nversion: 1.0.0\n---\nbody\n"
	if err := os.WriteFile(filepath.Join(skillSrc, "SKILL.md"), []byte(skillMD), 0o644); err != nil {
		t.Fatal(err)
	}

	globalDir := filepath.Join(home, "global-claude")
	if err := os.MkdirAll(globalDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(skillSrc, filepath.Join(globalDir, "demo-skill")); err != nil {
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
	statusCmd.SetContext(ui.WithCapture(t.Context(), &termkit.Capture{}))
	t.Cleanup(func() {
		statusCmd.SetOut(nil)
		statusCmd.SetContext(context.Background())
	})

	capture := ui.CaptureFromContext(statusCmd.Context())
	if capture == nil {
		t.Fatal("expected a capture on the command's context")
	}

	if err := runStatus(statusCmd, statusOpts{global: true, output: outputText}); err != nil {
		t.Fatalf("runStatus: unexpected error: %v", err)
	}

	if buf.Len() != 0 {
		t.Fatalf("expected no output written to the command's writer under capture, got:\n%s", buf.String())
	}

	tables := capture.Tables()
	if len(tables) != 1 {
		t.Fatalf("expected exactly one captured table, got %d", len(tables))
	}
	if len(tables[0].Rows) != 1 {
		t.Fatalf("expected exactly one captured row, got %d: %+v", len(tables[0].Rows), tables[0].Rows)
	}
	if len(tables[0].IDs) != 1 || tables[0].IDs[0] != "demo-skill" {
		t.Fatalf("expected row IDs [demo-skill], got %+v", tables[0].IDs)
	}
	if len(tables[0].Actions) != 1 || tables[0].Actions[0].Key != "u" || tables[0].Actions[0].Label != "uninstall" {
		t.Fatalf("expected one {Key:u Label:uninstall} action, got %+v", tables[0].Actions)
	}
}

// TestStatusRowActionScope checks that the uninstall scope is resolved from
// the captured table a status row action fired on: a "Global — <tool>"
// section targets that tool globally, and a "Project" section keeps
// uninstall's project default.
func TestStatusRowActionScope(t *testing.T) {
	t.Parallel()

	scope := statusRowActionScope(termkit.Data{Title: "Global — claude-code"})
	if scope[cmdx.FlagGlobal] != "true" {
		t.Errorf("global section scope[%s] = %q, want true", cmdx.FlagGlobal, scope[cmdx.FlagGlobal])
	}
	if scope[cmdx.FlagFor] != "claude-code" {
		t.Errorf("global section scope[%s] = %q, want claude-code", cmdx.FlagFor, scope[cmdx.FlagFor])
	}

	if got := statusRowActionScope(termkit.Data{Title: "Project"}); got != nil {
		t.Errorf("project section scope = %+v, want nil", got)
	}
}
