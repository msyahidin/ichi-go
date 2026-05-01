package river

import (
	"context"
	"fmt"

	riverqueue "github.com/riverqueue/river"
	"ichi-go/internal/infra/queue"
)

// GenericJobArgs carries a raw payload for existing ConsumeFunc-based consumers.
// ConsumerName routes the job to the correct handler in BridgeWorker.
type GenericJobArgs struct {
	ConsumerName string `json:"consumer_name"`
	Payload      []byte `json:"payload"`
}

func (GenericJobArgs) Kind() string { return "generic_job" }

// BridgeWorker is a single River worker that dispatches to any number of
// ConsumeFunc handlers by ConsumerName. Only one instance is registered so
// "generic_job" is never added twice (which would panic).
type BridgeWorker struct {
	riverqueue.WorkerDefaults[GenericJobArgs]
	handlers map[string]queue.ConsumeFunc
}

// NewBridgeWorker builds a BridgeWorker from a map of consumerName → handler.
func NewBridgeWorker(handlers map[string]queue.ConsumeFunc) *BridgeWorker {
	return &BridgeWorker{handlers: handlers}
}

func (w *BridgeWorker) Work(ctx context.Context, job *riverqueue.Job[GenericJobArgs]) error {
	handler, ok := w.handlers[job.Args.ConsumerName]
	if !ok {
		return fmt.Errorf("bridge worker: no handler registered for consumer %q", job.Args.ConsumerName)
	}
	if handler == nil {
		return fmt.Errorf("bridge worker: handler for consumer %q is nil", job.Args.ConsumerName)
	}
	return handler(ctx, job.Args.Payload)
}
