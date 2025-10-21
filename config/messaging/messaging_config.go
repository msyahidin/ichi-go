package config

import "github.com/spf13/viper"

type MessagingConfig struct {
	Enabled  bool           `mapstructure:"enabled"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
}

func SetDefault() {
	viper.SetDefault("messaging.enabled", false)
	RabbitMQSetDefault()
}
