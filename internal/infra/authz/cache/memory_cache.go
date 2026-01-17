package cache

import (
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2/expirable"
)

// MemoryCache implements L1 in-memory LRU cache for RBAC decisions
type MemoryCache struct {
	cache *lru.LRU[string, bool]
	mu    sync.RWMutex
}

// NewMemoryCache creates a new LRU memory cache
func NewMemoryCache(size int, ttl time.Duration) (*MemoryCache, error) {
	// Create LRU cache with expiration
	cache := lru.NewLRU[string, bool](size, nil, ttl)

	return &MemoryCache{
		cache: cache,
	}, nil
}

// Get retrieves a value from cache
func (m *MemoryCache) Get(key string) (bool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, found := m.cache.Get(key)
	return value, found
}

// Set stores a value in cache
func (m *MemoryCache) Set(key string, value bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache.Add(key, value)
}

// Delete removes a key from cache
func (m *MemoryCache) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache.Remove(key)
}

// Clear removes all entries from cache
func (m *MemoryCache) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache.Purge()
}

// Len returns the number of items in cache
func (m *MemoryCache) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.cache.Len()
}

// Contains checks if a key exists in cache
func (m *MemoryCache) Contains(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.cache.Contains(key)
}
