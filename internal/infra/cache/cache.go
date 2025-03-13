package cache

import (
	"context"
)

type Cache interface {
	Ping(ctx context.Context) error
	Set(ctx context.Context, key string, data interface{}, options Options) (bool, error)
	Get(ctx context.Context, key string, data interface{}) (interface{}, error)
	GetRaw(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) (bool, error)
}
