package messaging

import (
	"github.com/spf13/viper"
	"ichi-go/internal/infra/messaging/rabbitmq"
)

type Config struct {
	Enabled  bool            `mapstructure:"enabled"`
	RabbitMQ rabbitmq.Config `mapstructure:"rabbitmq"`
}

func SetDefault() {
	viper.SetDefault("messaging.enabled", false)
	rabbitmq.RabbitMQSetDefault()
}
