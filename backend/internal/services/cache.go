package services

import (
	"encoding/json"
	"sync"
	"time"
)

// CacheItem represents a cached item with expiration and LRU tracking
type CacheItem struct {
	Value        []byte
	Expiration   time.Time
	LastAccessed time.Time // LRU eviction用のアクセス時刻
}

// IsExpired checks if the cache item has expired
func (c *CacheItem) IsExpired() bool {
	return time.Now().After(c.Expiration)
}

// MemoryCache is a simple in-memory cache with TTL and LRU support
type MemoryCache struct {
	items map[string]*CacheItem
	mu    sync.RWMutex

	// Configuration
	defaultTTL     time.Duration
	cleanupTicker  *time.Ticker
	stopCleanup    chan struct{}
	maxItems       int
	cleanupRunning bool
}

// CacheConfig holds configuration for the cache
type CacheConfig struct {
	DefaultTTL      time.Duration
	CleanupInterval time.Duration
	MaxItems        int
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		DefaultTTL:      10 * time.Minute,
		CleanupInterval: 5 * time.Minute,
		MaxItems:        1000,
	}
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(config CacheConfig) *MemoryCache {
	cache := &MemoryCache{
		items:       make(map[string]*CacheItem),
		defaultTTL:  config.DefaultTTL,
		maxItems:    config.MaxItems,
		stopCleanup: make(chan struct{}),
	}

	// Start cleanup goroutine
	cache.startCleanup(config.CleanupInterval)

	return cache
}

// startCleanup starts the background cleanup goroutine
func (c *MemoryCache) startCleanup(interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cleanupRunning {
		return
	}
	c.cleanupRunning = true

	// tickerの初期化とgoroutine起動をロック内で行い、データレースを防止
	c.cleanupTicker = time.NewTicker(interval)
	ticker := c.cleanupTicker // ローカル変数にコピーしてgoroutine内で安全に使用

	go func() {
		for {
			select {
			case <-ticker.C:
				c.deleteExpired()
			case <-c.stopCleanup:
				ticker.Stop()
				return
			}
		}
	}()
}

// deleteExpired removes all expired items from the cache
func (c *MemoryCache) deleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, item := range c.items {
		if item.IsExpired() {
			delete(c.items, key)
		}
	}
}

// Set stores a value in the cache with the default TTL
func (c *MemoryCache) Set(key string, value any) error {
	return c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value in the cache with a custom TTL
func (c *MemoryCache) SetWithTTL(key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict items (LRU eviction: expired item first, then least recently used)
	if len(c.items) >= c.maxItems {
		c.evictOne()
	}

	now := time.Now()
	c.items[key] = &CacheItem{
		Value:        data,
		Expiration:   now.Add(ttl),
		LastAccessed: now,
	}

	return nil
}

// evictOne removes one item from the cache using LRU policy (must be called with lock held)
func (c *MemoryCache) evictOne() {
	// First, try to remove an expired item
	for key, item := range c.items {
		if item.IsExpired() {
			delete(c.items, key)
			return
		}
	}

	// If no expired items, remove the least recently used item (LRU)
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, item := range c.items {
		if first || item.LastAccessed.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.LastAccessed
			first = false
		}
	}

	if oldestKey != "" {
		delete(c.items, oldestKey)
	}
}

// Get retrieves a value from the cache and updates last accessed time for LRU
func (c *MemoryCache) Get(key string, dest any) (bool, error) {
	c.mu.Lock()
	item, exists := c.items[key]
	if exists && !item.IsExpired() {
		// LRU: アクセス時刻を更新
		item.LastAccessed = time.Now()
	}
	c.mu.Unlock()

	if !exists {
		return false, nil
	}

	if item.IsExpired() {
		c.Delete(key)
		return false, nil
	}

	if err := json.Unmarshal(item.Value, dest); err != nil {
		return false, err
	}

	return true, nil
}

// Delete removes a value from the cache
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// DeleteByPrefix removes all values with keys starting with the given prefix
func (c *MemoryCache) DeleteByPrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.items {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.items, key)
		}
	}
}

// Clear removes all items from the cache
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*CacheItem)
}

// Close stops the cleanup goroutine and clears the cache atomically
func (c *MemoryCache) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cleanupRunning {
		close(c.stopCleanup)
		c.cleanupRunning = false
	}
	// クリア処理も同じロック内で実行してレースコンディションを防止
	c.items = make(map[string]*CacheItem)
}

// Stats returns cache statistics
func (c *MemoryCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var expired int
	for _, item := range c.items {
		if item.IsExpired() {
			expired++
		}
	}

	return CacheStats{
		ItemCount:    len(c.items),
		ExpiredCount: expired,
		MaxItems:     c.maxItems,
	}
}

// CacheStats holds cache statistics
type CacheStats struct {
	ItemCount    int `json:"item_count"`
	ExpiredCount int `json:"expired_count"`
	MaxItems     int `json:"max_items"`
}
