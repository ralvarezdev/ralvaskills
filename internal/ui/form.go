package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// argsFieldName is the reserved values() key for the positional-arguments
// field, distinct from any real flag name.
const argsFieldName = "\x00args"

// flagsPlaceholder is the "[flags]" token in a cobra Use string, which is
// not itself a positional argument.
const flagsPlaceholder = "flags"

// defaultArgsLabel is the fallback label for a positional-args field whose
// bracket text couldn't be pulled out of the command's Use string.
const defaultArgsLabel = "args"

const (
	formCharLimit         = 256
	defaultFormInputWidth = 40
	minFormInputWidth     = 20
	formWidthMargin       = 8
)

// formField is one prompt in the parameter form: a text input, or a
// checkbox for a bool flag. key is the pflag name to Set (or argsFieldName
// for the positional field); label is what's shown on screen.
type formField struct {
	key      string
	label    string
	help     string
	isBool   bool
	boolVal  bool
	required bool
	input    textinput.Model
}

// newTextField creates a text field seeded with value.
func newTextField(key, label, help, value string, required bool) *formField {
	input := textinput.New()
	input.Placeholder = help
	input.CharLimit = formCharLimit
	input.Width = defaultFormInputWidth
	input.SetValue(value)

	return &formField{key: key, label: label, help: help, required: required, input: input}
}

// newBoolField creates a checkbox field starting at value.
func newBoolField(key, label, help string, value bool) *formField {
	return &formField{key: key, label: label, help: help, isBool: true, boolVal: value}
}

// commandFields builds the parameter form's fields from cmd's local flags
// plus one positional-args field when cmd takes args. It returns nil when
// there is nothing to ask, so the session can skip the form entirely.
func commandFields(cmd *cobra.Command) []*formField {
	var fields []*formField

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Name == "help" {
			return
		}
		label := "--" + flag.Name
		if flag.Value.Type() == "bool" {
			fields = append(fields, newBoolField(flag.Name, label, flag.Usage, flag.DefValue == "true"))
			return
		}
		fields = append(fields, newTextField(flag.Name, label, flag.Usage, flag.DefValue, false))
	})

	if takesArgs(cmd) {
		label := argsLabel(cmd)
		fields = append(fields, newTextField(argsFieldName, label, "positional arguments", "", argsRequired(cmd)))
	}

	return fields
}

// argsLabel pulls the bracketed argument text out of cmd's Use string, e.g.
// "install [name...] [flags]" -> "name...", falling back to a generic label
// if none is found.
func argsLabel(cmd *cobra.Command) string {
	for field := range strings.FieldsSeq(cmd.Use) {
		trimmed := strings.Trim(field, "[]<>")
		if trimmed == flagsPlaceholder || trimmed == "" {
			continue
		}
		if strings.ContainsAny(field, "[]<>") {
			return trimmed
		}
	}
	return defaultArgsLabel
}

// argsRequired reports whether cmd's positional argument is required, based
// on its Use string using <...> rather than [...] for the bracket.
func argsRequired(cmd *cobra.Command) bool {
	for field := range strings.FieldsSeq(cmd.Use) {
		trimmed := strings.Trim(field, "[]<>")
		if trimmed == flagsPlaceholder || trimmed == "" {
			continue
		}
		if strings.HasPrefix(field, "<") {
			return true
		}
	}
	return false
}

// formModel drives the parameter form after a leaf command is chosen, ending
// in a confirmation step so a stray enter never runs anything by accident.
type formModel struct {
	title     string
	fields    []*formField
	index     int
	submit    bool
	cancelled bool
	err       string
	confirm   bool
	cap       Capability
	width     int
}

// newFormModel builds the parameter form for title (the command's full
// path) with fields.
func newFormModel(capa Capability, title string, fields []*formField) *formModel {
	model := &formModel{title: title, fields: fields, cap: capa, width: defaultPickerWidth}
	model.focus(0)
	return model
}

// Init implements tea.Model.
func (m *formModel) Init() tea.Cmd { return textinput.Blink }

// Update implements tea.Model.
//
//nolint:ireturn // signature dictated by the tea.Model interface
func (m *formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, isResize := msg.(tea.WindowSizeMsg); isResize {
		m.width = size.Width
		m.resizeInputs()
		return m, nil
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		if handled, cmd := m.handleKey(key); handled {
			return m, cmd
		}
	}

	if len(m.fields) == 0 {
		return m, nil
	}

	var cmd tea.Cmd
	m.fields[m.index].input, cmd = m.fields[m.index].input.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m *formModel) View() string {
	lines := []string{m.cap.accent(m.title, true), ""}

	for i, field := range m.fields {
		cursor := "  "
		if i == m.index {
			cursor = m.cap.accent("> ", true)
		}

		if field.isBool {
			mark := " "
			if field.boolVal {
				mark = "x"
			}
			lines = append(lines, cursor+field.label+"  ["+mark+"]  "+m.cap.muted(field.help))
			continue
		}

		lines = append(lines, cursor+field.label, "    "+field.input.View())
	}

	if m.err != "" {
		lines = append(lines, "", m.cap.paint(m.err, lipgloss.NewStyle().Foreground(ColorDanger)))
	}

	if m.confirm {
		lines = append(lines, "", "Will run:  "+m.cap.accent(m.preview(), true), "",
			m.cap.muted("(enter: run · esc: back to the menu)"))
		return strings.Join(lines, "\n")
	}

	lines = append(lines, "", m.cap.muted("(tab: next · space: toggle · enter: next · esc: back to the menu)"))
	return strings.Join(lines, "\n")
}

