//go:build windows

package skill

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

// createLink creates a directory junction. Unlike symlinks, junctions do
// not require Administrator privileges or Developer Mode on Windows. They
// only work for directories on the same volume, both acceptable for skill
// installs (registry cache and project skills dir live on the user volume).
func createLink(target, link string) error {
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("resolve target %s: %w", target, err)
	}
	out, err := exec.Command("cmd", "/c", "mklink", "/J", link, absTarget).CombinedOutput()
	if err != nil {
		return fmt.Errorf("mklink /J: %w: %s", err, out)
	}
	return nil
}
