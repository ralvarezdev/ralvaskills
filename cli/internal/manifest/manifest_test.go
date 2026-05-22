package manifest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ralvarezdev/ralvaskills/cli/internal"
	"github.com/ralvarezdev/ralvaskills/cli/internal/skill"
)

func TestWriteReadMod_roundtrip(t *testing.T) {
	dir := t.TempDir()

	m := Mod{
		Version: "1",
		Tools:   []internal.ToolID{internal.ToolClaudeCode},
		Pinned:  []string{"go-architect"},
		Skills:  map[string]string{"go-architect": "*", "tdd": "1.2.0"},
	}
	if err := WriteMod(dir, m); err != nil {
		t.Fatalf("WriteMod: %v", err)
	}

	got, err := ReadMod(dir)
	if err != nil {
		t.Fatalf("ReadMod: %v", err)
	}
	if got.Version != m.Version {
		t.Errorf("Version: got %q want %q", got.Version, m.Version)
	}
	if len(got.Skills) != 2 {
		t.Errorf("Skills len: got %d want 2", len(got.Skills))
	}
	if got.Skills["tdd"] != "1.2.0" {
		t.Errorf("Skills[tdd]: got %q want %q", got.Skills["tdd"], "1.2.0")
	}
	if len(got.Pinned) != 1 || got.Pinned[0] != "go-architect" {
		t.Errorf("Pinned: got %v", got.Pinned)
	}
}

func TestWriteMod_atomic(t *testing.T) {
	dir := t.TempDir()

	// Write once, then overwrite — original content must not be visible mid-write.
	m1 := Mod{Version: "1", Skills: map[string]string{"a": "*"}}
	m2 := Mod{Version: "1", Skills: map[string]string{"b": "*"}}

	if err := WriteMod(dir, m1); err != nil {
		t.Fatalf("first WriteMod: %v", err)
	}
	if err := WriteMod(dir, m2); err != nil {
		t.Fatalf("second WriteMod: %v", err)
	}

	got, err := ReadMod(dir)
	if err != nil {
		t.Fatalf("ReadMod: %v", err)
	}
	if _, ok := got.Skills["b"]; !ok {
		t.Error("expected skill 'b' after overwrite")
	}
	if _, ok := got.Skills["a"]; ok {
		t.Error("unexpected stale skill 'a' after overwrite")
	}
}

func TestReadMod_missingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := ReadMod(dir)
	if err == nil {
		t.Fatal("expected error for missing rsk.mod, got nil")
	}
}

func TestWriteReadLock_roundtrip(t *testing.T) {
	dir := t.TempDir()

	l := Lock{Skills: []LockEntry{
		{Name: "go-architect", Version: "1.0.0", Source: skill.SourceLocal, Path: "/skills/go-architect"},
		{Name: "tdd", Version: "2.1.0", Source: skill.SourceOfficial, Path: "/cache/tdd"},
	}}
	if err := WriteLock(dir, l); err != nil {
		t.Fatalf("WriteLock: %v", err)
	}

	got, err := ReadLock(dir)
	if err != nil {
		t.Fatalf("ReadLock: %v", err)
	}
	if len(got.Skills) != 2 {
		t.Fatalf("Skills len: got %d want 2", len(got.Skills))
	}
	if got.Skills[0].Name != "go-architect" {
		t.Errorf("first entry: got %q want %q", got.Skills[0].Name, "go-architect")
	}
	if got.Skills[1].Version != "2.1.0" {
		t.Errorf("second version: got %q want %q", got.Skills[1].Version, "2.1.0")
	}
}

func TestReadLock_missingFileReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	got, err := ReadLock(dir)
	if err != nil {
		t.Fatalf("expected no error for missing lock, got %v", err)
	}
	if len(got.Skills) != 0 {
		t.Errorf("expected empty lock, got %v", got.Skills)
	}
}

func TestUpsertRemoveLockEntry(t *testing.T) {
	l := Lock{}
	l = UpsertLockEntry(l, LockEntry{Name: "a", Version: "1.0.0"})
	l = UpsertLockEntry(l, LockEntry{Name: "b", Version: "2.0.0"})

	if len(l.Skills) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(l.Skills))
	}

	// Upsert existing updates in place.
	l = UpsertLockEntry(l, LockEntry{Name: "a", Version: "1.1.0"})
	if len(l.Skills) != 2 {
		t.Fatalf("upsert grew slice to %d", len(l.Skills))
	}
	if l.Skills[0].Version != "1.1.0" {
		t.Errorf("upsert: got version %q want %q", l.Skills[0].Version, "1.1.0")
	}

	// Remove.
	l = RemoveLockEntry(l, "a")
	if len(l.Skills) != 1 || l.Skills[0].Name != "b" {
		t.Errorf("after remove: %v", l.Skills)
	}

	// Remove non-existent is no-op.
	l = RemoveLockEntry(l, "nonexistent")
	if len(l.Skills) != 1 {
		t.Errorf("remove non-existent changed len to %d", len(l.Skills))
	}
}

func TestWritePinned(t *testing.T) {
	dir := t.TempDir()

	if err := WritePinned(dir, []string{"go-architect", "tdd"}); err != nil {
		t.Fatalf("WritePinned: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, ClaudeFileName))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	got := string(content)
	want := "@skills/go-architect/SKILL.md\n@skills/tdd/SKILL.md\n"
	if got != want {
		t.Errorf("CLAUDE.md content:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestWritePinned_empty(t *testing.T) {
	dir := t.TempDir()
	if err := WritePinned(dir, nil); err != nil {
		t.Fatalf("WritePinned empty: %v", err)
	}
	info, err := os.Stat(filepath.Join(dir, ClaudeFileName))
	if err != nil {
		t.Fatalf("CLAUDE.md not created: %v", err)
	}
	if info.Size() != 0 {
		t.Errorf("expected empty file, got %d bytes", info.Size())
	}
}

func TestAppendImport_idempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), ClaudeFileName)

	if err := AppendImport(path); err != nil {
		t.Fatalf("first AppendImport: %v", err)
	}
	if err := AppendImport(path); err != nil {
		t.Fatalf("second AppendImport: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	count := 0
	for i := 0; i < len(content); i++ {
		if i+len(ClaudeImportLine) <= len(content) && string(content[i:i+len(ClaudeImportLine)]) == ClaudeImportLine {
			count++
		}
	}
	if count != 1 {
		t.Errorf("importLine appears %d times, want 1", count)
	}
}

func TestRemoveImport(t *testing.T) {
	path := filepath.Join(t.TempDir(), ClaudeFileName)
	os.WriteFile(path, []byte("# header\n"+ClaudeImportLine+"\n# footer\n"), LocalFilePermission)

	if err := RemoveImport(path); err != nil {
		t.Fatalf("RemoveImport: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(content) == "" {
		t.Fatal("RemoveImport removed everything")
	}
	for _, line := range []string{string(content)} {
		if line == ClaudeImportLine {
			t.Error("importLine still present after RemoveImport")
		}
	}
}

func TestRemoveImport_missingFile(t *testing.T) {
	err := RemoveImport(filepath.Join(t.TempDir(), "nonexistent.md"))
	if err != nil {
		t.Errorf("expected nil for missing file, got %v", err)
	}
}
