package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ralvarezdev/termkit"

	"github.com/ralvarezdev/ralvaskills/v2/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/v2/internal/manifest"
	"github.com/ralvarezdev/ralvaskills/v2/internal/ui"
)

// TestRunClaudeToolsListJSONOutput checks that -o json on `rsk claude tools
// list` emits well-formed JSON with the allow/deny rule lists, and that an
// invalid --output value is rejected the same way list/catalog/status reject
// theirs.
//
//nolint:paralleltest // t.Chdir changes the process-wide working directory
func TestRunClaudeToolsListJSONOutput(t *testing.T) {
	projDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projDir, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(projDir)
	if err := manifest.WriteMod(filepath.Join(projDir, ".rsk"), manifest.Mod{}); err != nil {
		t.Fatalf("write rsk.mod: %v", err)
	}

	claudeTool, err := claudeToolGet()
	if err != nil {
		t.Fatal(err)
	}
	if err = claudeTool.WritePermissions(projDir, []string{"Bash"}, []string{"Write(**)"}); err != nil {
		t.Fatalf("seed permissions: %v", err)
	}

	// setupCommands (invoked from this package's init) already registered
	// --output on claudeToolsListCmd, so just flip its value and restore the
	// default afterward rather than re-declaring the flag.
	f := claudeToolsListCmd.Flags()
	t.Cleanup(func() { _ = f.Set(cmdx.FlagOutput, string(outputText)) })
	if err = f.Set(cmdx.FlagOutput, string(outputJSON)); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	claudeToolsListCmd.SetOut(&buf)
	t.Cleanup(func() { claudeToolsListCmd.SetOut(nil) })

	if err = runClaudeToolsList(claudeToolsListCmd, nil); err != nil {
		t.Fatalf("runClaudeToolsList with -o json: unexpected error: %v", err)
	}

	var perms claudeToolsPermissions
	if err = json.Unmarshal(buf.Bytes(), &perms); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if len(perms.Allow) != 1 || perms.Allow[0] != "Bash" {
		t.Fatalf("unexpected allow list: %+v", perms.Allow)
	}
	if len(perms.Deny) != 1 || perms.Deny[0] != "Write(**)" {
		t.Fatalf("unexpected deny list: %+v", perms.Deny)
	}

	if err = f.Set(cmdx.FlagOutput, "bogus"); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	if err = runClaudeToolsList(claudeToolsListCmd, nil); err == nil {
		t.Fatal("expected error for invalid --output value, got nil")
	}
}

// TestRunClaudeToolsListCapture guards the picker's captured/scrollable path
// for `rsk claude tools list`: under a ui.WithCapture-wrapped context, the
// table must land in the capture and nothing must be written to the
// command's own writer.
//
//nolint:paralleltest // t.Chdir changes the process-wide working directory
func TestRunClaudeToolsListCapture(t *testing.T) {
	projDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projDir, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(projDir)
	if err := manifest.WriteMod(filepath.Join(projDir, ".rsk"), manifest.Mod{}); err != nil {
		t.Fatalf("write rsk.mod: %v", err)
	}

	f := claudeToolsListCmd.Flags()
	t.Cleanup(func() { _ = f.Set(cmdx.FlagOutput, string(outputText)) })
	if err := f.Set(cmdx.FlagOutput, string(outputText)); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	claudeToolsListCmd.SetOut(&buf)
	claudeToolsListCmd.SetContext(ui.WithCapture(t.Context(), &termkit.Capture{}))
	t.Cleanup(func() {
		claudeToolsListCmd.SetOut(nil)
		claudeToolsListCmd.SetContext(context.Background())
	})

	capture := ui.CaptureFromContext(claudeToolsListCmd.Context())
	if capture == nil {
		t.Fatal("expected a capture on the command's context")
	}

	if err := runClaudeToolsList(claudeToolsListCmd, nil); err != nil {
		t.Fatalf("runClaudeToolsList: unexpected error: %v", err)
	}

	if buf.Len() != 0 {
		t.Fatalf("expected no output written to the command's writer under capture, got:\n%s", buf.String())
	}

	tables := capture.Tables()
	if len(tables) != 1 {
		t.Fatalf("expected exactly one captured table, got %d", len(tables))
	}
	if len(tables[0].Rows) != len(availableClaudeTools) {
		t.Fatalf("expected %d captured rows, got %d", len(availableClaudeTools), len(tables[0].Rows))
	}
	if len(tables[0].IDs) != len(availableClaudeTools) {
		t.Fatalf("expected %d captured row IDs, got %d", len(availableClaudeTools), len(tables[0].IDs))
	}
	for i, id := range tables[0].IDs {
		if id != availableClaudeTools[i] {
			t.Fatalf("row ID %d = %q, want %q", i, id, availableClaudeTools[i])
		}
	}

	// The three row actions registered in setupCommands must reach the
	// captured table, sorted by key, so pressing one in the result view runs
	// allow/deny/remove for the tool under the cursor.
	wantActions := []struct{ key, label string }{{"a", "allow"}, {"d", "deny"}, {"x", "remove"}}
	if len(tables[0].Actions) != len(wantActions) {
		t.Fatalf(
			"expected %d captured row actions, got %d: %+v",
			len(wantActions), len(tables[0].Actions), tables[0].Actions,
		)
	}
	for i, want := range wantActions {
		if tables[0].Actions[i].Key != want.key || tables[0].Actions[i].Label != want.label {
			t.Fatalf("action %d = %+v, want {Key:%s Label:%s}", i, tables[0].Actions[i], want.key, want.label)
		}
	}
}
