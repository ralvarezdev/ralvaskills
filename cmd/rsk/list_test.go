package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ralvarezdev/termkit"
	"github.com/ralvarezdev/termkit/output"

	"github.com/ralvarezdev/ralvaskills/v2/internal/config"
)

// TestRunListGlobalLabelsRowsByTool guards against the regression where a
// skill installed for multiple tools printed as indistinguishable duplicate
// rows in `rsk list --global` — each row must show which tool it belongs to.
func TestRunListGlobalLabelsRowsByTool(t *testing.T) {
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

	claudeDir := filepath.Join(home, "targets", "claude-code")
	opencodeDir := filepath.Join(home, "targets", "opencode")
	for _, dir := range []string{claudeDir, opencodeDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(skillSrc, filepath.Join(dir, "demo-skill")); err != nil {
			t.Fatal(err)
		}
	}

	cfg := config.Config{
		RegistryURL: "https://example.invalid",
		GlobalTargets: map[string]string{
			"claude-code": claudeDir,
			"opencode":    opencodeDir,
		},
		DefaultTargetScope: "all",
		OfficialCache:      filepath.Join(home, "official"),
		VersionsCache:      filepath.Join(home, "versions"),
	}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	var buf bytes.Buffer
	listCmd.SetOut(&buf)
	t.Cleanup(func() { listCmd.SetOut(nil) })

	if err := runListGlobal(listCmd, listOpts{global: true, format: output.FormatText}); err != nil {
		t.Fatalf("runListGlobal: unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "claude-code") || !strings.Contains(out, "opencode") {
		t.Fatalf("expected output to label rows by tool (claude-code/opencode), got:\n%s", out)
	}

	lines := make([]string, 0)
	for line := range strings.SplitSeq(out, "\n") {
		if strings.Contains(line, "demo-skill") {
			lines = append(lines, line)
		}
	}
	if len(lines) != 2 {
		t.Fatalf("expected 2 rows for demo-skill (one per tool), got %d:\n%v", len(lines), lines)
	}
	if lines[0] == lines[1] {
		t.Fatalf("expected the two demo-skill rows to differ (tool label), got identical rows: %q", lines[0])
	}
}

// TestRunListGlobalCapture guards the picker's captured/scrollable path for
// `rsk list --global`: under a termkit.WithCapture-wrapped context, the table must
// land in the capture (for the TUI's result screen to render) and nothing
// must be written to the command's own writer, since a captured run's output
// would otherwise print behind the TUI.
func TestRunListGlobalCapture(t *testing.T) {
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

	claudeDir := filepath.Join(home, "targets", "claude-code")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(skillSrc, filepath.Join(claudeDir, "demo-skill")); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		RegistryURL:        "https://example.invalid",
		GlobalTargets:      map[string]string{"claude-code": claudeDir},
		DefaultTargetScope: "all",
		OfficialCache:      filepath.Join(home, "official"),
		VersionsCache:      filepath.Join(home, "versions"),
	}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	var buf bytes.Buffer
	listCmd.SetOut(&buf)
	listCmd.SetContext(termkit.WithCapture(t.Context(), &termkit.Capture{}))
	t.Cleanup(func() {
		listCmd.SetOut(nil)
		listCmd.SetContext(context.Background())
	})

	capture := termkit.CaptureFromContext(listCmd.Context())
	if capture == nil {
		t.Fatal("expected a capture on the command's context")
	}

	if err := runListGlobal(listCmd, listOpts{global: true, format: output.FormatText}); err != nil {
		t.Fatalf("runListGlobal: unexpected error: %v", err)
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
