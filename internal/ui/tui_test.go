package ui

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/termkit"
)

// newTestTree builds root -> {view (table view), say (plain output), fail,
// group -> leaf} for exercising RunCaptured.
func newTestTree() (root, view, say, fail *cobra.Command) {
	root = &cobra.Command{Use: "rsk", SilenceUsage: true, SilenceErrors: true}

	view = &cobra.Command{
		Use: "view",
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "ignored: the table replaces text output")
			capture := CaptureFromContext(cmd.Context())
			if capture == nil {
				return errors.New("no capture on the context")
			}
			capture.AddTable(termkit.Data{Headers: []string{"NAME"}, Rows: [][]any{{"x"}}})
			if cmd.Flags().Changed("json") {
				capture.SetRaw(`{"a":1}`)
			}
			return nil
		},
	}
	view.Flags().Bool("json", false, "raw output")
	MarkTableView(view)

	say = &cobra.Command{
		Use:  "say [word...]",
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "said %v\n", args)
			fmt.Fprintf(cmd.ErrOrStderr(), "loud=%v\n", cmd.Flags().Changed("loud"))
			return nil
		},
	}
	say.Flags().Bool("loud", false, "")

	fail = &cobra.Command{
		Use: "fail",
		RunE: func(*cobra.Command, []string) error {
			return termkit.NewFieldError("arg:name", errors.New("bad name"))
		},
	}

	group := &cobra.Command{Use: "group"}
	group.AddCommand(&cobra.Command{Use: "leaf", RunE: func(*cobra.Command, []string) error { return nil }})
	root.AddCommand(view, say, fail, group)
	return root, view, say, fail
}

func TestRunCapturedTableViewRecordsTablesAndDropsText(t *testing.T) {
	t.Parallel()

	root, view, _, _ := newTestTree()
	capture, err := RunCaptured(root)(context.Background(), view, []string{"rsk", "view"})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := len(capture.Tables()); got != 1 {
		t.Fatalf("tables = %d, want 1", got)
	}
	if got := capture.Messages(); len(got) != 0 {
		t.Fatalf("messages = %v, want none for a table view", got)
	}
}

func TestRunCapturedTurnsRawPayloadIntoMessage(t *testing.T) {
	t.Parallel()

	root, view, _, _ := newTestTree()
	capture, err := RunCaptured(root)(context.Background(), view, []string{"rsk", "view", "--json"})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := capture.Messages(); len(got) != 1 || got[0] != `{"a":1}` {
		t.Fatalf("messages = %v, want the raw payload", got)
	}
}

func TestRunCapturedCollectsTextFromPlainCommands(t *testing.T) {
	t.Parallel()

	root, _, say, _ := newTestTree()
	run := RunCaptured(root)

	capture, err := run(context.Background(), say, []string{"rsk", "say", "--loud", "--", "-a", "b"})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	want := "said [-a b]\nloud=true"
	if got := capture.Messages(); len(got) != 1 || got[0] != want {
		t.Fatalf("messages = %q, want [%q]", got, want)
	}

	// Flags do not leak into the next run of the same command.
	capture, err = run(context.Background(), say, []string{"rsk", "say"})
	if err != nil {
		t.Fatalf("second run: %v", err)
	}
	if got := capture.Messages(); len(got) != 1 || got[0] != "said []\nloud=false" {
		t.Fatalf("second run messages = %q, want loud=false", got)
	}
}

func TestRunCapturedKeepsFieldErrors(t *testing.T) {
	t.Parallel()

	root, _, _, fail := newTestTree()
	_, err := RunCaptured(root)(context.Background(), fail, []string{"rsk", "fail"})
	fieldErrs := termkit.FieldErrors(err)
	if len(fieldErrs) != 1 || fieldErrs[0].Field != "arg:name" {
		t.Fatalf("field errors = %+v (err %v), want one for arg:name", fieldErrs, err)
	}
}

func TestRunCapturedDoesNotPinContext(t *testing.T) {
	t.Parallel()

	root, view, _, _ := newTestTree()
	if _, err := RunCaptured(root)(context.Background(), view, []string{"rsk", "view"}); err != nil {
		t.Fatalf("run: %v", err)
	}
	if CaptureFromContext(view.Context()) != nil {
		t.Fatal("the command kept the run's capture; the next run would record into it")
	}
}

func TestInSession(t *testing.T) {
	t.Parallel()

	if InSession(nil) || InSession(context.Background()) { //nolint:staticcheck // nil context is the case under test
		t.Error("a plain context is not a session")
	}
	if !InSession(WithCapture(context.Background(), &termkit.Capture{})) {
		t.Error("a captured context is a session")
	}
}

func TestNewSessionConfigNoticeAndRunnable(t *testing.T) {
	t.Parallel()

	root, _, _, _ := newTestTree() //nolint:dogsled // only the root matters here
	cfg := newSessionConfig(SessionOptions{Root: root, Notice: "2 updates available"})
	if cfg.Root != root || cfg.Run == nil || cfg.Title != "rsk" {
		t.Fatalf("config = %+v, want root, run and title set", cfg)
	}
	if want := "rsk — pick a command · 2 updates available"; cfg.PickerTitle != want {
		t.Errorf("picker title = %q, want %q", cfg.PickerTitle, want)
	}
}
