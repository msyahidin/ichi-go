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

	// Queue: named + unnamed providers for all enabled connections
	provideQueueInfra(injector, cfg)

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

// provideQueueInfra registers named DI providers for every enabled queue connection,
// then provides unnamed backward-compat aliases pointing at the default connection.
func provideQueueInfra(injector do.Injector, cfg *config.Config) {
	queueCfg := cfg.Queue()

	for _, nc := range queueCfg.EnabledConnections() {
		nc := nc // capture loop var

		switch nc.Config.Driver {
		case "amqp":
			do.ProvideNamed(injector, "queue.conn."+nc.Name,
				func(i do.Injector) (*rabbitmq.Connection, error) {
					conn, err := rabbitmq.NewConnection(nc.Config.AMQP)
					if err != nil {
						logger.Warnf("amqp[%s]: connect failed: %v — starting without queue", nc.Name, err)
						return nil, nil
					}
					logger.Debugf("initialized amqp connection: %s", nc.Name)
					return conn, nil
				})

			do.ProvideNamed(injector, "queue.producer."+nc.Name,
				func(i do.Injector) (rabbitmq.MessageProducer, error) {
					conn, _ := do.InvokeNamed[*rabbitmq.Connection](i, "queue.conn."+nc.Name)
					if conn == nil {
						return nil, nil
					}
					if nc.Config.AMQP.Publisher.ExchangeName == "" {
						return nil, fmt.Errorf("amqp[%s]: publisher exchange_name not configured", nc.Name)
					}
					p, err := rabbitmq.NewProducer(conn, nc.Config.AMQP)
					if err != nil {
						return nil, fmt.Errorf("amqp[%s]: producer: %w", nc.Name, err)
					}
					logger.Debugf("initialized amqp producer: %s (exchange=%s)", nc.Name, nc.Config.AMQP.Publisher.ExchangeName)
					return p, nil
				})

			do.ProvideNamed(injector, "queue.dispatcher."+nc.Name,
				func(i do.Injector) (queue.Dispatcher, error) {
					producer, _ := do.InvokeNamed[rabbitmq.MessageProducer](i, "queue.producer."+nc.Name)
					return queue.NewDispatcher("amqp", producer, nil)
				})

		case "database":
			do.ProvideNamed(injector, "queue.river."+nc.Name,
				func(i do.Injector) (*riverqueue.Client[*sql.Tx], error) {
					dbKey := nc.Config.Database.Connection
					bunDB, err := do.InvokeNamed[*bun.DB](i, "db."+dbKey)
					if err != nil || bunDB == nil {
						return nil, fmt.Errorf("queue[%s]: database %q not found: %w", nc.Name, dbKey, err)
					}
					registrations := queue.GetRegisteredConsumers(i)
					client, err := buildRiverClient(bunDB, nc.Config.Database, registrations)
					if err != nil {
						return nil, fmt.Errorf("queue[%s]: %w", nc.Name, err)
					}
					logger.Debugf("initialized River client: %s (db=%s)", nc.Name, dbKey)
					return client, nil
				})

			do.ProvideNamed(injector, "queue.dispatcher."+nc.Name,
				func(i do.Injector) (queue.Dispatcher, error) {
					rc, err := do.InvokeNamed[*riverqueue.Client[*sql.Tx]](i, "queue.river."+nc.Name)
					if err != nil {
						return nil, fmt.Errorf("queue[%s]: river client: %w", nc.Name, err)
					}
					return queue.NewDispatcher("database", nil, rc)
				})
		}
	}

	// Unnamed queue.Dispatcher → default connection (existing application callsites unchanged).
	do.Provide(injector, func(i do.Injector) (queue.Dispatcher, error) {
		if !queueCfg.AnyEnabled() {
			logger.Debugf("Queue disabled — no Dispatcher")
			return nil, nil
		}
		def := queueCfg.Default
		d, err := do.InvokeNamed[queue.Dispatcher](i, "queue.dispatcher."+def)
		if err != nil {
			return nil, fmt.Errorf("queue dispatcher (default=%s): %w", def, err)
		}
		logger.Debugf("initialized queue.Dispatcher (default=%s)", def)
		return d, nil
	})

	// Unnamed *rabbitmq.Connection → default AMQP connection (registry.go uses this).
	// Returns nil when the default connection is not AMQP or is disabled.
	do.Provide(injector, func(i do.Injector) (*rabbitmq.Connection, error) {
		def, ok := queueCfg.DefaultConnection()
		if !ok || !def.Enabled || def.Driver != "amqp" {
			return nil, nil
		}
		return do.InvokeNamed[*rabbitmq.Connection](i, "queue.conn."+queueCfg.Default)
	})

	// Unnamed rabbitmq.MessageProducer → default AMQP producer.
	// Returns nil when the default connection is not AMQP or is disabled.
	do.Provide(injector, func(i do.Injector) (rabbitmq.MessageProducer, error) {
		def, ok := queueCfg.DefaultConnection()
		if !ok || !def.Enabled || def.Driver != "amqp" {
			return nil, nil
		}
		return do.InvokeNamed[rabbitmq.MessageProducer](i, "queue.producer."+queueCfg.Default)
	})
}

// buildRiverClient constructs a River client in poll-only mode, sharing bun's *sql.DB.
func buildRiverClient(bunDB *bun.DB, cfg queue.DatabaseBackendConfig, registrations []queue.ConsumerRegistration) (*riverqueue.Client[*sql.Tx], error) {
	workers := riverqueue.NewWorkers()
	riverimpl.RegisterBridgeWorkers(workers, registrations)

	pollInterval := cfg.PollInterval
	if pollInterval == 0 {
		pollInterval = time.Second
	}
	rescueAfter := cfg.RescueStuckJobsAfter
	if rescueAfter == 0 {
		rescueAfter = time.Hour
	}
	maxWorkers := cfg.MaxWorkers
	if maxWorkers == 0 {
		maxWorkers = 50
	}

	return riverqueue.NewClient(riverdatabasesql.New(bunDB.DB), &riverqueue.Config{
		Queues: map[string]riverqueue.QueueConfig{
			riverqueue.QueueDefault: {MaxWorkers: maxWorkers},
			"emails":                {MaxWorkers: 10},
			"notifications":         {MaxWorkers: 20},
		},
		Workers:              workers,
		FetchPollInterval:    pollInterval,
		RescueStuckJobsAfter: rescueAfter,
	})
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
