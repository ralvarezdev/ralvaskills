package ui

import (
	"github.com/ralvarezdev/termkit/history"
	"github.com/ralvarezdev/termkit/prefs"
	"github.com/ralvarezdev/termkit/session"
)

// defaultStorePaths returns termkit's conventional history and prefs file
// locations for rsk, or "" for one that cannot be resolved.
func defaultStorePaths() (historyPath, prefsPath string) {
	if path, err := history.DefaultPath(appName); err == nil {
		historyPath = path
	}
	if path, err := prefs.DefaultPath(appName); err == nil {
		prefsPath = path
	}
	return historyPath, prefsPath
}

// attachStores opens the recent-commands and prefs files and hands them to
// cfg. A missing path or a file that cannot be read leaves that feature off:
// remembering things is a convenience, never a reason to refuse to start.
func attachStores(cfg *session.Config, historyPath, prefsPath string) {
	if historyPath != "" {
		if store, err := history.Open(historyPath); err == nil {
			cfg.History = store
		}
	}
	if prefsPath != "" {
		if store, err := prefs.Open(prefsPath); err == nil {
			cfg.Prefill = store
			cfg.Theme = store
		}
	}
}
