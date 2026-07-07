package manifest

import (
	"testing"

	"github.com/ralvarezdev/ralvaskills/internal/schema"
	"github.com/ralvarezdev/ralvaskills/internal/skill"
	"github.com/ralvarezdev/ralvaskills/internal/tool"
)

func TestWriteReadMod_roundtrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	m := Mod{
		Tools:  []tool.ID{tool.ClaudeID},
		Pinned: []string{"go-architect"},
		Skills: map[string]string{"go-architect": "*", "tdd": "1.2.0"},
	}
	if err := WriteMod(dir, m); err != nil {
		t.Fatalf("WriteMod: %v", err)
	}

	got, err := ReadMod(dir)
	if err != nil {
		t.Fatalf("ReadMod: %v", err)
	}
	if got.SchemaVersion != schema.Mod {
		t.Errorf("SchemaVersion: got %d want %d", got.SchemaVersion, schema.Mod)
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

func TestWriteMod_overwrite(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	m1 := Mod{Skills: map[string]string{"a": "*"}}
	m2 := Mod{Skills: map[string]string{"b": "*"}}

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
	t.Parallel()
	dir := t.TempDir()
	_, err := ReadMod(dir)
	if err == nil {
		t.Fatal("expected error for missing rsk.mod, got nil")
	}
}

func TestWriteReadLock_roundtrip(t *testing.T) {
	t.Parallel()
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
	if got.SchemaVersion != schema.Lock {
		t.Errorf("SchemaVersion: got %d want %d", got.SchemaVersion, schema.Lock)
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
	t.Parallel()
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
	t.Parallel()
	l := Lock{}
	l = UpsertLockEntry(l, LockEntry{Name: "a", Version: "1.0.0"})
	l = UpsertLockEntry(l, LockEntry{Name: "b", Version: "2.0.0"})

	if len(l.Skills) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(l.Skills))
	}

	l = UpsertLockEntry(l, LockEntry{Name: "a", Version: "1.1.0"})
	if len(l.Skills) != 2 {
		t.Fatalf("upsert grew slice to %d", len(l.Skills))
	}
	if l.Skills[0].Version != "1.1.0" {
		t.Errorf("upsert: got version %q want %q", l.Skills[0].Version, "1.1.0")
	}

	l = RemoveLockEntry(l, "a")
	if len(l.Skills) != 1 || l.Skills[0].Name != "b" {
		t.Errorf("after remove: %v", l.Skills)
	}

	l = RemoveLockEntry(l, "nonexistent")
	if len(l.Skills) != 1 {
		t.Errorf("remove non-existent changed len to %d", len(l.Skills))
	}
}
