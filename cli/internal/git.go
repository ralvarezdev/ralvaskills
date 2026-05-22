package internal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ralvarezdev/ralvaskills/cli/internal/config"
)

func GitPull(dir string, out io.Writer) error {
	c := exec.Command("git", "-C", dir, "pull")
	c.Stdout = out
	c.Stderr = os.Stderr
	return c.Run()
}

func GitClone(url, dest string, out io.Writer) error {
	if err := os.MkdirAll(filepath.Dir(dest), config.GitDirPermission); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	c := exec.Command("git", "clone", url, dest)
	c.Stdout = out
	c.Stderr = os.Stderr
	return c.Run()
}
