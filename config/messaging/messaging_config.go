package messaging

import (
	"github.com/spf13/viper"
	"ichi-go/internal/infra/messaging/rabbitmq"
)

type MessagingConfig struct {
	Enabled  bool                    `mapstructure:"enabled"`
	RabbitMQ rabbitmq.RabbitMQConfig `mapstructure:"rabbitmq"`
}

func SetDefault() {
	viper.SetDefault("messaging.enabled", false)
	rabbitmq.RabbitMQSetDefault()
}
