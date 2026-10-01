package fsx

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/ralvarezdev/ralvaskills/v2/internal/fsperm"
)

const testTempPattern = ".fsxtest-*.tmp"

func writeString(w io.Writer) error {
	_, err := io.WriteString(w, "new content\n")
	return err
}

func TestWriteAtomic_createsWithNewPerm(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "fresh.txt")

	if err := WriteAtomic(path, testTempPattern, writeString); err != nil {
		t.Fatalf("WriteAtomic: %v", err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if got := fi.Mode().Perm(); got != fsperm.File {
		t.Errorf("new file perm = %o, want %o", got, fsperm.File)
	}
}

func TestWriteAtomic_preservesExistingPerm(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "existing.md")

	if err := os.WriteFile(path, []byte("old\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	if err := WriteAtomic(path, testTempPattern, writeString); err != nil {
		t.Fatalf("WriteAtomic: %v", err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	// CreateTemp makes 0600 and rename carries it over, so without an explicit
	// Chmod this silently tightens a world-readable file.
	if got := fi.Mode().Perm(); got != 0o644 {
		t.Errorf("perm after overwrite = %o, want 0640 preserved as 0644", got)
	}
}

func TestWriteAtomicPerm_newPermHonoured(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "secret.txt")

	if err := WriteAtomicPerm(path, testTempPattern, 0o600, writeString); err != nil {
		t.Fatalf("WriteAtomicPerm: %v", err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if got := fi.Mode().Perm(); got != 0o600 {
		t.Errorf("perm = %o, want 0600", got)
	}
}

func TestWriteAtomicPerm_existingWinsOverNewPerm(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "keep.txt")

	if err := os.WriteFile(path, []byte("old\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	// newPerm is 0600 but the file already exists at 0644, so 0644 must win.
	if err := WriteAtomicPerm(path, testTempPattern, 0o600, writeString); err != nil {
		t.Fatalf("WriteAtomicPerm: %v", err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if got := fi.Mode().Perm(); got != 0o644 {
		t.Errorf("perm = %o, want 0644 (existing mode wins)", got)
	}
}

func TestWriteAtomic_writeErrorLeavesNoTempFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "untouched.txt")

	if err := os.WriteFile(path, []byte("original\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	wantErr := os.ErrInvalid
	err := WriteAtomic(path, testTempPattern, func(io.Writer) error { return wantErr })
	if err == nil {
		t.Fatal("expected an error from the write callback")
	}

	// The target must be untouched and no temp file may survive.
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("read target: %v", readErr)
	}
	if string(data) != "original\n" {
		t.Errorf("target content = %q, want unchanged", data)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	if len(entries) != 1 {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("directory has %v, want only the target file", names)
	}
}

func TestWriteAtomic_replacesSymlinkItself(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "target.md")
	link := filepath.Join(dir, "CLAUDE.md")

	if err := os.WriteFile(target, []byte("target\n"), 0o644); err != nil {
		t.Fatalf("seed target: %v", err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}

	if err := WriteAtomic(link, testTempPattern, writeString); err != nil {
		t.Fatalf("WriteAtomic: %v", err)
	}

	// Documented behavior: rename replaces the link. The symlink policy that
	// protects user files lives in the callers (tool.refuseIfLink); this test
	// pins the underlying primitive so the two cannot be confused.
	fi, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		t.Error("link should have been replaced by a regular file")
	}
}
