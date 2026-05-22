package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// ConfirmYN prints a styled [y/N] prompt and returns true for "y" or "yes".
func ConfirmYN(w io.Writer, prompt string) bool {
	fmt.Fprintf(w, "\n"+Padding+"%s %s ", PromptStyle.Render(prompt), MutedStyle.Render("[y/N]"))
	
	r := bufio.NewReader(os.Stdin)
	line, _ := r.ReadString('\n')
	fmt.Fprintln(w)
	
	line = strings.ToLower(strings.TrimSpace(strings.TrimRight(line, "\r\n")))
	return line == "y" || line == "yes"
}
