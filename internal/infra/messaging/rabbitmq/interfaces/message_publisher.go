package interfaces

import "context"

// MessagePublisher defines how to publish messages
type MessagePublisher interface {
	Publish(ctx context.Context, routingKey string, message interface{}) error
	Close() error
}
