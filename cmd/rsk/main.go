package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ralvarezdev/ralvaskills/v2/internal/config"
	updatecheck "github.com/ralvarezdev/ralvaskills/v2/internal/update"
)

func main() {
	if os.Getenv(rskUpdateCheckWorkerEnv) == "1" {
		runUpdateCheckWorker()
		return
	}
	Execute()
}

// runUpdateCheckWorker computes and persists the update-availability cache,
// then exits immediately. It's the detached subprocess maybeSpawnUpdateCheck
// spawns, so it runs entirely independently of — and never delays — the bare
// `rsk` invocation that triggered it. Any failure here is silent: this is a
// best-effort background refresh, not a user-facing operation.
func runUpdateCheckWorker() {
	cfg, err := config.Load()
	if err != nil {
		os.Exit(0)
	}

	result, err := updatecheck.Check(context.Background(), cfg)
	if err != nil {
		os.Exit(0)
	}

	if err = updatecheck.Save(cfg, result); err != nil {
		fmt.Fprintln(os.Stderr, "update check: save failed:", err)
	}
	os.Exit(0)
}
