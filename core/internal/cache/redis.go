package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache is a thin wrapper around go-redis for read-through caching.
type RedisCache struct {
	client *redis.Client
}

// New creates a RedisCache from a Redis URL (e.g. "redis://localhost:6379").
// Returns nil if the URL is empty or connection fails -- callers must nil-check.
func New(redisURL string) *RedisCache {
	if redisURL == "" {
		return nil
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		// Non-fatal: callers fall through to Postgres on nil cache.
		return nil
	}

	return &RedisCache{client: client}
}

// Get retrieves a value by key. Returns ("", false) on miss or error.
func (c *RedisCache) Get(ctx context.Context, key string) (string, bool) {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return "", false
	}
	return val, true
}

// Set stores a value with a TTL.
func (c *RedisCache) Set(ctx context.Context, key, value string, ttl time.Duration) {
	c.client.Set(ctx, key, value, ttl)
}

// Del removes one or more keys. Used for cache invalidation.
func (c *RedisCache) Del(ctx context.Context, keys ...string) {
	c.client.Del(ctx, keys...)
}
