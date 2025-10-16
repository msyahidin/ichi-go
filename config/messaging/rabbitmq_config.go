package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type RabbitMQConfig struct {
	Enabled    bool                   `mapstructure:"enabled"`
	Connection RabbitConnectionConfig `yaml:"connection" mapstructure:"connection"`
	Exchanges  []ExchangeConfig       `yaml:"exchanges" mapstructure:"exchanges"`
	Consumers  []ConsumerConfig       `yaml:"consumers" mapstructure:"consumers"`
	Publisher  PublisherConfig        `yaml:"publisher" mapstructure:"publisher"`
}

type RabbitConnectionConfig struct {
	Host           string `yaml:"host" mapstructure:"host"`
	Port           int    `yaml:"port" mapstructure:"port"`
	Username       string `yaml:"username" mapstructure:"username"`
	Password       string `yaml:"password" mapstructure:"password"`
	ConnectionName string `yaml:"connection_name" mapstructure:"connection_name"`
}

type ExchangeConfig struct {
	Name       string `yaml:"name" mapstructure:"name"`
	Type       string `yaml:"type" mapstructure:"type"` //direct, topic, fanout, headers
	Durable    bool   `yaml:"durable" mapstructure:"durable"`
	AutoDelete bool   `yaml:"auto_delete" mapstructure:"auto_delete"`
	Internal   bool   `yaml:"internal" mapstructure:"internal"`
	NoWait     bool   `yaml:"no_wait" mapstructure:"no_wait"`
}

type ConsumerConfig struct {
	Name    string `yaml:"name" mapstructure:"name"`
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`

	Queue QueueConfig `yaml:"queue" mapstructure:"queue"`

	ExchangeName string   `yaml:"exchange_name" mapstructure:"exchange_name"`
	RoutingKeys  []string `yaml:"routing_keys" mapstructure:"routing_keys"`

	PrefetchCount  int `yaml:"prefetch_count" mapstructure:"prefetch_count"`
	WorkerPoolSize int `yaml:"worker_pool_size" mapstructure:"worker_pool_size"`

	AutoAck   bool `yaml:"auto_ack" mapstructure:"auto_ack"`
	Exclusive bool `yaml:"exclusive" mapstructure:"exclusive"`

	// Consumer tag for RabbitMQ management
	ConsumerTag string `yaml:"consumer_tag" mapstructure:"consumer_tag"`
}

type QueueConfig struct {
	Name       string `yaml:"name" mapstructure:"name"`
	Durable    bool   `yaml:"durable" mapstructure:"durable"`
	AutoDelete bool   `yaml:"auto_delete" mapstructure:"auto_delete"`
	Exclusive  bool   `yaml:"exclusive" mapstructure:"exclusive"`
	NoWait     bool   `yaml:"no_wait" mapstructure:"no_wait"`
}

type PublisherConfig struct {
	ExchangeName string `yaml:"exchange_name" mapstructure:"exchange_name"`
}

func GetConsumerByName(config *RabbitMQConfig, name string) (*ConsumerConfig, error) {
	for _, consumer := range config.Consumers {
		if consumer.Name == name {
			return &consumer, nil
		}
	}
	return nil, fmt.Errorf("consumer '%s' not found", name)
}

func GetExchangeByName(config *RabbitMQConfig, name string) (*ExchangeConfig, error) {
	for _, exchange := range config.Exchanges {
		if exchange.Name == name {
			return &exchange, nil
		}
	}
	return nil, fmt.Errorf("consumer '%s' not found", name)
}

func GetEnabledConsumers(config *RabbitMQConfig) []ConsumerConfig {
	var enabled []ConsumerConfig
	for _, consumer := range config.Consumers {
		if consumer.Enabled {
			enabled = append(enabled, consumer)
		}
	}
	return enabled
}

func RabbitMQSetDefault() {
	viper.SetDefault("messaging.rabbitmq.host", "localhost")
	viper.SetDefault("messaging.rabbitmq.port", 5672)
	viper.SetDefault("messaging.rabbitmq.username", "admin")
	viper.SetDefault("messaging.rabbitmq.password", "admin")
	viper.SetDefault("messaging.rabbitmq.enabled", true)
}

func GetRabbitMQURI(c RabbitConnectionConfig) string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
	)
}
