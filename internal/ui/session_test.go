package ui

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ralvarezdev/termkit"
	"github.com/spf13/cobra"
)

// tableViewCmd builds a leaf command marked with MarkTableView whose RunE
// records a table into whatever capture is on its context, mirroring how
// list/status/catalog/claude tools list behave once captured.
func tableViewCmd(t *testing.T) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{
		Use: "widgets",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if capture := CaptureFromContext(cmd.Context()); capture != nil {
				capture.AddTable(termkit.Data{Headers: []string{"Name"}, Rows: [][]any{{"one"}, {"two"}}})
			}
			return nil
		},
	}
	MarkTableView(cmd)
	root := &cobra.Command{Use: "rsk"}
	root.AddCommand(cmd)
	return cmd
}

func TestCommandDisplayName(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "rsk"}
	tools := &cobra.Command{Use: "tools"}
	list := &cobra.Command{Use: "list"}
	tools.AddCommand(list)
	root.AddCommand(tools)

	if got, want := commandDisplayName(list), "tools list"; got != want {
		t.Errorf("commandDisplayName() = %q, want %q", got, want)
	}
}

func TestResetFlags(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "widgets"}
	cmd.Flags().String("name", "default", "")
	cmd.Flags().StringSlice("tags", nil, "")
	if err := cmd.Flags().Set("name", "changed"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("tags", "a,b"); err != nil {
		t.Fatal(err)
	}
	type leftoverKey struct{}
	cmd.SetContext(context.WithValue(context.Background(), leftoverKey{}, "leftover"))

	resetFlags(cmd)

	if v, _ := cmd.Flags().GetString("name"); v != "default" {
		t.Errorf("name flag = %q, want default", v)
	}
	if v, _ := cmd.Flags().GetStringSlice("tags"); len(v) != 0 {
		t.Errorf("tags flag = %v, want empty", v)
	}
	if cmd.Flags().Lookup("name").Changed {
		t.Error("name flag still marked Changed after reset")
	}
	if cmd.Context() != context.Background() {
		t.Error("resetFlags did not reset the command's context to context.Background()")
	}
}

func TestIsTableView(t *testing.T) {
	t.Parallel()

	marked := &cobra.Command{Use: "list"}
	MarkTableView(marked)
	unmarked := &cobra.Command{Use: "install"}

	if !IsTableView(marked) {
		t.Error("IsTableView(marked) = false, want true")
	}
	if IsTableView(unmarked) {
		t.Error("IsTableView(unmarked) = true, want false")
	}
	if IsTableView(nil) {
		t.Error("IsTableView(nil) = true, want false")
	}
}

// TestSessionOpenTableViewRunsCaptured checks that open() routes a
// MarkTableView command with no parameters through startCaptured rather than
// the tea.Exec path, and that the resulting capturedMsg, once fed back into
// Update, opens the scrollable result screen with the table it recorded.
func TestSessionOpenTableViewRunsCaptured(t *testing.T) {
	t.Parallel()

	cmd := tableViewCmd(t)
	session := newSessionModel([]*cobra.Command{cmd}, "")
	session.width, session.height = 80, 24

	teaCmd := session.open(cmd)
	if !session.running {
		t.Fatal("open() on a table-view command did not set session.running")
	}
	if teaCmd == nil {
		t.Fatal("open() on a table-view command returned a nil tea.Cmd")
	}

	msg := teaCmd()
	captured, ok := msg.(capturedMsg)
	if !ok {
		t.Fatalf("open()'s tea.Cmd produced %T, want capturedMsg", msg)
	}
	if captured.err != nil {
		t.Fatalf("capturedMsg.err = %v, want nil", captured.err)
	}
	if len(captured.tables) != 1 || len(captured.tables[0].Rows) != 2 {
		t.Fatalf("capturedMsg.tables = %+v, want one table with 2 rows", captured.tables)
	}

	session.Update(captured)
	if session.running {
		t.Error("session.running still true after the captured run finished")
	}
	if session.screen != screenResult {
		t.Fatalf("session.screen = %v, want screenResult", session.screen)
	}
	if session.result == nil {
		t.Fatal("session.result is nil after a successful captured run")
	}
}

// TestSessionCapturedErrorReportsOnPickerStatusLine checks that a failed
// captured run is surfaced through the picker's own commandFinishedMsg
// status line rather than opening the result screen.
func TestSessionCapturedErrorReportsOnPickerStatusLine(t *testing.T) {
	t.Parallel()

	cmd := tableViewCmd(t)
	session := newSessionModel([]*cobra.Command{cmd}, "")
	session.width, session.height = 80, 24

	wantErr := errors.New("boom")
	session.captured(capturedMsg{line: "widgets", err: wantErr})

	if session.running {
		t.Error("session.running still true after a failed captured run")
	}
	if session.screen == screenResult {
		t.Error("session.screen is screenResult after a failed captured run")
	}
	if session.picker.lastErr == nil {
		t.Fatal("picker.lastErr is nil after a failed captured run")
	}
}

// TestSessionUpdateResultReturnsToPicker checks that closing the result
// screen (esc/q) records the run on the picker's status line and returns to
// the picker screen.
func TestSessionUpdateResultReturnsToPicker(t *testing.T) {
	t.Parallel()

	cmd := tableViewCmd(t)
	session := newSessionModel([]*cobra.Command{cmd}, "")
	session.width, session.height = 80, 24
	session.captured(capturedMsg{
		line:   "widgets",
		tables: []termkit.Data{{Headers: []string{"Name"}, Rows: [][]any{{"one"}}}},
	})
	if session.screen != screenResult {
		t.Fatalf("session.screen = %v, want screenResult", session.screen)
	}

	session.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if session.screen != screenPicker {
		t.Fatalf("session.screen = %v, want screenPicker after closing the result", session.screen)
	}
	if session.result != nil {
		t.Error("session.result not cleared after closing the result screen")
	}
	if session.picker.lastErr != nil {
		t.Errorf("picker.lastErr = %v, want nil after a successful run", session.picker.lastErr)
	}
}

// TestSessionRunningGatesKeys checks that while a captured command is
// running, keys other than ctrl+c are swallowed rather than reaching the
// picker (which would otherwise mutate the cobra tree mid-run).
func TestSessionRunningGatesKeys(t *testing.T) {
	t.Parallel()

	cmd := tableViewCmd(t)
	session := newSessionModel([]*cobra.Command{cmd}, "")
	session.width, session.height = 80, 24
	session.running = true

	if _, cmd := session.Update(tea.KeyMsg{Type: tea.KeyEnter}); cmd != nil {
		t.Error("Update() returned a non-nil tea.Cmd for a key while running")
	}
	if !session.running {
		t.Error("session.running was cleared by an unrelated key")
	}

	if _, quitCmd := session.Update(tea.KeyMsg{Type: tea.KeyCtrlC}); quitCmd == nil {
		t.Error("ctrl+c while running did not return a tea.Cmd (expected tea.Quit)")
	}
	if !session.picker.cancelled {
		t.Error("ctrl+c while running did not mark the picker cancelled")
	}
}
