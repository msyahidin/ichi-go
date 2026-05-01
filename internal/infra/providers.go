package infra

import (
	"database/sql"
	"fmt"
	"time"

	riverqueue "github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverdatabasesql"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
	"github.com/uptrace/bun"

	"ichi-go/config"
	"ichi-go/internal/infra/authz/adapter"
	"ichi-go/internal/infra/authz/cache"
	"ichi-go/internal/infra/authz/enforcer"
	infraCache "ichi-go/internal/infra/cache"
	"ichi-go/internal/infra/database"
	"ichi-go/internal/infra/queue"
	"ichi-go/internal/infra/queue/rabbitmq"
	riverimpl "ichi-go/internal/infra/queue/river"
	"ichi-go/pkg/logger"
	"ichi-go/pkg/rbac"
)

func Setup(injector do.Injector, cfg *config.Config) {
	do.ProvideValue(injector, cfg)

	// Core infrastructure
	provideDatabases(injector, cfg)
	do.Provide(injector, provideCache(cfg))
	do.Provide(injector, provideMessaging(cfg))
	do.Provide(injector, provideMessageProducer(cfg))

	// Queue dispatcher (RabbitMQ or River depending on config)
	do.Provide(injector, provideRiverClient(cfg))
	do.Provide(injector, provideQueueDispatcher(cfg))

	// RBAC infrastructure
	do.Provide(injector, provideRBACConfig(cfg))
	do.Provide(injector, provideRedisCache(cfg))
	do.Provide(injector, provideCasbinAdapter(cfg))
	do.Provide(injector, provideEnforcer(cfg))
}

func provideDatabases(injector do.Injector, cfg *config.Config) {
	for name, dbCfg := range cfg.Databases() {
		name, dbCfg := name, dbCfg // capture loop vars

		switch dbCfg.Driver {
		case "mysql":
			do.ProvideNamed(injector, "db."+name, func(i do.Injector) (*bun.DB, error) {
				db, err := database.NewMySQLClient(&dbCfg)
				if err != nil {
					return nil, fmt.Errorf("mysql[%s]: %w", name, err)
				}
				logger.Debugf("initialized database connection: %s (mysql)", name)
				return db, nil
			})

		case "postgres":
			do.ProvideNamed(injector, "db."+name, func(i do.Injector) (*bun.DB, error) {
				db, err := database.NewPostgresClient(&dbCfg)
				if err != nil {
					return nil, fmt.Errorf("postgres[%s]: %w", name, err)
				}
				logger.Debugf("initialized database connection: %s (postgres)", name)
				return db, nil
			})

		default:
			logger.Warnf("unknown database driver %q for connection %q — skipping", dbCfg.Driver, name)
		}
	}

	// Unnamed *bun.DB → primary connection (keeps all existing repos working unchanged)
	primary := cfg.PrimaryDatabase()
	do.Provide(injector, func(i do.Injector) (*bun.DB, error) {
		return do.InvokeNamed[*bun.DB](i, "db."+primary)
	})
}

