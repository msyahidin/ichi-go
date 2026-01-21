package queue

import (
	"ichi-go/internal/infra/queue/rabbitmq"
)

// Config holds queue system configuration.
// Renamed from "messaging" for clarity.
type Config struct {
	Enabled  bool            `mapstructure:"enabled"`
	RabbitMQ rabbitmq.Config `mapstructure:"rabbitmq"`
}

func NewConfig() Config {
	return Config{
		Enabled:  false,
		RabbitMQ: rabbitmq.NewConfig(),
	}
}

//// SetDefault sets default configuration.
//func SetDefault() {
//	viper.SetDefault("queue.enabled", false)
//	rabbitmq.RabbitMQSetDefault()
//}
