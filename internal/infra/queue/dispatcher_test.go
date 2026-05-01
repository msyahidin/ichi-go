package queue_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"ichi-go/internal/infra/queue"
	"ichi-go/internal/infra/queue/rabbitmq"
	mocks "ichi-go/internal/infra/queue/rabbitmq/mocks"
)

type emailJob struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
}

func (j emailJob) Kind() string { return "email.send" }

func TestNewDispatcher_UnknownDriver(t *testing.T) {
	_, err := queue.NewDispatcher("unknown_driver", nil, nil)
	assert.ErrorContains(t, err, "unknown queue driver")
}

func TestNewDispatcher_AMQP_NilProducer(t *testing.T) {
	_, err := queue.NewDispatcher("amqp", nil, nil)
	assert.ErrorContains(t, err, "producer is nil")
}

func TestNewDispatcher_Database_NilClient(t *testing.T) {
	_, err := queue.NewDispatcher("database", nil, nil)
	assert.ErrorContains(t, err, "client is nil")
}

func TestAMQPDispatcher_Dispatch(t *testing.T) {
	producer := mocks.NewMockMessageProducer(t)

	expectedPayload, _ := json.Marshal(emailJob{UserID: 1, Email: "a@b.com"})
	producer.On("Publish",
		mock.Anything,
		"email.send",
		expectedPayload,
		mock.AnythingOfType("rabbitmq.PublishOptions"),
	).Return(nil)

	d, err := queue.NewDispatcher("amqp", producer, nil)
	assert.NoError(t, err)

	err = d.Dispatch(context.Background(), emailJob{UserID: 1, Email: "a@b.com"})
	assert.NoError(t, err)
	producer.AssertExpectations(t)
}

func TestAMQPDispatcher_Dispatch_WithDelay(t *testing.T) {
	producer := mocks.NewMockMessageProducer(t)

	producer.On("Publish",
		mock.Anything,
		"email.send",
		mock.Anything,
		mock.MatchedBy(func(opts rabbitmq.PublishOptions) bool {
			return opts.Delay == 5*time.Minute
		}),
	).Return(nil)

	d, err := queue.NewDispatcher("amqp", producer, nil)
	assert.NoError(t, err)

	err = d.Dispatch(context.Background(), emailJob{UserID: 1, Email: "a@b.com"},
		queue.Delay(5*time.Minute))
	assert.NoError(t, err)
	producer.AssertExpectations(t)
}
