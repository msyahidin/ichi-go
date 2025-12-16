package rabbitmq

import (
	"context"
)

// MessagePublisher defines how to publish messages
type MessagePublisher interface {
	Publish(ctx context.Context, routingKey string, message interface{}, opts PublishOptions) error
	Close() error
}
