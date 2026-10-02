package install_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ralvarezdev/ralvaskills/v3/internal/fsperm"
	"github.com/ralvarezdev/ralvaskills/v3/internal/install"
	"github.com/ralvarezdev/ralvaskills/v3/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/v3/internal/skill"
	"github.com/ralvarezdev/ralvaskills/v3/internal/tool"
)

func TestApplyProjectWritesManifestLockAndPin(t *testing.T) {
	t.Parallel()

	project := t.TempDir()
	rskDir := filepath.Join(project, ".rsk")
	if err := manifest.WriteMod(rskDir, manifest.Mod{Tools: []tool.ID{tool.ClaudeID}}); err != nil {
		t.Fatalf("seed manifest: %v", err)
	}

	src := filepath.Join(t.TempDir(), "go-architect")
	if err := os.MkdirAll(src, fsperm.Dir); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	target := filepath.Join(project, ".claude", "skills")

	results, err := install.Apply([]skill.Skill{{
		Name:    "go-architect",
		Version: "1.2.3",
		Path:    src,
		Source:  skill.SourceLocal,
	}}, install.Options{
		Scope:       install.ScopeProject,
		Targets:     []string{target},
		RskDir:      rskDir,
		Pin:         true,
		Constraints: map[string]string{"go-architect": "1.2.3"},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if len(results) != 1 || results[0].Err != nil {
		t.Fatalf("results = %+v", results)
	}
	if _, linkErr := os.Lstat(filepath.Join(target, "go-architect")); linkErr != nil {
		t.Errorf("link missing: %v", linkErr)
	}

	m, err := manifest.ReadMod(rskDir)
	if err != nil {
		t.Fatalf("ReadMod: %v", err)
	}
	if m.Skills["go-architect"] != "1.2.3" {
		t.Errorf("skills = %+v, want go-architect pinned to 1.2.3", m.Skills)
	}
	if !slices.Contains(m.Pinned, "go-architect") {
		t.Errorf("pinned = %+v, want go-architect", m.Pinned)
	}

	lock, err := manifest.ReadLock(rskDir)
	if err != nil {
		t.Fatalf("ReadLock: %v", err)
	}
	if len(lock.Skills) != 1 || lock.Skills[0].Version != "1.2.3" {
		t.Errorf("lock = %+v, want one go-architect entry at 1.2.3", lock.Skills)
	}
}

func TestApplyGlobalWritesNoManifest(t *testing.T) {
	t.Parallel()

	project := t.TempDir()
	target := filepath.Join(project, "global-skills")
	src := filepath.Join(t.TempDir(), "tdd")
	if err := os.MkdirAll(src, fsperm.Dir); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}

	results, err := install.Apply([]skill.Skill{{
		Name:    "tdd",
		Version: "1.0.0",
		Path:    src,
	}}, install.Options{
		Scope:   install.ScopeGlobal,
		Targets: []string{target},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if results[0].Err != nil {
		t.Fatalf("results = %+v", results)
	}
	if _, linkErr := os.Lstat(filepath.Join(target, "tdd")); linkErr != nil {
		t.Errorf("link missing: %v", linkErr)
	}
	if _, statErr := os.Stat(filepath.Join(project, ".rsk")); !os.IsNotExist(statErr) {
		t.Error("global apply wrote a project manifest")
	}
}
