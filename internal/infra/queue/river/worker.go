package river

import (
	"context"

	riverqueue "github.com/riverqueue/river"
	"ichi-go/internal/infra/queue"
)

// GenericJobArgs carries a raw payload for existing ConsumeFunc-based consumers.
// ConsumerName discriminates which handler processes the job.
type GenericJobArgs struct {
	ConsumerName string `json:"consumer_name"`
	Payload      []byte `json:"payload"`
}

func (GenericJobArgs) Kind() string { return "generic_job" }

// GenericJobWorker bridges a single river.Worker to any queue.ConsumeFunc handler.
type GenericJobWorker struct {
	riverqueue.WorkerDefaults[GenericJobArgs]
	handler queue.ConsumeFunc
}

// NewGenericJobWorker creates a GenericJobWorker wrapping the given ConsumeFunc.
func NewGenericJobWorker(handler queue.ConsumeFunc) *GenericJobWorker {
	return &GenericJobWorker{handler: handler}
}

func (w *GenericJobWorker) Work(ctx context.Context, job *riverqueue.Job[GenericJobArgs]) error {
	return w.handler(ctx, job.Args.Payload)
}
