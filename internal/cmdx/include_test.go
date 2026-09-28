package cmdx_test

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/ralvaskills/v2/internal/cmdx"
)

func newIncludeCmd(allowed ...string) *cobra.Command {
	cmd := &cobra.Command{Use: "x"}
	cmdx.RegisterInclude(cmd, "usage", allowed...)
	return cmd
}

func TestReadIncludes(t *testing.T) {
	t.Parallel()

	both := []string{cmdx.IncludePersonal, cmdx.IncludeOfficial}
	personalOnly := []string{cmdx.IncludePersonal}
	tests := []struct {
		name    string
		allowed []string
		args    []string
		want    cmdx.Includes
		wantErr string
	}{
		{"none", both, nil, cmdx.Includes{}, ""},
		{"include personal", both, []string{"--include", "personal"}, cmdx.Includes{Personal: true}, ""},
		{
			"include csv",
			both,
			[]string{"--include", "personal,official"},
			cmdx.Includes{Personal: true, Official: true},
			"",
		},
		{
			"include repeated",
			both,
			[]string{"--include", "personal", "--include", "official"},
			cmdx.Includes{Personal: true, Official: true},
			"",
		},
		{"legacy personal", personalOnly, []string{"--personal"}, cmdx.Includes{Personal: true}, ""},
		{"legacy official", both, []string{"--official"}, cmdx.Includes{Official: true}, ""},
		{"legacy both", both, []string{"--personal", "--official"}, cmdx.Includes{Personal: true, Official: true}, ""},
		{"bad value", both, []string{"--include", "bogus"}, cmdx.Includes{}, `invalid value "bogus"`},
		{"official not allowed", personalOnly, []string{"--include", "official"}, cmdx.Includes{}, "allowed: personal"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := newIncludeCmd(tt.allowed...)
			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Fatal(err)
			}
			got, err := cmdx.ReadIncludes(cmd, tt.allowed...)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("got %+v, %v; want %+v", got, err, tt.want)
			}
		})
	}
}

func TestRegisterIncludeHidesLegacyAliases(t *testing.T) {
	t.Parallel()

	cmd := newIncludeCmd(cmdx.IncludePersonal, cmdx.IncludeOfficial)
	if cmd.Flags().Lookup(cmdx.FlagInclude).Hidden {
		t.Error("--include must be visible")
	}
	for _, name := range []string{cmdx.FlagPersonal, cmdx.FlagOfficial} {
		if f := cmd.Flags().Lookup(name); f == nil || !f.Hidden {
			t.Errorf("--%s must exist and be hidden", name)
		}
	}
	if newIncludeCmd(cmdx.IncludePersonal).Flags().Lookup(cmdx.FlagOfficial) != nil {
		t.Error("--official must not exist where official is not allowed")
	}
}
