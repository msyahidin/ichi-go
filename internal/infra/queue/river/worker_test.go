package river_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	riverqueue "github.com/riverqueue/river"
	"ichi-go/internal/infra/queue"
	riverworker "ichi-go/internal/infra/queue/river"
)

func TestBridgeWorker_Work_CallsHandler(t *testing.T) {
	called := false
	var receivedPayload []byte

	worker := riverworker.NewBridgeWorker(map[string]queue.ConsumeFunc{
		"payment_handler": func(ctx context.Context, payload []byte) error {
			called = true
			receivedPayload = payload
			return nil
		},
	})

	job := &riverqueue.Job[riverworker.GenericJobArgs]{
		Args: riverworker.GenericJobArgs{
			ConsumerName: "payment_handler",
			Payload:      []byte(`{"amount":100}`),
		},
	}

	err := worker.Work(context.Background(), job)
	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, []byte(`{"amount":100}`), receivedPayload)
}

func TestBridgeWorker_Work_PropagatesError(t *testing.T) {
	worker := riverworker.NewBridgeWorker(map[string]queue.ConsumeFunc{
		"payment_handler": func(ctx context.Context, payload []byte) error {
			return errors.New("transient db error")
		},
	})

	job := &riverqueue.Job[riverworker.GenericJobArgs]{
		Args: riverworker.GenericJobArgs{ConsumerName: "payment_handler", Payload: []byte(`{}`)},
	}

	err := worker.Work(context.Background(), job)
	assert.EqualError(t, err, "transient db error")
}

func TestBridgeWorker_Work_UnknownConsumer(t *testing.T) {
	worker := riverworker.NewBridgeWorker(map[string]queue.ConsumeFunc{})

	job := &riverqueue.Job[riverworker.GenericJobArgs]{
		Args: riverworker.GenericJobArgs{ConsumerName: "missing_handler", Payload: []byte(`{}`)},
	}

	err := worker.Work(context.Background(), job)
	assert.ErrorContains(t, err, "no handler registered for consumer")
}

func TestRegisterBridgeWorkers_SingleWorker(t *testing.T) {
	workers := riverqueue.NewWorkers()
	registrations := []queue.ConsumerRegistration{
		{Name: "payment_handler", ConsumeFunc: func(ctx context.Context, payload []byte) error { return nil }},
		{Name: "email_handler", ConsumeFunc: func(ctx context.Context, payload []byte) error { return nil }},
	}
	// Multiple registrations must not panic (single Kind "generic_job" registered once)
	assert.NotPanics(t, func() {
		err := riverworker.RegisterBridgeWorkers(workers, registrations)
		require.NoError(t, err)
	})
}

func TestBridgeWorker_Work_NilHandler(t *testing.T) {
	// Build a BridgeWorker with a map entry whose value is explicitly nil.
	handlers := map[string]queue.ConsumeFunc{
		"nil_handler": nil,
	}
	worker := riverworker.NewBridgeWorker(handlers)

	job := &riverqueue.Job[riverworker.GenericJobArgs]{
		Args: riverworker.GenericJobArgs{
			ConsumerName: "nil_handler",
			Payload:      []byte(`{}`),
		},
	}

	err := worker.Work(context.Background(), job)
	assert.ErrorContains(t, err, "handler for consumer",
		"expected a descriptive error, not a panic, when the handler is nil")
}
