package consumers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ichi-go/internal/applications/notification/channels"
	"ichi-go/internal/applications/notification/dto"
)

// ============================================================================
// Helpers
// ============================================================================

// newCtx returns a background context for tests.
func newCtx() context.Context { return context.Background() }

// asChannels converts []*mockChannel to []channels.NotificationChannel.
func asChannels(chs ...*mockChannel) []channels.NotificationChannel {
	out := make([]channels.NotificationChannel, len(chs))
	for i, c := range chs {
		out[i] = c
	}
	return out
}

// ============================================================================
// maskUserID — pure function, table-driven
// ============================================================================

func TestMaskUserID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1234567", "***567"},
		{"1234", "***234"},
		{"42", "***"},        // ≤3 chars: fully redacted
		{"abc", "***"},       // exactly 3 chars: fully redacted
		{"", "***"},          // empty: fully redacted
		{"usr_123456", "***456"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, maskUserID(tt.input))
		})
	}
}

// ============================================================================
// UserNotificationConsumer.Consume() — all tests use nil Redis (guard bypassed)
// ============================================================================

func TestUserConsume_InvalidJSON(t *testing.T) {
	ch := newMockChannel(dto.ChannelEmail)
	c := &UserNotificationConsumer{channels: asChannels(ch), redis: nil}

	err := c.Consume(newCtx(), []byte("not-valid-json"))

	require.NoError(t, err) // permanent discard — bad JSON should not requeue
	ch.AssertNotCalled(t, "Send")
}

func TestUserConsume_WrongDeliveryMode(t *testing.T) {
	ch := newMockChannel(dto.ChannelEmail)
	c := &UserNotificationConsumer{channels: asChannels(ch), redis: nil}

	// blast event arriving at the user consumer (wrong mode)
	event := makeTestBlastEvent("evt-001", dto.ChannelEmail)
	err := c.Consume(newCtx(), marshalEvent(event))

	require.NoError(t, err)
	ch.AssertNotCalled(t, "Send")
}

func TestUserConsume_MissingUserID(t *testing.T) {
	ch := newMockChannel(dto.ChannelEmail)
	c := &UserNotificationConsumer{channels: asChannels(ch), redis: nil}

	event := makeTestUserEvent("", "evt-002", dto.ChannelEmail)
	err := c.Consume(newCtx(), marshalEvent(event))

	require.NoError(t, err)
	ch.AssertNotCalled(t, "Send")
}

func TestUserConsume_HappyPath(t *testing.T) {
	ch := newMockChannel(dto.ChannelEmail)
	c := &UserNotificationConsumer{channels: asChannels(ch), redis: nil}

	event := makeTestUserEvent("42", "evt-003", dto.ChannelEmail)
	ch.On("Send", mock.Anything, mock.Anything).Return(nil)

	err := c.Consume(newCtx(), marshalEvent(event))

	require.NoError(t, err)
	ch.AssertCalled(t, "Send", mock.Anything, mock.Anything)
}

func TestUserConsume_ChannelNotTargeted(t *testing.T) {
	// Consumer has email channel, but event targets only push
	emailCh := newMockChannel(dto.ChannelEmail)
	c := &UserNotificationConsumer{channels: asChannels(emailCh), redis: nil}

	event := makeTestUserEvent("42", "evt-004", dto.ChannelPush)
	err := c.Consume(newCtx(), marshalEvent(event))

	require.NoError(t, err)
	emailCh.AssertNotCalled(t, "Send")
}

func TestUserConsume_ChannelSendFails(t *testing.T) {
	ch := newMockChannel(dto.ChannelEmail)
	c := &UserNotificationConsumer{channels: asChannels(ch), redis: nil}

	event := makeTestUserEvent("42", "evt-005", dto.ChannelEmail)
	ch.On("Send", mock.Anything, mock.Anything).Return(errors.New("smtp timeout"))

	err := c.Consume(newCtx(), marshalEvent(event))

	// dispatch returns error when ALL targeted channels fail (triggers requeue)
	require.Error(t, err)
}

func TestUserConsume_EmptyEventID_NilRedis(t *testing.T) {
	// With nil Redis the idempotency guard is bypassed — empty EventID is processed normally
	ch := newMockChannel(dto.ChannelEmail)
	c := &UserNotificationConsumer{channels: asChannels(ch), redis: nil}

	event := makeTestUserEvent("42", "", dto.ChannelEmail) // EventID intentionally empty
	ch.On("Send", mock.Anything, mock.Anything).Return(nil)

	err := c.Consume(newCtx(), marshalEvent(event))

	require.NoError(t, err)
	ch.AssertCalled(t, "Send", mock.Anything, mock.Anything)
}
