package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// PickerItem is one selectable row in the top-level command picker.
type PickerItem struct {
	Cmd *cobra.Command
}

// Title implements list.Item.
func (i PickerItem) Title() string { return i.Cmd.Name() }

// Description implements list.Item.
func (i PickerItem) Description() string { return i.Cmd.Short }

// FilterValue implements list.Item.
func (i PickerItem) FilterValue() string { return i.Cmd.Name() }

// commandFinishedMsg reports the outcome of the most recently run command, so
// the menu can surface it without interrupting the dashboard loop.
type commandFinishedMsg struct {
	name string
	err  error
}

// pickerModel drives the top-level command list. Enter runs the selected
// command inline (releasing the terminal so it can print/prompt normally)
// and returns to the menu afterward; ctrl+c/q/esc leave the picker.
type pickerModel struct {
	list          list.Model
	cancelled     bool
	lastRun       string
	lastErr       error
	cap           Capability
	width, height int
	notice        string
}

// NewPicker builds a Bubble Tea program that lets the user pick and run one
// of cmds' visible subcommands, returning to the same menu after each run.
// notice, when non-empty, is shown above the list — e.g. an
// updates-available flag — and may be "".
func NewPicker(cmds []*cobra.Command, in io.Reader, out io.Writer, notice string) *tea.Program {
	model := newPickerModel(cmds, notice)
	return tea.NewProgram(model, tea.WithAltScreen(), tea.WithInput(in), tea.WithOutput(out))
}

func newPickerModel(cmds []*cobra.Command, notice string) *pickerModel {
	capa := NewCapability()
	m := &pickerModel{
		list:   newPickerList(cmds, capa),
		cap:    capa,
		width:  defaultPickerWidth,
		height: defaultPickerHeight,
		notice: notice,
	}
	m.fit()
	return m
}

func newPickerList(cmds []*cobra.Command, capa Capability) list.Model {
	items := make([]list.Item, 0, len(cmds))
	for _, c := range cmds {
		if c.Hidden || !c.Runnable() && len(c.Commands()) == 0 {
			continue
		}
		items = append(items, PickerItem{Cmd: c})
	}

	model := list.New(items, pickerDelegate(capa), defaultPickerWidth, defaultPickerHeight)
	model.Title = "rsk — pick a command"
	styleList(&model, capa)
	model.SetFilteringEnabled(true)
	model.SetShowHelp(true)
	model.SetShowStatusBar(true)
	model.SetShowPagination(true)
	return model
}

// pickerDelegate renders list rows in the theme palette instead of bubbles'
// default magenta.
func pickerDelegate(capa Capability) list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	if !capa.color {
		return delegate
	}

	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(ColorAccent).BorderLeftForeground(ColorAccent)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(ColorAccent).BorderLeftForeground(ColorAccent)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(ColorMuted)
	return delegate
}

// styleList puts the list chrome (title, filter prompt) in the theme palette.
func styleList(model *list.Model, capa Capability) {
	if !capa.color {
		return
	}
	model.Styles.Title = model.Styles.Title.UnsetBackground().Bold(true).Foreground(ColorAccent)
	model.Styles.FilterPrompt = model.Styles.FilterPrompt.Foreground(ColorAccent)
	model.Styles.FilterCursor = model.Styles.FilterCursor.Foreground(ColorAccent)
}

// Init implements tea.Model.
func (m *pickerModel) Init() tea.Cmd { return nil }

