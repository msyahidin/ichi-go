package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"ichi-go/internal/infra/messaging/rabbitmq/interfaces"
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

func NewPublisher(connection *Connection, config RabbitMQConfig) (interfaces.MessagePublisher, error) {
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

	// Declare all exchanges from config
	for _, exchange := range p.config.Exchanges {
		err = ch.ExchangeDeclare(
			exchange.Name,
			exchange.Type,
			exchange.Durable,
			exchange.AutoDelete,
			exchange.Internal,
			exchange.NoWait,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", exchange.Name, err)
		}
	}

	p.channel = ch
	return nil
}

func (p *Publisher) Publish(ctx context.Context, routingKey string, message interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = p.channel.PublishWithContext(ctx, p.exchangeName, routingKey, false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
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
