package ui

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestTakesArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		use  string
		want bool
	}{
		{"no args", "list [flags]", false},
		{"optional args", "install [name...] [flags]", true},
		{"required arg", "pin <name> [flags]", true},
		{"bare name", "status", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := &cobra.Command{Use: tt.use}
			if got := takesArgs(cmd); got != tt.want {
				t.Errorf("takesArgs(%q) = %v, want %v", tt.use, got, tt.want)
			}
		})
	}
}

func TestChildLevel(t *testing.T) {
	t.Parallel()

	visible := &cobra.Command{Use: "list", Short: "list things", Run: func(*cobra.Command, []string) {}}
	hidden := &cobra.Command{Use: "secret", Hidden: true, Run: func(*cobra.Command, []string) {}}
	domain := &cobra.Command{Use: "tools", Short: "manage tools"}
	domain.AddCommand(&cobra.Command{Use: "add", Run: func(*cobra.Command, []string) {}})
	notRunnableLeaf := &cobra.Command{Use: "group"} // no Run, no children: filtered out

	cmds := []*cobra.Command{visible, hidden, domain, notRunnableLeaf}

	root := childLevel(cmds, true)
	if root[0].isBack {
		t.Fatal("root level should not have a back row")
	}
	if len(root) != 2 {
		t.Fatalf("root level = %#v, want 2 items (list, tools)", root)
	}
	if root[0].title != "list" || root[0].cmd != visible {
		t.Errorf("root[0] = %+v, want list", root[0])
	}
	if root[1].title != "tools" || len(root[1].cmd.Commands()) == 0 {
		t.Errorf("root[1] = %+v, want tools domain", root[1])
	}

	nested := childLevel(cmds, false)
	if !nested[0].isBack {
		t.Fatal("non-root level should start with a back row")
	}
	if len(nested) != 3 {
		t.Fatalf("nested level = %#v, want 3 items (back, list, tools)", nested)
	}
}
