package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	cacheConfig "ichi-go/config/cache"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
	once   sync.Once
)

func New(cacheConfig *cacheConfig.CacheConfig) *redis.Client {
	once.Do(func() {
		options := &redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cacheConfig.Host, cacheConfig.Port),
			Username: cacheConfig.Username,
			Password: cacheConfig.Password,
			DB:       cacheConfig.Db,
			PoolSize: cacheConfig.PoolSize,
		}

		if cacheConfig.UseTLS {
			options.TLSConfig = &tls.Config{
				InsecureSkipVerify: cacheConfig.SkipVerify,
			}
		}

		Client = redis.NewClient(options)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := Client.Ping(ctx).Err(); err != nil {
			panic(fmt.Sprintf("failed to connect to redis: %v", err))
		}
	})

	return Client
}
