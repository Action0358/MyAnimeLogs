package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryCache_SetAndGet(t *testing.T) {
	config := CacheConfig{
		DefaultTTL:      1 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxItems:        100,
	}
	cache := NewMemoryCache(config)
	defer cache.Close()

	// Test setting and getting a string
	err := cache.Set("key1", "value1")
	require.NoError(t, err)

	var result string
	found, err := cache.Get("key1", &result)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "value1", result)
}

func TestMemoryCache_GetNotFound(t *testing.T) {
	config := CacheConfig{
		DefaultTTL:      1 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxItems:        100,
	}
	cache := NewMemoryCache(config)
	defer cache.Close()

	var result string
	found, err := cache.Get("nonexistent", &result)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestMemoryCache_Expiration(t *testing.T) {
	config := CacheConfig{
		DefaultTTL:      50 * time.Millisecond,
		CleanupInterval: 1 * time.Minute,
		MaxItems:        100,
	}
	cache := NewMemoryCache(config)
	defer cache.Close()

	err := cache.Set("key1", "value1")
	require.NoError(t, err)

	// Should be found immediately
	var result string
	found, err := cache.Get("key1", &result)
	require.NoError(t, err)
	assert.True(t, found)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should not be found after expiration
	found, err = cache.Get("key1", &result)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestMemoryCache_SetWithTTL(t *testing.T) {
	config := CacheConfig{
		DefaultTTL:      1 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxItems:        100,
	}
	cache := NewMemoryCache(config)
	defer cache.Close()

	// Set with custom TTL
	err := cache.SetWithTTL("key1", "value1", 50*time.Millisecond)
	require.NoError(t, err)

	// Should be found immediately
	var result string
	found, err := cache.Get("key1", &result)
	require.NoError(t, err)
	assert.True(t, found)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should not be found after expiration
	found, err = cache.Get("key1", &result)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestMemoryCache_Delete(t *testing.T) {
	config := CacheConfig{
		DefaultTTL:      1 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxItems:        100,
	}
	cache := NewMemoryCache(config)
	defer cache.Close()

	err := cache.Set("key1", "value1")
	require.NoError(t, err)

	cache.Delete("key1")

	var result string
	found, err := cache.Get("key1", &result)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestMemoryCache_DeleteByPrefix(t *testing.T) {
	config := CacheConfig{
		DefaultTTL:      1 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxItems:        100,
	}
	cache := NewMemoryCache(config)
	defer cache.Close()

	// Set multiple keys with same prefix
	err := cache.Set("prefix:key1", "value1")
	require.NoError(t, err)
	err = cache.Set("prefix:key2", "value2")
	require.NoError(t, err)
	err = cache.Set("other:key3", "value3")
	require.NoError(t, err)

	// Delete by prefix
	cache.DeleteByPrefix("prefix:")

	// prefix keys should be deleted
	var result string
	found, _ := cache.Get("prefix:key1", &result)
	assert.False(t, found)
	found, _ = cache.Get("prefix:key2", &result)
	assert.False(t, found)

	// other key should still exist
	found, _ = cache.Get("other:key3", &result)
	assert.True(t, found)
}

func TestMemoryCache_Clear(t *testing.T) {
	config := CacheConfig{
		DefaultTTL:      1 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxItems:        100,
	}
	cache := NewMemoryCache(config)
	defer cache.Close()

	err := cache.Set("key1", "value1")
	require.NoError(t, err)
	err = cache.Set("key2", "value2")
	require.NoError(t, err)

	cache.Clear()

	var result string
	found, _ := cache.Get("key1", &result)
	assert.False(t, found)
	found, _ = cache.Get("key2", &result)
	assert.False(t, found)
}

func TestMemoryCache_Stats(t *testing.T) {
	config := CacheConfig{
		DefaultTTL:      1 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxItems:        100,
	}
	cache := NewMemoryCache(config)
	defer cache.Close()

	err := cache.Set("key1", "value1")
	require.NoError(t, err)
	err = cache.Set("key2", "value2")
	require.NoError(t, err)

	stats := cache.Stats()
	assert.Equal(t, 2, stats.ItemCount)
	assert.Equal(t, 0, stats.ExpiredCount)
	assert.Equal(t, 100, stats.MaxItems)
}

func TestMemoryCache_MaxItems(t *testing.T) {
	config := CacheConfig{
		DefaultTTL:      1 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxItems:        3,
	}
	cache := NewMemoryCache(config)
	defer cache.Close()

	// Add items up to max
	err := cache.Set("key1", "value1")
	require.NoError(t, err)
	err = cache.Set("key2", "value2")
	require.NoError(t, err)
	err = cache.Set("key3", "value3")
	require.NoError(t, err)

	// Adding one more should evict one
	err = cache.Set("key4", "value4")
	require.NoError(t, err)

	stats := cache.Stats()
	assert.LessOrEqual(t, stats.ItemCount, 3)
}

func TestMemoryCache_ComplexTypes(t *testing.T) {
	config := CacheConfig{
		DefaultTTL:      1 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxItems:        100,
	}
	cache := NewMemoryCache(config)
	defer cache.Close()

	// Test with struct
	type TestStruct struct {
		Name  string
		Value int
	}

	original := TestStruct{Name: "test", Value: 42}
	err := cache.Set("struct_key", original)
	require.NoError(t, err)

	var result TestStruct
	found, err := cache.Get("struct_key", &result)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, original, result)

	// Test with slice
	originalSlice := []string{"a", "b", "c"}
	err = cache.Set("slice_key", originalSlice)
	require.NoError(t, err)

	var resultSlice []string
	found, err = cache.Get("slice_key", &resultSlice)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, originalSlice, resultSlice)
}

func TestCacheItem_IsExpired(t *testing.T) {
	// Not expired
	item := &CacheItem{
		Value:      []byte("test"),
		Expiration: time.Now().Add(1 * time.Hour),
	}
	assert.False(t, item.IsExpired())

	// Expired
	item = &CacheItem{
		Value:      []byte("test"),
		Expiration: time.Now().Add(-1 * time.Hour),
	}
	assert.True(t, item.IsExpired())
}
