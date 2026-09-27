// Package git provides thin wrappers around the git CLI for rsk operations.
package git

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

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

// Behind fetches dir's remote and reports how many commits HEAD is behind its
// upstream tracking branch. Not every clone has a remote configured, or a
// branch tracking one (a bare `git init`, a detached HEAD, a fetch that fails
// because there's no network) — any of those make "how far behind" simply
// unmeasurable, so they report 0, nil rather than erroring. This is meant for
// a passive, best-effort drift check: a false "nothing to report" is far
// preferable to surfacing a spurious error.
func Behind(ctx context.Context, dir string) (int, error) {
	if err := exec.CommandContext(ctx, "git", "-C", dir, "fetch", "--quiet").Run(); err != nil {
		return 0, nil //nolint:nilerr // no remote/network — unmeasurable, not an error; see doc comment above
	}

	out, err := exec.CommandContext(ctx, "git", "-C", dir, "rev-list", "--count", "HEAD..@{u}").Output()
	if err != nil {
		return 0, nil //nolint:nilerr // no upstream tracking branch — unmeasurable, not an error; see doc comment above
	}

	n, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, fmt.Errorf("parse rev-list count %q: %w", out, err)
	}
	return n, nil
}
