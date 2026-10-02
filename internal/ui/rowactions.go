package ui

import (
	"github.com/spf13/cobra"

	"github.com/ralvarezdev/termkit"
	"github.com/ralvarezdev/termkit/session"

	"github.com/ralvarezdev/ralvaskills/v3/internal/cmdx"
)

// rowActions is the registry of row actions each table view offers. The
// registry itself lives in termkit/session; this is the process-wide instance
// the command tree registers into and the session resolves against.
var rowActions = session.NewRowActions()

// defaultScopeFlags are the flags a row action always carries over from the
// view it fired on, so an action on a --global/--for/--personal view runs
// scoped the same way.
var defaultScopeFlags = []string{cmdx.FlagGlobal, cmdx.FlagFor, cmdx.FlagPersonal}

// MarkRowAction records that, while source's captured result view is showing,
// pressing key runs target with the selected row's ID as its sole argument —
// the same as picking target from the menu and typing that name into its own
// form. label appears in the result view's help line.
func MarkRowAction(source *cobra.Command, key, label string, target *cobra.Command) {
	mark(source, key, label, target, "")
}

// MarkRowActionScoped is MarkRowAction with a scope resolver: when the source
// view's tables cover more than one scope, scope picks the flag values target
// runs with from the table the action fired on. A nil scope behaves exactly
// like MarkRowAction.
func MarkRowActionScoped(
	source *cobra.Command, key, label string, target *cobra.Command, scope session.Scope,
) {
	rowActions.MarkScoped(source, key, label, target, scope)
	rowActions.Inherit(source, key, defaultScopeFlags...)
}

// MarkRowActionArgs records a row action whose argv is built from the row (see
// session.ArgsFunc), so the target's form opens prefilled from the row's
// values rather than only the row's ID. Like MarkRowAction, it carries over
// the default scope flags.
func MarkRowActionArgs(
	source *cobra.Command, key, label string, target *cobra.Command, fn session.ArgsFunc,
) {
	rowActions.MarkArgs(source, key, label, target, fn)
	rowActions.Inherit(source, key, defaultScopeFlags...)
}

// RowActionsFor returns the termkit.RowAction hints registered for source,
// sorted by key, for building a captured termkit.Data's Actions.
func RowActionsFor(source *cobra.Command) []termkit.RowAction {
	return rowActions.For(source)
}

// SessionRowAction is the session's Config.RowAction: it runs the command
// registered for the result's source command and the fired key, passing the
// selected row's ID as its sole positional argument and carrying over the scope
// flags the source command was run with, plus any the fired table's scope adds.
func SessionRowAction(ev session.RowActionEvent) (*cobra.Command, []string, bool) {
	return rowActions.Resolve(ev)
}

// mark registers an action (positionally, or via flag when non-empty) and
// carries over the default scope flags.
func mark(source *cobra.Command, key, label string, target *cobra.Command, flag string) {
	rowActions.MarkFlag(source, key, label, target, flag)
	rowActions.Inherit(source, key, defaultScopeFlags...)
}
