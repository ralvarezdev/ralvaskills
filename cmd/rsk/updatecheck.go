package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/ralvarezdev/ralvaskills/internal/config"
	updatecheck "github.com/ralvarezdev/ralvaskills/internal/update"
)

// rskUpdateCheckWorkerEnv marks a detached background process spawned by
// maybeSpawnUpdateCheck to refresh the update-availability cache. main()
// checks for it before any normal cobra dispatch, so the check itself never
// competes with (or delays) a real invocation.
const rskUpdateCheckWorkerEnv = "RSK_UPDATE_CHECK_WORKER"

// updateNotice returns a one-line "updates available" notice for a bare
// `rsk` invocation, or "" when there's nothing to say: no config yet, no
// cache yet, a stale cache, or an empty one. It only ever reads the local
// cache file — never the network — so it can't slow down startup.
func updateNotice() string {
	cfg, err := config.Load()
	if err != nil {
		return ""
	}

	result, checkedAt, ok := updatecheck.Load(cfg)
	if !ok || updatecheck.Stale(checkedAt) || len(result.Outdated) == 0 {
		return ""
	}

	n := len(result.Outdated)
	if n == 1 {
		return "1 update available — run 'rsk update' to apply."
	}
	return fmt.Sprintf("%d updates available — run 'rsk update' to apply.", n)
}

// maybeSpawnUpdateCheck kicks off a detached background refresh of the
// update-availability cache when it's missing or stale, so the *next*
// invocation benefits without the *current* one ever waiting on the
// network. Spawn failures (missing config, can't resolve the binary path,
// etc.) are silently ignored — this is a best-effort nicety, never something
// that should surface an error or block startup.
func maybeSpawnUpdateCheck() {
	cfg, err := config.Load()
	if err != nil {
		return
	}

	_, checkedAt, ok := updatecheck.Load(cfg)
	if ok && !updatecheck.Stale(checkedAt) {
		return
	}

	exe, err := os.Executable()
	if err != nil {
		return
	}

	c := exec.CommandContext(context.Background(), exe)
	c.Env = append(os.Environ(), rskUpdateCheckWorkerEnv+"=1")
	c.Stdin = nil
	c.Stdout = nil
	c.Stderr = nil
	if err = c.Start(); err != nil {
		return
	}
}
