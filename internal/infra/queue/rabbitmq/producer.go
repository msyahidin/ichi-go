package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"ichi-go/pkg/logger"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Producer publishes messages to RabbitMQ.
type Producer struct {
	connection   *Connection
	config       Config
	exchangeName string
	channel      *amqp.Channel
	confirms     chan amqp.Confirmation
	mu           sync.Mutex
}

// PublishOptions configures message publishing.
type PublishOptions struct {
	Headers   amqp.Table    // Custom metadata
	Delay     time.Duration // Delivery delay
	Mandatory bool          // Return error if no queue is bound
}

// NewProducer creates message producer.
func NewProducer(connection *Connection, config Config) (MessageProducer, error) {
	p := &Producer{
		connection:   connection,
		config:       config,
		exchangeName: config.Publisher.ExchangeName,
		confirms:     make(chan amqp.Confirmation, 1),
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

	// Enable publisher confirms for reliability
	if err := ch.Confirm(false); err != nil {
		logger.Warnf("⚠️  Failed to enable publisher confirms: %v (continuing anyway)", err)
	} else {
		logger.Infof("✅ Publisher confirms enabled")
	}

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

		logger.Infof("✅ Exchange declared successfully")
		logger.Infof("   Name: %s", exchange.Name)
		logger.Infof("   Type: %s", exchange.Type)
		logger.Infof("   Durable: %v", exchange.Durable)
	}

	p.channel = ch

	logger.Infof("✅ Producer configured to publish to exchange: '%s'", p.exchangeName)
	logger.Infof("✅ Producer setup complete")

	return nil
}

// Publish sends message to queue with enhanced diagnostics.
func (p *Producer) Publish(ctx context.Context, routingKey string, message interface{}, opts PublishOptions) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Detailed logging for diagnostics
	logger.Infof("📤 Publishing message:")
	logger.Infof("   Exchange: '%s'", p.exchangeName)
	logger.Infof("   Routing Key: '%s'", routingKey)
	logger.Infof("   Message Type: %T", message)

	// Serialize
	body, err := json.Marshal(message)
	if err != nil {
		logger.Errorf("❌ Failed to marshal message: %v", err)
		return fmt.Errorf("failed to marshal: %w", err)
	}

	logger.Infof("   Body Length: %d bytes", len(body))
	logger.Debugf("   Body Content: %s", string(body))

	// Add timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Initialize headers
	if opts.Headers == nil {
		opts.Headers = amqp.Table{}
	}

	// Add delay
	if opts.Delay > 0 {
		opts.Headers["x-delay"] = int32(opts.Delay.Milliseconds())
		logger.Infof("   Delay: %v", opts.Delay)
	}

	// Add timestamp to headers for tracking
	opts.Headers["published_at"] = time.Now().Format(time.RFC3339)

	// Publish with mandatory flag to detect routing failures
	mandatory := opts.Mandatory
	logger.Infof("   Mandatory: %v (will error if no queue bound)", mandatory)
	logger.Infof("🔄 Calling PublishWithContext...")

	err = p.channel.PublishWithContext(
		ctx,
		p.exchangeName,
		routingKey,
		mandatory, // Set to true to get errors if message can't be routed
		false,     // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Headers:      opts.Headers,
		},
	)

	if err != nil {
		logger.Errorf("❌ PublishWithContext failed: %v", err)
		return fmt.Errorf("failed to publish: %w", err)
	}

	logger.Infof("✅ Message published successfully")
	logger.Infof("   Exchange: '%s'", p.exchangeName)
	logger.Infof("   Routing Key: '%s'", routingKey)
	logger.Infof("   Next: RabbitMQ will route to queues bound with this routing key")

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
