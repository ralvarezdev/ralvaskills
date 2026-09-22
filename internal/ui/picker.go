package ui

import (
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

// pickerModel drives the top-level command list. Enter selects and quits;
// ctrl+c/q/esc cancel and quit.
type pickerModel struct {
	list          list.Model
	selected      *cobra.Command
	cancelled     bool
	cap           Capability
	width, height int
}

// NewPicker builds a Bubble Tea program that lets the user pick one of
// cmds' visible, runnable subcommands. Run it and inspect the returned
// PickedCommand and PickerCancelled.
func NewPicker(cmds []*cobra.Command, in io.Reader, out io.Writer) *tea.Program {
	model := newPickerModel(cmds)
	return tea.NewProgram(model, tea.WithAltScreen(), tea.WithInput(in), tea.WithOutput(out))
}

func newPickerModel(cmds []*cobra.Command) *pickerModel {
	capa := NewCapability()
	m := &pickerModel{
		list:   newPickerList(cmds, capa),
		cap:    capa,
		width:  defaultPickerWidth,
		height: defaultPickerHeight,
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
	if size, isResize := msg.(tea.WindowSizeMsg); isResize {
		m.width, m.height = size.Width, size.Height
		m.fit()
		return m, nil
	}

	key, isKey := msg.(tea.KeyMsg)
	if isKey && m.list.FilterState() != list.Filtering {
		switch key.String() {
		case "ctrl+c", "q":
			m.cancelled = true
			return m, tea.Quit
		case "esc":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			if item, ok := m.list.SelectedItem().(PickerItem); ok {
				m.selected = item.Cmd
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m *pickerModel) View() string {
	return m.cap.Header(m.width, m.height) + "\n" + m.list.View()
}

// fit sizes the list to the terminal minus the header.
func (m *pickerModel) fit() {
	used := lipgloss.Height(m.cap.Header(m.width, m.height))
	m.list.SetSize(m.width, max(m.height-used, minPickerListHeight))
}

// RunPicker launches the interactive command picker over cmds and, on a
// selection, either runs it (when it takes no positional args) or prints its
// help (when it does, since huh-based prompting inside each command already
// covers the args it needs — the picker does not try to collect them itself).
// Returns ui.ErrAborted if the user cancels.
func RunPicker(cmds []*cobra.Command, out io.Writer) error {
	program := NewPicker(cmds, os.Stdin, out)

	final, err := program.Run()
	if err != nil {
		return fmt.Errorf("run picker: %w", err)
	}

	result, ok := final.(*pickerModel)
	if !ok || result.cancelled || result.selected == nil {
		return ErrAborted
	}

	selected := result.selected
	if len(selected.Commands()) > 0 || takesArgs(selected) {
		return selected.Help()
	}

	if selected.RunE != nil {
		return selected.RunE(selected, nil)
	}
	if selected.Run != nil {
		selected.Run(selected, nil)
		return nil
	}
	return selected.Help()
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
