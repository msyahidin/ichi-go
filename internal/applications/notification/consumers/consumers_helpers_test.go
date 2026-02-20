package consumers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ichi-go/internal/applications/notification/dto"
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

func marshalEvent(t *testing.T, e dto.NotificationEvent) []byte {
	t.Helper()
	b, err := json.Marshal(e)
	require.NoError(t, err)
	return b
}
