package update

import (
	"os"
	"testing"
	"time"

	"github.com/ralvarezdev/ralvaskills/internal/config"
)

func testConfig(t *testing.T) config.Config {
	t.Helper()
	return config.Config{VersionsCache: t.TempDir()}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	t.Parallel()

	cfg := testConfig(t)
	want := Result{Mode: ModeRegistry, Outdated: []string{"foo", "bar"}}

	if err := Save(cfg, want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, checkedAt, ok := Load(cfg)
	if !ok {
		t.Fatal("Load: ok = false, want true")
	}
	if got.Mode != want.Mode {
		t.Errorf("Mode = %q, want %q", got.Mode, want.Mode)
	}
	outdatedMatch := len(got.Outdated) == len(want.Outdated) &&
		got.Outdated[0] == want.Outdated[0] &&
		got.Outdated[1] == want.Outdated[1]
	if !outdatedMatch {
		t.Errorf("Outdated = %v, want %v", got.Outdated, want.Outdated)
	}
	if checkedAt.IsZero() {
		t.Error("CheckedAt is zero, want a recent timestamp")
	}
	if Stale(checkedAt) {
		t.Error("Stale(just-saved) = true, want false")
	}
}

func TestSaveLoadEmptyResult(t *testing.T) {
	t.Parallel()

	cfg := testConfig(t)
	if err := Save(cfg, Result{Mode: ModeLocal}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, _, ok := Load(cfg)
	if !ok {
		t.Fatal("Load: ok = false, want true")
	}
	if len(got.Outdated) != 0 {
		t.Errorf("Outdated = %v, want empty", got.Outdated)
	}
}

func TestLoadMissingCache(t *testing.T) {
	t.Parallel()

	cfg := testConfig(t)
	if _, _, ok := Load(cfg); ok {
		t.Error("Load on missing cache: ok = true, want false")
	}
}

func TestLoadMalformedCache(t *testing.T) {
	t.Parallel()

	cfg := testConfig(t)
	if err := Save(cfg, Result{Mode: ModeRegistry}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := os.WriteFile(cachePath(cfg), []byte("not json"), 0o644); err != nil {
		t.Fatalf("write garbage: %v", err)
	}

	if _, _, ok := Load(cfg); ok {
		t.Error("Load on malformed cache: ok = true, want false")
	}
}

func TestStale(t *testing.T) {
	t.Parallel()

	if Stale(time.Now()) {
		t.Error("Stale(now) = true, want false")
	}
	if !Stale(time.Now().Add(-StaleAfter - time.Minute)) {
		t.Error("Stale(25h ago) = false, want true")
	}
}
