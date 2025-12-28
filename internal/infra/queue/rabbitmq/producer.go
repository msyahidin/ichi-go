package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"ichi-go/pkg/logger"
	"sync"
	"time"
)

// Producer publishes messages to RabbitMQ.
type Producer struct {
	connection   *Connection
	config       Config
	exchangeName string
	channel      *amqp.Channel
	mu           sync.Mutex
}

// PublishOptions configures message publishing.
type PublishOptions struct {
	Headers amqp.Table    // Custom metadata
	Delay   time.Duration // Delivery delay
}

// NewProducer creates message producer.
func NewProducer(connection *Connection, config Config) (MessageProducer, error) {
	p := &Producer{
		connection:   connection,
		config:       config,
		exchangeName: config.Publisher.ExchangeName,
	}

	if err := p.setup(); err != nil {
		return nil, err
	}

	return p, nil
}

// setup initializes channel and exchanges.
func (p *Producer) setup() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	logger.Infof("🔧 Setting up producer...")

	ch, err := p.connection.GetConnection().Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	logger.Infof("✅ Channel opened")

	// Declare exchanges
	for _, exchange := range p.config.Exchanges {
		logger.Infof("📢 Declaring exchange: name=%s, type=%s, durable=%v",
			exchange.Name, exchange.Type, exchange.Durable)

		args := amqp.Table{}

		if exchange.Type == "x-delayed-message" {
			if delayedType, ok := exchange.Args["x-delayed-type"]; ok {
				args["x-delayed-type"] = delayedType
			}
		}

		err = ch.ExchangeDeclare(
			exchange.Name,
			exchange.Type,
			exchange.Durable,
			exchange.AutoDelete,
			exchange.Internal,
			exchange.NoWait,
			args,
		)
		if err != nil {
			ch.Close()
			logger.Errorf("❌ Failed to declare exchange %s: %v", exchange.Name, err)
			return fmt.Errorf("failed to declare exchange %s: %w", exchange.Name, err)
		}

		logger.Infof("✅ Exchange declared: %s", exchange.Name)
	}

	p.channel = ch
	logger.Infof("✅ Producer setup complete")
	return nil
}

// Publish sends message to queue.
//
// Flow:
// 1. Serialize to JSON
// 2. Publish to exchange
// 3. Exchange routes to queue(s)
// 4. Consumer processes
//
// Examples:
//
//	// Simple
//	producer.Publish(ctx, "user.welcome", msg, PublishOptions{})
//
//	// Delayed
//	opts := PublishOptions{Delay: 5 * time.Minute}
//	producer.Publish(ctx, "reminder", msg, opts)
//	opts := PublishOptions{Headers: amqp.Table{"x-correlation-id": id}}
//	producer.Publish(ctx, "order", msg, opts)
func (p *Producer) Publish(ctx context.Context, routingKey string, message interface{}, opts PublishOptions) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Add logging BEFORE serialization
	logger.Infof("📤 Publishing message to exchange=%s, routingKey=%s", p.exchangeName, routingKey)

	// Serialize
	body, err := json.Marshal(message)
	if err != nil {
		logger.Errorf("❌ Failed to marshal message: %v", err)
		return fmt.Errorf("failed to marshal: %w", err)
	}

	logger.Debugf("📦 Message body: %s", string(body))

	// Add timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Initialize headers
	if opts.Headers == nil {
		opts.Headers = amqp.Table{}
	}

	// Add delay
	if opts.Delay > 0 {
		opts.Headers["x-delay"] = int32(opts.Delay.Milliseconds())
	}

	// Publish
	logger.Debugf("🔄 Publishing to channel (exchange=%s, key=%s)", p.exchangeName, routingKey)
	err = p.channel.PublishWithContext(
		ctx,
		p.exchangeName,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Headers:      opts.Headers,
		},
	)

	if err != nil {
		logger.Errorf("❌ Failed to publish: %v", err)
		return fmt.Errorf("failed to publish: %w", err)
	}

	logger.Infof("✅ Message published successfully to %s/%s", p.exchangeName, routingKey)
	return nil
}

// Close releases resources.
func (p *Producer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.channel != nil && !p.channel.IsClosed() {
		return p.channel.Close()
	}
	return nil
}
