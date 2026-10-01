package main

import (
	"testing"

	"github.com/ralvarezdev/ralvaskills/v3/internal/skill"
	"github.com/ralvarezdev/ralvaskills/v3/internal/tool"
)

func TestParseSourceFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    skill.Source
		wantErr bool
	}{
		{name: "empty disables the filter", input: "", want: ""},
		{name: "local", input: "local", want: skill.SourceLocal},
		{name: "official", input: "official", want: skill.SourceOfficial},
		{name: "registry is not selectable", input: "registry", wantErr: true},
		{name: "unknown is rejected", input: "bogus", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseSourceFilter(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseSourceFilter(%q): expected error, got %q", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseSourceFilter(%q): unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("parseSourceFilter(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToolsFromFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    []tool.ID
		wantErr bool
	}{
		{name: "claude-code", input: "claude-code", want: []tool.ID{tool.ClaudeID}},
		{name: "opencode", input: "opencode", want: []tool.ID{tool.OpenCodeID}},
		{name: "all", input: "all", want: []tool.ID{tool.ClaudeID, tool.OpenCodeID}},
		{name: "empty is rejected", input: "", wantErr: true},
		{name: "unknown is rejected", input: "bogus", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := toolsFromFlag(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("toolsFromFlag(%q): expected error, got %v", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("toolsFromFlag(%q): unexpected error: %v", tt.input, err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("toolsFromFlag(%q) = %v, want %v", tt.input, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("toolsFromFlag(%q) = %v, want %v", tt.input, got, tt.want)
				}
			}
		})
	}
}
