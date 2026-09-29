package ui

import (
	"sort"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ralvarezdev/termkit"
	"github.com/ralvarezdev/termkit/session"

	"github.com/ralvarezdev/ralvaskills/v2/internal/cmdx"
)

// rowActionEntry pairs the label shown in a result view's help line with the
// command run for a MarkRowAction's key.
type rowActionEntry struct {
	label string
	cmd   *cobra.Command
	// scope, when set, returns extra flag values to set on cmd before it
	// runs, given the captured table the action fired on. It lets a view
	// whose tables cover different scopes — status renders one table per
	// target directory — run the target under the scope of the table the
	// selected row came from.
	scope RowActionScope
}

// RowActionScope returns the flag values a row action's target should run
// with, for the captured table the action fired on. A nil return means the
// target runs with only the flags copied from the source view's own run.
type RowActionScope func(table termkit.Data) map[string]string

// rowActions maps a table-view command to the row actions available while
// its captured result is showing, keyed by the key pressed. Registration
// happens during setup and reads during a run, so the map is guarded.
var (
	rowActionsMu sync.RWMutex
	rowActions   = map[*cobra.Command]map[string]rowActionEntry{}
)

// MarkRowAction records that, while source's captured result view is
// showing, pressing key runs target with the selected row's ID as its sole
// argument — the same as picking target from the menu and typing that name
// into its own form. label appears in the result view's help line.
func MarkRowAction(source *cobra.Command, key, label string, target *cobra.Command) {
	MarkRowActionScoped(source, key, label, target, nil)
}

// MarkRowActionScoped is MarkRowAction with a scope resolver: when the
// source view's tables cover more than one scope, scope picks the flag
// values target runs with from the table the action fired on (see
// RowActionScope). A nil scope behaves exactly like MarkRowAction.
func MarkRowActionScoped(
	source *cobra.Command, key, label string, target *cobra.Command, scope RowActionScope,
) {
	rowActionsMu.Lock()
	defer rowActionsMu.Unlock()
	if rowActions[source] == nil {
		rowActions[source] = map[string]rowActionEntry{}
	}
	rowActions[source][key] = rowActionEntry{label: label, cmd: target, scope: scope}
}

// RowActionsFor returns the termkit.RowAction hints registered for source
// via MarkRowAction, sorted by key, for building a captured termkit.Data's
// Actions.
func RowActionsFor(source *cobra.Command) []termkit.RowAction {
	rowActionsMu.RLock()
	defer rowActionsMu.RUnlock()
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

// rowActionFor resolves the command registered via MarkRowAction for
// source's key, along with its scope resolver, or ok=false if none is
// registered.
func rowActionFor(source *cobra.Command, key string) (target *cobra.Command, scope RowActionScope, ok bool) {
	rowActionsMu.RLock()
	defer rowActionsMu.RUnlock()
	e, found := rowActions[source][key]
	return e.cmd, e.scope, found
}

// SessionRowAction is the session's Config.RowAction: it runs the command
// registered for the result's source command and the fired key, passing the
// selected row's ID as its sole positional argument and carrying over the
// scope flags the source command was run with (--global/--for/--personal),
// plus any the fired table's scope adds.
func SessionRowAction(ev session.RowActionEvent) (*cobra.Command, []string, bool) {
	target, scope, ok := rowActionFor(ev.Source, ev.Key)
	if !ok || ev.ID == "" {
		return nil, nil, false
	}

	args := copySharedFlagArgs(ev.Source, target, ev.SourceArgv, cmdx.FlagGlobal, cmdx.FlagFor, cmdx.FlagPersonal)
	args = append(args, scopeFlagArgs(target, scope, ev.Table)...)

	return target, append(args, ev.ID), true
}

// copySharedFlagArgs returns the --flag[=value] arguments for each named flag
// srcArgv set on source and target also defines, so running target for a row
// action carries over the same scope filters source was browsed with.
func copySharedFlagArgs(source, target *cobra.Command, srcArgv []string, names ...string) []string {
	if err := source.Flags().Parse(srcArgv); err != nil {
		return nil
	}

	var args []string
	for _, name := range names {
		srcFlag := source.Flags().Lookup(name)
		tgtFlag := target.Flags().Lookup(name)
		if srcFlag == nil || tgtFlag == nil || !srcFlag.Changed {
			continue
		}
		args = append(args, flagArg(tgtFlag, srcFlag.Value.String()))
	}
	return args
}

// scopeFlagArgs returns the --flag[=value] arguments for the flags scope
// resolves from table, skipping any flag target doesn't define. Names are
// sorted so the argv is deterministic.
func scopeFlagArgs(target *cobra.Command, scope RowActionScope, table termkit.Data) []string {
	if scope == nil {
		return nil
	}

	values := scope(table)
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)

	var args []string
	for _, name := range names {
		flag := target.Flags().Lookup(name)
		if flag == nil {
			continue
		}
		args = append(args, flagArg(flag, values[name]))
	}
	return args
}

// flagArg builds one --name or --name=value argument for flag, using the bare
// boolean form when the flag is a bool.
func flagArg(flag *pflag.Flag, value string) string {
	if flag.Value.Type() == "bool" {
		if value == "true" {
			return "--" + flag.Name
		}
		return "--" + flag.Name + "=false"
	}
	return "--" + flag.Name + "=" + value
}
