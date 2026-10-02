package mcp_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ralvarezdev/ralvaskills/v3/internal/fsperm"
	"github.com/ralvarezdev/ralvaskills/v3/internal/mcp"
)

func TestScanFixtures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		fixture    string
		signals    []string
		candidates []string
	}{
		{
			name:       "go module only",
			fixture:    "gomod-only",
			signals:    []string{"language:go"},
			candidates: []string{"go-architect"},
		},
		{
			name:       "go with grpc",
			fixture:    "go-grpc",
			signals:    []string{"language:go", "protocol:grpc"},
			candidates: []string{"go-architect", "grpc-architect", "protobuf-architect"},
		},
		{
			name:       "nextjs",
			fixture:    "nextjs",
			signals:    []string{"language:typescript", "framework:nextjs"},
			candidates: []string{"nextjs-architect", "react-architect"},
		},
		{
			name:       "fullstack",
			fixture:    "fullstack",
			signals:    []string{"language:go", "infra:docker", "infra:ci"},
			candidates: []string{"go-architect", "docker-architect", "ci-cd-architect"},
		},
		{
			name:       "nested testdata is ignored",
			fixture:    "nested-testdata",
			signals:    []string{"language:go"},
			candidates: []string{"go-architect"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := filepath.Join("testdata", "fixtures", tc.fixture)
			profile, err := mcp.Scan(t.Context(), root)
			if err != nil {
				t.Fatalf("Scan(%s): %v", tc.fixture, err)
			}

			if got := signalKinds(profile.Signals); !slices.Equal(got, tc.signals) {
				t.Errorf("signals = %v, want %v", got, tc.signals)
			}
			if got := candidateSkills(profile.Candidates); !slices.Equal(got, tc.candidates) {
				t.Errorf("candidates = %v, want %v", got, tc.candidates)
			}
			assertProofs(t, profile, root)
			assertBecause(t, profile)
		})
	}
}

func TestScanEmptyProject(t *testing.T) {
	t.Parallel()

	profile, err := mcp.Scan(t.Context(), t.TempDir())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(profile.Signals) != 0 {
		t.Errorf("signals = %v, want none", profile.Signals)
	}
	if len(profile.Candidates) != 0 {
		t.Errorf("candidates = %v, want none", profile.Candidates)
	}
}

func TestScanMissingRoot(t *testing.T) {
	t.Parallel()

	if _, err := mcp.Scan(t.Context(), filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Fatal("Scan of a missing root returned nil error")
	}
}

func TestScanRootIsFile(t *testing.T) {
	t.Parallel()

	file := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(file, []byte("x"), fsperm.File); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := mcp.Scan(t.Context(), file); err == nil {
		t.Fatal("Scan of a file returned nil error")
	}
}

func signalKinds(signals []mcp.Signal) []string {
	kinds := make([]string, 0, len(signals))
	for _, s := range signals {
		kinds = append(kinds, s.Kind.String())
	}
	return kinds
}

func candidateSkills(candidates []mcp.Candidate) []string {
	skills := make([]string, 0, len(candidates))
	for _, c := range candidates {
		skills = append(skills, c.Skill)
	}
	return skills
}

// assertProofs checks that every signal cites an existing file relative to root.
func assertProofs(t *testing.T, profile mcp.Profile, root string) {
	t.Helper()

	for _, s := range profile.Signals {
		if s.Proof == "" {
			t.Errorf("signal %s has an empty proof", s.Kind)
			continue
		}
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(s.Proof))); err != nil {
			t.Errorf("signal %s proof %q does not exist: %v", s.Kind, s.Proof, err)
		}
	}
}

// assertBecause checks that every candidate names the signals that proposed it.
func assertBecause(t *testing.T, profile mcp.Profile) {
	t.Helper()

	kinds := make(map[mcp.SignalKind]struct{}, len(profile.Signals))
	for _, s := range profile.Signals {
		kinds[s.Kind] = struct{}{}
	}
	for _, c := range profile.Candidates {
		if len(c.Because) == 0 {
			t.Errorf("candidate %s has no because", c.Skill)
		}
		for _, kind := range c.Because {
			if _, ok := kinds[kind]; !ok {
				t.Errorf("candidate %s cites %s, not a detected signal", c.Skill, kind)
			}
		}
	}
}
