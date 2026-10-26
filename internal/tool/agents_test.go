package tool

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureAgentsFile_createsSkeleton(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "my-service")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := ensureAgentsFile(dir); err != nil {
		t.Fatalf("ensureAgentsFile: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, AgentsFileName))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	content := string(data)
	if !strings.HasPrefix(content, "# my-service — agent guide\n") {
		t.Errorf("skeleton should start with the project name, got: %q", content)
	}
	for _, section := range []string{"## Commands", "## Layout", "## Conventions", "## Boundaries"} {
		if !strings.Contains(content, section) {
			t.Errorf("skeleton missing %q: %q", section, content)
		}
	}
}

func TestEnsureAgentsFile_keepsExistingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, AgentsFileName)
	if err := os.WriteFile(path, []byte("# hand written\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := ensureAgentsFile(dir); err != nil {
		t.Fatalf("ensureAgentsFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "# hand written\n" {
		t.Errorf("existing AGENTS.md was modified: %q", data)
	}
}

func TestEnsureAgentsImport_prependsOnce(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), ClaudeFileName)
	if err := os.WriteFile(path, []byte("# notes\n"+claudeImportLine+"\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := ensureAgentsImport(path); err != nil {
		t.Fatalf("first ensureAgentsImport: %v", err)
	}
	if err := ensureAgentsImport(path); err != nil {
		t.Fatalf("second ensureAgentsImport: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	content := string(data)
	if want := agentsImportLine + "\n# notes\n" + claudeImportLine + "\n"; content != want {
		t.Errorf("content:\ngot:  %q\nwant: %q", content, want)
	}
}

func TestEnsureAgentsImport_keepsLaterImport(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), ClaudeFileName)
	seeded := "# notes\n\n" + agentsImportLine + "\n"
	if err := os.WriteFile(path, []byte(seeded), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := ensureAgentsImport(path); err != nil {
		t.Fatalf("ensureAgentsImport: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != seeded {
		t.Errorf("a file that already imports AGENTS.md was reordered: %q", data)
	}
}

func TestEnsureAgentsImport_refusesSymlink(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "notes.md")
	link := filepath.Join(dir, ClaudeFileName)
	writeSymlink(t, target, link, "notes\n")

	if err := ensureAgentsImport(link); err == nil {
		t.Fatal("expected an error for a symlinked CLAUDE.md")
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(data) != "notes\n" {
		t.Errorf("target content = %q, want it unmodified", data)
	}
}

func TestAppendClaudeImport_addsMissingNewline(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), ClaudeFileName)
	if err := os.WriteFile(path, []byte("# notes"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := appendClaudeImport(path); err != nil {
		t.Fatalf("appendClaudeImport: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if want := "# notes\n" + claudeImportLine + "\n"; string(data) != want {
		t.Errorf("content:\ngot:  %q\nwant: %q", data, want)
	}
}

func TestSyncPinned_bootstrapsAgentsGuide(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "fresh")
	if err := os.MkdirAll(filepath.Join(dir, ".rsk"), 0o750); err != nil {
		t.Fatalf("mkdir .rsk: %v", err)
	}

	ct := &ClaudeTool{}
	if err := ct.SyncPinned(dir, []string{"go-architect"}); err != nil {
		t.Fatalf("SyncPinned: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, AgentsFileName)); err != nil {
		t.Fatalf("AGENTS.md not created: %v", err)
	}

	root, err := os.ReadFile(filepath.Join(dir, ClaudeFileName))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	want := agentsImportLine + "\n" + claudeImportLine + "\n"
	if string(root) != want {
		t.Errorf("CLAUDE.md content:\ngot:  %q\nwant: %q", root, want)
	}

	// Idempotent: a second sync must not stack a second AGENTS.md import.
	if err = ct.SyncPinned(dir, []string{"go-architect"}); err != nil {
		t.Fatalf("second SyncPinned: %v", err)
	}
	root2, err := os.ReadFile(filepath.Join(dir, ClaudeFileName))
	if err != nil {
		t.Fatalf("reread CLAUDE.md: %v", err)
	}
	if strings.Count(string(root2), agentsImportLine) != 1 {
		t.Errorf("AGENTS.md import duplicated: %q", root2)
	}
}
