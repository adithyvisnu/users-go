package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	DefaultTTL   = 5 * time.Minute
	UserTTL      = 10 * time.Minute
	UserListTTL  = 2 * time.Minute
)

// Cache wraps a Redis client and exposes typed helpers.
type Cache struct {
	client *redis.Client
}

// New creates a new Cache backed by the given Redis client.
func New(client *redis.Client) *Cache {
	return &Cache{client: client}
}

// ─── Generic helpers ──────────────────────────────────────────────────────────

// Get retrieves a value and JSON-unmarshals it into dest.
// Returns (false, nil) on a cache miss (redis.Nil).
func (c *Cache) Get(ctx context.Context, key string, dest any) (bool, error) {
	raw, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return false, nil // cache miss — not an error
	}
	if err != nil {
		return false, fmt.Errorf("cache get %q: %w", key, err)
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return false, fmt.Errorf("cache unmarshal %q: %w", key, err)
	}
	return true, nil
}

// Set JSON-marshals value and stores it with the given TTL.
func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal %q: %w", key, err)
	}
	if err := c.client.Set(ctx, key, raw, ttl).Err(); err != nil {
		return fmt.Errorf("cache set %q: %w", key, err)
	}
	return nil
}

// Delete removes one or more keys atomically.
func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("cache delete: %w", err)
	}
	return nil
}

// DeletePattern removes all keys matching a glob pattern.
// Use sparingly — SCAN-based, safe on large keyspaces.
func (c *Cache) DeletePattern(ctx context.Context, pattern string) error {
	var cursor uint64
	for {
		keys, next, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("cache scan %q: %w", pattern, err)
		}
		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("cache delete batch: %w", err)
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}

// ─── Key builders ─────────────────────────────────────────────────────────────

// UserKey returns the cache key for a single user.
//   users-api:tenant:<tenantID>:user:<userID>
func UserKey(tenantID, userID string) string {
	return fmt.Sprintf("users-api:tenant:%s:user:%s", tenantID, userID)
}

// UserListKey returns the cache key for a paginated user list.
//   users-api:tenant:<tenantID>:users:page:<page>:limit:<limit>
func UserListKey(tenantID string, page, limit int) string {
	return fmt.Sprintf("users-api:tenant:%s:users:page:%d:limit:%d", tenantID, page, limit)
}

// UserListPattern returns a glob that matches all list keys for a tenant.
// Used to bust the entire list cache when a user is mutated.
func UserListPattern(tenantID string) string {
	return fmt.Sprintf("users-api:tenant:%s:users:*", tenantID)
}

// Ping checks connectivity.
func (c *Cache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
