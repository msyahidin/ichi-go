package queue

import (
	"ichi-go/internal/infra/queue/rabbitmq"

	"github.com/spf13/viper"
)

// Config holds queue system configuration.
// Renamed from "messaging" for clarity.
type Config struct {
	Enabled  bool            `mapstructure:"enabled"`
	RabbitMQ rabbitmq.Config `mapstructure:"rabbitmq"`
}

// SetDefault sets default configuration.
func SetDefault() {
	viper.SetDefault("queue.enabled", false)
	rabbitmq.RabbitMQSetDefault()
}
