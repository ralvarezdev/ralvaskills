package skill

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ralvarezdev/ralvaskills/v3/internal/fsperm"
)

func TestWalkReadsDescription(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dir := filepath.Join(root, "go-architect")
	if err := os.MkdirAll(dir, fsperm.Dir); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := "---\nname: go-architect\nversion: 1.2.3\ndescription: Go 1.26 standards.\n---\n\n# body\n"
	if err := os.WriteFile(filepath.Join(dir, SkillFileName), []byte(body), fsperm.File); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	skills, err := Walk(root, SourceLocal)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("got %d skills, want 1", len(skills))
	}
	if skills[0].Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", skills[0].Version, "1.2.3")
	}
	if skills[0].Description != "Go 1.26 standards." {
		t.Errorf("Description = %q, want %q", skills[0].Description, "Go 1.26 standards.")
	}
}

func TestReadFrontmatterQuotedDescription(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	body := "---\nversion: 0.1.0\ndescription: \"Quoted: with colon.\"\n---\n"
	if err := os.WriteFile(filepath.Join(root, SkillFileName), []byte(body), fsperm.File); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	fm, err := ReadFrontmatter(root)
	if err != nil {
		t.Fatalf("ReadFrontmatter: %v", err)
	}
	if fm.Description != "Quoted: with colon." {
		t.Errorf("Description = %q, want %q", fm.Description, "Quoted: with colon.")
	}
}
