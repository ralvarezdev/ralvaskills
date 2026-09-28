package ui

import (
	"context"
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ralvarezdev/termkit"
	"github.com/spf13/cobra"

	"github.com/ralvarezdev/ralvaskills/v2/internal/cmdx"
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
	// resultCmd is the command whose captured run produced the current
	// result screen, used to resolve a termkit.RowActionMsg fired from it
	// (see handleRowAction) — nil outside screenResult. resultTables are the
	// captured tables it rendered, so a scoped row action can be resolved
	// against the table it fired on.
	resultCmd    *cobra.Command
	resultTables []termkit.Data
}

// capturedMsg reports a read-only command that ran in-process under a
// termkit.Capture, carrying the tables, messages, and raw payload the TUI
// renders.
type capturedMsg struct {
	cmd      *cobra.Command
	line     string
	tables   []termkit.Data
	messages []string
	raw      string
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

	if action, ok := msg.(termkit.RowActionMsg); ok {
		cmd := s.handleRowAction(action)
		return s, cmd
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
		termkit.ResetFlags(cmd)
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
		return capturedMsg{
			cmd: cmd, line: line, tables: capture.Tables(), messages: capture.Messages(), raw: capture.Raw(), err: err,
		}
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
	s.resultCmd = msg.cmd
	s.resultTables = msg.tables
	s.result = termkit.NewResultView(termkit.Result{
		Breadcrumb: s.resultCrumb,
		Messages:   msg.messages,
		Tables:     msg.tables,
		Raw:        msg.raw,
	}, s.width, s.height)
	s.screen = screenResult
}

// handleRowAction runs the command registered via MarkRowAction for the
// result screen's source command and the fired RowAction's key, passing the
// selected row's ID as its sole argument — the same as picking that command
// from the menu and typing the name into its own form — and carrying over
// any --global/--for/--personal scope the source command was run with, so
// e.g. uninstalling from `list --global --for claude-code` targets the same
// scope instead of silently defaulting to the project.
//
// When the action was registered with MarkRowActionScoped, the table the
// action fired on further narrows that scope (status renders one table per
// target directory), so e.g. uninstalling from the "Global — opencode"
// section targets that tool.
func (s *sessionModel) handleRowAction(msg termkit.RowActionMsg) tea.Cmd {
	if s.resultCmd == nil || msg.ID == "" {
		return nil
	}
	target, scope, ok := rowActionFor(s.resultCmd, msg.Key)
	if !ok {
		return nil
	}

	termkit.ResetFlags(target)
	copySharedFlags(s.resultCmd, target, cmdx.FlagGlobal, cmdx.FlagFor, cmdx.FlagPersonal)
	applyRowActionScope(target, scope, s.tableAt(msg.Table))

	s.result, s.screen, s.resultCmd, s.resultTables = nil, screenPicker, nil, nil
	return s.run(target, []string{msg.ID})
}

// tableAt returns the captured table at index i, or the zero Data when the
// index is out of range.
func (s *sessionModel) tableAt(i int) termkit.Data {
	if i < 0 || i >= len(s.resultTables) {
		return termkit.Data{}
	}
	return s.resultTables[i]
}

// applyRowActionScope sets scope's flag values on target, skipping any flag
// target doesn't define and any value it rejects, so a scoped row action
// runs against the same scope as the row it fired on.
func applyRowActionScope(target *cobra.Command, scope RowActionScope, table termkit.Data) {
	if scope == nil {
		return
	}
	for name, value := range scope(table) {
		flag := target.Flags().Lookup(name)
		if flag == nil {
			continue
		}
		if err := flag.Value.Set(value); err != nil {
			continue
		}
		flag.Changed = true
	}
}

// copySharedFlags copies each named flag's value from source to target when
// both define it and the user actually set it on source, so running target
// for a row action carries over the same scope filters source was browsed
// with instead of guessing.
func copySharedFlags(source, target *cobra.Command, names ...string) {
	for _, name := range names {
		sourceFlag := source.Flags().Lookup(name)
		targetFlag := target.Flags().Lookup(name)
		if sourceFlag == nil || targetFlag == nil || !sourceFlag.Changed {
			continue
		}
		//nolint:errcheck // re-parsing a value already validated on source's own flag cannot fail
		targetFlag.Value.Set(sourceFlag.Value.String())
		targetFlag.Changed = true
	}
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
		s.result, s.screen, s.resultTables = nil, screenPicker, nil
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
