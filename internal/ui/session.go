package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// sessionScreen is the screen the session is currently showing.
type sessionScreen int

const (
	screenPicker sessionScreen = iota
	screenForm
)

// sessionModel ties the nested picker and the parameter form together into
// one long-lived program: picking a leaf either runs it directly (no
// parameters to collect) or opens the form for it, and both paths return to
// the picker screen once the run finishes.
type sessionModel struct {
	picker        *pickerModel
	form          *formModel
	screen        sessionScreen
	pending       *cobra.Command
	cap           Capability
	width, height int
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

// Update implements tea.Model: window sizes go to every screen, everything
// else to the one showing.
func (s *sessionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, isResize := msg.(tea.WindowSizeMsg); isResize {
		s.width, s.height = size.Width, size.Height
		s.picker.Update(msg)
		if s.form != nil {
			s.form.Update(msg)
		}
		return s, nil
	}

	var cmd tea.Cmd
	if s.screen == screenForm {
		cmd = s.updateForm(msg)
	} else {
		cmd = s.updatePicker(msg)
	}
	return s, cmd
}

// View implements tea.Model.
func (s *sessionModel) View() string {
	if s.screen == screenForm {
		return s.form.View()
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
// nothing to ask.
func (s *sessionModel) open(cmd *cobra.Command) tea.Cmd {
	fields := commandFields(cmd)
	if fields == nil {
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
	return s.run(cmd, args)
}

// run releases the terminal to run cmd inline via tea.Exec, returning to the
// picker screen once the user acknowledges the result.
func (s *sessionModel) run(cmd *cobra.Command, args []string) tea.Cmd {
	return tea.Exec(&commandExec{cmd: cmd, args: args}, func(err error) tea.Msg {
		return commandFinishedMsg{name: strings.TrimPrefix(cmd.CommandPath(), cmd.Root().Name()+" "), err: err}
	})
}
