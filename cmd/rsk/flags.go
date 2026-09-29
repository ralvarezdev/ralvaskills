package main

import (
	"github.com/spf13/cobra"

	"github.com/ralvarezdev/ralvaskills/v2/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/v2/internal/skill"
	"github.com/ralvarezdev/ralvaskills/v2/internal/tool"
)

// forToolFlag reads --for as a registered tool ID. Empty means the flag was
// not given; an unknown tool is reported against the flag so the interactive
// session can show it on the field. Commands that also accept the "all" scope
// parse with cmdx.ParseTargetScope instead.
func forToolFlag(cmd *cobra.Command) (tool.ID, error) {
	raw := cmdx.String(cmd, cmdx.FlagFor)
	if raw == "" {
		return "", nil
	}
	id, err := tool.ParseID(raw)
	if err != nil {
		return "", fieldError(cmdx.FlagFor, "%v", err)
	}
	return id, nil
}

// parseSourceFilter parses --source: an empty value disables the source filter,
// otherwise the value must name a source the catalog can walk (local or
// official). The registry source is deliberately not selectable here.
func parseSourceFilter(raw string) (skill.Source, error) {
	switch skill.Source(raw) {
	case "", skill.SourceLocal, skill.SourceOfficial:
		return skill.Source(raw), nil
	default:
		return "", fieldError(
			cmdx.FlagSource,
			"--source must be '%s' or '%s'",
			skill.SourceLocal,
			skill.SourceOfficial,
		)
	}
}

// registerForCompletion wires --for's shell completion from the registered
// tools, optionally including the "all" scope for commands that accept it.
func registerForCompletion(cmd *cobra.Command, includeAll bool) {
	values := tool.Names()
	if includeAll {
		values = append(values, cmdx.ForAll)
	}
	err := cmd.RegisterFlagCompletionFunc(cmdx.FlagFor,
		func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
			return values, cobra.ShellCompDirectiveNoFileComp
		})
	if err != nil {
		panic("register --for completion on " + cmd.Name() + ": " + err.Error())
	}
}
