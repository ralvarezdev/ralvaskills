package tool

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ralvarezdev/ralvaskills/v3/internal/fsx"
)

// AgentsFileName is the tool-neutral agent guide at the project root. Claude
// Code needs it imported from CLAUDE.md, because it only loads AGENTS.md on its
// own when no CLAUDE.md exists. OpenCode and goose read it natively.
const AgentsFileName = "AGENTS.md"

// agentsImportLine is the import CLAUDE.md uses to load the agent guide. Claude
// Code resolves it relative to CLAUDE.md, so both files must stay at the root.
const agentsImportLine = "@" + AgentsFileName

// agentsSkeleton is the starting point written when a project has no AGENTS.md.
// The section order is deliberate: what the repo is, how to run it, where
// things live, what is non-obvious, what is off limits.
const agentsSkeleton = "# %s — agent guide\n\n" +
	"One paragraph on what this repository is, who it serves, and what is out of scope.\n\n" +
	"## Commands\n\n" +
	"- Build:\n" +
	"- Test:\n" +
	"- Lint:\n\n" +
	"## Layout\n\n" +
	"- `dir/`: what it owns.\n\n" +
	"## Conventions\n\n" +
	"- A rule a good engineer would still get wrong from the code alone.\n\n" +
	"## Boundaries\n\n" +
	"- Always:\n" +
	"- Ask first:\n" +
	"- Never:\n"

// ensureAgentsFile creates a skeleton AGENTS.md at the project root when the
// project has none. An existing file is never touched: it is the human-written
// source of truth, and rsk has nothing to add to it.
func ensureAgentsFile(projectDir string) error {
	path := filepath.Join(projectDir, AgentsFileName)
	if _, err := os.Lstat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat %s: %w", path, err)
	}

	name := filepath.Base(projectDir)
	if name == "." || name == string(filepath.Separator) {
		name = "project"
	}
	return fsx.WriteAtomic(path, rskTempPattern, func(w io.Writer) error {
		_, err := fmt.Fprintf(w, agentsSkeleton, name)
		return err
	})
}

// ensureAgentsImport makes @AGENTS.md the first line of the project's
// CLAUDE.md, so Claude Code loads the agent guide before any Claude-specific
// content. Existing content keeps its order below that line, and a file that
// already imports AGENTS.md anywhere is left alone.
func ensureAgentsImport(claudeMDPath string) error {
	data, err := os.ReadFile(claudeMDPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", claudeMDPath, err)
	}
	if hasMarkdownLine(string(data), agentsImportLine) {
		return nil
	}

	if err = refuseIfLink(claudeMDPath); err != nil {
		return err
	}

	return fsx.WriteAtomic(claudeMDPath, rskTempPattern, func(w io.Writer) error {
		if _, writeErr := fmt.Fprintln(w, agentsImportLine); writeErr != nil {
			return writeErr
		}
		_, writeErr := w.Write(data)
		return writeErr
	})
}

// hasMarkdownLine reports whether s contains line as a whole line, ignoring
// surrounding whitespace.
func hasMarkdownLine(s, line string) bool {
	for l := range strings.SplitSeq(s, "\n") {
		if strings.TrimSpace(l) == line {
			return true
		}
	}
	return false
}
