package cache

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v4"
	"ichi-go/pkg/logger"
	"time"
)

type CacheImpl struct {
	redisClient *redis.Client
}

func NewCache(redisClient *redis.Client) *CacheImpl {
	return &CacheImpl{redisClient: redisClient}
}

func (c *CacheImpl) Ping(ctx context.Context) error {
	_, err := c.redisClient.Ping(ctx).Result()
	if err != nil {
		return err
	}

	return nil
}

type Options struct {
	Compress   bool
	Expiration time.Duration
}

func (c *CacheImpl) Set(ctx context.Context, key string, data interface{}, options Options) (bool, error) {
	serializedData, err := msgpack.Marshal(&data)
	if err != nil {
		if errors.Is(redis.Nil, err) {
			return false, nil
		}

		logger.Errorf("Failed for marshaling data: %v", err)
		return false, err
	}

	if options.Compress {
		serializedData, err = CompressData(serializedData)
		if err != nil {
			logger.Errorf("Failed for compress data: %v", err)
			return false, err
		}
	}

	err = c.redisClient.Set(ctx, key, serializedData, options.Expiration).Err()
	if err != nil {
		logger.Errorf("Failed save data on Redis: %s # err %s", key, err)
		return false, nil
	}

	return true, nil
}

func (c *CacheImpl) Get(ctx context.Context, key string, data interface{}) (interface{}, error) {

	redisData, err := c.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(redis.Nil, err) {
			return nil, nil
		}
		logger.Errorf("Failed get data from Redis:  %v", err)
		return nil, err
	}

	decompressedData, err := DecompressData(redisData, len(redisData))

	err = msgpack.Unmarshal(decompressedData, data)
	if err != nil {
		logger.Errorf("Failed for unMarshaling data:  %v", err)
		return nil, err
	}

	return data, nil
}

func (c *CacheImpl) GetRaw(ctx context.Context, key string) ([]byte, error) {
	redisData, err := c.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(redis.Nil, err) {
			return nil, nil
		}
		logger.Errorf("Failed get data from Redis:  %v", err)
		return nil, err
	}
	return redisData, nil
}

func (c *CacheImpl) Delete(ctx context.Context, key string) (bool, error) {
	_, err := c.redisClient.Del(ctx, key).Result()
	if err != nil {
		logger.Errorf("Failed for delete data on redis:  %v", err)
		return false, err
	}

	return true, nil
}
