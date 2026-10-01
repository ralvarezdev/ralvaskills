package main

import (
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/termkit"

	"github.com/ralvarezdev/ralvaskills/v2/internal/cmdx"
)

// The five tests marked paralleltest below are deliberately sequential. They
// read the package-level cobra command tree, which only looks read-only:
// Commands() sorts c.commands in place, and Find lazily merges and caches
// inherited flag sets (mergePersistentFlags) as it walks. Parallel tests in
// this package race on that shared state — `go test -race -count=2 ./cmd/rsk/...`
// reports a DATA RACE between TestToolsCommandIsVisible and
// TestCatalogBundleSubcommands on unmodified main. A mutex around just the two
// Find call sites only moved the race to Commands(), so the whole group stays
// sequential until the tree is rebuilt per test.
//
//nolint:paralleltest // shared cobra tree mutates on read; see note above
func TestVisibleTopLevelCommandsAreGrouped(t *testing.T) {
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

func TestToolsCommandIsVisible(t *testing.T) {
	t.Parallel()
	for _, sub := range []string{"list", "allow", "deny", "remove"} {
		found, _, err := rootCmd.Find([]string{"tools", sub})
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

//nolint:paralleltest // shared cobra tree mutates on read; see note above
func TestRequireClaudeTarget(t *testing.T) {
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

//nolint:paralleltest // shared cobra tree mutates on read; see note above
func TestDestructiveCommandsHaveYesFlag(t *testing.T) {
	for _, cmd := range []*cobra.Command{destroyCmd, installCmd, uninstallCmd, updateCmd} {
		f := cmd.Flags().Lookup(cmdx.FlagYes)
		if f == nil || f.Shorthand != "y" {
			t.Errorf("%s: missing --yes/-y", cmd.Name())
		}
	}
}

func TestPinIsVisibleWithRemove(t *testing.T) {
	t.Parallel()
	if pinCmd.Hidden || pinCmd.GroupID != groupPinning {
		t.Errorf("pin: hidden=%v group=%q, want visible in %q", pinCmd.Hidden, pinCmd.GroupID, groupPinning)
	}
	if pinCmd.Flags().Lookup(cmdx.FlagRemove) == nil {
		t.Error("pin is missing --remove")
	}
}

//nolint:paralleltest // shared cobra tree mutates on read; see note above
func TestSourceScopeFlagsRegistered(t *testing.T) {
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
	if catalogCmd.Flags().Lookup(cmdx.FlagSource) == nil {
		t.Error("catalog: --source (local|official) must be kept")
	}
}

func TestCatalogBundleSubcommands(t *testing.T) {
	t.Parallel()

	if catalogCmd.Flags().Lookup("bundle") != nil {
		t.Error("catalog: --bundle must be gone (use `catalog bundles` / `catalog bundle <name>`)")
	}

	for _, path := range [][]string{
		{"catalog", "bundles"},
		{"catalog", "bundle", "go-grpc"},
	} {
		found, _, err := rootCmd.Find(path)
		if err != nil || found == nil {
			t.Errorf("rsk %s not reachable: %v", strings.Join(path, " "), err)
			continue
		}
		if !termkit.IsTableView(found) {
			t.Errorf("rsk %s is not marked a table view", strings.Join(path, " "))
		}
	}

	if catalogBundlesCmd.Flags().Lookup(cmdx.FlagOutput) == nil {
		t.Error("catalog bundles: missing --output")
	}
	if catalogBundleCmd.Flags().Lookup(cmdx.FlagSource) == nil {
		t.Error("catalog bundle: missing --source")
	}
	if f := catalogBundleCmd.Flags().Lookup(cmdx.FlagInclude); f == nil || f.Hidden {
		t.Error("catalog bundle: --include missing or hidden")
	}
}

//nolint:paralleltest // shared cobra tree mutates on read; see note above
func TestSessionRunnable(t *testing.T) {
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
