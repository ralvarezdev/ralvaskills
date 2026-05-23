package manifest

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ralvarezdev/ralvaskills/internal/skill"
)

const (
	// ProjectFolderName is the name of the project-local rsk directory that
	// marks a directory as an rsk project root.
	ProjectFolderName = ".rsk"

	// ClaudeFileName is the name of the generated file inside .rsk/ that lists
	// pinned skills for Claude Code to import.
	ClaudeFileName = "CLAUDE.md"

	// OpenCodeFileName is the name of the OpenCode configuration file written
	// inside the project root by rsk.
	OpenCodeFileName = "opencode.json"

	// ModFileName is the filename of the project skill manifest.
	ModFileName = "rsk.mod"

	// LockFileName is the filename of the resolved skill lockfile.
	LockFileName = "rsk.lock"

	// TempFilePattern is the glob pattern for temporary files created during
	// atomic write operations inside the .rsk directory.
	TempFilePattern = ".rsk-*.tmp"
)

var (
	// ProjectSkillsPrefix is the forward-slash path prefix for .rsk/skills/
	// entries. Uses "/" rather than filepath.Separator because opencode.json
	// stores Unix-style paths regardless of the host OS.
	ProjectSkillsPrefix = ProjectFolderName + "/" + skill.SkillsFolderName + "/"

	// RelativeClaudeFilePath is the path of CLAUDE.md relative to the project
	// root (.rsk/CLAUDE.md).
	RelativeClaudeFilePath = filepath.Join(ProjectFolderName, ClaudeFileName)
)

// ProjectFolderPath returns the .rsk directory path for the current working
// directory. Returns an error if rsk.mod is not found, indicating the directory
// has not been initialized as an rsk project.
func ProjectFolderPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	rskDir := filepath.Join(cwd, ProjectFolderName)
	if _, statErr := os.Stat(ModPath(rskDir)); os.IsNotExist(statErr) {
		return "", fmt.Errorf("no rsk.mod found — run 'rsk project init' first")
	}
	return rskDir, nil
}

// ProjectSkillsPath returns the canonical path of the project-local skills
// directory (.rsk/skills/).
func ProjectSkillsPath(rskDir string) string {
	return filepath.Join(rskDir, skill.SkillsFolderName)
}

// LockPath returns the canonical path of rsk.lock inside rskDir.
func LockPath(rskDir string) string {
	return filepath.Join(rskDir, LockFileName)
}

// ModPath returns the canonical path of rsk.mod inside rskDir.
func ModPath(rskDir string) string {
	return filepath.Join(rskDir, ModFileName)
}

// ClaudeConfigFilePath returns the canonical path of CLAUDE.md inside rskDir.
func ClaudeConfigFilePath(rskDir string) string {
	return filepath.Join(rskDir, ClaudeFileName)
}

// OpenCodeConfigFilePath returns the canonical path of opencode.json inside
// rskDir.
func OpenCodeConfigFilePath(rskDir string) string {
	return filepath.Join(rskDir, OpenCodeFileName)
}
