package cmdx_test

import (
	"testing"

	"github.com/ralvarezdev/ralvaskills/v2/internal/cmdx"
)

func TestParseTargetScope(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    cmdx.TargetScope
		wantErr bool
	}{
		{name: "empty is the unset scope", input: "", want: ""},
		{name: "registered tool", input: "claude-code", want: cmdx.TargetScope("claude-code")},
		{name: "second registered tool", input: "opencode", want: cmdx.TargetScope("opencode")},
		{name: "all sentinel", input: "all", want: cmdx.ScopeAll},
		{name: "unknown tool", input: "bogus", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := cmdx.ParseTargetScope(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseTargetScope(%q): expected error, got %q", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseTargetScope(%q): unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("ParseTargetScope(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
