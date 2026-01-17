package cache

import (
	"context"
	"fmt"
	"time"

	"ichi-go/pkg/logger"
	"ichi-go/pkg/rbac"
)

// DecisionCache provides multi-tier caching for RBAC permission decisions
// L1: In-memory LRU cache (fast, per-instance)
// L2: Redis cache (shared, compressed)
type DecisionCache struct {
	memoryCache *MemoryCache
	redisCache  *RedisCache
	config      *rbac.CacheConfig
	stats       *CacheStats
}

// CacheStats tracks cache performance metrics
type CacheStats struct {
	L1Hits   int64
	L1Misses int64
	L2Hits   int64
	L2Misses int64
}

// NewDecisionCache creates a new multi-tier decision cache
func NewDecisionCache(redisCache *RedisCache, config *rbac.CacheConfig) (*DecisionCache, error) {
	if config == nil {
		return nil, fmt.Errorf("cache config is required")
	}

	if !config.Enabled {
		logger.Infof("RBAC decision cache is disabled")
		return &DecisionCache{
			config: config,
			stats:  &CacheStats{},
		}, nil
	}

	// Parse TTLs
	memoryTTL, err := time.ParseDuration(config.MemoryTTL)
	if err != nil {
		return nil, fmt.Errorf("invalid memory_ttl: %w", err)
	}

	// Create L1 memory cache
	memoryCache, err := NewMemoryCache(config.MaxSize, memoryTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to create memory cache: %w", err)
	}

	logger.Infof("RBAC decision cache initialized: L1=%dms L2=%s maxSize=%d",
		memoryTTL.Milliseconds(), config.RedisTTL, config.MaxSize)

	return &DecisionCache{
		memoryCache: memoryCache,
		redisCache:  redisCache,
		config:      config,
		stats:       &CacheStats{},
	}, nil
}

// Get retrieves a cached decision
// Returns: (allowed bool, found bool, error)
func (c *DecisionCache) Get(ctx context.Context, key string) (bool, bool, error) {
	if !c.config.Enabled {
		return false, false, nil
	}

	// Try L1 cache (memory)
	if c.memoryCache != nil {
		if value, found := c.memoryCache.Get(key); found {
			c.stats.L1Hits++
			logger.WithContext(ctx).Debugf("L1 cache hit: %s", key)
			return value, true, nil
		}
		c.stats.L1Misses++
	}

	// Try L2 cache (Redis)
	if c.redisCache != nil {
		value, found, err := c.redisCache.GetDecision(ctx, key)
		if err != nil {
			logger.WithContext(ctx).Errorf("L2 cache error: %v", err)
			return false, false, err
		}

		if found {
			c.stats.L2Hits++
			logger.WithContext(ctx).Debugf("L2 cache hit: %s", key)

			// Populate L1 cache
			if c.memoryCache != nil {
				c.memoryCache.Set(key, value)
			}

			return value, true, nil
		}
		c.stats.L2Misses++
	}

	return false, false, nil
}

// Set stores a decision in cache
func (c *DecisionCache) Set(ctx context.Context, key string, allowed bool) error {
	if !c.config.Enabled {
		return nil
	}

	// Store in L1 (memory)
	if c.memoryCache != nil {
		c.memoryCache.Set(key, allowed)
	}

	// Store in L2 (Redis)
	if c.redisCache != nil {
		ttl, err := time.ParseDuration(c.config.RedisTTL)
		if err != nil {
			return fmt.Errorf("invalid redis_ttl: %w", err)
		}

		if err := c.redisCache.SetDecision(ctx, key, allowed, ttl); err != nil {
			logger.WithContext(ctx).Errorf("Failed to set L2 cache: %v", err)
			return err
		}
	}

	logger.WithContext(ctx).Debugf("Cached decision: %s = %v", key, allowed)
	return nil
}

// Delete removes a specific key from cache
func (c *DecisionCache) Delete(ctx context.Context, key string) error {
	if !c.config.Enabled {
		return nil
	}

	// Delete from L1
	if c.memoryCache != nil {
		c.memoryCache.Delete(key)
	}

	// Delete from L2
	if c.redisCache != nil {
		if err := c.redisCache.Delete(ctx, key); err != nil {
			return err
		}
	}

	logger.WithContext(ctx).Debugf("Deleted cache key: %s", key)
	return nil
}

// DeletePattern removes all keys matching a pattern
func (c *DecisionCache) DeletePattern(ctx context.Context, pattern string) error {
	if !c.config.Enabled {
		return nil
	}

	// Clear L1 (pattern matching not efficient, so clear all)
	if c.memoryCache != nil {
		c.memoryCache.Clear()
		logger.WithContext(ctx).Debugf("Cleared L1 cache for pattern: %s", pattern)
	}

	// Delete from L2 with pattern
	if c.redisCache != nil {
		if err := c.redisCache.DeletePattern(ctx, pattern); err != nil {
			return err
		}
	}

	logger.WithContext(ctx).Infof("Deleted cache pattern: %s", pattern)
	return nil
}

// Clear removes all entries from cache
func (c *DecisionCache) Clear(ctx context.Context) error {
	if !c.config.Enabled {
		return nil
	}

	// Clear L1
	if c.memoryCache != nil {
		c.memoryCache.Clear()
	}

	// Clear L2 (Redis - use pattern to clear only RBAC keys)
	if c.redisCache != nil {
		if err := c.redisCache.DeletePattern(ctx, "rbac:*"); err != nil {
			return err
		}
	}

	logger.WithContext(ctx).Infof("Cleared all RBAC cache")
	return nil
}

// GetStats returns cache performance statistics
func (c *DecisionCache) GetStats() CacheStats {
	return *c.stats
}

// GetHitRatio returns the overall cache hit ratio (0.0 - 1.0)
func (c *DecisionCache) GetHitRatio() float64 {
	totalHits := c.stats.L1Hits + c.stats.L2Hits
	totalRequests := totalHits + c.stats.L2Misses

	if totalRequests == 0 {
		return 0.0
	}

	return float64(totalHits) / float64(totalRequests)
}

// ResetStats resets cache statistics
func (c *DecisionCache) ResetStats() {
	c.stats = &CacheStats{}
}

// MakeCacheKey generates a cache key for a permission check
func MakeCacheKey(tenantID, userID, resource, action string) string {
	return fmt.Sprintf("rbac:decision:%s:%s:%s:%s", tenantID, userID, resource, action)
}

// MakeTenantPattern generates a cache key pattern for a tenant
func MakeTenantPattern(tenantID string) string {
	return fmt.Sprintf("rbac:decision:%s:*", tenantID)
}

// MakeUserPattern generates a cache key pattern for a user in a tenant
func MakeUserPattern(tenantID, userID string) string {
	return fmt.Sprintf("rbac:decision:%s:%s:*", tenantID, userID)
}
