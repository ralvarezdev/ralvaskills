package main

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/termkit"

	"github.com/ralvarezdev/ralvaskills/v2/internal/cmdx"
	"github.com/ralvarezdev/ralvaskills/v2/internal/ui"
)

// sessionCmd returns a command whose context marks it as running inside the
// interactive session.
func sessionCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "x"}
	cmd.Flags().String(cmdx.FlagFor, "", "")
	cmd.SetContext(ui.WithCapture(context.Background(), &termkit.Capture{}))
	return cmd
}

func wantFieldError(t *testing.T, err error, field string) {
	t.Helper()
	fieldErrs := termkit.FieldErrors(err)
	if len(fieldErrs) != 1 || fieldErrs[0].Field != field {
		t.Fatalf("field errors = %+v (err %v), want one for %q", fieldErrs, err, field)
	}
}

func TestDestructiveCommandsAskForConfirmationInTheSession(t *testing.T) {
	t.Parallel()

	for _, cmd := range []*cobra.Command{destroyCmd, installCmd, uninstallCmd, updateCmd} {
		if !termkit.IsDestructive(cmd) {
			t.Errorf("%s is not marked destructive", cmd.Name())
		}
	}
	for _, cmd := range []*cobra.Command{listCmd, statusCmd, catalogCmd, pinCmd, newCmd} {
		if termkit.IsDestructive(cmd) {
			t.Errorf("%s is marked destructive", cmd.Name())
		}
	}
}

func TestConfirmProceedSkipsThePromptInTheSession(t *testing.T) {
	t.Parallel()

	var out strings.Builder
	if !confirmProceed(sessionCmd(), &out, "Proceed?") {
		t.Fatal("the session already confirmed; the command must go ahead")
	}
	if out.Len() != 0 {
		t.Fatalf("wrote %q; nothing should be prompted in the session", out.String())
	}
}

func TestSessionRunsBlameMissingInputOnTheirField(t *testing.T) {
	t.Parallel()

	var out strings.Builder

	_, err := nameFromArgsOrPrompt(sessionCmd(), nil, fieldArgRule, "Tool rule to remove")
	wantFieldError(t, err, fieldArgRule)

	_, err = promptInstallName(sessionCmd(), &out)
	wantFieldError(t, err, fieldArgName)

	_, err = resolveUninstallArgs(sessionCmd(), &out, nil)
	wantFieldError(t, err, fieldArgName)

	_, err = resolveForFlag(sessionCmd(), &out)
	wantFieldError(t, err, cmdx.FlagFor)

	if out.Len() != 0 {
		t.Fatalf("wrote %q; nothing should be prompted in the session", out.String())
	}
}

func TestInstallValidationBlamesTheFlag(t *testing.T) {
	t.Parallel()

	wantFieldError(t, validateInstallOpts(installOpts{forTool: "claude-code"}), cmdx.FlagFor)
	wantFieldError(t, validateInstallOpts(installOpts{global: true, pin: true}), cmdx.FlagPin)
	wantFieldError(t, validateInstallOpts(installOpts{version: "v1"}), cmdx.FlagVersion)
}
