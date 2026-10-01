package tool

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeSymlink creates target with content and a symlink at link pointing to it.
func writeSymlink(t *testing.T, target, link, content string) {
	t.Helper()
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		t.Fatalf("seed target: %v", err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}
}

func TestAppendClaudeImport_refusesSymlink(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "AGENTS.md")
	link := filepath.Join(dir, "CLAUDE.md")
	writeSymlink(t, target, link, "my project notes\n")

	err := appendClaudeImport(link)
	if err == nil {
		t.Fatal("expected an error for a symlinked CLAUDE.md")
	}
	if !strings.Contains(err.Error(), "symlink") {
		t.Errorf("error should mention the symlink, got: %v", err)
	}

	// The whole point: the target must be byte-for-byte untouched.
	data, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("read target: %v", readErr)
	}
	if string(data) != "my project notes\n" {
		t.Errorf("target content = %q, want it unmodified", data)
	}

	// And the link must still be a link, not silently replaced.
	fi, lstatErr := os.Lstat(link)
	if lstatErr != nil {
		t.Fatalf("lstat: %v", lstatErr)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Error("symlink was replaced by a regular file")
	}
}

func TestAppendClaudeImport_refusesLinkToOwnGeneratedFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rskDir := filepath.Join(dir, ".rsk")
	if err := os.MkdirAll(rskDir, 0o750); err != nil {
		t.Fatalf("mkdir .rsk: %v", err)
	}

	// The self-import loop: CLAUDE.md → .rsk/CLAUDE.md, which is the file rsk
	// generates. Appending here makes the generated file import itself.
	generated := filepath.Join(rskDir, ClaudeFileName)
	link := filepath.Join(dir, ClaudeFileName)
	writeSymlink(t, generated, link, "@../.claude/skills/x/SKILL.md\n")

	if err := appendClaudeImport(link); err == nil {
		t.Fatal("expected an error when CLAUDE.md points at .rsk/CLAUDE.md")
	}

	data, err := os.ReadFile(generated)
	if err != nil {
		t.Fatalf("read generated: %v", err)
	}
	if strings.Contains(string(data), claudeImportLine) {
		t.Errorf("generated file imports itself: %q", data)
	}
}

func TestRemoveClaudeImport_refusesSymlink(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "AGENTS.md")
	link := filepath.Join(dir, "CLAUDE.md")
	writeSymlink(t, target, link, "notes\n"+claudeImportLine+"\n")

	if err := removeClaudeImport(link); err == nil {
		t.Fatal("expected an error for a symlinked CLAUDE.md")
	}

	// The rename path would otherwise replace the link with a regular file.
	fi, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Error("symlink was replaced by a regular file")
	}

	data, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("read target: %v", readErr)
	}
	if !strings.Contains(string(data), claudeImportLine) {
		t.Errorf("target was modified: %q", data)
	}
}

func TestWritePinnedClaude_refusesSymlink(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rskDir := filepath.Join(dir, ".rsk")
	if err := os.MkdirAll(rskDir, 0o750); err != nil {
		t.Fatalf("mkdir .rsk: %v", err)
	}

	// .rsk/CLAUDE.md itself symlinked away — the rename would replace the link.
	elsewhere := filepath.Join(dir, "elsewhere.md")
	writeSymlink(t, elsewhere, filepath.Join(rskDir, ClaudeFileName), "old\n")

	if err := writePinnedClaude(rskDir, []string{"go-architect"}); err == nil {
		t.Fatal("expected an error when .rsk/CLAUDE.md is a symlink")
	}

	data, err := os.ReadFile(elsewhere)
	if err != nil {
		t.Fatalf("read elsewhere: %v", err)
	}
	if string(data) != "old\n" {
		t.Errorf("target modified: %q", data)
	}
}

func TestRefuseIfLink_allowsRegularAndMissing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	regular := filepath.Join(dir, "regular.md")
	if err := os.WriteFile(regular, []byte("hi\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := refuseIfLink(regular); err != nil {
		t.Errorf("regular file should be allowed, got: %v", err)
	}

	// A nonexistent path is what `rsk new` hits on a fresh project.
	if err := refuseIfLink(filepath.Join(dir, "missing.md")); err != nil {
		t.Errorf("missing file should be allowed, got: %v", err)
	}
}

func TestRefuseIfLink_refusesDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sub := filepath.Join(dir, "subdir")
	if err := os.Mkdir(sub, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := refuseIfLink(sub); err == nil {
		t.Error("expected an error for a directory")
	}
}

func TestSyncPinned_regularFileEndToEnd(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".rsk"), 0o750); err != nil {
		t.Fatalf("mkdir .rsk: %v", err)
	}
	claudeMD := filepath.Join(dir, ClaudeFileName)
	if err := os.WriteFile(claudeMD, []byte("# notes\n"), 0o644); err != nil {
		t.Fatalf("seed CLAUDE.md: %v", err)
	}
	if err := os.Chmod(claudeMD, 0o644); err != nil {
		t.Fatalf("chmod CLAUDE.md: %v", err)
	}

	ct := &ClaudeTool{}
	if err := ct.SyncPinned(dir, []string{"go-architect"}); err != nil {
		t.Fatalf("SyncPinned: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".rsk", ClaudeFileName))
	if err != nil {
		t.Fatalf("read pinned: %v", err)
	}
	if !strings.Contains(string(data), "go-architect") {
		t.Errorf("pinned file missing skill: %q", data)
	}

	root, err := os.ReadFile(filepath.Join(dir, ClaudeFileName))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	if !strings.Contains(string(root), claudeImportLine) {
		t.Errorf("CLAUDE.md missing import line: %q", root)
	}

	// Idempotent: a second sync must not append a duplicate.
	if err = ct.SyncPinned(dir, []string{"go-architect"}); err != nil {
		t.Fatalf("second SyncPinned: %v", err)
	}
	root2, err := os.ReadFile(filepath.Join(dir, ClaudeFileName))
	if err != nil {
		t.Fatalf("reread CLAUDE.md: %v", err)
	}
	if strings.Count(string(root2), claudeImportLine) != 1 {
		t.Errorf("import line duplicated: %q", root2)
	}

	// R3: the two-step sync (writePinnedClaude then appendClaudeImport) must
	// not leave CLAUDE.md tighter than it started.
	fi, err := os.Stat(claudeMD)
	if err != nil {
		t.Fatalf("stat CLAUDE.md: %v", err)
	}
	if got := fi.Mode().Perm(); got != 0o644 {
		t.Errorf("CLAUDE.md perm = %o, want 0644 preserved", got)
	}
}
