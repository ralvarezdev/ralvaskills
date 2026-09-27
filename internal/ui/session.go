package ui

import (
	"context"
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ralvarezdev/termkit"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// sessionScreen is the screen the session is currently showing.
type sessionScreen int

const (
	screenPicker sessionScreen = iota
	screenForm
	screenResult
)

// sessionModel ties the nested picker, the parameter form, and the captured
// result view together into one long-lived program: picking a leaf either
// runs it directly (no parameters to collect) or opens the form for it, a
// read-only table-view command runs in-process and opens the scrollable
// result screen, and every other command releases the terminal via
// tea.Exec — all three paths return to the picker screen once they finish.
type sessionModel struct {
	picker        *pickerModel
	form          *formModel
	result        *termkit.ResultView
	screen        sessionScreen
	pending       *cobra.Command
	cap           Capability
	width, height int
	// running is true while a captured command executes in-process; keys are
	// ignored then so the cobra tree isn't mutated mid-run.
	running     bool
	runningLine string
	resultLine  string
	resultCrumb string
}

// capturedMsg reports a read-only command that ran in-process under a
// termkit.Capture, carrying the tables and messages the TUI renders.
type capturedMsg struct {
	line     string
	tables   []termkit.Data
	messages []string
	err      error
}

// newSessionModel builds the session rooted at cmds' top-level commands.
func newSessionModel(cmds []*cobra.Command, notice string) *sessionModel {
	picker := newPickerModel(cmds, notice)
	return &sessionModel{
		picker: picker,
		screen: screenPicker,
		cap:    picker.cap,
		width:  defaultPickerWidth,
		height: defaultPickerHeight,
	}
}

// Init implements tea.Model.
func (s *sessionModel) Init() tea.Cmd { return s.picker.Init() }

// Update implements tea.Model: window sizes go to every screen, a finished
// captured run is handled here directly, and everything else goes to the
// one screen currently showing.
//
//nolint:ireturn // signature dictated by the tea.Model interface
func (s *sessionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, isResize := msg.(tea.WindowSizeMsg); isResize {
		s.width, s.height = size.Width, size.Height
		s.picker.Update(msg)
		if s.form != nil {
			s.form.Update(msg)
		}
		if s.result != nil {
			s.result.Update(msg)
		}
		return s, nil
	}

	if captured, ok := msg.(capturedMsg); ok {
		s.captured(captured)
		return s, nil
	}

	// While a captured command runs in-process it mutates the cobra tree, so
	// keys must not reach the picker. ctrl+c still aborts the session.
	if s.running {
		if key, ok := msg.(tea.KeyMsg); ok && key.String() == keyCtrlC {
			s.picker.cancelled = true
			return s, tea.Quit
		}
		return s, nil
	}

	var cmd tea.Cmd
	switch s.screen {
	case screenForm:
		cmd = s.updateForm(msg)
	case screenResult:
		cmd = s.updateResult(msg)
	case screenPicker:
		cmd = s.updatePicker(msg)
	}
	return s, cmd
}

// View implements tea.Model.
func (s *sessionModel) View() string {
	if s.running {
		return MutedStyle.Render("running " + s.runningLine + "…")
	}
	switch s.screen {
	case screenForm:
		return s.form.View()
	case screenResult:
		if s.result != nil {
			return s.result.View()
		}
	case screenPicker:
		return s.picker.View()
	}
	return s.picker.View()
}

// updatePicker forwards to the picker and opens what it selects.
func (s *sessionModel) updatePicker(msg tea.Msg) tea.Cmd {
	_, cmd := s.picker.Update(msg)

	if s.picker.cancelled {
		return tea.Quit
	}

	selected := s.picker.selected
	if selected == nil {
		return cmd
	}
	s.picker.selected = nil

	return s.open(selected)
}

// open shows the parameter form for cmd, or runs it directly when it has
// nothing to ask — in-process and captured when cmd is a marked table view,
// via tea.Exec otherwise.
func (s *sessionModel) open(cmd *cobra.Command) tea.Cmd {
	fields := commandFields(cmd)
	if fields == nil {
		if IsTableView(cmd) {
			return s.startCaptured(cmd, nil)
		}
		return s.run(cmd, nil)
	}

	s.pending = cmd
	s.form = newFormModel(s.cap, cmd.CommandPath(), fields)
	s.form.Update(tea.WindowSizeMsg{Width: s.width, Height: s.height})
	s.screen = screenForm
	return s.form.Init()
}

// updateForm forwards to the form; esc goes back to the picker, submitting
// applies the values and runs the command.
func (s *sessionModel) updateForm(msg tea.Msg) tea.Cmd {
	_, cmd := s.form.Update(msg)

	switch {
	case s.form.cancelled:
		s.form, s.screen = nil, screenPicker
		return nil
	case s.form.submit:
		return s.apply()
	}

	return cmd
}

// apply sets the pending command's flags from the form's values and runs it;
// a Set error (e.g. an int flag given non-numeric text) keeps the form open
// instead of running.
func (s *sessionModel) apply() tea.Cmd {
	values := s.form.values()
	var args []string

	for _, field := range s.form.fields {
		if field.key == argsFieldName {
			args = splitArgs(values[argsFieldName])
			continue
		}

		value := values[field.key]
		if err := s.pending.Flags().Set(field.key, value); err != nil {
			s.form.submit = false
			s.form.err = err.Error()
			return nil
		}
	}

	cmd := s.pending
	s.pending, s.form, s.screen = nil, nil, screenPicker
	if IsTableView(cmd) {
		return s.startCaptured(cmd, args)
	}
	return s.run(cmd, args)
}

// run releases the terminal to run cmd inline via tea.Exec, returning to the
// picker screen once the user acknowledges the result.
func (s *sessionModel) run(cmd *cobra.Command, args []string) tea.Cmd {
	return tea.Exec(&commandExec{cmd: cmd, args: args}, func(err error) tea.Msg {
		return commandFinishedMsg{name: commandDisplayName(cmd), err: err}
	})
}

// startCaptured runs a read-only command in-process under a termkit.Capture,
// so its table and messages come back as data for the result screen instead
// of being printed behind the TUI's back.
func (s *sessionModel) startCaptured(cmd *cobra.Command, args []string) tea.Cmd {
	line := commandDisplayName(cmd)
	if len(args) > 0 {
		line += " " + strings.Join(args, " ")
	}
	s.running = true
	s.runningLine = line
	s.resultCrumb = cmd.CommandPath()
	return func() tea.Msg {
		capture := &termkit.Capture{}
		resetFlags(cmd)
		cmd.SetContext(WithCapture(context.Background(), capture))
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		err := runLeaf(cmd, args)
		// Clear the context so a later tea.Exec run of the same command isn't
		// captured too.
		cmd.SetContext(context.Background())
		if err != nil {
			err = fmt.Errorf("run %s: %w", cmd.Name(), err)
		}
		return capturedMsg{line: line, tables: capture.Tables(), messages: capture.Messages(), err: err}
	}
}

// captured handles a finished in-process run: an error is reported through
// the picker's own status line (the same commandFinishedMsg it already
// handles for tea.Exec runs), while a success opens the scrollable result
// screen.
func (s *sessionModel) captured(msg capturedMsg) {
	s.running = false
	if msg.err != nil {
		s.picker.Update(commandFinishedMsg{name: msg.line, err: msg.err})
		return
	}

	s.resultLine = msg.line
	var table *termkit.Data
	if len(msg.tables) > 0 {
		table = &msg.tables[0]
	}
	s.result = termkit.NewResultView(termkit.Result{
		Breadcrumb: s.resultCrumb,
		Messages:   msg.messages,
		Table:      table,
	}, s.width, s.height)
	s.screen = screenResult
}

// updateResult forwards keys to the result screen and, when it closes,
// returns to the picker and records the successful run on its status line.
func (s *sessionModel) updateResult(msg tea.Msg) tea.Cmd {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == keyCtrlC {
		s.picker.cancelled = true
		return tea.Quit
	}
	if s.result == nil {
		s.screen = screenPicker
		return nil
	}
	_, cmd := s.result.Update(msg)
	if s.result.Closed() {
		s.picker.Update(commandFinishedMsg{name: s.resultLine, err: nil})
		s.result, s.screen = nil, screenPicker
	}
	return cmd
}

// commandDisplayName is cmd's path with the root's own name trimmed off, the
// same shape commandExec's tea.Exec path already reports on the picker's
// status line.
func commandDisplayName(cmd *cobra.Command) string {
	return strings.TrimPrefix(cmd.CommandPath(), cmd.Root().Name()+" ")
}

// runLeaf runs cmd with args exactly as commandExec.runCommand does for the
// tea.Exec path, so both call sites resolve RunE/Run/Help the same way.
func runLeaf(cmd *cobra.Command, args []string) error {
	if cmd.RunE != nil {
		return cmd.RunE(cmd, args)
	}
	if cmd.Run != nil {
		cmd.Run(cmd, args)
		return nil
	}
	return cmd.Help()
}

// resetFlags puts cmd's own flags back to their defaults. cobra keeps flag
// values between Execute/RunE calls in one process, so without this a
// second in-process run of the same command would silently reuse the first
// run's values (or stay stuck printing help after a --help run).
func resetFlags(cmd *cobra.Command) {
	// Reset a context left by a previous run so the current context (with or
	// without a capture) applies to this command again.
	cmd.SetContext(context.Background())
	cmd.NonInheritedFlags().VisitAll(func(flag *pflag.Flag) {
		if slice, ok := flag.Value.(pflag.SliceValue); ok {
			//nolint:errcheck // replacing with an empty slice cannot fail
			slice.Replace(nil)
		} else {
			//nolint:errcheck // re-parsing the flag's own default cannot fail
			flag.Value.Set(flag.DefValue)
		}
		flag.Changed = false
	})
}
