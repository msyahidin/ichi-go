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

		// Override NoWait to false — topology setup must be synchronous so that
		// declaration errors surface immediately rather than silently closing the
		// channel on the next operation.
		if ex.NoWait {
			logger.Warnf("⚠️  Exchange '%s' has no_wait=true in config; overriding to false for topology setup", ex.Name)
		}

		logger.Infof("📢 Declaring exchange: name=%s, type=%s, durable=%v", ex.Name, ex.Type, ex.Durable)

		if err := ch.ExchangeDeclare(
			ex.Name,
			ex.Type,
			ex.Durable,
			ex.AutoDelete,
			ex.Internal,
			false, // NoWait always false — see comment above
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

		if len(consumer.RoutingKeys) == 0 {
			// Fanout exchanges route all messages regardless of routing key —
			// bind with an empty key so the queue is explicitly connected.
			// For all other exchange types, an unbound queue receives nothing.
			if getExchangeType(config, consumer.ExchangeName) == "fanout" {
				logger.Infof("🔗 Fanout binding: queue '%s' -> exchange '%s' (no routing key needed)", q.Name, consumer.ExchangeName)
				if err := ch.QueueBind(q.Name, "", consumer.ExchangeName, false, nil); err != nil {
					return fmt.Errorf("failed to bind queue '%s' to fanout exchange '%s': %w", q.Name, consumer.ExchangeName, err)
				}
			} else {
				logger.Warnf("⚠️  Consumer '%s': queue '%s' has no routing keys — it will receive no messages from exchange '%s'",
					consumer.Name, q.Name, consumer.ExchangeName)
			}
			continue
		}

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

// buildExchangeArgs returns a defensive copy of the exchange's configured args.
// Ranging over a nil map is a no-op in Go, so this is always safe.
func buildExchangeArgs(ex ExchangeConfig) amqp.Table {
	result := amqp.Table{}
	for k, v := range ex.Args {
		result[k] = v
	}
	return result
}

// getExchangeType returns the type of the named exchange from config, or "" if not found.
func getExchangeType(config Config, name string) string {
	for _, ex := range config.Exchanges {
		if ex.Name == name {
			return ex.Type
		}
	}
	return ""
}
