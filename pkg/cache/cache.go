package cache

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

type item struct {
	value      any
	expiration int64
}

// MemoryCache provides in-memory thread-safe caching with TTL.
type MemoryCache struct {
	items map[string]item
	mu    sync.RWMutex
}

// NewMemoryCache creates a new MemoryCache instance.
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		items: make(map[string]item),
	}
}

// Set stores a key-value pair with a duration TTL.
func (c *MemoryCache) Set(key string, value any, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var exp int64
	if duration > 0 {
		exp = time.Now().Add(duration).UnixNano()
	}

	c.items[key] = item{
		value:      value,
		expiration: exp,
	}
}

// Get retrieves a key from cache, returning false if expired or missing.
func (c *MemoryCache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	it, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if it.expiration > 0 && time.Now().UnixNano() > it.expiration {
		return nil, false
	}

	return it.value, true
}

// Delete removes a key from cache.
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Flush clears all cache items.
func (c *MemoryCache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]item)
}

// ClearStorageCache removes cached file assets from storage directory.
func ClearStorageCache(cacheDir string) error {
	if cacheDir == "" {
		cacheDir = "storage/cache"
	}
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return nil
	}
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		_ = os.RemoveAll(filepath.Join(cacheDir, entry.Name()))
	}
	return nil
}

