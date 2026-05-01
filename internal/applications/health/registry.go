package health

import (
	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
	"github.com/uptrace/bun"

	"ichi-go/config"
	"ichi-go/internal/applications/health/controller"
	"ichi-go/internal/applications/health/service"
	"ichi-go/internal/infra/queue/rabbitmq"
	"ichi-go/pkg/health"
)

func Register(injector do.Injector, serviceName string, e *echo.Echo, cfg *config.Config) {
	// Create health checkers
	db := do.MustInvoke[*bun.DB](injector)
	redisClient := do.MustInvoke[*redis.Client](injector)

	dbChecker := health.NewDatabaseChecker(db)
	redisChecker := health.NewRedisChecker(redisClient)

	// RabbitMQ checker (optional - may be nil if queue is disabled)
	checkers := []health.Checker{dbChecker, redisChecker}

	// RabbitMQ checker — only when the default queue connection is actually AMQP/RabbitMQ.
	// AnyEnabled() could also be true for database-backed (River) queues, so we inspect
	// the resolved *rabbitmq.Connection instead.
	if cfg.Queue().AnyEnabled() {
		mqConn, err := do.Invoke[*rabbitmq.Connection](injector)
		if err == nil && mqConn != nil {
			// Default connection is AMQP — attach RabbitMQ liveness check.
			mqChecker := health.NewRabbitMQChecker(mqConn)
			checkers = append(checkers, mqChecker)
		}
		// When the default connection is database-backed (River), connectivity is
		// already covered by the database checker above.
		// TODO: add health.NewRiverQueueChecker once implemented.
	}

	aggregateChecker := health.NewAggregateChecker(checkers...)

	// Create service
	healthService := service.NewHealthService(aggregateChecker, cfg.App().Version)

	// Create controller
	healthController := controller.NewHealthController(healthService)

	healthController.RegisterRoutes(e, serviceName)
}
