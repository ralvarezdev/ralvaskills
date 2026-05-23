// Package git provides thin wrappers around the git CLI for rsk operations.
package git

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ralvarezdev/ralvaskills/internal/fsperm"
)

// Pull runs "git pull" in dir, streaming output to out.
func Pull(ctx context.Context, dir string, out io.Writer) error {
	c := exec.CommandContext(ctx, "git", "-C", dir, "pull")
	c.Stdout = out
	c.Stderr = os.Stderr
	return c.Run()
}

// Clone clones url into dest, creating parent directories as needed, and
// streams output to out.
func Clone(ctx context.Context, url, dest string, out io.Writer) error {
	if err := os.MkdirAll(filepath.Dir(dest), fsperm.Dir); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	c := exec.CommandContext(ctx, "git", "clone", url, dest)
	c.Stdout = out
	c.Stderr = os.Stderr
	return c.Run()
}
