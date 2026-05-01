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

func TestGenericJobWorker_Work_CallsHandler(t *testing.T) {
	called := false
	var receivedPayload []byte

	worker := riverworker.NewGenericJobWorker(func(ctx context.Context, payload []byte) error {
		called = true
		receivedPayload = payload
		return nil
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

func TestGenericJobWorker_Work_PropagatesError(t *testing.T) {
	worker := riverworker.NewGenericJobWorker(func(ctx context.Context, payload []byte) error {
		return errors.New("transient db error")
	})

	job := &riverqueue.Job[riverworker.GenericJobArgs]{
		Args: riverworker.GenericJobArgs{ConsumerName: "payment_handler", Payload: []byte(`{}`)},
	}

	err := worker.Work(context.Background(), job)
	assert.EqualError(t, err, "transient db error")
}

func TestRegisterBridgeWorkers_NoError(t *testing.T) {
	workers := riverqueue.NewWorkers()
	registrations := []queue.ConsumerRegistration{
		{
			Name:        "payment_handler",
			Description: "test",
			ConsumeFunc: func(ctx context.Context, payload []byte) error { return nil },
		},
	}
	assert.NotPanics(t, func() {
		riverworker.RegisterBridgeWorkers(workers, registrations)
	})
}
