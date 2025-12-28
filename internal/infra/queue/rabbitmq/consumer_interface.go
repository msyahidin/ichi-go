package rabbitmq

import "context"

// ConsumeFunc processes messages from queue.
//
// Error Handling:
// - Return ERROR for transient failures (will retry):
//   - Database timeout, network errors, service unavailable
//
// - Return NIL for permanent failures (will skip):
//   - Invalid JSON, unknown event, validation failure
//
// Example:
//
//	func (c *Consumer) Consume(ctx context.Context, body []byte) error {
//	    var event Event
//	    if err := json.Unmarshal(body, &event); err != nil {
//	        return nil // Don't retry bad JSON
//	    }
//
//	    if err := c.process(ctx, event); err != nil {
//	        if isTransient(err) {
//	            return err // Retry
//	        }
//	        return nil // Skip
//	    }
//	    return nil
//	}
//
// Best Practices:
// - Keep fast (< 30 sec)
// - Make idempotent
// - Use structured logging
type ConsumeFunc func(ctx context.Context, body []byte) error

// MessageConsumer consumes messages from queue.
//
// Manages worker pool for concurrent processing.
type MessageConsumer interface {
	// Consume starts consuming messages.
	//
	// Blocks until context cancelled.
	//
	// Flow:
	// 1. Worker receives message
	// 2. Calls handler(ctx, body)
	// 3. On nil: ack (remove)
	// 4. On error: nack and requeue
	Consume(ctx context.Context, handler ConsumeFunc) error

	// Close releases resources.
	Close() error
}
