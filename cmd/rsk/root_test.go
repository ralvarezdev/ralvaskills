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
	session := &cobra.Command{Use: "x"}
	session.SetContext(ui.WithSession(context.Background()))
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
		{"tea.Exec session skips", session, false, false},
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
