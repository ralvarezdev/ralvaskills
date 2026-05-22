package source

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

// buildTarball builds an in-memory .tar.gz with the given entries.
// Each entry is (name, content); directories end with "/".
func buildTarball(t *testing.T, entries []struct{ name, content string }) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	for _, e := range entries {
		if len(e.name) > 0 && e.name[len(e.name)-1] == '/' {
			tw.WriteHeader(&tar.Header{
				Typeflag: tar.TypeDir,
				Name:     e.name,
				Mode:     0o750,
			})
			continue
		}
		body := []byte(e.content)
		tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     e.name,
			Size:     int64(len(body)),
			Mode:     0o644,
		})
		tw.Write(body)
	}
	tw.Close()
	gw.Close()
	return buf.Bytes()
}

func TestExtractTarball_normal(t *testing.T) {
	dest := t.TempDir()
	data := buildTarball(t, []struct{ name, content string }{
		{"go-architect/", ""},
		{"go-architect/SKILL.md", "version: 1.0.0\n"},
		{"go-architect/prompt.md", "hello\n"},
	})

	if err := extractTarball(bytes.NewReader(data), dest, "go-architect"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dest, "SKILL.md")); err != nil {
		t.Errorf("SKILL.md not extracted: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dest, "prompt.md")); err != nil {
		t.Errorf("prompt.md not extracted: %v", err)
	}
}

func TestExtractTarball_zipSlip(t *testing.T) {
	dest := t.TempDir()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	body := []byte("evil")
	tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     "skill/../../evil.txt",
		Size:     int64(len(body)),
		Mode:     0o644,
	})
	tw.Write(body)
	tw.Close()
	gw.Close()

	err := extractTarball(bytes.NewReader(buf.Bytes()), dest, "skill")
	if err == nil {
		t.Fatal("expected error for zip-slip path, got nil")
	}

	// Confirm the file was NOT written outside dest.
	parent := filepath.Dir(dest)
	if _, statErr := os.Stat(filepath.Join(parent, "evil.txt")); statErr == nil {
		t.Error("zip-slip: evil.txt was written outside destination directory")
	}
}

func TestExtractTarball_rejectsSymlink(t *testing.T) {
	dest := t.TempDir()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeSymlink,
		Name:     "skill/link",
		Linkname: "/etc/passwd",
	})
	tw.Close()
	gw.Close()

	if err := extractTarball(bytes.NewReader(buf.Bytes()), dest, "skill"); err == nil {
		t.Fatal("expected error for symlink entry, got nil")
	}
}

func TestExtractTarball_stripPrefix(t *testing.T) {
	dest := t.TempDir()
	data := buildTarball(t, []struct{ name, content string }{
		{"myskill/SKILL.md", "version: 2.0.0\n"},
	})

	if err := extractTarball(bytes.NewReader(data), dest, "myskill"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dest, "SKILL.md"))
	if err != nil {
		t.Fatalf("SKILL.md not found: %v", err)
	}
	if string(content) != "version: 2.0.0\n" {
		t.Errorf("unexpected content: %q", string(content))
	}
}
