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

	ch, err := c.connection.GetConnection().Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Set QoS (prefetch) for this consumer
	err = ch.Qos(c.consumerConfig.PrefetchCount, 0, false)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Declare exchange
	err = ch.ExchangeDeclare(
		c.exchangeConfig.Name,
		c.exchangeConfig.Type,
		c.exchangeConfig.Durable,
		c.exchangeConfig.AutoDelete,
		c.exchangeConfig.Internal,
		c.exchangeConfig.NoWait,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	_, err = ch.QueueDeclare(
		c.consumerConfig.Queue.Name,
		c.consumerConfig.Queue.Durable,
		c.consumerConfig.Queue.AutoDelete,
		c.consumerConfig.Queue.Exclusive,
		c.consumerConfig.Queue.NoWait,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange with routing keys
	for _, routingKey := range c.consumerConfig.RoutingKeys {
		err = ch.QueueBind(
			c.consumerConfig.Queue.Name,
			routingKey,
			c.exchangeConfig.Name,
			false, nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue: %w", err)
		}
		logger.Infof("bound queue '%s' to exchange '%s' with routing key '%s'",
			c.consumerConfig.Queue.Name, c.exchangeConfig.Name, routingKey)
	}

	c.channel = ch
	logger.Infof("Consumer '%s' ready (prefetch: %d, workers: %d)",
		c.consumerConfig.Name,
		c.consumerConfig.PrefetchCount,
		c.consumerConfig.WorkerPoolSize)

	return nil
}

func (c *Consumer) Consume(ctx context.Context, handler MessageHandler) error {
	c.mu.Lock()
	deliveries, err := c.channel.Consume(
		c.consumerConfig.Queue.Name,
		c.consumerConfig.ConsumerTag,
		c.consumerConfig.AutoAck,
		c.consumerConfig.Exclusive,
		false, false, nil,
	)
	c.mu.Unlock()

	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	logger.Infof("Consumer '%s' listening on queue '%s'",
		c.consumerConfig.Name, c.consumerConfig.Queue.Name)

	// Create worker pool
	var wg sync.WaitGroup
	for i := 0; i < c.consumerConfig.WorkerPoolSize; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			logger.Infof("Worker #%d started for '%s'", workerID, c.consumerConfig.Name)

			for {
				select {
				case <-ctx.Done():
					logger.Infof("👷 Worker #%d stopping", workerID)
					return

				case delivery, ok := <-deliveries:
					if !ok {
						logger.Infof("Worker #%d channel closed", workerID)
						return
					}

					// Process message
					if err := handler(ctx, delivery.Body); err != nil {
						if !c.consumerConfig.AutoAck {
							delivery.Nack(false, true) // Requeue on error
						}
					} else {
						if !c.consumerConfig.AutoAck {
							delivery.Ack(false)
						}
					}
				}
			}
		}(i)
	}

	wg.Wait()
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
