package ui

import (
	"bufio"
	"cmp"
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

const (
	// backRowTitle labels the row that returns to the parent level.
	backRowTitle = ".. back"

	// breadcrumbSep joins the level names in the list title.
	breadcrumbSep = " › "

	// singleChildLevelRows is the row count (back row + one child) of a level
	// with exactly one visible child.
	singleChildLevelRows = 2
)

// keyCtrlC is the abort key, checked in enough places (the picker, the form,
// the session's running/result screens) to warrant a shared constant.
const keyCtrlC = "ctrl+c"

// pickItem is one selectable row in the picker: either a domain to descend
// into (cmd has children) or a leaf to run (cmd is runnable, no children),
// plus the back row.
type pickItem struct {
	title  string
	desc   string
	cmd    *cobra.Command
	isBack bool
}

// Title implements list.Item.
func (i pickItem) Title() string { return i.title }

// Description implements list.Item.
func (i pickItem) Description() string { return i.desc }

// FilterValue implements list.Item.
func (i pickItem) FilterValue() string { return i.title }

// backRow is the row that returns to the parent level.
func backRow() pickItem {
	return pickItem{title: backRowTitle, desc: "back to the previous level", isBack: true}
}

// childLevel builds the rows shown for cmds: a back row (unless this is the
// root level), then one row per visible, runnable-or-nested child, ordered by
// cobra group (in the order the parent declared them, ungrouped last) and
// alphabetically within a group.
func childLevel(cmds []*cobra.Command, isRoot bool) []pickItem {
	items := make([]pickItem, 0, len(cmds)+1)
	if !isRoot {
		items = append(items, backRow())
	}
	rank := groupRanks(cmds)
	rows := make([]pickItem, 0, len(cmds))
	for _, c := range cmds {
		if c.Hidden || (!c.Runnable() && len(c.Commands()) == 0) {
			continue
		}
		if isRoot && isCobraBuiltin(c) {
			continue
		}
		rows = append(rows, pickItem{title: c.Name(), desc: c.Short, cmd: c})
	}
	slices.SortStableFunc(rows, func(a, b pickItem) int {
		return cmp.Compare(rank(a.cmd), rank(b.cmd))
	})
	return append(items, rows...)
}

// groupRanks returns a function mapping a command to the position of its
// cobra group among the parent's declared groups; ungrouped commands (and
// commands without a parent) rank after every group.
func groupRanks(cmds []*cobra.Command) func(*cobra.Command) int {
	var order []*cobra.Group
	if len(cmds) > 0 && cmds[0].HasParent() {
		order = cmds[0].Parent().Groups()
	}
	return func(c *cobra.Command) int {
		i := slices.IndexFunc(order, func(g *cobra.Group) bool { return g.ID == c.GroupID })
		if i < 0 || c.GroupID == "" {
			return len(order)
		}
		return i
	}
}

// isCobraBuiltin reports whether c is one of cobra's auto-generated top-level
// commands (help, completion), which are noise in the picker.
func isCobraBuiltin(c *cobra.Command) bool {
	return c.Name() == "help" || c.Name() == "completion"
}

// commandFinishedMsg reports the outcome of the most recently run command, so
// the menu can surface it without interrupting the dashboard loop.
type commandFinishedMsg struct {
	name string
	err  error
}

// pickerLevel is one screen of the picker's navigation stack. name is the
// breadcrumb segment that led here ("" for the root); cursor is the list index
// of the row the user descended from, restored when they come back.
type pickerLevel struct {
	name   string
	items  []pickItem
	cursor int
}

// pickerModel drives the command list as a stack of levels: the root lists
// top-level commands, descending into a domain shows its children. Enter on
// a leaf sets m.selected and quits so the session can decide what to do next
// (run directly, or collect its parameters first); ctrl+c/q/esc leave the
// picker.
type pickerModel struct {
	stack         []pickerLevel // stack[0] is the root; the last entry is on screen
	list          list.Model
	cancelled     bool
	selected      *cobra.Command
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
	model := newSessionModel(cmds, notice)
	return tea.NewProgram(model, tea.WithAltScreen(), tea.WithInput(in), tea.WithOutput(out))
}

func newPickerModel(cmds []*cobra.Command, notice string) *pickerModel {
	capa := NewCapability()
	m := &pickerModel{
		stack:  []pickerLevel{{items: childLevel(cmds, true)}},
		cap:    capa,
		width:  defaultPickerWidth,
		height: defaultPickerHeight,
		notice: notice,
	}
	m.show()
	return m
}

func newPickerList(items []pickItem, capa Capability) list.Model {
	entries := make([]list.Item, 0, len(items))
	for _, item := range items {
		entries = append(entries, item)
	}

	model := list.New(entries, pickerDelegate(capa), defaultPickerWidth, defaultPickerHeight)
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
	if !capa.Color {
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
	if !capa.Color {
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
//
//nolint:ireturn // signature dictated by the tea.Model interface
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
			case keyCtrlC:
				m.cancelled = true
				return m, tea.Quit
			case "q":
				return m, tea.Quit
			case "esc":
				if m.canPop() {
					m.pop()
					return m, nil
				}
				return m, tea.Quit
			case "enter":
				if item, ok := m.list.SelectedItem().(pickItem); ok {
					return m.choose(item)
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
	view := m.cap.Header(rskBannerSpec, m.width, m.height)
	if m.notice != "" {
		view += "\n" + m.cap.Paint(m.notice, lipgloss.NewStyle().Foreground(ColorWarning).Bold(true))
	}
	view += "\n" + m.list.View()
	if m.lastRun == "" {
		return view
	}
	return view + "\n" + m.cap.LastRunLine(m.lastRun, m.lastErr)
}

// fit sizes the list to the terminal minus the header (and notice, if any).
func (m *pickerModel) fit() {
	used := lipgloss.Height(m.cap.Header(rskBannerSpec, m.width, m.height))
	if m.notice != "" {
		used++
	}
	m.list.SetSize(m.width, max(m.height-used, minPickerListHeight))
}

// choose descends into a domain, pops on the back row, or selects a leaf and
// quits so the session can act on it.
//
//nolint:ireturn // return type mirrors Update's, which tea.Model dictates
func (m *pickerModel) choose(item pickItem) (tea.Model, tea.Cmd) {
	switch {
	case item.isBack:
		m.pop()
		return m, nil
	case len(item.cmd.Commands()) > 0:
		m.descend(item)
		return m, nil
	default:
		m.selected = item.cmd
		return m, tea.Quit
	}
}

// current returns the level on screen.
func (m *pickerModel) current() *pickerLevel { return &m.stack[len(m.stack)-1] }

// canPop reports whether popping is possible (not at the root).
func (m *pickerModel) canPop() bool { return len(m.stack) > 1 }

// descend pushes item's children as a new level. A group whose only visible
// child is itself a group (claude -> tools) is collapsed: the levels in
// between are skipped and the breadcrumb names the whole chain.
func (m *pickerModel) descend(item pickItem) {
	names := []string{item.title}
	items := childLevel(item.cmd.Commands(), false)
	for len(items) == singleChildLevelRows && items[1].cmd != nil && len(items[1].cmd.Commands()) > 0 {
		names = append(names, items[1].title)
		items = childLevel(items[1].cmd.Commands(), false)
	}
	m.current().cursor = m.list.GlobalIndex() // unfiltered position, valid even when descending from a filtered view
	m.stack = append(m.stack, pickerLevel{name: strings.Join(names, breadcrumbSep), items: items})
	m.show()
}

// pop returns to the previous level, restoring the cursor to the row the user
// descended from.
func (m *pickerModel) pop() {
	if !m.canPop() {
		return
	}
	m.stack = m.stack[:len(m.stack)-1]
	m.show()
	m.list.Select(m.current().cursor)
}

// show rebuilds the list for the current level and refreshes its sizing and
// breadcrumb title; it is the only place the list is (re)built.
func (m *pickerModel) show() {
	m.list = newPickerList(m.current().items, m.cap)
	m.fit()
	m.updateTitle()
}

// updateTitle refreshes the list title with the breadcrumb.
func (m *pickerModel) updateTitle() {
	title := "rsk — pick a command"
	if m.canPop() {
		names := make([]string, 0, len(m.stack)-1)
		for _, level := range m.stack[1:] {
			names = append(names, level.name)
		}
		title = "rsk — " + strings.Join(names, breadcrumbSep)
	}
	m.list.Title = title
}

// commandExec adapts a cobra.Command to tea.ExecCommand so tea.Exec can run
// it with the terminal released, then pauses for a keypress before handing
// the terminal back to the picker so the user has time to read the output.
// args carries the positional arguments resolved by the session's parameter
// form, if any.
type commandExec struct {
	cmd         *cobra.Command
	args        []string
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
	// RunE is invoked directly below, bypassing cobra's own Find/execute
	// path that normally propagates the root's context onto whichever
	// command it dispatches to — so c.cmd.Context() would otherwise stay
	// nil and any RunE that calls cmd.Context() (e.g. to build an HTTP
	// request) would panic.
	if c.cmd.Context() == nil {
		ctx := c.cmd.Root().Context()
		if ctx == nil {
			ctx = context.Background()
		}
		c.cmd.SetContext(ctx)
	}

	fmt.Fprintf(c.out, "\n$ rsk %s\n\n", c.cmd.Name())
	err := c.runCommand()
	if err != nil {
		fmt.Fprintf(c.out, "\nrsk: %s\n", err)
	}

	fmt.Fprint(c.out, "\n(press enter to return to the menu)")
	//nolint:errcheck // just waiting for Enter; a closed/EOF stdin here doesn't change whether the command itself succeeded
	bufio.NewReader(c.in).ReadString('\n')
	return err
}

// runCommand runs c.cmd with c.args — the session only ever hands it a leaf
// command, resolved through the parameter form when it needed one.
func (c *commandExec) runCommand() error {
	if c.cmd.RunE != nil {
		return c.cmd.RunE(c.cmd, c.args)
	}
	if c.cmd.Run != nil {
		c.cmd.Run(c.cmd, c.args)
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
		if trimmed == flagsPlaceholder {
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

	if result, ok := final.(*sessionModel); ok && result.picker.cancelled {
		return ErrAborted
	}
	return nil
}
