package tool

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ralvarezdev/ralvaskills/internal/fsperm"
)

func TestClaudeID_registered(t *testing.T) {
	_, ok := Get(ClaudeID)
	if !ok {
		t.Errorf("ClaudeID %q not registered", ClaudeID)
	}
}

func TestOpenCodeID_registered(t *testing.T) {
	_, ok := Get(OpenCodeID)
	if !ok {
		t.Errorf("OpenCodeID %q not registered", OpenCodeID)
	}
}

func TestAll_order(t *testing.T) {
	tools := All()
	if len(tools) < 2 {
		t.Fatalf("expected at least 2 tools, got %d", len(tools))
	}
	for i := 1; i < len(tools); i++ {
		if tools[i].ID() <= tools[i-1].ID() {
			t.Errorf("All() not sorted at index %d: %q >= %q", i, tools[i].ID(), tools[i-1].ID())
		}
	}
}

func TestWritePinnedClaude_roundtrip(t *testing.T) {
	dir := t.TempDir()

	if err := writePinnedClaude(dir, []string{"go-architect", "tdd"}); err != nil {
		t.Fatalf("writePinnedClaude: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, ClaudeFileName))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	got := string(content)
	want := "@../.claude/skills/go-architect/SKILL.md\n@../.claude/skills/tdd/SKILL.md\n"
	if got != want {
		t.Errorf("CLAUDE.md content:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestWritePinnedClaude_empty(t *testing.T) {
	dir := t.TempDir()
	if err := writePinnedClaude(dir, nil); err != nil {
		t.Fatalf("writePinnedClaude empty: %v", err)
	}
	info, err := os.Stat(filepath.Join(dir, ClaudeFileName))
	if err != nil {
		t.Fatalf("CLAUDE.md not created: %v", err)
	}
	if info.Size() != 0 {
		t.Errorf("expected empty file, got %d bytes", info.Size())
	}
}

func TestAppendClaudeImport_idempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), ClaudeFileName)

	if err := appendClaudeImport(path); err != nil {
		t.Fatalf("first appendClaudeImport: %v", err)
	}
	if err := appendClaudeImport(path); err != nil {
		t.Fatalf("second appendClaudeImport: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if strings.Count(string(content), claudeImportLine) != 1 {
		t.Errorf("importLine appears %d times, want 1", strings.Count(string(content), claudeImportLine))
	}
}

func TestRemoveClaudeImport(t *testing.T) {
	path := filepath.Join(t.TempDir(), ClaudeFileName)
	if err := os.WriteFile(path, []byte("# header\n"+claudeImportLine+"\n# footer\n"), fsperm.File); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := removeClaudeImport(path); err != nil {
		t.Fatalf("removeClaudeImport: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(content) == 0 {
		t.Fatal("removeClaudeImport removed everything")
	}
	if strings.Contains(string(content), claudeImportLine) {
		t.Error("importLine still present after removeClaudeImport")
	}
}

func TestRemoveClaudeImport_missingFile(t *testing.T) {
	err := removeClaudeImport(filepath.Join(t.TempDir(), "nonexistent.md"))
	if err != nil {
		t.Errorf("expected nil for missing file, got %v", err)
	}
}
