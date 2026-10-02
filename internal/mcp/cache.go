package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ralvarezdev/ralvaskills/v3/internal/fsperm"
	"github.com/ralvarezdev/ralvaskills/v3/internal/fsx"
	"github.com/ralvarezdev/ralvaskills/v3/internal/source"
)

const (
	// indexCacheFileName is the file the registry index is cached under, inside
	// the registry cache directory.
	indexCacheFileName = "index.json"

	// indexCacheTempPattern is the temp-file pattern for atomic cache writes.
	indexCacheTempPattern = ".rsk-index-*.tmp"

	// DefaultIndexTTL is how long a cached registry index is served before it is
	// refetched.
	DefaultIndexTTL = 24 * time.Hour
)

// IndexFetchFunc fetches the registry index. It lets the cache be tested without
// a server and lets the caller wire the real source.Registry.
type IndexFetchFunc func(context.Context) (map[string]*source.IndexEntry, error)

// IndexSnapshot is a registry index plus how current it is.
type IndexSnapshot struct {
	Skills    map[string]*source.IndexEntry
	FetchedAt time.Time
	Stale     bool
}

// IndexCache serves the registry index from a disk file, refetching through
// fetch when the file is missing or older than ttl. A failed fetch falls back to
// a stale file so a server can still start offline.
type IndexCache struct {
	fetch IndexFetchFunc
	path  string
	ttl   time.Duration
}

// cachedIndex is the on-disk shape: the published skills map plus the instant it
// was fetched, which drives the TTL.
type cachedIndex struct {
	FetchedAt time.Time                     `json:"fetched_at"`
	Skills    map[string]*source.IndexEntry `json:"skills"`
}

// NewIndexCache returns a cache rooted at cacheDir that serves entries younger
// than ttl and refetches through fetch otherwise. fetch must not be nil.
func NewIndexCache(cacheDir string, ttl time.Duration, fetch IndexFetchFunc) *IndexCache {
	return &IndexCache{
		path:  filepath.Join(cacheDir, indexCacheFileName),
		ttl:   ttl,
		fetch: fetch,
	}
}

// Load returns the registry index from cache or, when needed, from fetch.
//
// A cache entry younger than ttl is returned as is. Otherwise fetch is called
// and its result persisted; a failed fetch falls back to a stale cache entry
// (Stale=true) instead of failing, so the server starts without network. When
// there is no usable cache and the fetch fails, the fetch error is returned.
func (c *IndexCache) Load(ctx context.Context) (IndexSnapshot, error) {
	cached, cacheErr := readIndexCache(c.path)
	if cacheErr == nil && time.Since(cached.FetchedAt) < c.ttl {
		return IndexSnapshot{Skills: cached.Skills, FetchedAt: cached.FetchedAt}, nil
	}

	skills, fetchErr := c.fetch(ctx)
	if fetchErr == nil {
		now := time.Now()
		// Persisting is best-effort: a write failure must not fail the load.
		//nolint:errcheck // the fetched index is returned regardless of the cache write
		_ = writeIndexCache(c.path, now, skills)
		return IndexSnapshot{Skills: skills, FetchedAt: now}, nil
	}

	if cacheErr == nil {
		return IndexSnapshot{Skills: cached.Skills, FetchedAt: cached.FetchedAt, Stale: true}, nil
	}
	return IndexSnapshot{}, fmt.Errorf("fetch registry index: %w", fetchErr)
}

func readIndexCache(path string) (cachedIndex, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return cachedIndex{}, fmt.Errorf("read index cache %s: %w", path, err)
	}

	var cached cachedIndex
	if unmarshalErr := json.Unmarshal(data, &cached); unmarshalErr != nil {
		return cachedIndex{}, fmt.Errorf("decode index cache %s: %w", path, unmarshalErr)
	}
	if cached.Skills == nil {
		return cachedIndex{}, fmt.Errorf("decode index cache %s: no skills", path)
	}
	return cached, nil
}

func writeIndexCache(path string, fetchedAt time.Time, skills map[string]*source.IndexEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), fsperm.Dir); err != nil {
		return fmt.Errorf("create index cache dir: %w", err)
	}

	cached := cachedIndex{FetchedAt: fetchedAt, Skills: skills}
	return fsx.WriteAtomic(path, indexCacheTempPattern, func(w io.Writer) error {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(cached)
	})
}
