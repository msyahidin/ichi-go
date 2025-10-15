package config

import (
	"github.com/spf13/viper"
	"strconv"
)

type RabbitMQConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

func RabbitMQSetDefault() {
	viper.SetDefault("messages.rabbitmq.host", "localhost")
	viper.SetDefault("messages.rabbitmq.port", 5672)
	viper.SetDefault("messages.rabbitmq.username", "admin")
	viper.SetDefault("messages.rabbitmq.password", "admin")
	viper.SetDefault("messages.rabbitmq.enabled", true)
}

func (c RabbitMQConfig) GetRabbitMQURI() string {
	return "amqp://" + c.Username + ":" + c.Password + "@" + c.Host + ":" + strconv.Itoa(c.Port) + "/"
}
