package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	"ichi-go/config"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
	once   sync.Once
)

func New() *redis.Client {
	cfg := config.Cfg.Cache
	once.Do(func() {
		options := &redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Username: cfg.Username,
			Password: cfg.Password,
			DB:       cfg.Db,
			PoolSize: cfg.PoolSize,
		}

		if cfg.UseTLS {
			options.TLSConfig = &tls.Config{
				InsecureSkipVerify: cfg.SkipVerify,
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
