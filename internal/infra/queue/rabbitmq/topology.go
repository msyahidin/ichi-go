package rabbitmq

import (
	"fmt"
	"ichi-go/pkg/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)

// SetupTopology declares all exchanges, queues, and bindings once after connection is established.
// This must be called before creating any producers or consumers.
func SetupTopology(conn *Connection, config Config) error {
	ch, err := conn.GetConnection().Channel()
	if err != nil {
		return fmt.Errorf("failed to open topology channel: %w", err)
	}
	defer ch.Close()

	logger.Infof("🗺️  Setting up RabbitMQ topology...")

	// Declare all exchanges
	for _, ex := range config.Exchanges {
		args := buildExchangeArgs(ex)

		logger.Infof("📢 Declaring exchange: name=%s, type=%s, durable=%v", ex.Name, ex.Type, ex.Durable)

		if err := ch.ExchangeDeclare(
			ex.Name,
			ex.Type,
			ex.Durable,
			ex.AutoDelete,
			ex.Internal,
			ex.NoWait,
			args,
		); err != nil {
			return fmt.Errorf("failed to declare exchange '%s': %w", ex.Name, err)
		}

		logger.Infof("✅ Exchange declared: %s", ex.Name)
	}

	// Declare queues and bindings for enabled consumers
	for _, consumer := range config.Consumers {
		if !consumer.Enabled {
			continue
		}

		logger.Infof("📦 Declaring queue: name=%s, durable=%v", consumer.Queue.Name, consumer.Queue.Durable)

		q, err := ch.QueueDeclare(
			consumer.Queue.Name,
			consumer.Queue.Durable,
			consumer.Queue.AutoDelete,
			consumer.Queue.Exclusive,
			consumer.Queue.NoWait,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue '%s': %w", consumer.Queue.Name, err)
		}

		logger.Infof("✅ Queue declared: %s (messages: %d, consumers: %d)", q.Name, q.Messages, q.Consumers)

		for _, key := range consumer.RoutingKeys {
			logger.Infof("🔗 Binding queue '%s' -> exchange '%s' (key: '%s')", q.Name, consumer.ExchangeName, key)

			if err := ch.QueueBind(q.Name, key, consumer.ExchangeName, false, nil); err != nil {
				return fmt.Errorf("failed to bind queue '%s' to exchange '%s' with key '%s': %w",
					q.Name, consumer.ExchangeName, key, err)
			}
		}
	}

	logger.Infof("✅ Topology setup complete")
	return nil
}

// buildExchangeArgs constructs the amqp.Table for an exchange declaration,
// preserving plugin-specific args (e.g. x-delayed-type for x-delayed-message).
func buildExchangeArgs(ex ExchangeConfig) amqp.Table {
	args := amqp.Table{}
	if ex.Type == "x-delayed-message" {
		if delayedType, ok := ex.Args["x-delayed-type"]; ok {
			args["x-delayed-type"] = delayedType
		}
	}
	return args
}