// preview renders the invocation the current field values resolve to:
// "rsk <path> <args> <flags>".
func (m *formModel) preview() string {
	parts := []string{"rsk", m.title}

	for _, field := range m.fields {
		if field.key != argsFieldName {
			continue
		}
		if args := strings.TrimSpace(field.input.Value()); args != "" {
			parts = append(parts, args)
		}
	}

	for _, field := range m.fields {
		if field.key == argsFieldName {
			continue
		}
		if field.isBool {
			if field.boolVal {
				parts = append(parts, field.label)
			}
			continue
		}
		if value := strings.TrimSpace(field.input.Value()); value != "" {
			parts = append(parts, field.label, value)
		}
	}

	return strings.Join(parts, " ")
}

// resizeInputs keeps the text inputs as wide as the terminal allows.
func (m *formModel) resizeInputs() {
	width := max(m.width-formWidthMargin, minFormInputWidth)
	for _, field := range m.fields {
		field.input.Width = width
	}
}

// handleKey processes navigation and toggles. It reports whether the key was
// consumed by the form rather than the focused input.
func (m *formModel) handleKey(key tea.KeyMsg) (bool, tea.Cmd) {
	switch key.String() {
	case "ctrl+c", "esc":
		m.cancelled = true
		return true, tea.Quit
	case "tab", "down":
		m.focus(m.index + 1)
		return true, nil
	case "shift+tab", "up":
		m.focus(m.index - 1)
		return true, nil
	case "enter":
		return m.enter()
	}

	if len(m.fields) > 0 && m.fields[m.index].isBool &&
		(key.String() == "left" || key.String() == "right" || key.String() == " ") {
		m.fields[m.index].boolVal = !m.fields[m.index].boolVal
		return true, nil
	}

	return false, nil
}

// enter advances a field, or on the last one validates and opens the
// confirmation, or on the confirmation submits.
func (m *formModel) enter() (bool, tea.Cmd) {
	if m.confirm {
		m.submit = true
		return true, tea.Quit
	}

	if m.index < len(m.fields)-1 {
		m.focus(m.index + 1)
		return true, nil
	}

	if err := m.validate(); err != nil {
		m.err = err.Error()
		return true, nil
	}

	m.confirm = true
	m.err = ""
	return true, nil
}

// focus moves the highlight to index, wrapping around.
func (m *formModel) focus(index int) {
	if len(m.fields) == 0 {
		return
	}
	if index < 0 {
		index = len(m.fields) - 1
	}
	if index >= len(m.fields) {
		index = 0
	}

	for i, field := range m.fields {
		if field.isBool {
			continue
		}
		if i == index {
			field.input.Focus()
		} else {
			field.input.Blur()
		}
	}
	m.index = index
}

// validate enforces required fields; anything type-specific (e.g. an int
// flag) surfaces as a pflag Set error when the session applies the values.
func (m *formModel) validate() error {
	for _, field := range m.fields {
		if field.isBool {
			continue
		}
		if field.required && strings.TrimSpace(field.input.Value()) == "" {
			return fmt.Errorf("%s is required", field.label)
		}
	}
	return nil
}

// values collects the form state as a flag-name -> value map, keyed by
// argsFieldName for the positional field.
func (m *formModel) values() map[string]string {
	values := make(map[string]string, len(m.fields))
	for _, field := range m.fields {
		if field.isBool {
			values[field.key] = strconv.FormatBool(field.boolVal)
			continue
		}
		values[field.key] = strings.TrimSpace(field.input.Value())
	}
	return values
}

// splitArgs splits a free-text argument line on spaces, honoring quotes.
func splitArgs(raw string) []string {
	var (
		args  []string
		buf   []rune
		quote rune
	)

	flush := func() {
		if len(buf) > 0 {
			args = append(args, string(buf))
			buf = nil
		}
	}

	for _, char := range raw {
		switch {
		case quote != 0:
			if char == quote {
				quote = 0
			} else {
				buf = append(buf, char)
			}
		case char == '\'' || char == '"':
			quote = char
		case char == ' ' || char == '\t':
			flush()
		default:
			buf = append(buf, char)
		}
	}
	flush()

	return args
}
