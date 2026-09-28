package ui

import (
	"sort"

	"github.com/ralvarezdev/termkit"
	"github.com/spf13/cobra"
)

// rowActionEntry pairs the label shown in a result view's help line with the
// command run for a MarkRowAction's key.
type rowActionEntry struct {
	label string
	cmd   *cobra.Command
}

// rowActions maps a table-view command to the row actions available while
// its captured result is showing, keyed by the key pressed.
var rowActions = map[*cobra.Command]map[string]rowActionEntry{}

// MarkRowAction records that, while source's captured result view is
// showing, pressing key runs target with the selected row's ID as its sole
// argument — the same as picking target from the menu and typing that name
// into its own form. label appears in the result view's help line.
func MarkRowAction(source *cobra.Command, key, label string, target *cobra.Command) {
	if rowActions[source] == nil {
		rowActions[source] = map[string]rowActionEntry{}
	}
	rowActions[source][key] = rowActionEntry{label: label, cmd: target}
}

// RowActionsFor returns the termkit.RowAction hints registered for source
// via MarkRowAction, sorted by key, for building a captured termkit.Data's
// Actions.
func RowActionsFor(source *cobra.Command) []termkit.RowAction {
	entries := rowActions[source]
	if len(entries) == 0 {
		return nil
	}
	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]termkit.RowAction, len(keys))
	for i, k := range keys {
		out[i] = termkit.RowAction{Key: k, Label: entries[k].label}
	}
	return out
}

// rowActionTargetFor resolves the command registered via MarkRowAction for
// source's key, or ok=false if none is registered.
func rowActionTargetFor(source *cobra.Command, key string) (target *cobra.Command, ok bool) {
	e, found := rowActions[source][key]
	return e.cmd, found
}
