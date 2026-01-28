package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"myanimelogs/internal/common"

	"github.com/redis/go-redis/v9"
)

// RedisCache is a Redis-based cache implementation
type RedisCache struct {
	client *redis.Client
	logger *common.Logger
}

// NewRedisCache creates a new Redis cache client
func NewRedisCache(logger *common.Logger) (*RedisCache, error) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL environment variable is not set")
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse REDIS_URL: %w", err)
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis cache connected successfully")

	return &RedisCache{
		client: client,
		logger: logger,
	}, nil
}

// Set stores a value in Redis with the specified TTL
func (r *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Get retrieves a value from Redis
func (r *RedisCache) Get(ctx context.Context, key string, dest any) (bool, error) {
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil // Key not found
	}
	if err != nil {
		return false, fmt.Errorf("failed to get cache: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return false, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return true, nil
}

// Delete removes a key from Redis
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete cache: %w", err)
	}
	return nil
}

// DeleteByPrefix removes all keys matching the prefix
// Keys are collected first, then deleted in a single DEL command to reduce network overhead
func (r *RedisCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	var keys []string
	iter := r.client.Scan(ctx, 0, prefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	// Delete all keys in a single command if any were found
	if len(keys) > 0 {
		if err := r.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
		r.logger.Debug("Deleted keys by prefix", "prefix", prefix, "count", len(keys))
	}

	return nil
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}
