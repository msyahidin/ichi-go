package queue

import (
	"time"

	"ichi-go/internal/infra/queue/rabbitmq"

	"github.com/spf13/viper"
)

// QueueSchema mirrors the `queue:` YAML block — same pattern as DatabaseSchema.
type QueueSchema struct {
	Default     string                     `mapstructure:"default"`
	Connections map[string]ConnectionConfig `mapstructure:"connections"`
}

// ConnectionConfig is a union of all supported backends.
// Only the fields that match Driver are populated at runtime.
type ConnectionConfig struct {
	Enabled  bool                  `mapstructure:"enabled"`
	Driver   string                `mapstructure:"driver"` // "amqp" | "database"
	AMQP     rabbitmq.Config       `mapstructure:"amqp"`
	Database DatabaseBackendConfig `mapstructure:"database"`
}

// DatabaseBackendConfig holds River queue settings for the "database" driver.
// Connection refers to a key in database.connections — the River client shares that pool.
type DatabaseBackendConfig struct {
	Connection           string        `mapstructure:"connection"`
	MaxWorkers           int           `mapstructure:"max_workers"`
	PollInterval         time.Duration `mapstructure:"poll_interval"`
	RescueStuckJobsAfter time.Duration `mapstructure:"rescue_stuck_jobs_after"`
}

// NamedConnection pairs a connection name with its resolved config.
type NamedConnection struct {
	Name   string
	Config ConnectionConfig
}

// AnyEnabled reports whether at least one connection is enabled.
func (s *QueueSchema) AnyEnabled() bool {
	for _, c := range s.Connections {
		if c.Enabled {
			return true
		}
	}
	return false
}

// EnabledConnections returns all connections with Enabled == true.
func (s *QueueSchema) EnabledConnections() []NamedConnection {
	var out []NamedConnection
	for name, c := range s.Connections {
		if c.Enabled {
			out = append(out, NamedConnection{Name: name, Config: c})
		}
	}
	return out
}

// DefaultConnection returns the config for the default connection name.
func (s *QueueSchema) DefaultConnection() (ConnectionConfig, bool) {
	c, ok := s.Connections[s.Default]
	return c, ok
}

// DefaultAMQPConfig returns the rabbitmq.Config for the default connection.
// Returns false when the default connection is not an AMQP driver.
func (s *QueueSchema) DefaultAMQPConfig() (rabbitmq.Config, bool) {
	c, ok := s.DefaultConnection()
	if !ok || c.Driver != "amqp" {
		return rabbitmq.Config{}, false
	}
	return c.AMQP, true
}

// SetDefault registers Viper defaults for the queue block.
func SetDefault() {
	viper.SetDefault("queue.default", "amqp")
	viper.SetDefault("queue.connections.amqp.enabled", false)
	viper.SetDefault("queue.connections.amqp.driver", "amqp")
	viper.SetDefault("queue.connections.database.enabled", false)
	viper.SetDefault("queue.connections.database.driver", "database")
	viper.SetDefault("queue.connections.database.database.connection", "postgres")
	viper.SetDefault("queue.connections.database.database.max_workers", 50)
	viper.SetDefault("queue.connections.database.database.poll_interval", time.Second)
	viper.SetDefault("queue.connections.database.database.rescue_stuck_jobs_after", time.Hour)
	rabbitmq.RabbitMQSetDefault()
}
