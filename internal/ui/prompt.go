package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Ask prints a styled prompt and reads one trimmed line from stdin.
// Returns defaultVal if the user presses Enter with no input.
func Ask(w io.Writer, label, defaultVal string) (string, error) {
	if defaultVal != "" {
		fmt.Fprintf(w, "\n"+Padding+"%s %s: ", PromptStyle.Render(label), MutedStyle.Render("["+defaultVal+"]"))
	} else {
		fmt.Fprintf(w, "\n"+Padding+"%s: ", PromptStyle.Render(label))
	}
	reader := bufio.NewReader(os.Stdin)
	raw, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read input: %w", err)
	}
	input := strings.TrimSpace(raw)
	if input == "" {
		return defaultVal, nil
	}
	return input, nil
}

// Select prints a numbered list of choices and prompts for a selection.
// Returns the 0-based index; falls back to defaultIdx on invalid input.
func Select(w io.Writer, label string, choices []string, defaultIdx int) (int, error) {
	fmt.Fprintln(w)
	for i, choice := range choices {
		if i == defaultIdx {
			fmt.Fprintf(w, Padding+"> [%d] %s\n", i+1, BoldStyle.Render(choice))
		} else {
			fmt.Fprintf(w, Padding+"  [%d] %s\n", i+1, choice)
		}
	}
	selection, err := Ask(w, label, strconv.Itoa(defaultIdx+1))
	if err != nil {
		return 0, err
	}
	picked, parseErr := strconv.Atoi(selection)
	if parseErr != nil || picked < 1 || picked > len(choices) {
		return defaultIdx, nil
	}
	return picked - 1, nil
}
