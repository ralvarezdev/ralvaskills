package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// run executes git in dir, failing the test on error.
func run(ctx context.Context, t *testing.T, dir string, args ...string) {
	t.Helper()
	c := exec.CommandContext(ctx, "git", args...)
	c.Dir = dir
	c.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
	)
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// commit writes a file and commits it in dir.
func commit(ctx context.Context, t *testing.T, dir, file, msg string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, file), []byte(msg), 0o644); err != nil {
		t.Fatalf("write %s: %v", file, err)
	}
	run(ctx, t, dir, "add", file)
	run(ctx, t, dir, "commit", "-m", msg)
}

//nolint:paralleltest // subtests share and mutate cloneA/cloneB sequentially by design; see comment below
func TestBehind(t *testing.T) {
	ctx := t.Context()

	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	cloneA := filepath.Join(root, "a") // the clone we call Behind() on
	cloneB := filepath.Join(root, "b") // pushes new commits so a falls behind

	run(ctx, t, root, "init", "--bare", remote)

	run(ctx, t, root, "clone", remote, cloneA)
	commit(ctx, t, cloneA, "seed.txt", "seed")
	run(ctx, t, cloneA, "push", "origin", "HEAD")

	run(ctx, t, root, "clone", remote, cloneB)

	// Subtests intentionally run sequentially, not in parallel: they share
	// and mutate cloneA/cloneB across steps (each depends on the git state
	// the previous one left behind), so t.Parallel() here would race on the
	// same .git directories.
	//nolint:paralleltest // sequential by design, see comment above
	t.Run("up to date", func(t *testing.T) {
		n, err := Behind(ctx, cloneA)
		if err != nil {
			t.Fatalf("Behind: %v", err)
		}
		if n != 0 {
			t.Errorf("Behind(up-to-date clone) = %d, want 0", n)
		}
	})

	//nolint:paralleltest // sequential by design, see comment above
	t.Run("behind after remote gets new commits", func(t *testing.T) {
		commit(ctx, t, cloneB, "one.txt", "one")
		commit(ctx, t, cloneB, "two.txt", "two")
		run(ctx, t, cloneB, "push", "origin", "HEAD")

		n, err := Behind(ctx, cloneA)
		if err != nil {
			t.Fatalf("Behind: %v", err)
		}
		if n != 2 {
			t.Errorf("Behind(clone 2 commits behind) = %d, want 2", n)
		}
	})

	//nolint:paralleltest // sequential by design, see comment above
	t.Run("no upstream tracking branch", func(t *testing.T) {
		noUpstream := filepath.Join(root, "no-upstream")
		run(ctx, t, root, "init", noUpstream)
		commit(ctx, t, noUpstream, "solo.txt", "solo")

		n, err := Behind(ctx, noUpstream)
		if err != nil {
			t.Fatalf("Behind(no upstream): unexpected error: %v", err)
		}
		if n != 0 {
			t.Errorf("Behind(no upstream) = %d, want 0", n)
		}
	})
}
