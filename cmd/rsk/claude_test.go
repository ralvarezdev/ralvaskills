package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ralvarezdev/ralvaskills/v2/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/v2/internal/manifest"
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
