package ui

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/ralvaskills/v2/internal/cmdx"
)

func TestArgsLabel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		use  string
		want string
	}{
		{"bracketed name", "install [name...] [flags]", "name..."},
		{"angle bracket", "pin <name> [flags]", "name"},
		{"no args", "list [flags]", "args"},
		{"no flags token", "status", "args"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := &cobra.Command{Use: tt.use}
			if got := argsLabel(cmd); got != tt.want {
				t.Errorf("argsLabel(%q) = %q, want %q", tt.use, got, tt.want)
			}
		})
	}
}

func TestArgsRequired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		use  string
		want bool
	}{
		{"optional bracket", "install [name...] [flags]", false},
		{"required bracket", "pin <name> [flags]", true},
		{"no args", "list [flags]", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := &cobra.Command{Use: tt.use}
			if got := argsRequired(cmd); got != tt.want {
				t.Errorf("argsRequired(%q) = %v, want %v", tt.use, got, tt.want)
			}
		})
	}
}

func TestSplitArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{"empty", "", nil},
		{"single", "foo", []string{"foo"}},
		{"multiple", "foo bar baz", []string{"foo", "bar", "baz"}},
		{"extra whitespace", "  foo   bar  ", []string{"foo", "bar"}},
		{"double quoted", `foo "bar baz"`, []string{"foo", "bar baz"}},
		{"single quoted", `foo 'bar baz'`, []string{"foo", "bar baz"}},
		{"tabs", "foo\tbar", []string{"foo", "bar"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := splitArgs(tt.raw)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitArgs(%q) = %#v, want %#v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestCommandFields_noParams(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "list [flags]"}
	if got := commandFields(cmd); got != nil {
		t.Errorf("commandFields() = %#v, want nil", got)
	}
}

func TestCommandFields_flagsAndArgs(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "install [name...] [flags]"}
	cmd.Flags().Bool("global", false, "install system-wide")
	cmd.Flags().String("source", "official", "source to install from")
	cmd.Flags().Bool("help", false, "help for install")

	fields := commandFields(cmd)
	if len(fields) != 3 {
		t.Fatalf("commandFields() len = %d, want 3 (got %#v)", len(fields), fields)
	}

	byKey := make(map[string]*formField, len(fields))
	for _, f := range fields {
		byKey[f.key] = f
	}

	if _, ok := byKey["help"]; ok {
		t.Error("commandFields() included the help flag")
	}

	global, ok := byKey["global"]
	if !ok {
		t.Fatal("commandFields() missing the global flag")
	}
	if !global.isBool || global.boolVal {
		t.Errorf("global field = %+v, want isBool=true boolVal=false", global)
	}

	source, ok := byKey["source"]
	if !ok {
		t.Fatal("commandFields() missing the source flag")
	}
	if source.isBool || source.input.Value() != "official" {
		t.Errorf("source field = %+v, want isBool=false value=official", source)
	}

	args, ok := byKey[argsFieldName]
	if !ok {
		t.Fatal("commandFields() missing the positional-args field")
	}
	if args.required {
		t.Error("args field should not be required for [name...]")
	}
	if args.label != "name..." {
		t.Errorf("args label = %q, want %q", args.label, "name...")
	}
}

func TestCommandFields_outputFlagDroppedForTableViews(t *testing.T) {
	t.Parallel()

	build := func() *cobra.Command {
		cmd := &cobra.Command{Use: "list [flags]"}
		cmd.Flags().String(cmdx.FlagOutput, "text", "Output format: text|json")
		cmd.Flags().Bool("all", false, "include all")
		return cmd
	}

	plain := build()
	if got := commandFields(plain); len(got) != 2 {
		t.Errorf("non-table command fields = %d, want 2 (output kept)", len(got))
	}

	view := build()
	MarkTableView(view)
	got := commandFields(view)
	if len(got) != 1 || got[0].key != "all" {
		t.Errorf("table view fields = %+v, want only --all", got)
	}

	onlyOutput := &cobra.Command{Use: "status"}
	onlyOutput.Flags().String(cmdx.FlagOutput, "text", "")
	MarkTableView(onlyOutput)
	if commandFields(onlyOutput) != nil {
		t.Error("table view with only --output should have no form")
	}
}
