package ui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ralvarezdev/termkit"
	"github.com/spf13/cobra"
)

// tableViewCmd builds a leaf command marked with MarkTableView whose RunE
// records a table into whatever capture is on its context, mirroring how
// list/status/catalog/tools list behave once captured.
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

// TestSessionHandleRowActionRunsTargetWithScopeCopied checks that firing a
// RowActionMsg registered via MarkRowAction closes the result screen, copies
// the source command's --global scope onto the target before running it, and
// returns a non-nil tea.Cmd (the target's tea.Exec run).
func TestSessionHandleRowActionRunsTargetWithScopeCopied(t *testing.T) {
	t.Parallel()

	source := tableViewCmd(t)
	source.Flags().Bool("global", false, "")
	if err := source.Flags().Set("global", "true"); err != nil {
		t.Fatal(err)
	}

	target := &cobra.Command{Use: "uninstall", RunE: func(*cobra.Command, []string) error { return nil }}
	target.Flags().Bool("global", false, "")
	MarkRowAction(source, "u", "uninstall", target)

	session := newSessionModel([]*cobra.Command{source}, "")
	session.width, session.height = 80, 24
	session.captured(capturedMsg{cmd: source, line: "catalog"})
	if session.resultCmd != source {
		t.Fatal("session.resultCmd was not set from capturedMsg.cmd")
	}

	teaCmd := session.handleRowAction(termkit.RowActionMsg{ID: "go-grpc", Key: "u"})

	if teaCmd == nil {
		t.Fatal("handleRowAction returned a nil tea.Cmd for a registered action")
	}
	if session.screen != screenPicker {
		t.Errorf("session.screen = %v, want screenPicker after a row action", session.screen)
	}
	if session.result != nil || session.resultCmd != nil {
		t.Error("session.result/resultCmd should be cleared after a row action")
	}
	if v, _ := target.Flags().GetBool("global"); !v {
		t.Error("handleRowAction did not copy --global from the source command onto the target")
	}
}

// TestSessionHandleRowActionNoopWithoutRegistration checks that a
// RowActionMsg for a key with no MarkRowAction registration is ignored.
func TestSessionHandleRowActionNoopWithoutRegistration(t *testing.T) {
	t.Parallel()

	source := tableViewCmd(t)
	session := newSessionModel([]*cobra.Command{source}, "")
	session.width, session.height = 80, 24
	session.captured(capturedMsg{cmd: source, line: "catalog"})

	if cmd := session.handleRowAction(termkit.RowActionMsg{ID: "go-grpc", Key: "x"}); cmd != nil {
		t.Error("handleRowAction should return nil for an unregistered key")
	}
	if session.screen != screenResult {
		t.Error("session.screen should stay screenResult when the row action is a no-op")
	}
}

// TestSessionHandleRowActionAppliesTableScope checks that a row action
// registered with MarkRowActionScoped sets its target's flags from the table
// the action fired on, so a view whose tables cover more than one scope (e.g.
// status's one table per target directory) runs the target against the scope
// of the selected row rather than the source view's own flags.
func TestSessionHandleRowActionAppliesTableScope(t *testing.T) {
	t.Parallel()

	source := tableViewCmd(t)
	target := &cobra.Command{Use: "uninstall", RunE: func(*cobra.Command, []string) error { return nil }}
	target.Flags().Bool("global", false, "")
	target.Flags().String("for", "", "")
	MarkRowActionScoped(source, "u", "uninstall", target, func(table termkit.Data) map[string]string {
		if tool, ok := strings.CutPrefix(table.Title, "Global — "); ok {
			return map[string]string{"global": "true", "for": tool}
		}
		return nil
	})

	session := newSessionModel([]*cobra.Command{source}, "")
	session.width, session.height = 80, 24
	session.captured(capturedMsg{
		cmd:  source,
		line: "status",
		tables: []termkit.Data{
			{Title: "Global — opencode", Headers: []string{"Name"}, Rows: [][]any{{"a"}}, IDs: []string{"a"}},
			{Title: "Project", Headers: []string{"Name"}, Rows: [][]any{{"b"}}, IDs: []string{"b"}},
		},
	})

	if cmd := session.handleRowAction(termkit.RowActionMsg{Table: 0, ID: "a", Key: "u"}); cmd == nil {
		t.Fatal("handleRowAction returned a nil tea.Cmd for a registered scoped action")
	}
	if v, _ := target.Flags().GetBool("global"); !v {
		t.Error("scoped action did not set --global from the fired table")
	}
	if v, _ := target.Flags().GetString("for"); v != "opencode" {
		t.Errorf("scoped action set --for = %q, want opencode", v)
	}
	if session.resultTables != nil {
		t.Error("session.resultTables should be cleared after a row action")
	}
}
