package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/infamousrusty/tagsha/internal/metrics"
	"github.com/redis/go-redis/v9"
)

// Cache wraps Redis with JSON serialisation and stale-fallback support.
type Cache struct {
	client     *redis.Client
	defaultTTL time.Duration
}

// New creates and validates a new Redis-backed Cache.
func New(redisURL string, ttl time.Duration) (*Cache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parsing Redis URL: %w", err)
	}

	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connecting to Redis: %w", err)
	}
	return &Cache{client: client, defaultTTL: ttl}, nil
}

// Get retrieves and deserialises a cached value. Returns (false, nil) on miss.
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	val, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		metrics.CacheOperationsTotal.WithLabelValues("get", "miss").Inc()
		return false, nil
	}
	if err != nil {
		metrics.CacheOperationsTotal.WithLabelValues("get", "error").Inc()
		metrics.ErrorsTotal.WithLabelValues("cache").Inc()
		return false, fmt.Errorf("cache get: %w", err)
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		metrics.CacheOperationsTotal.WithLabelValues("get", "unmarshal_error").Inc()
		return false, fmt.Errorf("cache unmarshal: %w", err)
	}
	metrics.CacheOperationsTotal.WithLabelValues("get", "hit").Inc()
	return true, nil
}

// Set serialises and stores a value with the given TTL.
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.defaultTTL
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal: %w", err)
	}
	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		metrics.CacheOperationsTotal.WithLabelValues("set", "error").Inc()
		return fmt.Errorf("cache set: %w", err)
	}
	metrics.CacheOperationsTotal.WithLabelValues("set", "ok").Inc()
	return nil
}

// SetStale stores a long-lived fallback copy keyed with a "stale:" prefix.
func (c *Cache) SetStale(ctx context.Context, key string, value interface{}) error {
	return c.Set(ctx, "stale:"+key, value, time.Hour)
}

// GetStale retrieves a stale fallback value.
func (c *Cache) GetStale(ctx context.Context, key string, dest interface{}) (bool, error) {
	return c.Get(ctx, "stale:"+key, dest)
}

// Ping checks Redis connectivity.
func (c *Cache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// TagsKey returns the canonical cache key for a repository tag list.
func TagsKey(owner, repo string) string {
	return fmt.Sprintf("tagsha:tags:%s:%s", owner, repo)
}
