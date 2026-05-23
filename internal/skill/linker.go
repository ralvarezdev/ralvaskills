package skill

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ralvarezdev/ralvaskills/internal/fsperm"
)

// Link creates a symlink at targetDir/<skill.Name> pointing to skill.Path,
// replacing any existing symlink at that location.
func Link(s Skill, targetDir string) error {
	if err := os.MkdirAll(targetDir, fsperm.Dir); err != nil {
		return fmt.Errorf("create target dir %s: %w", targetDir, err)
	}

	linkPath := filepath.Join(targetDir, s.Name)
	if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove existing link %s: %w", linkPath, err)
	}

	if err := os.Symlink(s.Path, linkPath); err != nil {
		return fmt.Errorf("symlink %s → %s: %w", linkPath, s.Path, err)
	}
	return nil
}

// Unlink removes the symlink at targetDir/<name>. Returns nil if the symlink
// does not exist.
func Unlink(name, targetDir string) error {
	linkPath := filepath.Join(targetDir, name)
	if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove link %s: %w", linkPath, err)
	}
	return nil
}

// IsLinked reports whether a symlink named name exists in targetDir.
func IsLinked(name, targetDir string) bool {
	fi, err := os.Lstat(filepath.Join(targetDir, name))
	return err == nil && fi.Mode()&os.ModeSymlink != 0
}
