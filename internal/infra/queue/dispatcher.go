package queue

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	riverqueue "github.com/riverqueue/river"
	"ichi-go/internal/infra/queue/rabbitmq"
)

// NewDispatcher builds the active Dispatcher based on the configured driver name.
// Pass nil for unused arguments (e.g. nil riverClient when driver is "rabbitmq").
func NewDispatcher(driver string, producer rabbitmq.MessageProducer, riverClient *riverqueue.Client[*sql.Tx]) (Dispatcher, error) {
	switch driver {
	case "amqp":
		if producer == nil {
			return nil, fmt.Errorf("amqp dispatcher: producer is nil (queue connection unavailable)")
		}
		return &rabbitMQDispatcher{producer: producer}, nil

	case "database":
		if riverClient == nil {
			return nil, fmt.Errorf("database dispatcher: client is nil (postgres connection unavailable)")
		}
		return &riverDispatcher{client: riverClient}, nil

	default:
		return nil, fmt.Errorf("unknown queue driver: %q (valid: amqp, database)", driver)
	}
}

// rabbitMQDispatcher implements Dispatcher using the existing RabbitMQ producer.
// Serialises the job to JSON and publishes with routing_key = job.Kind().
type rabbitMQDispatcher struct {
	producer rabbitmq.MessageProducer
}

func (d *rabbitMQDispatcher) Dispatch(ctx context.Context, job JobArgs, opts ...DispatchOption) error {
	o := ApplyOptions(opts...)

	// RabbitMQ does not support routing by queue name, attempt limiting, or priority
	// via the generic Dispatch interface. Fail fast so callers see the mismatch immediately.
	if o.Queue != "" || o.MaxAttempts != 0 || o.Priority != 0 {
		return fmt.Errorf(
			"rabbitmq dispatcher: unsupported options for job %q (Queue=%q, MaxAttempts=%d, Priority=%d); "+
				"AMQP dispatch only supports Delay",
			job.Kind(), o.Queue, o.MaxAttempts, o.Priority,
		)
	}

	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("rabbitmq dispatcher: failed to marshal job %q: %w", job.Kind(), err)
	}

	return d.producer.Publish(ctx, job.Kind(), payload, rabbitmq.PublishOptions{
		Delay: o.Delay,
	})
}

// riverDispatcher implements Dispatcher using riverqueue with riverdatabasesql (poll-only).
// Shares bun's *sql.DB — no extra connection pool needed.
type riverDispatcher struct {
	client *riverqueue.Client[*sql.Tx]
}

func (d *riverDispatcher) Dispatch(ctx context.Context, job JobArgs, opts ...DispatchOption) error {
	o := ApplyOptions(opts...)

	insertOpts := &riverqueue.InsertOpts{
		Queue:       o.Queue,
		MaxAttempts: o.MaxAttempts,
		Priority:    o.Priority,
	}
	if o.Delay > 0 {
		insertOpts.ScheduledAt = time.Now().Add(o.Delay)
	}

	// river.JobArgs requires Kind() string — queue.JobArgs has the same method,
	// so any job that satisfies queue.JobArgs also satisfies river.JobArgs.
	riverArgs, ok := job.(riverqueue.JobArgs)
	if !ok {
		return fmt.Errorf("river dispatcher: job %T does not implement river.JobArgs (needs Kind() string)", job)
	}

	_, err := d.client.Insert(ctx, riverArgs, insertOpts)
	return err
}