func provideCache(cfg *config.Config) func(do.Injector) (*redis.Client, error) {
	return func(i do.Injector) (*redis.Client, error) {
		client := infraCache.New(cfg.Cache())
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

func provideRiverClient(cfg *config.Config) func(do.Injector) (*riverqueue.Client[*sql.Tx], error) {
	return func(i do.Injector) (*riverqueue.Client[*sql.Tx], error) {
		queueCfg := cfg.Queue()
		if !queueCfg.Enabled || queueCfg.Driver != "river" {
			logger.Debugf("River client not needed (driver=%s)", queueCfg.Driver)
			return nil, nil
		}

		dbKey := queueCfg.River.Database
		bunDB, err := do.InvokeNamed[*bun.DB](i, "db."+dbKey)
		if err != nil || bunDB == nil {
			return nil, fmt.Errorf("river: database %q not found in DI (check databases config): %w", dbKey, err)
		}

		workers := riverqueue.NewWorkers()
		registrations := queue.GetRegisteredConsumers(i)
		riverimpl.RegisterBridgeWorkers(workers, registrations)

		riverCfg := queueCfg.River
		pollInterval := riverCfg.PollInterval
		if pollInterval == 0 {
			pollInterval = time.Second
		}
		rescueAfter := riverCfg.RescueStuckJobsAfter
		if rescueAfter == 0 {
			rescueAfter = time.Hour
		}
		maxWorkers := riverCfg.MaxWorkers
		if maxWorkers == 0 {
			maxWorkers = 50
		}

		// riverdatabasesql wraps the *sql.DB that bun already owns — poll-only mode.
		client, err := riverqueue.NewClient(riverdatabasesql.New(bunDB.DB), &riverqueue.Config{
			Queues: map[string]riverqueue.QueueConfig{
				riverqueue.QueueDefault: {MaxWorkers: maxWorkers},
				"emails":                {MaxWorkers: 10},
				"notifications":         {MaxWorkers: 20},
			},
			Workers:              workers,
			FetchPollInterval:    pollInterval,
			RescueStuckJobsAfter: rescueAfter,
		})
		if err != nil {
			return nil, fmt.Errorf("river: failed to create client: %w", err)
		}

		logger.Debugf("initialized River client (db=%s, maxWorkers=%d, poll-only)", dbKey, maxWorkers)
		return client, nil
	}
}

func provideQueueDispatcher(cfg *config.Config) func(do.Injector) (queue.Dispatcher, error) {
	return func(i do.Injector) (queue.Dispatcher, error) {
		if !cfg.Queue().Enabled {
			logger.Debugf("Queue disabled — no Dispatcher")
			return nil, nil
		}

		producer, _ := do.Invoke[rabbitmq.MessageProducer](i)
		riverClient, _ := do.Invoke[*riverqueue.Client[*sql.Tx]](i)

		d, err := queue.NewDispatcher(cfg.Queue().Driver, producer, riverClient)
		if err != nil {
			return nil, err
		}
		logger.Debugf("initialized queue.Dispatcher (driver=%s)", cfg.Queue().Driver)
		return d, nil
	}
}

// RBAC Infrastructure Providers

func provideRBACConfig(cfg *config.Config) func(do.Injector) (*rbac.Config, error) {
	return func(i do.Injector) (*rbac.Config, error) {
		// Get RBAC config from application config
		rbacCfg := cfg.RBAC()
		if rbacCfg == nil {
			return nil, fmt.Errorf("failed to load RBAC config")
		}
		logger.Debugf("initialized RBAC config: mode=%s", rbacCfg.Mode)
		return rbacCfg, nil
	}
}

func provideRedisCache(cfg *config.Config) func(do.Injector) (*cache.RedisCache, error) {
	return func(i do.Injector) (*cache.RedisCache, error) {
		redisClient, err := do.Invoke[*redis.Client](i)
		if err != nil || redisClient == nil {
			logger.Warnf("Redis not available for RBAC cache: %v", err)
			return nil, nil
		}

		rbacCfg := do.MustInvoke[*rbac.Config](i)
		redisCache, err := cache.NewRedisCache(redisClient, rbacCfg.Cache.Compression)
		if err != nil {
			return nil, fmt.Errorf("failed to create Redis cache: %w", err)
		}

		logger.Debugf("initialized RBAC Redis cache (compression=%v)", rbacCfg.Cache.Compression)
		return redisCache, nil
	}
}

func provideCasbinAdapter(cfg *config.Config) func(do.Injector) (*adapter.BunAdapter, error) {
	return func(i do.Injector) (*adapter.BunAdapter, error) {
		db := do.MustInvoke[*bun.DB](i)

		bunAdapter, err := adapter.NewBunAdapter(db)
		if err != nil {
			return nil, fmt.Errorf("failed to create Casbin adapter: %w", err)
		}

		logger.Debugf("initialized Casbin Bun adapter")
		return bunAdapter, nil
	}
}

func provideEnforcer(cfg *config.Config) func(do.Injector) (*enforcer.Enforcer, error) {
	return func(i do.Injector) (*enforcer.Enforcer, error) {
		db := do.MustInvoke[*bun.DB](i)
		rbacCfg := do.MustInvoke[*rbac.Config](i)

		// Create enforcer
		enf, err := enforcer.New(db, rbacCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create RBAC enforcer: %w", err)
		}

		// Load policies
		if err := enf.ReloadPolicy(); err != nil {
			logger.Warnf("Failed to load initial policies: %v", err)
		}

		logger.Infof("✅ RBAC enforcer initialized (mode=%s, policies=%d)",
			rbacCfg.Mode, enf.GetPolicyCount())

		return enf, nil
	}
}
