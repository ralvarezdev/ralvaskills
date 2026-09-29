package ui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/termkit"
	"github.com/ralvarezdev/termkit/session"

	"github.com/ralvarezdev/ralvaskills/v2/internal/cmdx"
)

// addScopeFlags gives cmd the --global/--for/--personal flags a scope command
// carries.
func addScopeFlags(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().Bool(cmdx.FlagGlobal, false, "global")
	cmd.Flags().String(cmdx.FlagFor, "", "tool")
	cmd.Flags().Bool(cmdx.FlagPersonal, false, "personal")

	return cmd
}

// TestSessionRowActionCarriesScopeAndRow pins the resolver's argv: the row ID
// becomes the target's positional argument, and the scope flags the source run
// was browsed with (--global/--for) carry over.
func TestSessionRowActionCarriesScopeAndRow(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "rsk"}
	source := addScopeFlags(&cobra.Command{Use: "list", RunE: func(*cobra.Command, []string) error { return nil }})
	target := addScopeFlags(&cobra.Command{Use: "uninstall", RunE: func(*cobra.Command, []string) error { return nil }})
	root.AddCommand(source, target)

	MarkRowAction(source, "u", "uninstall", target)

	targetCmd, args, ok := SessionRowAction(session.RowActionEvent{
		Source:     source,
		SourceArgv: []string{"rsk", "list", "--global", "--for", "claude-code"},
		ID:         "go-architect",
		Key:        "u",
	})
	if !ok || targetCmd != target {
		t.Fatalf("resolved = %v, ok=%v; want the registered uninstall target", targetCmd, ok)
	}
	want := []string{"--global", "--for=claude-code", "go-architect"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %v, want %v", args, want)
	}

	// An unknown key or an empty row ID resolves nothing.
	if _, _, resolved := SessionRowAction(session.RowActionEvent{Source: source, ID: "x", Key: "z"}); resolved {
		t.Fatal("an unregistered key must not resolve")
	}
	if _, _, resolved := SessionRowAction(session.RowActionEvent{Source: source, ID: "", Key: "u"}); resolved {
		t.Fatal("an empty row ID must not resolve")
	}
}

// TestSessionRowActionAppliesTableScope pins the scoped variant: the fired
// table's title resolves --global/--for even when the source run carried no
// scope flags.
func TestSessionRowActionAppliesTableScope(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "rsk"}
	source := addScopeFlags(&cobra.Command{Use: "status", RunE: func(*cobra.Command, []string) error { return nil }})
	target := addScopeFlags(&cobra.Command{Use: "uninstall", RunE: func(*cobra.Command, []string) error { return nil }})
	root.AddCommand(source, target)

	MarkRowActionScoped(source, "u", "uninstall", target, func(table termkit.Data) map[string]string {
		if tool, ok := strings.CutPrefix(table.Title, "Global — "); ok {
			return map[string]string{cmdx.FlagGlobal: "true", cmdx.FlagFor: tool}
		}

		return nil
	})

	_, args, ok := SessionRowAction(session.RowActionEvent{
		Source: source,
		Table:  termkit.Data{Title: "Global — opencode"},
		ID:     "fastapi-architect",
		Key:    "u",
	})
	if !ok {
		t.Fatal("scoped action did not resolve")
	}
	if want := []string{"--for=opencode", "--global", "fastapi-architect"}; !reflect.DeepEqual(args, want) {
		t.Fatalf("scoped args = %v, want %v", args, want)
	}
}
