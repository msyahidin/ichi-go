package services

import (
	"context"
	"errors"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ichi-go/internal/applications/notification/dto"
	"ichi-go/internal/infra/queue/rabbitmq"
)

// ============================================================================
// Mock
// ============================================================================

type mockProducer struct {
	mock.Mock
}

func (m *mockProducer) Publish(ctx context.Context, routingKey string, message interface{}, opts rabbitmq.PublishOptions) error {
	args := m.Called(ctx, routingKey, message, opts)
	return args.Error(0)
}

func (m *mockProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

// ============================================================================
// Helpers
// ============================================================================

func makeEvent(eventID, eventType string) dto.NotificationEvent {
	return dto.NotificationEvent{
		EventID:   eventID,
		EventType: eventType,
		Channels:  []dto.Channel{dto.ChannelEmail},
	}
}

func newService(blast, user rabbitmq.MessageProducer) *NotificationService {
	return NewNotificationService(blast, user)
}

// ============================================================================
// Blast() tests
// ============================================================================

func TestBlast_HappyPath(t *testing.T) {
	blast := new(mockProducer)
	user := new(mockProducer)
	svc := newService(blast, user)

	event := makeEvent("evt-001", "order.shipped")

	blast.On("Publish", mock.Anything, blastRoutingKey, mock.MatchedBy(func(e dto.NotificationEvent) bool {
		return e.DeliveryMode == dto.DeliveryModeBlast && e.EventID == "evt-001"
	}), mock.MatchedBy(func(opts rabbitmq.PublishOptions) bool {
		h := opts.Headers
		return h["x-delivery-mode"] == string(dto.DeliveryModeBlast) &&
			h["x-event-id"] == "evt-001" &&
			h["x-event-type"] == "order.shipped"
	})).Return(nil)

	err := svc.Blast(context.Background(), event)

	require.NoError(t, err)
	blast.AssertExpectations(t)
	user.AssertNotCalled(t, "Publish")
}

func TestBlast_DeliveryModeForced(t *testing.T) {
	blast := new(mockProducer)
	svc := newService(blast, nil)

	// Caller incorrectly sets user mode — Blast() must override it.
	event := makeEvent("evt-002", "promo.sale")
	event.DeliveryMode = dto.DeliveryModeUser

	blast.On("Publish", mock.Anything, blastRoutingKey, mock.MatchedBy(func(e dto.NotificationEvent) bool {
		return e.DeliveryMode == dto.DeliveryModeBlast
	}), mock.Anything).Return(nil)

	err := svc.Blast(context.Background(), event)
	require.NoError(t, err)
	blast.AssertExpectations(t)
}

func TestBlast_NilProducer(t *testing.T) {
	svc := newService(nil, nil)
	err := svc.Blast(context.Background(), makeEvent("evt-003", "promo.sale"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "blast producer unavailable")
}

func TestBlast_EmptyEventID(t *testing.T) {
	blast := new(mockProducer)
	svc := newService(blast, nil)

	event := makeEvent("", "order.shipped")
	err := svc.Blast(context.Background(), event)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "EventID must not be empty")
	blast.AssertNotCalled(t, "Publish")
}

func TestBlast_PublishError(t *testing.T) {
	blast := new(mockProducer)
	svc := newService(blast, nil)

	publishErr := errors.New("connection reset")
	blast.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(publishErr)

	err := svc.Blast(context.Background(), makeEvent("evt-004", "order.shipped"))

	require.Error(t, err)
	assert.Equal(t, publishErr, err)
}

// ============================================================================
// SendToUser() tests
// ============================================================================

func TestSendToUser_HappyPath(t *testing.T) {
	blast := new(mockProducer)
	user := new(mockProducer)
	svc := newService(blast, user)

	event := makeEvent("evt-010", "otp.login")
	expectedRoutingKey := "user.42"

	user.On("Publish", mock.Anything, expectedRoutingKey, mock.MatchedBy(func(e dto.NotificationEvent) bool {
		return e.DeliveryMode == dto.DeliveryModeUser &&
			e.UserID == "42" &&
			e.EventID == "evt-010"
	}), mock.MatchedBy(func(opts rabbitmq.PublishOptions) bool {
		h := opts.Headers
		return h["x-delivery-mode"] == string(dto.DeliveryModeUser) &&
			h["x-user-id"] == "42" &&
			h["x-event-id"] == "evt-010"
	})).Return(nil)

	err := svc.SendToUser(context.Background(), "42", event)

	require.NoError(t, err)
	user.AssertExpectations(t)
	blast.AssertNotCalled(t, "Publish")
}

func TestSendToUser_RoutingKeyFormat(t *testing.T) {
	user := new(mockProducer)
	svc := newService(nil, user)

	// userID with a hyphen — verify the prefix is concatenated exactly
	user.On("Publish", mock.Anything, "user.user-42", mock.Anything, mock.Anything).Return(nil)

	err := svc.SendToUser(context.Background(), "user-42", makeEvent("evt-011", "otp.login"))
	require.NoError(t, err)
	user.AssertExpectations(t)
}

func TestSendToUser_NilProducer(t *testing.T) {
	svc := newService(nil, nil)
	err := svc.SendToUser(context.Background(), "42", makeEvent("evt-012", "otp.login"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user producer unavailable")
}

func TestSendToUser_EmptyUserID(t *testing.T) {
	user := new(mockProducer)
	svc := newService(nil, user)

	err := svc.SendToUser(context.Background(), "", makeEvent("evt-013", "otp.login"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "userID must not be empty")
	user.AssertNotCalled(t, "Publish")
}

func TestSendToUser_EmptyEventID(t *testing.T) {
	user := new(mockProducer)
	svc := newService(nil, user)

	err := svc.SendToUser(context.Background(), "42", makeEvent("", "otp.login"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "EventID must not be empty")
	user.AssertNotCalled(t, "Publish")
}

func TestSendToUser_PublishError(t *testing.T) {
	user := new(mockProducer)
	svc := newService(nil, user)

	publishErr := errors.New("channel closed")
	user.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(publishErr)

	err := svc.SendToUser(context.Background(), "42", makeEvent("evt-014", "otp.login"))

	require.Error(t, err)
	assert.Equal(t, publishErr, err)
}

// ============================================================================
// Header verification helpers
// ============================================================================

// assertBlastHeaders verifies the AMQP table contains the expected blast headers.
// This is a standalone helper to use with mock.MatchedBy when you need to verify headers directly.
func assertBlastHeaders(t *testing.T, headers amqp.Table, eventID, eventType string) {
	t.Helper()
	assert.Equal(t, string(dto.DeliveryModeBlast), headers["x-delivery-mode"])
	assert.Equal(t, eventID, headers["x-event-id"])
	assert.Equal(t, eventType, headers["x-event-type"])
}
