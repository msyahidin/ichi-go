package rabbitmq

import "context"

// MessageProducer publishes messages to queue.
// Renamed from "MessagePublisher".
//
// Usage:
//
//	type UserService struct {
//	    producer MessageProducer
//	}
//
//	func (s *UserService) CreateUser(ctx, req) error {
//	    user, _ := s.repo.Create(ctx, req)
//	    msg := WelcomeMessage{UserID: user.ID}
//	    s.producer.Publish(ctx, "user.welcome", msg, PublishOptions{})
//	    return nil
//	}
//
// Thread-safe.
type MessageProducer interface {
	// Publish sends message to queue.
	//
	// Returns immediately (fire-and-forget).
	// No delivery guarantee.
	Publish(ctx context.Context, routingKey string, message interface{}, opts PublishOptions) error

	// Close releases resources.
	Close() error
}
