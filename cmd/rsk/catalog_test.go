package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ralvarezdev/termkit"

	"github.com/ralvarezdev/ralvaskills/v2/internal/config"
	"github.com/ralvarezdev/ralvaskills/v2/internal/skill"
)

// TestRunCatalogSkillsCapture guards the picker's captured/scrollable path
// for `rsk catalog`: under a termkit.WithCapture-wrapped context, the skill table
// must land in the capture and nothing must be written to the command's own
// writer.
func TestRunCatalogSkillsCapture(t *testing.T) {
	home := t.TempDir()
	t.Setenv(config.EnvConfigHome, home)

	repoPath := filepath.Join(home, "repo")
	skillDir := filepath.Join(repoPath, skill.SkillsFolderName, "demo-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	skillMD := "---\nname: demo-skill\nversion: 1.0.0\n---\nbody\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillMD), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		RepoPath:           repoPath,
		GlobalTargets:      map[string]string{"claude-code": filepath.Join(home, "global-claude")},
		DefaultTargetScope: "all",
		OfficialCache:      filepath.Join(home, "official"),
		VersionsCache:      filepath.Join(home, "versions"),
	}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	var buf bytes.Buffer
	catalogCmd.SetOut(&buf)
	catalogCmd.SetContext(termkit.WithCapture(t.Context(), &termkit.Capture{}))
	t.Cleanup(func() {
		catalogCmd.SetOut(nil)
		catalogCmd.SetContext(context.Background())
	})

	capture := termkit.CaptureFromContext(catalogCmd.Context())
	if capture == nil {
		t.Fatal("expected a capture on the command's context")
	}

	if err := runCatalog(catalogCmd, catalogOpts{output: outputText}); err != nil {
		t.Fatalf("runCatalog: unexpected error: %v", err)
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
}

// TestRunCatalogSkillsCaptureJSON guards the picker's captured/scrollable
// path for `rsk catalog -o json`: under capture, JSON output must land in
// the capture's raw payload (rendered as a scrollable viewport by
// termkit.ResultView) instead of silently falling back to a table.
func TestRunCatalogSkillsCaptureJSON(t *testing.T) {
	home := t.TempDir()
	t.Setenv(config.EnvConfigHome, home)

	repoPath := filepath.Join(home, "repo")
	skillDir := filepath.Join(repoPath, skill.SkillsFolderName, "demo-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	skillMD := "---\nname: demo-skill\nversion: 1.0.0\n---\nbody\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillMD), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		RepoPath:           repoPath,
		GlobalTargets:      map[string]string{"claude-code": filepath.Join(home, "global-claude")},
		DefaultTargetScope: "all",
		OfficialCache:      filepath.Join(home, "official"),
		VersionsCache:      filepath.Join(home, "versions"),
	}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	var buf bytes.Buffer
	catalogCmd.SetOut(&buf)
	catalogCmd.SetContext(termkit.WithCapture(t.Context(), &termkit.Capture{}))
	t.Cleanup(func() {
		catalogCmd.SetOut(nil)
		catalogCmd.SetContext(context.Background())
	})

	capture := termkit.CaptureFromContext(catalogCmd.Context())
	if capture == nil {
		t.Fatal("expected a capture on the command's context")
	}

	if err := runCatalog(catalogCmd, catalogOpts{output: outputJSON}); err != nil {
		t.Fatalf("runCatalog: unexpected error: %v", err)
	}

	if buf.Len() != 0 {
		t.Fatalf("expected no output written to the command's writer under capture, got:\n%s", buf.String())
	}
	if len(capture.Tables()) != 0 {
		t.Fatalf("expected no captured tables for -o json, got %d", len(capture.Tables()))
	}

	raw := capture.Raw()
	if raw == "" {
		t.Fatal("expected a non-empty raw JSON payload")
	}
	var entries []skillEntry
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		t.Fatalf("capture.Raw() is not valid JSON: %v\n%s", err, raw)
	}
	if len(entries) != 1 || entries[0].Name != "demo-skill" {
		t.Fatalf("decoded entries = %+v, want one entry named demo-skill", entries)
	}
}
