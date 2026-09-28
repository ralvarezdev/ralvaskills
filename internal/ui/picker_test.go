package ui

import (
	"slices"
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

func leafCmd(name string) *cobra.Command {
	return &cobra.Command{Use: name, Run: func(*cobra.Command, []string) {}}
}

func TestChildLevelHidesCobraBuiltinsAtRoot(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "rsk"}
	root.AddCommand(leafCmd("list"), leafCmd("help"), leafCmd("completion"))

	items := childLevel(root.Commands(), true)
	if len(items) != 1 || items[0].title != "list" {
		t.Fatalf("root level = %+v, want only list", items)
	}
}

func TestChildLevelOrdersByGroup(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "rsk"}
	root.AddGroup(&cobra.Group{ID: "first", Title: "First:"}, &cobra.Group{ID: "second", Title: "Second:"})
	add := func(name, group string) {
		c := leafCmd(name)
		c.GroupID = group
		root.AddCommand(c)
	}
	add("alpha", "second")
	add("bravo", "")
	add("charlie", "first")
	add("delta", "second")
	add("echo", "first")

	got := make([]string, 0, 5)
	for _, item := range childLevel(root.Commands(), true) {
		got = append(got, item.title)
	}
	want := []string{"charlie", "echo", "alpha", "delta", "bravo"}
	if !slices.Equal(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}
}

func groupTree() (root, claude *cobra.Command) {
	root = &cobra.Command{Use: "rsk"}
	claude = &cobra.Command{Use: "claude"}
	tools := &cobra.Command{Use: "tools"}
	tools.AddCommand(leafCmd("allow"), leafCmd("deny"), leafCmd("list"), leafCmd("remove"))
	claude.AddCommand(tools)
	root.AddCommand(leafCmd("aaa"), claude)
	return root, claude
}

func TestDescendCollapsesSingleChildGroup(t *testing.T) {
	t.Parallel()

	root, claude := groupTree()
	m := newPickerModel(root.Commands(), "")
	m.list.Select(1) // claude
	m.choose(pickItem{title: "claude", cmd: claude})

	if len(m.stack) != 2 {
		t.Fatalf("stack = %d, want 2 (root + collapsed claude/tools)", len(m.stack))
	}
	if got, want := m.list.Title, "rsk — claude › tools"; got != want {
		t.Errorf("title = %q, want %q", got, want)
	}
	if got := len(m.current().items); got != 5 { // back + 4 leaves
		t.Errorf("collapsed level has %d rows, want 5", got)
	}

	m.pop()
	if len(m.stack) != 1 || m.list.Index() != 1 {
		t.Errorf("after pop: stack=%d index=%d, want 1 and 1", len(m.stack), m.list.Index())
	}
}

func TestDescendDoesNotCollapseMultiChildGroup(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "rsk"}
	grp := &cobra.Command{Use: "grp"}
	sub := &cobra.Command{Use: "sub"}
	sub.AddCommand(leafCmd("x"))
	grp.AddCommand(sub, leafCmd("y"))
	root.AddCommand(grp)

	m := newPickerModel(root.Commands(), "")
	m.choose(pickItem{title: "grp", cmd: grp})
	if got, want := m.list.Title, "rsk — grp"; got != want {
		t.Errorf("title = %q, want %q", got, want)
	}
}

func TestPopRestoresCursor(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "rsk"}
	grp := &cobra.Command{Use: "grp"}
	grp.AddCommand(leafCmd("x"), leafCmd("y"))
	root.AddCommand(leafCmd("a"), leafCmd("b"), grp, leafCmd("z"))

	m := newPickerModel(root.Commands(), "")
	idx := -1
	for i, it := range m.stack[0].items {
		if it.cmd == grp {
			idx = i
		}
	}
	m.list.Select(idx)
	selected, ok := m.list.SelectedItem().(pickItem)
	if !ok {
		t.Fatal("no selected item")
	}
	m.choose(selected)
	m.list.Select(2)
	m.pop()
	if m.list.Index() != idx {
		t.Errorf("cursor after pop = %d, want %d", m.list.Index(), idx)
	}
}
