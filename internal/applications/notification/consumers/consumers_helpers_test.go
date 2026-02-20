package consumers

import (
	"context"
	"encoding/json"

	"github.com/stretchr/testify/mock"

	"ichi-go/internal/applications/notification/dto"
	"ichi-go/internal/infra/queue/rabbitmq"
)

// ============================================================================
// mockChannel — implements channels.NotificationChannel
// ============================================================================

type mockChannel struct {
	mock.Mock
	name dto.Channel
}

func newMockChannel(name dto.Channel) *mockChannel {
	return &mockChannel{name: name}
}

func (m *mockChannel) Name() dto.Channel { return m.name }

func (m *mockChannel) Send(ctx context.Context, event dto.NotificationEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// ============================================================================
// mockProducer — implements rabbitmq.MessageProducer
// ============================================================================

type mockConsumerProducer struct {
	mock.Mock
}

func (m *mockConsumerProducer) Publish(ctx context.Context, routingKey string, message interface{}, opts rabbitmq.PublishOptions) error {
	args := m.Called(ctx, routingKey, message, opts)
	return args.Error(0)
}

func (m *mockConsumerProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

// ============================================================================
// Event builders
// ============================================================================

func makeTestUserEvent(userID, eventID string, chs ...dto.Channel) dto.NotificationEvent {
	return dto.NotificationEvent{
		EventID:      eventID,
		EventType:    "otp.login",
		DeliveryMode: dto.DeliveryModeUser,
		UserID:       userID,
		Channels:     chs,
	}
}

func makeTestBlastEvent(eventID string, chs ...dto.Channel) dto.NotificationEvent {
	return dto.NotificationEvent{
		EventID:      eventID,
		EventType:    "promo.sale",
		DeliveryMode: dto.DeliveryModeBlast,
		Channels:     chs,
	}
}

func marshalEvent(e dto.NotificationEvent) []byte {
	b, _ := json.Marshal(e)
	return b
}
