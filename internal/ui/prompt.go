package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
)

// stdinReader is shared across all plain-text prompt calls so that buffered
// look-ahead from one read is not lost for the next read on the same stdin.
var stdinReader = bufio.NewReader(os.Stdin)

// Ask prints a styled prompt and reads one trimmed line from stdin.
// Returns defaultVal if the user presses Enter with no input.
// Uses a huh Input field when running in a TTY; falls back to plain text otherwise.
func Ask(w io.Writer, label, defaultVal string) (string, error) {
	if IsTTY() {
		return askHuh(label, defaultVal)
	}
	return askPlain(w, label, defaultVal)
}

// Select prints a numbered list of choices and prompts for a selection.
// Returns the 0-based index; falls back to defaultIdx on invalid input.
// Uses a huh Select field when running in a TTY; falls back to plain text otherwise.
func Select(w io.Writer, label string, choices []string, defaultIdx int) (int, error) {
	if IsTTY() {
		return selectHuh(label, choices, defaultIdx)
	}
	return selectPlain(w, label, choices, defaultIdx)
}

// MultiSelect presents a list of choices and returns the 0-based indices of
// all selected items. defaultIndices is pre-selected when the user accepts
// the prompt without changing anything.
// Uses a huh MultiSelect field when running in a TTY; falls back to plain text otherwise.
func MultiSelect(w io.Writer, label string, choices []string, defaultIndices []int) ([]int, error) {
	if IsTTY() {
		return multiSelectHuh(label, choices, defaultIndices)
	}
	return multiSelectPlain(w, label, choices, defaultIndices)
}

// — TTY paths (huh) —

func askHuh(label, defaultVal string) (string, error) {
	result := defaultVal
	f := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title(label).
			Value(&result),
	)).WithTheme(huh.ThemeCharm())
	if err := f.Run(); err != nil {
		return "", huhErr(err)
	}
	return result, nil
}

func selectHuh(label string, choices []string, defaultIdx int) (int, error) {
	options := make([]huh.Option[int], len(choices))
	for i, choice := range choices {
		options[i] = huh.NewOption(choice, i)
	}
	result := defaultIdx
	f := huh.NewForm(huh.NewGroup(
		huh.NewSelect[int]().
			Title(label).
			Options(options...).
			Value(&result),
	)).WithTheme(huh.ThemeCharm())
	if err := f.Run(); err != nil {
		return 0, huhErr(err)
	}
	return result, nil
}

func multiSelectHuh(label string, choices []string, defaultIndices []int) ([]int, error) {
	options := make([]huh.Option[int], len(choices))
	for i, choice := range choices {
		options[i] = huh.NewOption(choice, i)
	}
	result := make([]int, len(defaultIndices))
	copy(result, defaultIndices)
	f := huh.NewForm(huh.NewGroup(
		huh.NewMultiSelect[int]().
			Title(label).
			Options(options...).
			Value(&result),
	)).WithTheme(huh.ThemeCharm())
	if err := f.Run(); err != nil {
		return nil, huhErr(err)
	}
	return result, nil
}

// huhErr converts huh.ErrUserAborted to ErrAborted so callers and root.Execute
// can distinguish a user cancellation from a real failure.
func huhErr(err error) error {
	if errors.Is(err, huh.ErrUserAborted) {
		return ErrAborted
	}
	return err
}

// — non-TTY paths (plain bufio) —

func askPlain(w io.Writer, label, defaultVal string) (string, error) {
	if defaultVal != "" {
		_, _ = fmt.Fprintf(w, "\n"+Padding+"%s %s: ", PromptStyle.Render(label), MutedStyle.Render("["+defaultVal+"]"))
	} else {
		_, _ = fmt.Fprintf(w, "\n"+Padding+"%s: ", PromptStyle.Render(label))
	}
	raw, err := stdinReader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read input: %w", err)
	}
	input := strings.TrimSpace(raw)
	if input == "" {
		return defaultVal, nil
	}
	return input, nil
}

func selectPlain(w io.Writer, label string, choices []string, defaultIdx int) (int, error) {
	fmt.Fprintln(w)
	for i, choice := range choices {
		if i == defaultIdx {
			fmt.Fprintf(w, Padding+"> [%d] %s\n", i+1, BoldStyle.Render(choice))
		} else {
			fmt.Fprintf(w, Padding+"  [%d] %s\n", i+1, choice)
		}
	}
	selection, err := askPlain(w, label, strconv.Itoa(defaultIdx+1))
	if err != nil {
		return 0, err
	}
	picked, err := strconv.Atoi(selection)
	if err == nil && picked >= 1 && picked <= len(choices) {
		return picked - 1, nil
	}
	return defaultIdx, nil
}

func multiSelectPlain(w io.Writer, label string, choices []string, defaultIndices []int) ([]int, error) {
	fmt.Fprintln(w)
	defaults := make([]string, len(defaultIndices))
	for i, idx := range defaultIndices {
		defaults[i] = strconv.Itoa(idx + 1)
	}
	for i, choice := range choices {
		fmt.Fprintf(w, Padding+"  [%d] %s\n", i+1, choice)
	}
	selection, err := askPlain(w, label, strings.Join(defaults, ","))
	if err != nil {
		return nil, err
	}
	return parseMultiSelectIndices(selection, len(choices)), nil
}

func parseMultiSelectIndices(s string, count int) []int {
	seen := make(map[int]bool)
	var result []int
	for part := range strings.SplitSeq(s, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || n < 1 || n > count || seen[n] {
			continue
		}
		seen[n] = true
		result = append(result, n-1)
	}
	return result
}
