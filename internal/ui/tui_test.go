package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
			capture := termkit.CaptureFromContext(cmd.Context())
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
	termkit.MarkTableView(view)

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

// TestRunCapturedKeepsRawPayload pins that a raw payload stays in the capture
// for the session to render, instead of being converted to a message here.
func TestRunCapturedKeepsRawPayload(t *testing.T) {
	t.Parallel()

	root, view, _, _ := newTestTree()
	capture, err := RunCaptured(root)(context.Background(), view, []string{"rsk", "view", "--json"})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if got := capture.Raw(); got != `{"a":1}` {
		t.Fatalf("raw = %q, want the payload", got)
	}
	if msgs := capture.Messages(); len(msgs) != 0 {
		t.Fatalf("messages = %v, want none (the session shows Raw)", msgs)
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
	if termkit.CaptureFromContext(view.Context()) != nil {
		t.Fatal("the command kept the run's capture; the next run would record into it")
	}
}

func TestInSession(t *testing.T) {
	t.Parallel()

	if InSession(nil) || InSession(context.Background()) { //nolint:staticcheck // nil context is the case under test
		t.Error("a plain context is not a session")
	}
	if !InSession(termkit.WithCapture(context.Background(), &termkit.Capture{})) {
		t.Error("a captured context is a session")
	}
}

func TestNewSessionConfigNoticeAndRunnable(t *testing.T) {
	t.Parallel()

	root, _, _, _ := newTestTree() //nolint:dogsled // only the root matters here
	cfg := newSessionConfig(SessionOptions{Root: root, Notice: "2 updates available"}, "", "")
	if cfg.Root != root || cfg.Run == nil || cfg.Title != "rsk" {
		t.Fatalf("config = %+v, want root, run and title set", cfg)
	}
	if want := "rsk — pick a command · 2 updates available"; cfg.PickerTitle != want {
		t.Errorf("picker title = %q, want %q", cfg.PickerTitle, want)
	}
}

func TestNewSessionConfigWithoutStoresLeavesThemOff(t *testing.T) {
	t.Parallel()

	root, _, _, _ := newTestTree() //nolint:dogsled // only the root matters here
	cfg := newSessionConfig(SessionOptions{Root: root}, "", "")
	if cfg.History != nil || cfg.Prefill != nil || cfg.Theme != nil {
		t.Fatalf("history/prefill/theme = %v/%v/%v, want all off", cfg.History, cfg.Prefill, cfg.Theme)
	}
}

func TestNewSessionConfigPersistsHistoryAndPrefs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	historyPath := filepath.Join(dir, "history.json")
	prefsPath := filepath.Join(dir, "prefs.json")
	root, _, _, _ := newTestTree() //nolint:dogsled // only the root matters here

	cfg := newSessionConfig(SessionOptions{Root: root}, historyPath, prefsPath)
	if cfg.History == nil || cfg.Prefill == nil || cfg.Theme == nil {
		t.Fatalf("history/prefill/theme = %v/%v/%v, want all on", cfg.History, cfg.Prefill, cfg.Theme)
	}
	cfg.History.Record("say")
	cfg.Prefill.SetValue("say", "loud", "true")
	cfg.Theme.SetTheme("catppuccin")

	again := newSessionConfig(SessionOptions{Root: root}, historyPath, prefsPath)
	if got := again.History.Recent(); len(got) != 1 || got[0] != "say" {
		t.Errorf("recent = %v, want [say]", got)
	}
	if got := again.Prefill.Values("say")["loud"]; got != "true" {
		t.Errorf("prefilled loud = %q, want true", got)
	}
	if got := again.Theme.Theme(); got != "catppuccin" {
		t.Errorf("theme = %q, want catppuccin", got)
	}
}

func TestNewSessionConfigIgnoresUnreadableStores(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(bad, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	root, _, _, _ := newTestTree() //nolint:dogsled // only the root matters here

	cfg := newSessionConfig(SessionOptions{Root: root}, bad, bad)
	if cfg.History != nil || cfg.Prefill != nil || cfg.Theme != nil {
		t.Fatal("a malformed store must leave its feature off, not fail the session")
	}
}

func TestHelpMarkdownListsVisibleCommandsByGroup(t *testing.T) {
	t.Parallel()

	root, _, _, _ := newTestTree() //nolint:dogsled // only the root matters here
	root.Long = "About rsk."
	root.AddGroup(&cobra.Group{ID: "g", Title: "Things:"})
	noop := func(*cobra.Command, []string) {}
	visible := &cobra.Command{Use: "shown", Short: "does a thing", GroupID: "g", Run: noop}
	hidden := &cobra.Command{Use: "secret", Short: "nope", GroupID: "g", Hidden: true, Run: noop}
	root.AddCommand(visible, hidden)

	got := helpMarkdown(root, "1 update available")
	for _, want := range []string{"About rsk.", "**1 update available**", "### Things", "- `shown` — does a thing"} {
		if !strings.Contains(got, want) {
			t.Errorf("help markdown is missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "secret") {
		t.Errorf("help markdown lists a hidden command:\n%s", got)
	}
}
