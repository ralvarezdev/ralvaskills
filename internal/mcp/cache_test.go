package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ralvarezdev/ralvaskills/v3/internal/catalog"
	"github.com/ralvarezdev/ralvaskills/v3/internal/fsperm"
	"github.com/ralvarezdev/ralvaskills/v3/internal/mcp"
	"github.com/ralvarezdev/ralvaskills/v3/internal/source"
)

func TestIndexCacheFreshSkipsFetch(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeIndexCacheFile(t, dir, time.Now(), map[string]*source.IndexEntry{"a": {Name: "a", Latest: "1.0.0"}})

	cache := mcp.NewIndexCache(dir, mcp.DefaultIndexTTL, func(context.Context) (map[string]*source.IndexEntry, error) {
		t.Error("fetch called for a fresh cache")
		return nil, errors.New("unexpected fetch")
	})

	snap, err := cache.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if snap.Stale {
		t.Error("fresh cache reported stale")
	}
	if snap.Skills["a"] == nil {
		t.Errorf("Skills = %+v, want the cached entry", snap.Skills)
	}
}

func TestIndexCacheStaleRefetches(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeIndexCacheFile(t, dir, time.Now().Add(-25*time.Hour), map[string]*source.IndexEntry{"old": {Name: "old"}})

	calls := 0
	fetched := map[string]*source.IndexEntry{"new": {Name: "new", Latest: "2.0.0"}}
	cache := mcp.NewIndexCache(dir, mcp.DefaultIndexTTL, func(context.Context) (map[string]*source.IndexEntry, error) {
		calls++
		return fetched, nil
	})

	snap, err := cache.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if calls != 1 {
		t.Errorf("fetch calls = %d, want 1", calls)
	}
	if snap.Stale {
		t.Error("refetched snapshot reported stale")
	}
	if snap.Skills["new"] == nil {
		t.Errorf("Skills = %+v, want the fetched entry", snap.Skills)
	}

	// The fetch must have replaced the stale cache on disk.
	refreshed, err := cache.Load(t.Context())
	if err != nil {
		t.Fatalf("Load (cached): %v", err)
	}
	if calls != 1 {
		t.Errorf("fetch calls after refresh = %d, want 1", calls)
	}
	if refreshed.Skills["new"] == nil {
		t.Errorf("cached skills = %+v, want the fetched entry", refreshed.Skills)
	}
}

func TestIndexCacheStaleFallbackOffline(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeIndexCacheFile(t, dir, time.Now().Add(-30*time.Hour), map[string]*source.IndexEntry{"old": {Name: "old"}})

	cache := mcp.NewIndexCache(dir, mcp.DefaultIndexTTL, func(context.Context) (map[string]*source.IndexEntry, error) {
		return nil, errors.New("network down")
	})

	snap, err := cache.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !snap.Stale {
		t.Error("offline fallback did not report stale")
	}
	if snap.Skills["old"] == nil {
		t.Errorf("Skills = %+v, want the stale entry", snap.Skills)
	}
}

func TestIndexCacheMissAndFetchFails(t *testing.T) {
	t.Parallel()

	cache := mcp.NewIndexCache(
		t.TempDir(),
		mcp.DefaultIndexTTL,
		func(context.Context) (map[string]*source.IndexEntry, error) {
			return nil, errors.New("network down")
		},
	)
	if _, err := cache.Load(t.Context()); err == nil {
		t.Fatal("Load with no cache and a failing fetch returned nil error")
	}
}

func TestIndexCacheCorruptFileRefetches(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "index.json"), []byte("{not json"))

	cache := mcp.NewIndexCache(dir, mcp.DefaultIndexTTL, func(context.Context) (map[string]*source.IndexEntry, error) {
		return map[string]*source.IndexEntry{"a": {Name: "a"}}, nil
	})
	snap, err := cache.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if snap.Skills["a"] == nil {
		t.Errorf("Skills = %+v, want the fetched entry", snap.Skills)
	}
}

// writeIndexCacheFile writes the on-disk cache shape so a test can control
// fetched_at without waiting out the TTL.
func writeIndexCacheFile(t *testing.T, dir string, fetchedAt time.Time, skills map[string]*source.IndexEntry) {
	t.Helper()

	payload := struct {
		FetchedAt time.Time                     `json:"fetched_at"`
		Skills    map[string]*source.IndexEntry `json:"skills"`
	}{FetchedAt: fetchedAt, Skills: skills}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal cache: %v", err)
	}
	writeFile(t, filepath.Join(dir, "index.json"), data)
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), fsperm.Dir); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, data, fsperm.File); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestIndexCacheFetchesRegistryIndexAndServesOffline(t *testing.T) {
	t.Parallel()

	body := `{"skills":{"go-architect":{"name":"go-architect","description":"Go standards",` +
		`"latest":"1.2.0","versions":{"1.2.0":{"version":"1.2.0","archive_url":"http://example/g.tar.gz"}}}}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))

	dir := t.TempDir()
	reg := source.NewRegistry(srv.URL, dir)
	snap, err := mcp.NewIndexCache(dir, mcp.DefaultIndexTTL, reg.Index).Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	entries := catalog.FromIndex(snap.Skills)
	if len(entries) != 1 || entries[0].Name != "go-architect" || entries[0].Latest != "1.2.0" {
		t.Fatalf("catalog = %+v, want go-architect at 1.2.0", entries)
	}

	// With the server down, a fresh cache must still serve the index.
	srv.Close()
	if _, offlineErr := mcp.NewIndexCache(dir, mcp.DefaultIndexTTL, reg.Index).Load(t.Context()); offlineErr != nil {
		t.Errorf("offline load: %v", offlineErr)
	}
}
