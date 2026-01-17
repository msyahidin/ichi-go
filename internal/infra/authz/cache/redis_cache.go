package cache

import (
	"context"
	"fmt"
	"time"

	infraCache "ichi-go/internal/infra/cache"
	"ichi-go/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements L2 Redis cache for RBAC decisions
type RedisCache struct {
	client      *redis.Client
	compression bool
}

// DecisionValue represents a cached decision with metadata
type DecisionValue struct {
	Allowed   bool      `json:"allowed"`
	CachedAt  time.Time `json:"cached_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(redisClient *redis.Client, compression bool) (*RedisCache, error) {
	if redisClient == nil {
		return nil, fmt.Errorf("redis client is required")
	}

	return &RedisCache{
		client:      redisClient,
		compression: compression,
	}, nil
}

// GetDecision retrieves a cached decision from Redis
func (r *RedisCache) GetDecision(ctx context.Context, key string) (bool, bool, error) {
	var value DecisionValue

	// Use the infra cache with compression
	cache := infraCache.NewCache(r.client)

	_, err := cache.Get(ctx, key, &value)
	if err != nil {
		// Key not found
		return false, false, nil
	}

	// Check if expired
	if time.Now().After(value.ExpiresAt) {
		// Expired, delete it
		_, _ = cache.Delete(ctx, key)
		return false, false, nil
	}

	return value.Allowed, true, nil
}

// SetDecision stores a decision in Redis with compression
func (r *RedisCache) SetDecision(ctx context.Context, key string, allowed bool, ttl time.Duration) error {
	now := time.Now()

	value := DecisionValue{
		Allowed:   allowed,
		CachedAt:  now,
		ExpiresAt: now.Add(ttl),
	}

	// Use the infra cache with compression
	cache := infraCache.NewCache(r.client)

	options := infraCache.Options{
		Compress:   r.compression,
		Expiration: ttl,
	}

	if _, err := cache.Set(ctx, key, value, options); err != nil {
		return fmt.Errorf("failed to set decision in Redis: %w", err)
	}

	return nil
}

// Delete removes a key from Redis
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	cache := infraCache.NewCache(r.client)

	if _, err := cache.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to delete key from Redis: %w", err)
	}
	return nil
}

// Keys returns all keys matching a pattern
func (r *RedisCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	var keys []string
	var cursor uint64

	for {
		var scanKeys []string
		var err error

		scanKeys, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan keys: %w", err)
		}

		keys = append(keys, scanKeys...)

		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

// DeletePattern removes all keys matching a pattern
func (r *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	// Scan for matching keys
	keys, err := r.Keys(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	// Delete all matching keys
	cache := infraCache.NewCache(r.client)
	for _, key := range keys {
		if _, err := cache.Delete(ctx, key); err != nil {
			logger.WithContext(ctx).Errorf("Failed to delete key %s: %v", key, err)
		}
	}

	logger.WithContext(ctx).Infof("Deleted %d keys matching pattern: %s", len(keys), pattern)
	return nil
}

// GetPolicyCache retrieves cached policies for a tenant
func (r *RedisCache) GetPolicyCache(ctx context.Context, tenantID string) ([]byte, bool, error) {
	key := fmt.Sprintf("rbac:policies:%s", tenantID)

	cache := infraCache.NewCache(r.client)

	data, err := cache.GetRaw(ctx, key)
	if err != nil {
		return nil, false, nil
	}

	if data == nil {
		return nil, false, nil
	}

	// Decompress if needed (handled by infra cache)
	return data, true, nil
}

// SetPolicyCache stores policies for a tenant
func (r *RedisCache) SetPolicyCache(ctx context.Context, tenantID string, policies []byte, ttl time.Duration) error {
	key := fmt.Sprintf("rbac:policies:%s", tenantID)

	cache := infraCache.NewCache(r.client)

	options := infraCache.Options{
		Compress:   r.compression,
		Expiration: ttl,
	}

	if _, err := cache.Set(ctx, key, policies, options); err != nil {
		return fmt.Errorf("failed to cache policies: %w", err)
	}

	logger.WithContext(ctx).Infof("Cached policies for tenant %s (compressed: %v)", tenantID, r.compression)

	return nil
}
