package cmd

import (
	"path/filepath"
	"testing"

	"github.com/ralvarezdev/ralvaskills/cli/internal/skill"
)

func TestDedupSkills(t *testing.T) {
	input := []skill.Skill{
		{Name: "a"},
		{Name: "b"},
		{Name: "a"},
		{Name: "c"},
		{Name: "b"},
	}
	got := dedupSkills(input)
	if len(got) != 3 {
		t.Fatalf("expected 3 unique skills, got %d: %v", len(got), got)
	}
	names := map[string]bool{}
	for _, s := range got {
		if names[s.Name] {
			t.Errorf("duplicate skill %q in dedup output", s.Name)
		}
		names[s.Name] = true
	}
}

func TestDetectSkillSource(t *testing.T) {
	repo := filepath.FromSlash("/home/user/ralvaskills")
	official := filepath.FromSlash("/home/user/.ralvaskills/cache/anthropic")
	registry := filepath.FromSlash("/home/user/.ralvaskills/cache/registry")

	tests := []struct {
		name      string
		skillPath string
		want      skill.Source
	}{
		{
			name:      "local repo skill",
			skillPath: filepath.Join(repo, "skills", "go-architect"),
			want:      skill.SourceLocal,
		},
		{
			name:      "official cache skill",
			skillPath: filepath.Join(official, "skills", "claude-code"),
			want:      skill.SourceOfficial,
		},
		{
			name:      "registry cache skill",
			skillPath: filepath.Join(registry, "go-architect", "1.0.0"),
			want:      skill.SourceRegistry,
		},
		{
			name:      "unknown path defaults to local",
			skillPath: filepath.FromSlash("/some/random/path/skill"),
			want:      skill.SourceLocal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := detectSkillSource(tc.skillPath, repo, official, registry)
			if got != tc.want {
				t.Errorf("detectSkillSource(%q) = %v, want %v", tc.skillPath, got, tc.want)
			}
		})
	}
}

func TestJoinKeys(t *testing.T) {
	// joinKeys returns a comma-separated string of map keys.
	// Order is non-deterministic so we test with a single-key map.
	m := map[string]string{"claude-code": "/path"}
	got := joinKeys(m)
	if got != "claude-code" {
		t.Errorf("joinKeys = %q, want %q", got, "claude-code")
	}
}

func TestParseNameVersion(t *testing.T) {
	tests := []struct {
		input       string
		wantName    string
		wantVersion string
		wantErr     bool
	}{
		{"go-architect", "go-architect", "", false},
		{"go-architect@1.0.0", "go-architect", "1.0.0", false},
		{"foo@", "foo", "", false},
		{"@1.0.0", "", "", true},
		{"@", "", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			name, version, err := parseNameVersion(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("parseNameVersion(%q): expected error, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseNameVersion(%q): unexpected error: %v", tc.input, err)
			}
			if name != tc.wantName {
				t.Errorf("name = %q, want %q", name, tc.wantName)
			}
			if version != tc.wantVersion {
				t.Errorf("version = %q, want %q", version, tc.wantVersion)
			}
		})
	}
}
