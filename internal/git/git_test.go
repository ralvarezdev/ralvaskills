package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// run executes a git (or other) command in dir, failing the test on error.
func run(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	c := exec.Command(name, args...)
	c.Dir = dir
	c.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
	)
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

// commit writes a file and commits it in dir.
func commit(t *testing.T, dir, file, msg string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, file), []byte(msg), 0o644); err != nil {
		t.Fatalf("write %s: %v", file, err)
	}
	run(t, dir, "git", "add", file)
	run(t, dir, "git", "commit", "-m", msg)
}

func TestBehind(t *testing.T) {
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	cloneA := filepath.Join(root, "a") // the clone we call Behind() on
	cloneB := filepath.Join(root, "b") // pushes new commits so a falls behind

	run(t, root, "git", "init", "--bare", remote)

	run(t, root, "git", "clone", remote, cloneA)
	commit(t, cloneA, "seed.txt", "seed")
	run(t, cloneA, "git", "push", "origin", "HEAD")

	run(t, root, "git", "clone", remote, cloneB)

	ctx := context.Background()

	t.Run("up to date", func(t *testing.T) {
		n, err := Behind(ctx, cloneA)
		if err != nil {
			t.Fatalf("Behind: %v", err)
		}
		if n != 0 {
			t.Errorf("Behind(up-to-date clone) = %d, want 0", n)
		}
	})

	t.Run("behind after remote gets new commits", func(t *testing.T) {
		commit(t, cloneB, "one.txt", "one")
		commit(t, cloneB, "two.txt", "two")
		run(t, cloneB, "git", "push", "origin", "HEAD")

		n, err := Behind(ctx, cloneA)
		if err != nil {
			t.Fatalf("Behind: %v", err)
		}
		if n != 2 {
			t.Errorf("Behind(clone 2 commits behind) = %d, want 2", n)
		}
	})

	t.Run("no upstream tracking branch", func(t *testing.T) {
		noUpstream := filepath.Join(root, "no-upstream")
		run(t, root, "git", "init", noUpstream)
		commit(t, noUpstream, "solo.txt", "solo")

		n, err := Behind(ctx, noUpstream)
		if err != nil {
			t.Fatalf("Behind(no upstream): unexpected error: %v", err)
		}
		if n != 0 {
			t.Errorf("Behind(no upstream) = %d, want 0", n)
		}
	})
}
