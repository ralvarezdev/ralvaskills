package manifest

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ralvarezdev/ralvaskills/internal/skill"
)

const (
	// LocalConfigFolderName is the name of the folder in the local source where RSK configuration files are stored.
	LocalConfigFolderName = ".rsk"

	// ClaudeFileName is the name of the file in .rsk that lists pinned skills to import.
	ClaudeFileName = "CLAUDE.md"

	// OpenCodeConfigFileName is the name of the configuration file for OpenCode, which may be used to specify settings related to skill development and management.
	OpenCodeConfigFileName = "opencode.json"

	// ModManifestFile is the name of the manifest file that defines project-specific skills and their versions.
	ModManifestFile = "rsk.mod"

	// LocalManifestFile is the name of the manifest file in the local source.
	LockManifestFile = "rsk.lock"

	// LocalTempFilePattern is the glob pattern for temporary files used during manifest operations, which should be ignored or cleaned up as needed.
	LocalTempFilePattern = ".rsk-*.tmp"

	// LocalDirPermission is the permission for the local .rsk directory, set to 0o755 to allow read/write/execute for the owner and read/execute for group and others.
	LocalDirPermission = 0o750

	// LocalFilePermission is the permission for files created in .rsk directories.
	LocalFilePermission = 0o644
)

var (
	// LocalSkillsPrefix is the forward-slash path prefix for .rsk/skills/ entries.
	// Intentionally uses "/" not filepath.Separator: opencode.json stores Unix-style
	// paths regardless of host OS, so this must remain slash-separated.
	LocalSkillsPrefix = LocalConfigFolderName + "/" + skill.SkillsFolderName + "/"

	// RelativeClaudeFilePath is the relative file path of CLAUDE.md from the project root, used to determine where the pinned skills configuration file should be located within a project.
	RelativeClaudeFilePath = filepath.Join(LocalConfigFolderName, ClaudeFileName)
)

// LocalConfigFolderPath returns the canonical local config folder path based on the current current working directory
func LocalConfigFolderPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	rskDir := filepath.Join(cwd, LocalConfigFolderName)
	if _, err := os.Stat(ModPath(rskDir)); os.IsNotExist(err) {
		return "", fmt.Errorf("no rsk.mod found — run 'rsk project init' first")
	}
	return rskDir, nil
}

// LocalConfigSkillsPath returns the canonical path of the local skills directory inside rskDir, where locally developed skills are stored.
func LocalConfigSkillsPath(rskDir string) string {
	return filepath.Join(rskDir, skill.SkillsFolderName)
}

// LockPath returns the canonical path of rsk.lock inside rskDir.
func LockPath(rskDir string) string {
	return filepath.Join(rskDir, LockManifestFile)
}

// ModPath returns the canonical path of rsk.mod inside rskDir.
func ModPath(rskDir string) string {
	return filepath.Join(rskDir, ModManifestFile)
}

// ClaudeConfigFilePath returns the canonical path of CLAUDE.md inside rskDir.
func ClaudeConfigFilePath(rskDir string) string {
	return filepath.Join(rskDir, ClaudeFileName)
}

// OpenCodeConfigFilePath returns the canonical path of opencode.json inside rskDir.
func OpenCodeConfigFilePath(rskDir string) string {
	return filepath.Join(rskDir, OpenCodeConfigFileName)
}
