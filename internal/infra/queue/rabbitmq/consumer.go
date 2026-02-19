package rabbitmq

import (
	"context"
	"fmt"
	"ichi-go/pkg/logger"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	connection     *Connection
	consumerConfig ConsumerConfig
	exchangeConfig ExchangeConfig
	channel        *amqp.Channel
	mu             sync.Mutex
}

func NewConsumer(
	connection *Connection,
	consumerConfig ConsumerConfig,
	exchangeConfig ExchangeConfig,
) (MessageConsumer, error) {
	c := &Consumer{
		connection:     connection,
		consumerConfig: consumerConfig,
		exchangeConfig: exchangeConfig,
	}

	if err := c.setup(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Consumer) setup() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	logger.Infof("🔧 Setting up consumer '%s'...", c.consumerConfig.Name)

	ch, err := c.connection.GetConnection().Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Set QoS (prefetch) for this consumer
	logger.Debugf("⚙️  Setting QoS: prefetch_count=%d", c.consumerConfig.PrefetchCount)
	err = ch.Qos(c.consumerConfig.PrefetchCount, 0, false)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	c.channel = ch

	logger.Debugf("✅ Consumer '%s' setup complete:", c.consumerConfig.Name)
	logger.Debugf("   Queue: '%s'", c.consumerConfig.Queue.Name)
	logger.Debugf("   Exchange: '%s' (type: %s)", c.exchangeConfig.Name, c.exchangeConfig.Type)
	logger.Debugf("   Routing Keys: %v", c.consumerConfig.RoutingKeys)
	logger.Debugf("   Prefetch: %d", c.consumerConfig.PrefetchCount)
	logger.Debugf("   Workers: %d", c.consumerConfig.WorkerPoolSize)
	logger.Debugf("   Auto-Ack: %v", c.consumerConfig.AutoAck)

	return nil
}

func (c *Consumer) Consume(ctx context.Context, handler ConsumeFunc) error {
	c.mu.Lock()

	logger.Infof("🎧 Starting to consume from queue '%s'...", c.consumerConfig.Queue.Name)

	deliveries, err := c.channel.Consume(
		c.consumerConfig.Queue.Name,
		c.consumerConfig.ConsumerTag,
		c.consumerConfig.AutoAck,
		c.consumerConfig.Exclusive,
		false,
		false,
		nil,
	)
	c.mu.Unlock()

	if err != nil {
		return fmt.Errorf("failed to start consuming from queue '%s': %w", c.consumerConfig.Queue.Name, err)
	}

	logger.Debugf("✅ Consumer '%s' listening on queue '%s'",
		c.consumerConfig.Name, c.consumerConfig.Queue.Name)
	logger.Debugf("   Waiting for messages with routing keys: %v", c.consumerConfig.RoutingKeys)

	// Create worker pool
	var wg sync.WaitGroup
	for i := 0; i < c.consumerConfig.WorkerPoolSize; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			logger.Debugf("👷 Worker #%d started for consumer '%s'", workerID, c.consumerConfig.Name)

			messagesProcessed := 0

			for {
				select {
				case <-ctx.Done():
					logger.Debugf("👷 Worker #%d stopping (processed %d messages)", workerID, messagesProcessed)
					return

				case delivery, ok := <-deliveries:
					if !ok {
						logger.Debugf("👷 Worker #%d: delivery channel closed (processed %d messages)",
							workerID, messagesProcessed)
						return
					}

					messagesProcessed++

					logger.Debugf("📨 Worker #%d received message #%d:", workerID, messagesProcessed)
					logger.Debugf("   Routing Key: '%s'", delivery.RoutingKey)
					logger.Debugf("   Content Type: %s", delivery.ContentType)
					logger.Debugf("   Body Length: %d bytes", len(delivery.Body))
					logger.Debugf("   Body: %s", string(delivery.Body))

					if delivery.Timestamp.IsZero() {
						logger.Debugf("   Published: (no timestamp)")
					} else {
						logger.Debugf("   Published: %s", delivery.Timestamp.Format("2006-01-02 15:04:05"))
					}

					// Process message
					logger.Infof("⚙️  Processing message...")
					if err := handler(ctx, delivery.Body); err != nil {
						logger.Errorf("❌ Worker #%d: handler error: %v", workerID, err)

						if !c.consumerConfig.AutoAck {
							logger.Warnf("📤 Nacking message (will requeue)")
							if nackErr := delivery.Nack(false, true); nackErr != nil {
								logger.Errorf("❌ Failed to nack: %v", nackErr)
							}
						}
					} else {
						logger.Debugf("✅ Worker #%d: message processed successfully", workerID)

						if !c.consumerConfig.AutoAck {
							logger.Debugf("📤 Acking message")
							if ackErr := delivery.Ack(false); ackErr != nil {
								logger.Errorf("❌ Failed to ack: %v", ackErr)
							} else {
								logger.Debugf("✅ Message acknowledged")
							}
						}
					}
				}
			}
		}(i)
	}

	logger.Debugf("✅ All %d workers running for consumer '%s'",
		c.consumerConfig.WorkerPoolSize, c.consumerConfig.Name)

	wg.Wait()

	logger.Debugf("👋 All workers stopped for consumer '%s'", c.consumerConfig.Name)

	return nil
}

func (c *Consumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.channel != nil && !c.channel.IsClosed() {
		return c.channel.Close()
	}
	return nil
}
