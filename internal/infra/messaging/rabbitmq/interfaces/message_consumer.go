package interfaces

import "context"

type MessageHandler func(ctx context.Context, body []byte) error

type MessageConsumer interface {
	Consume(ctx context.Context, handler MessageHandler) error
	Close() error
}
