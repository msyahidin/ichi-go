package health

import (
	"context"
	"time"

	"ichi-go/internal/infra/queue/rabbitmq"

	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bun"
)

// DatabaseChecker checks database connectivity
type DatabaseChecker struct {
	db *bun.DB
}

func NewDatabaseChecker(db *bun.DB) *DatabaseChecker {
	return &DatabaseChecker{db: db}
}

func (c *DatabaseChecker) Name() string {
	return "database"
}

func (c *DatabaseChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		Name:      c.Name(),
		CheckedAt: time.Now(),
	}

	// Use underlying sql.DB for ping
	err := c.db.DB.PingContext(ctx)
	health.Latency = time.Since(start)

	if err != nil {
		health.Status = StatusUnhealthy
		health.Message = "Database connection failed"
		return health
	}

	health.Status = StatusHealthy
	return health
}

// RedisChecker checks Redis connectivity
type RedisChecker struct {
	client *redis.Client
}

func NewRedisChecker(client *redis.Client) *RedisChecker {
	return &RedisChecker{client: client}
}

func (c *RedisChecker) Name() string {
	return "redis"
}

func (c *RedisChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		Name:      c.Name(),
		CheckedAt: time.Now(),
	}

	_, err := c.client.Ping(ctx).Result()
	health.Latency = time.Since(start)

	if err != nil {
		health.Status = StatusUnhealthy
		health.Message = "Redis connection failed"
		return health
	}

	health.Status = StatusHealthy
	return health
}

// RabbitMQChecker checks RabbitMQ connectivity
type RabbitMQChecker struct {
	conn *rabbitmq.Connection
}

func NewRabbitMQChecker(conn *rabbitmq.Connection) *RabbitMQChecker {
	return &RabbitMQChecker{conn: conn}
}

func (c *RabbitMQChecker) Name() string {
	return "rabbitmq"
}

func (c *RabbitMQChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		Name:      c.Name(),
		CheckedAt: time.Now(),
	}

	if c.conn == nil {
		health.Status = StatusHealthy
		health.Message = "Queue disabled"
		return health
	}

	conn := c.conn.GetConnection()
	health.Latency = time.Since(start)

	if conn == nil || conn.IsClosed() {
		health.Status = StatusUnhealthy
		health.Message = "RabbitMQ connection closed"
		return health
	}

	health.Status = StatusHealthy
	return health
}

// AggregateChecker runs multiple checkers
type AggregateChecker struct {
	checkers []Checker
}

func NewAggregateChecker(checkers ...Checker) *AggregateChecker {
	return &AggregateChecker{checkers: checkers}
}

func (c *AggregateChecker) CheckAll(ctx context.Context) map[string]ComponentHealth {
	results := make(map[string]ComponentHealth)

	// Check all dependencies with timeout
	for _, checker := range c.checkers {
		results[checker.Name()] = checker.Check(ctx)
	}

	return results
}
