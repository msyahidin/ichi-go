package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
	"time"
)

type Publisher struct {
	connection   *Connection
	config       RabbitMQConfig
	exchangeName string
	channel      *amqp.Channel
	mu           sync.Mutex
}
type PublishOptions struct {
	Headers amqp.Table
	Delay   time.Duration
}

func NewPublisher(connection *Connection, config RabbitMQConfig) (MessagePublisher, error) {
	p := &Publisher{
		connection:   connection,
		config:       config,
		exchangeName: config.Publisher.ExchangeName,
	}

	if err := p.setup(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Publisher) setup() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch, err := p.connection.GetConnection().Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	for _, exchange := range p.config.Exchanges {
		args := amqp.Table{}

		if exchange.Type == "x-delayed-message" {
			args["x-delayed-type"] = exchange.Args["x-delayed-type"]
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
			return fmt.Errorf("failed to declare exchange %s: %w", exchange.Name, err)
		}
	}

	p.channel = ch
	return nil
}

func (p *Publisher) Publish(ctx context.Context, routingKey string, message interface{}, opts PublishOptions) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if opts.Headers == nil {
		opts.Headers = amqp.Table{}
	}
	if opts.Delay > 0 {
		opts.Headers["x-delay"] = int32(opts.Delay.Milliseconds())
	}

	err = p.channel.PublishWithContext(ctx, p.exchangeName, routingKey, false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Headers:      opts.Headers,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish: %w", err)
	}

	return nil
}

func (p *Publisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.channel != nil && !p.channel.IsClosed() {
		return p.channel.Close()
	}
	return nil
}
