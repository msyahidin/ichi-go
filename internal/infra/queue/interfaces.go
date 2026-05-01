package queue

import "context"

// JobArgs is implemented by every job struct. Kind() returns the unique job type
// identifier used for routing (e.g. "notification.send_email").
type JobArgs interface {
	Kind() string
}

// ConsumeFunc processes a raw message payload. Return a non-nil error to nack/retry;
// return nil to ack (including on permanent failures like bad JSON).
type ConsumeFunc func(ctx context.Context, payload []byte) error

// Dispatcher publishes jobs to the active queue backend (RabbitMQ or River).
type Dispatcher interface {
	Dispatch(ctx context.Context, job JobArgs, opts ...DispatchOption) error
}
