package ui

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ralvarezdev/termkit"
	"github.com/ralvarezdev/termkit/form"
	"github.com/ralvarezdev/termkit/session"
)

// appName is the binary's name: the session's title, and the directory name
// termkit's history and prefs stores live under.
const appName = "rsk"

// SessionOptions configures RunSession.
type SessionOptions struct {
	// Root is the cobra root whose subtree the picker lists.
	Root *cobra.Command
	// Runnable decides which commands the picker offers (nil offers every
	// runnable, visible command).
	Runnable func(*cobra.Command) bool
	// Notice is an optional one-line notice (e.g. updates available) shown in
	// the picker title and the help overlay.
	Notice string
}

// RunSession opens the interactive termkit session over opts.Root: a picker,
// a typed parameter form (flags and positional arguments), an in-process
// captured run, and the paged result. Recently used commands float to the top
// and form values are remembered between runs when the stores under the user
// config dir can be opened.
func RunSession(ctx context.Context, opts SessionOptions) error {
	historyPath, prefsPath := defaultStorePaths()
	if err := session.New(newSessionConfig(opts, historyPath, prefsPath)).Run(ctx, false); err != nil {
		return fmt.Errorf("run session: %w", err)
	}
	return nil
}

// newSessionConfig builds the session configuration for opts. Empty store
// paths (or stores that fail to open) simply leave history and prefs off; the
// session never depends on them.
func newSessionConfig(opts SessionOptions, historyPath, prefsPath string) session.Config {
	title := appName + " — pick a command"
	if opts.Notice != "" {
		title += " · " + opts.Notice
	}

	cfg := session.Config{
		Root:        opts.Root,
		Run:         RunCaptured(opts.Root),
		Runnable:    opts.Runnable,
		Title:       appName,
		PickerTitle: title,
		Help:        helpMarkdown(opts.Root, opts.Notice),
		RowAction:   SessionRowAction,
		Colors:      Theme,
		Fields:      form.Options{}.WithMultiNames("include"),
	}
	attachStores(&cfg, historyPath, prefsPath)
	return cfg
}

// helpMarkdown is the `?` overlay's Markdown: a short about section, the
// notice, and one bullet per visible command under its cobra group heading.
func helpMarkdown(root *cobra.Command, notice string) string {
	var b strings.Builder
	b.WriteString("## " + appName + "\n\n")
	b.WriteString(root.Long + "\n\n")
	if notice != "" {
		b.WriteString("**" + notice + "**\n\n")
	}

	for _, group := range root.Groups() {
		var lines []string
		for _, cmd := range root.Commands() {
			if cmd.GroupID == group.ID && !cmd.Hidden {
				lines = append(lines, fmt.Sprintf("- %#q — %s", cmd.Name(), cmd.Short))
			}
		}
		if len(lines) == 0 {
			continue
		}
		b.WriteString("### " + strings.TrimSuffix(group.Title, ":") + "\n\n")
		b.WriteString(strings.Join(lines, "\n") + "\n\n")
	}
	return b.String()
}

// RunCaptured returns the session's RunFunc: it executes the selected command
// through root (so cobra parses the form's argv and validates its arguments
// exactly as on the command line), with a termkit.Capture in the context so
// table views record their tables. Text a command writes to its writers is
// added to the result as a message, except for table views, whose text output
// the captured table replaces; a raw payload (-o json) becomes a message too,
// because the session shows messages and pages, not raw payloads.
func RunCaptured(root *cobra.Command) session.RunFunc {
	return func(ctx context.Context, target *cobra.Command, argv []string) (*termkit.Capture, error) {
		capture := &termkit.Capture{}
		var out bytes.Buffer

		ctx = WithCapture(ctx, capture)
		termkit.ResetFlags(target)
		// cobra hands a command the run's context only while it has none, and
		// ResetFlags leaves it with a stale one.
		target.SetContext(ctx)
		root.SetOut(&out)
		root.SetErr(&out)
		root.SetArgs(argv[1:])
		defer func() {
			root.SetOut(nil)
			root.SetErr(nil)
			root.SetArgs(nil)
			// ResetFlags also drops the run's context, and with it the capture.
			termkit.ResetFlags(target)
		}()

		if err := root.ExecuteContext(ctx); err != nil {
			return nil, err
		}

		if text := strings.TrimSpace(out.String()); text != "" && !IsTableView(target) {
			capture.AddMessage(text)
		}
		return capture, nil
	}
}
