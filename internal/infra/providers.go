package infra

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
	"github.com/uptrace/bun"
	"ichi-go/config"
	"ichi-go/internal/infra/cache"
	"ichi-go/internal/infra/database"
	"ichi-go/internal/infra/queue/rabbitmq"
	"ichi-go/pkg/logger"
)

func Setup(injector do.Injector, cfg *config.Config) {
	do.ProvideValue(injector, cfg)

	do.Provide(injector, provideDatabase(cfg))
	do.Provide(injector, provideCache(cfg))
	do.Provide(injector, provideMessaging(cfg))
	do.Provide(injector, provideMessageProducer(cfg))
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
		if !cfg.Queue().Enabled {
			logger.Infof("Queue system disabled")
			return nil, nil
		}
		conn, err := rabbitmq.NewConnection(cfg.Queue().RabbitMQ)
		if err != nil {
			logger.Warnf("Failed to connect to RabbitMQ: %v", err)
			logger.Warnf("Application will start without queue support")
			return nil, nil
		}
		logger.Debugf("initialized messaging")
		return conn, nil
	}
}

// FIXED: Add detailed logging to diagnose configuration issues
func provideMessageProducer(cfg *config.Config) func(do.Injector) (rabbitmq.MessageProducer, error) {
	return func(i do.Injector) (rabbitmq.MessageProducer, error) {
		if !cfg.Queue().Enabled {
			logger.Debugf("Queue disabled - no producer")
			return nil, nil
		}

		// DIAGNOSTIC: Log the full RabbitMQ config
		logger.Infof("🔍 Diagnosing producer configuration...")
		logger.Infof("   Queue Enabled: %v", cfg.Queue().Enabled)
		logger.Infof("   RabbitMQ Enabled: %v", cfg.Queue().RabbitMQ.Enabled)
		logger.Infof("   Publisher Exchange Name: '%s'", cfg.Queue().RabbitMQ.Publisher.ExchangeName)
		logger.Infof("   Configured Exchanges: %d", len(cfg.Queue().RabbitMQ.Exchanges))

		for i, ex := range cfg.Queue().RabbitMQ.Exchanges {
			logger.Infof("     [%d] Name: '%s', Type: '%s'", i, ex.Name, ex.Type)
		}

		// Check if exchange name is configured
		if cfg.Queue().RabbitMQ.Publisher.ExchangeName == "" {
			logger.Errorf("❌ CRITICAL: Publisher exchange name is empty in config!")
			logger.Errorf("   Check your config file has: queue.rabbitmq.producer.exchange_name")
			return nil, fmt.Errorf("publisher exchange name not configured")
		}

		conn, err := do.Invoke[*rabbitmq.Connection](i)
		if err != nil || conn == nil {
			logger.Warnf("Queue connection unavailable - no producer")
			return nil, nil
		}

		logger.Infof("🔧 Creating message producer...")
		logger.Infof("   Will publish to exchange: '%s'", cfg.Queue().RabbitMQ.Publisher.ExchangeName)

		producer, err := rabbitmq.NewProducer(conn, cfg.Queue().RabbitMQ)
		if err != nil {
			logger.Errorf("Failed to create producer: %v", err)
			return nil, err
		}

		logger.Infof("✅ Message producer created successfully")
		return producer, nil
	}
}
