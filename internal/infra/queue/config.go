package queue

import (
	"time"

	"ichi-go/internal/infra/queue/rabbitmq"

	"github.com/spf13/viper"
)

// Config holds queue system configuration.
type Config struct {
	Enabled  bool            `mapstructure:"enabled"`
	Driver   string          `mapstructure:"driver"` // "rabbitmq" | "river"
	RabbitMQ rabbitmq.Config `mapstructure:"rabbitmq"`
	River    RiverConfig     `mapstructure:"river"`
}

// RiverConfig holds riverqueue-specific settings.
type RiverConfig struct {
	// Database is the connection name from database.connections to use for River.
	// riverdatabasesql shares that connection's *sql.DB (bun) — no extra pool.
	Database             string        `mapstructure:"database"`
	MaxWorkers           int           `mapstructure:"max_workers"`
	PollInterval         time.Duration `mapstructure:"poll_interval"`
	RescueStuckJobsAfter time.Duration `mapstructure:"rescue_stuck_jobs_after"`
}

// SetDefault sets default configuration.
func SetDefault() {
	viper.SetDefault("queue.enabled", false)
	viper.SetDefault("queue.driver", "rabbitmq")
	viper.SetDefault("queue.river.database", "postgres")
	viper.SetDefault("queue.river.max_workers", 50)
	viper.SetDefault("queue.river.poll_interval", time.Second)
	viper.SetDefault("queue.river.rescue_stuck_jobs_after", time.Hour)
	rabbitmq.RabbitMQSetDefault()
}
