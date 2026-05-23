package manifest

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ralvarezdev/ralvaskills/internal"
)

const (
	// ModFileName is the filename of the project skill manifest.
	ModFileName = "rsk.mod"

	// LockFileName is the filename of the resolved skill lockfile.
	LockFileName = "rsk.lock"

	// TempFilePattern is the glob pattern for temporary files created during
	// atomic write operations inside the .rsk directory.
	TempFilePattern = ".rsk-*.tmp"
)

// ProjectFolderPath returns the .rsk directory path for the current working
// directory. Returns an error if rsk.mod is not found, indicating the directory
// has not been initialized as an rsk project.
func ProjectFolderPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	rskDir := filepath.Join(cwd, internal.ProjectFolderName)
	if _, statErr := os.Stat(ModPath(rskDir)); os.IsNotExist(statErr) {
		return "", fmt.Errorf("no rsk.mod found — run 'rsk new' first")
	}
	return rskDir, nil
}

// LockPath returns the canonical path of rsk.lock inside rskDir.
func LockPath(rskDir string) string {
	return filepath.Join(rskDir, LockFileName)
}

// ModPath returns the canonical path of rsk.mod inside rskDir.
func ModPath(rskDir string) string {
	return filepath.Join(rskDir, ModFileName)
}
