package main

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/termkit"

	"github.com/ralvarezdev/ralvaskills/v2/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/v2/internal/ui"
)

func TestVisibleTopLevelCommandsAreGrouped(t *testing.T) {
	t.Parallel()

	declared := make([]string, 0, len(rootGroups))
	for _, group := range rootCmd.Groups() {
		declared = append(declared, group.ID)
	}

	for _, cmd := range rootCmd.Commands() {
		if cmd.Hidden || cmd.Name() == "help" || cmd.Name() == "completion" {
			continue
		}
		if !slices.Contains(declared, cmd.GroupID) {
			t.Errorf("command %q has GroupID %q, want one of %v", cmd.Name(), cmd.GroupID, declared)
		}
	}
}

func TestLegacyClaudeToolsPathIsHiddenButReachable(t *testing.T) {
	t.Parallel()

	if !claudeCmd.Hidden {
		t.Error("legacy claude command must be hidden")
	}
	for _, sub := range []string{"list", "allow", "deny", "remove"} {
		found, _, err := rootCmd.Find([]string{"claude", "tools", sub})
		if err != nil || found.Name() != sub || !found.HasParent() || found.Parent().Name() != "tools" {
			t.Errorf("rsk claude tools %s not reachable: %v", sub, err)
		}
		found, _, err = rootCmd.Find([]string{"tools", sub})
		if err != nil || found.Name() != sub {
			t.Errorf("rsk tools %s not reachable: %v", sub, err)
		}
	}
	if canonicalTools.root.Hidden || canonicalTools.root.GroupID != groupToolsConf {
		t.Errorf(
			"tools: hidden=%v group=%q, want visible in %q",
			canonicalTools.root.Hidden,
			canonicalTools.root.GroupID,
			groupToolsConf,
		)
	}
}

func TestRequireClaudeTarget(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"":            "",
		"claude-code": "",
		"opencode":    "not supported yet",
		"all":         "not supported yet",
		"bogus":       "unknown tool",
	}
	for value, wantErr := range cases {
		cmd := newToolsCmds("rsk tools").list
		if err := cmd.Flags().Set("for", value); err != nil {
			t.Fatal(err)
		}
		err := requireClaudeTarget(cmd)
		switch {
		case wantErr == "" && err != nil:
			t.Errorf("--for %q: unexpected error %v", value, err)
		case wantErr != "" && (err == nil || !strings.Contains(err.Error(), wantErr)):
			t.Errorf("--for %q: error %v, want containing %q", value, err, wantErr)
		}
	}
}

func TestShouldConfirm(t *testing.T) {
	t.Parallel()

	plain := &cobra.Command{Use: "x"}
	plain.SetContext(context.Background())
	captured := &cobra.Command{Use: "x"}
	captured.SetContext(ui.WithCapture(context.Background(), &termkit.Capture{}))

	cases := []struct {
		name string
		cmd  *cobra.Command
		yes  bool
		want bool
	}{
		{"prompts by default", plain, false, true},
		{"--yes skips", plain, true, false},
		{"captured session skips", captured, false, false},
	}
	for _, tc := range cases {
		if got := shouldConfirm(tc.cmd, tc.yes); got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestDestructiveCommandsHaveYesFlag(t *testing.T) {
	t.Parallel()

	for _, cmd := range []*cobra.Command{destroyCmd, uninstallCmd} {
		f := cmd.Flags().Lookup(cmdx.FlagYes)
		if f == nil || f.Shorthand != "y" {
			t.Errorf("%s: missing --yes/-y", cmd.Name())
		}
	}
}

func TestLegacyUnpinIsHiddenButReachable(t *testing.T) {
	t.Parallel()

	found, _, err := rootCmd.Find([]string{"unpin"})
	if err != nil || found != unpinCmd || !unpinCmd.Hidden {
		t.Fatalf("unpin: found=%v hidden=%v err=%v, want hidden and reachable", found, unpinCmd.Hidden, err)
	}
	if pinCmd.Hidden || pinCmd.GroupID != groupPinning {
		t.Errorf("pin: hidden=%v group=%q, want visible in %q", pinCmd.Hidden, pinCmd.GroupID, groupPinning)
	}
	if pinCmd.Flags().Lookup(cmdx.FlagRemove) == nil {
		t.Error("pin is missing --remove")
	}
}

func TestSourceScopeFlagsRegistered(t *testing.T) {
	t.Parallel()

	for _, cmd := range []*cobra.Command{installCmd, uninstallCmd, updateCmd, listCmd, statusCmd, catalogCmd} {
		if f := cmd.Flags().Lookup(cmdx.FlagInclude); f == nil || f.Hidden {
			t.Errorf("%s: --include missing or hidden", cmd.Name())
		}
		if f := cmd.Flags().Lookup(cmdx.FlagPersonal); f == nil || !f.Hidden {
			t.Errorf("%s: --personal must be a hidden alias", cmd.Name())
		}
	}
	if f := updateCmd.Flags().Lookup(cmdx.FlagOfficial); f == nil || !f.Hidden {
		t.Error("update: --official must be a hidden alias")
	}
	if f := catalogCmd.Flags().Lookup(cmdx.FlagBundles); f == nil || !f.Hidden {
		t.Error("catalog: --bundles must be a hidden alias")
	}
	if f := catalogCmd.Flags().Lookup(cmdx.FlagBundle); f == nil || f.NoOptDefVal != bundleListSentinel {
		t.Error("catalog: --bundle must be an optional-value flag")
	}
	if catalogCmd.Flags().Lookup(cmdx.FlagSource) == nil {
		t.Error("catalog: --source (local|official) must be kept")
	}
}

func TestResolveBundleFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		flag      string
		alias     bool
		args      []string
		wantList  bool
		wantName  string
		wantError bool
	}{
		{"unset", "", false, nil, false, "", false},
		{"bare --bundle lists bundles", bundleListSentinel, false, nil, true, "", false},
		{"--bundle=NAME", "go-grpc", false, nil, false, "go-grpc", false},
		{"--bundle NAME (positional)", bundleListSentinel, false, []string{"go-grpc"}, false, "go-grpc", false},
		{"--bundles alias", "", true, nil, true, "", false},
		{"--bundles with name keeps conflict", "go-grpc", true, nil, true, "go-grpc", false},
		{"stray positional", "", false, []string{"x"}, false, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			list, name, err := resolveBundleFlag(tt.flag, tt.alias, tt.args)
			if (err != nil) != tt.wantError || list != tt.wantList || name != tt.wantName {
				t.Errorf(
					"got (%v, %q, %v), want (%v, %q, err=%v)",
					list,
					name,
					err,
					tt.wantList,
					tt.wantName,
					tt.wantError,
				)
			}
		})
	}
}

func TestSessionRunnable(t *testing.T) {
	t.Parallel()

	cases := []struct {
		cmd  *cobra.Command
		name string
		want bool
	}{
		{installCmd, "install", true},
		{listCmd, "list", true},
		{canonicalTools.allow, "tools allow", true},
		{initCmd, "init (interactive wizard)", false},
	}
	for _, tc := range cases {
		if got := sessionRunnable(tc.cmd); got != tc.want {
			t.Errorf("%s: sessionRunnable = %v, want %v", tc.name, got, tc.want)
		}
	}

	for _, name := range []string{"help", "completion"} {
		builtin := &cobra.Command{Use: name}
		leaf := &cobra.Command{Use: "bash"}
		builtin.AddCommand(leaf)
		(&cobra.Command{Use: "rsk"}).AddCommand(builtin)
		if sessionRunnable(leaf) {
			t.Errorf("%s subcommands must not be offered", name)
		}
	}
}
