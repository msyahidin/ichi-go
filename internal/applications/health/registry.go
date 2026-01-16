package health

import (
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
	"github.com/uptrace/bun"

	"ichi-go/config"
	"ichi-go/internal/applications/health/controller"
	"ichi-go/internal/applications/health/service"
	"ichi-go/internal/infra/queue/rabbitmq"
	"ichi-go/pkg/health"
)

func Register(injector do.Injector, e *echo.Echo, cfg *config.Config) {
	// Create health checkers
	db := do.MustInvoke[*bun.DB](injector)
	redisClient := do.MustInvoke[*redis.Client](injector)

	dbChecker := health.NewDatabaseChecker(db)
	redisChecker := health.NewRedisChecker(redisClient)

	// RabbitMQ checker (optional - may be nil if queue is disabled)
	checkers := []health.Checker{dbChecker, redisChecker}

	if cfg.Queue().Enabled {
		mqConn, err := do.Invoke[*rabbitmq.Connection](injector)
		if err == nil && mqConn != nil {
			mqChecker := health.NewRabbitMQChecker(mqConn)
			checkers = append(checkers, mqChecker)
		}
	}

	aggregateChecker := health.NewAggregateChecker(checkers...)

	// Create service
	healthService := service.NewHealthService(aggregateChecker, cfg.App().Version)

	// Create controller
	healthController := controller.NewHealthController(healthService)

	// Register routes (NO auth, NO versioning)
	healthController.RegisterRoutes(e)
}
