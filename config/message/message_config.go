package config

import "github.com/spf13/viper"

type MessageConfig struct {
	Enabled  bool           `mapstructure:"enabled"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
}

func SetDefault() {
	viper.SetDefault("messages.enabled", false)
	RabbitMQSetDefault()
}
