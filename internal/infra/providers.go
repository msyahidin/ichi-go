package infra

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
	"github.com/uptrace/bun"
	"ichi-go/config"
	"ichi-go/internal/infra/cache"
	"ichi-go/internal/infra/database"
	"ichi-go/internal/infra/messaging/rabbitmq"
	"ichi-go/pkg/logger"
)

func Setup(injector do.Injector, cfg *config.Config) {
	do.ProvideValue(injector, cfg)

	do.Provide(injector, provideDatabase(cfg))
	do.Provide(injector, provideCache(cfg))
	do.Provide(injector, provideMessaging(cfg))
}

func provideDatabase(cfg *config.Config) func(do.Injector) (*bun.DB, error) {
	return func(i do.Injector) (*bun.DB, error) {
		db, err := database.NewBunClient(cfg.Database())
		if err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
		logger.Debugf("initialized database")
		return db, nil
	}
}

func provideCache(cfg *config.Config) func(do.Injector) (*redis.Client, error) {
	return func(i do.Injector) (*redis.Client, error) {
		client := cache.New(cfg.Cache())
		if client == nil {
			return nil, fmt.Errorf("failed to create cache")
		}
		logger.Debugf("initialized cache")
		return client, nil
	}
}

func provideMessaging(cfg *config.Config) func(do.Injector) (*rabbitmq.Connection, error) {
	return func(i do.Injector) (*rabbitmq.Connection, error) {
		if !cfg.Messaging().Enabled {
			return nil, nil
		}
		conn, err := rabbitmq.NewConnection(cfg.Messaging().RabbitMQ)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
		}
		logger.Debugf("initialized messaging")
		return conn, nil
	}
}
