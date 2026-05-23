package skill

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ralvarezdev/ralvaskills/internal/fsperm"
)

// Link creates a link at targetDir/<skill.Name> pointing to skill.Path,
// replacing any existing link at that location. On Unix the link is a
// symlink; on Windows it is a directory junction (see createLink).
func Link(s Skill, targetDir string) error {
	if err := os.MkdirAll(targetDir, fsperm.Dir); err != nil {
		return fmt.Errorf("create target dir %s: %w", targetDir, err)
	}

	linkPath := filepath.Join(targetDir, s.Name)
	if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove existing link %s: %w", linkPath, err)
	}

	if err := createLink(s.Path, linkPath); err != nil {
		return fmt.Errorf("link %s → %s: %w", linkPath, s.Path, err)
	}
	return nil
}

// Unlink removes the link at targetDir/<name>. Returns nil if the link
// does not exist.
func Unlink(name, targetDir string) error {
	linkPath := filepath.Join(targetDir, name)
	if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove link %s: %w", linkPath, err)
	}
	return nil
}

// IsLinked reports whether a link named name exists in targetDir. Windows
// directory junctions are reported with ModeSymlink set by Lstat since
// Go 1.23, so the same check works on every platform.
func IsLinked(name, targetDir string) bool {
	fi, err := os.Lstat(filepath.Join(targetDir, name))
	return err == nil && fi.Mode()&os.ModeSymlink != 0
}