// Update implements tea.Model. Keys are intercepted before being delegated to
// the list, but only while the user is not filtering — while filtering, enter
// applies the filter rather than choosing a row.
func (m *pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.fit()
		return m, nil

	case commandFinishedMsg:
		m.lastRun, m.lastErr = msg.name, msg.err
		return m, nil

	case tea.KeyMsg:
		if m.list.FilterState() != list.Filtering {
			switch msg.String() {
			case "ctrl+c":
				m.cancelled = true
				return m, tea.Quit
			case "q", "esc":
				return m, tea.Quit
			case "enter":
				if item, ok := m.list.SelectedItem().(PickerItem); ok {
					return m, m.runSelected(item.Cmd)
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m *pickerModel) View() string {
	view := m.cap.Header(m.width, m.height)
	if m.notice != "" {
		view += "\n" + m.cap.paint(m.notice, lipgloss.NewStyle().Foreground(ColorWarning).Bold(true))
	}
	view += "\n" + m.list.View()
	if m.lastRun == "" {
		return view
	}
	if m.lastErr != nil {
		return view + "\n" + m.cap.paint(fmt.Sprintf("last run: %s — %s", m.lastRun, m.lastErr),
			lipgloss.NewStyle().Foreground(ColorDanger))
	}
	return view + "\n" + m.cap.muted(fmt.Sprintf("last run: %s — ok", m.lastRun))
}

// fit sizes the list to the terminal minus the header (and notice, if any).
func (m *pickerModel) fit() {
	used := lipgloss.Height(m.cap.Header(m.width, m.height))
	if m.notice != "" {
		used++
	}
	m.list.SetSize(m.width, max(m.height-used, minPickerListHeight))
}

// runSelected releases the terminal to run cmd inline — with real stdin/
// stdout, so its own prompts and output work exactly as they would outside
// the picker — then restores the picker once the user acknowledges the
// result, instead of exiting the program.
func (m *pickerModel) runSelected(cmd *cobra.Command) tea.Cmd {
	return tea.Exec(&commandExec{cmd: cmd}, func(err error) tea.Msg {
		return commandFinishedMsg{name: cmd.Name(), err: err}
	})
}

// commandExec adapts a cobra.Command to tea.ExecCommand so tea.Exec can run
// it with the terminal released, then pauses for a keypress before handing
// the terminal back to the picker so the user has time to read the output.
type commandExec struct {
	cmd         *cobra.Command
	in          io.Reader
	out, errOut io.Writer
}

func (c *commandExec) SetStdin(r io.Reader) {
	if c.in == nil {
		c.in = r
	}
}

func (c *commandExec) SetStdout(w io.Writer) {
	if c.out == nil {
		c.out = w
	}
}

func (c *commandExec) SetStderr(w io.Writer) {
	if c.errOut == nil {
		c.errOut = w
	}
}

func (c *commandExec) Run() error {
	c.cmd.SetIn(c.in)
	c.cmd.SetOut(c.out)
	c.cmd.SetErr(c.errOut)

	fmt.Fprintf(c.out, "\n$ rsk %s\n\n", c.cmd.Name())
	err := c.runCommand()
	if err != nil {
		fmt.Fprintf(c.out, "\nrsk: %s\n", err)
	}

	fmt.Fprint(c.out, "\n(press enter to return to the menu)")
	bufio.NewReader(c.in).ReadString('\n')
	return err
}

// runCommand runs c.cmd if it takes no positional arguments, or prints its
// help otherwise — the picker doesn't collect arguments itself, since
// commands like install/uninstall already prompt for what they need via
// their own huh-based flows when run directly.
func (c *commandExec) runCommand() error {
	if len(c.cmd.Commands()) > 0 || takesArgs(c.cmd) {
		return c.cmd.Help()
	}
	if c.cmd.RunE != nil {
		return c.cmd.RunE(c.cmd, nil)
	}
	if c.cmd.Run != nil {
		c.cmd.Run(c.cmd, nil)
		return nil
	}
	return c.cmd.Help()
}

// takesArgs reports whether cmd expects positional arguments, based on its
// declared Use string (e.g. "install [name...] [flags]" has a bracketed
// arg, but "[flags]" itself is not one).
func takesArgs(cmd *cobra.Command) bool {
	fields := strings.Fields(cmd.Use)
	for _, field := range fields[1:] {
		trimmed := strings.Trim(field, "[]<>.")
		if trimmed == "flags" {
			continue
		}
		if strings.ContainsAny(field, "[]<>") {
			return true
		}
	}
	return false
}

// RunPicker launches the interactive command picker over cmds. The user can
// run any number of commands inline before leaving; q/esc exit cleanly
// (exit 0), while ctrl+c is treated as an abort (exit 130, SIGINT
// convention) via ui.ErrAborted. notice, when non-empty, is shown above the
// list (e.g. an updates-available flag).
func RunPicker(cmds []*cobra.Command, out io.Writer, notice string) error {
	program := NewPicker(cmds, os.Stdin, out, notice)

	final, err := program.Run()
	if err != nil {
		return fmt.Errorf("run picker: %w", err)
	}

	if result, ok := final.(*pickerModel); ok && result.cancelled {
		return ErrAborted
	}
	return nil
}
