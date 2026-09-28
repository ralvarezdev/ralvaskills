package main

import (
	"slices"
	"testing"
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
