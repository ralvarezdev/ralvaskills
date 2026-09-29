package ui

import (
	"bytes"
	"strings"
	"testing"
)

//nolint:paralleltest // mutates package-level state
func TestQuietSuppressesInformationalLines(t *testing.T) {
	t.Cleanup(func() { SetQuiet(false) })

	var buf bytes.Buffer
	SetQuiet(true)
	Info(&buf, "hidden-info")
	Indent(&buf, "hidden-indent")
	Dim(&buf, "hidden-dim")
	Warn(&buf, "shown-warning")

	got := buf.String()
	if strings.Contains(got, "hidden") || !strings.Contains(got, "shown-warning") {
		t.Fatalf("quiet output = %q", got)
	}
}

//nolint:paralleltest // mutates package-level state
func TestVerboseEnablesDebug(t *testing.T) {
	t.Cleanup(func() { SetVerbose(false) })

	var buf bytes.Buffer
	Debug(&buf, "hidden-debug")
	if buf.Len() != 0 {
		t.Fatalf("Debug without verbose wrote %q", buf.String())
	}

	SetVerbose(true)
	Debug(&buf, "shown-debug")
	if !strings.Contains(buf.String(), "shown-debug") {
		t.Fatalf("Debug with verbose = %q", buf.String())
	}
}
