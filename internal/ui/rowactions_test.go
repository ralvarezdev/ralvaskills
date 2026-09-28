package ui

import (
	"testing"

	"github.com/ralvarezdev/termkit"
	"github.com/spf13/cobra"
)

func TestMarkRowActionAndRowActionsFor(t *testing.T) {
	t.Parallel()

	source := &cobra.Command{Use: "catalog"}
	install := &cobra.Command{Use: "install"}
	MarkRowAction(source, "i", "install", install)

	actions := RowActionsFor(source)
	if len(actions) != 1 || actions[0].Key != "i" || actions[0].Label != "install" {
		t.Fatalf("RowActionsFor(source) = %+v, want one {Key:i Label:install}", actions)
	}

	target, scope, ok := rowActionFor(source, "i")
	if !ok || target != install || scope != nil {
		t.Fatalf("rowActionFor(source, %q) = (%v, %v, %v), want (install, <nil>, true)", "i", target, scope, ok)
	}

	if _, _, found := rowActionFor(source, "u"); found {
		t.Error("rowActionFor with an unregistered key should return ok=false")
	}
}

// TestMarkRowActionScoped checks that the scope resolver registered with
// MarkRowActionScoped is returned by rowActionFor and that RowActionsFor
// still surfaces the keybinding.
func TestMarkRowActionScoped(t *testing.T) {
	t.Parallel()

	source := &cobra.Command{Use: "status"}
	target := &cobra.Command{Use: "uninstall"}
	scope := func(termkit.Data) map[string]string { return map[string]string{"global": "true"} }
	MarkRowActionScoped(source, "u", "uninstall", target, scope)

	got, gotScope, ok := rowActionFor(source, "u")
	if !ok || got != target || gotScope == nil {
		t.Fatalf("rowActionFor(source, %q) = (%v, %v, %v), want (uninstall, <scope>, true)", "u", got, gotScope, ok)
	}
	if v := gotScope(termkit.Data{})["global"]; v != "true" {
		t.Errorf("scope(Data{})[global] = %q, want true", v)
	}

	actions := RowActionsFor(source)
	if len(actions) != 1 || actions[0].Key != "u" || actions[0].Label != "uninstall" {
		t.Fatalf("RowActionsFor(source) = %+v, want one {Key:u Label:uninstall}", actions)
	}
}

func TestRowActionsForUnregisteredCommand(t *testing.T) {
	t.Parallel()

	if actions := RowActionsFor(&cobra.Command{Use: "status"}); actions != nil {
		t.Errorf("RowActionsFor(unregistered) = %+v, want nil", actions)
	}
}

func TestCopySharedFlags(t *testing.T) {
	t.Parallel()

	source := &cobra.Command{Use: "list"}
	source.Flags().Bool("global", false, "")
	source.Flags().String("for", "", "")
	source.Flags().Bool("personal", false, "")
	if err := source.Flags().Set("global", "true"); err != nil {
		t.Fatal(err)
	}
	if err := source.Flags().Set("for", "claude-code"); err != nil {
		t.Fatal(err)
	}

	target := &cobra.Command{Use: "uninstall"}
	target.Flags().Bool("global", false, "")
	target.Flags().String("for", "", "")
	// target has no "personal" flag, so copying it should be a no-op rather
	// than panicking.

	copySharedFlags(source, target, "global", "for", "personal")

	if v, _ := target.Flags().GetBool("global"); !v {
		t.Error("copySharedFlags did not carry over --global")
	}
	if v, _ := target.Flags().GetString("for"); v != "claude-code" {
		t.Errorf("copySharedFlags did not carry over --for, got %q", v)
	}
}

func TestCopySharedFlagsSkipsUnsetFlags(t *testing.T) {
	t.Parallel()

	source := &cobra.Command{Use: "catalog"}
	source.Flags().Bool("personal", false, "")
	// left unset: source.Flags().Changed("personal") is false

	target := &cobra.Command{Use: "install"}
	target.Flags().Bool("personal", true, "")

	copySharedFlags(source, target, "personal")

	if v, _ := target.Flags().GetBool("personal"); !v {
		t.Error("copySharedFlags should leave target's default when source never set the flag")
	}
}
