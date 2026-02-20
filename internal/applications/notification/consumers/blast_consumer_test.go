package consumers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"ichi-go/internal/applications/notification/dto"
)

// ============================================================================
// extractCampaignID — pure helper, table-driven
// ============================================================================

func TestExtractCampaignID(t *testing.T) {
	tests := []struct {
		name     string
		meta     map[string]string
		expected int64
	}{
		{"nil meta", nil, 0},
		{"empty meta", map[string]string{}, 0},
		{"valid campaign_id", map[string]string{"campaign_id": "42"}, 42},
		{"non-integer campaign_id", map[string]string{"campaign_id": "abc"}, 0},
		{"zero campaign_id", map[string]string{"campaign_id": "0"}, 0},
		{"unrelated key", map[string]string{"source": "api"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, extractCampaignID(tt.meta))
		})
	}
}

// ============================================================================
// BlastConsumer.Consume()
// ============================================================================

func TestBlastConsume_InvalidJSON(t *testing.T) {
	ch := newMockChannel(dto.ChannelEmail)
	c := &BlastConsumer{channels: asChannels(ch)}

	err := c.Consume(newCtx(), []byte("not-json"))

	require.NoError(t, err) // permanent discard
	ch.AssertNotCalled(t, "Send")
}

func TestBlastConsume_WrongDeliveryMode(t *testing.T) {
	ch := newMockChannel(dto.ChannelEmail)
	c := &BlastConsumer{channels: asChannels(ch)}

	// user event arriving at blast consumer
	event := makeTestUserEvent("42", "evt-001", dto.ChannelEmail)
	err := c.Consume(newCtx(), marshalEvent(event))

	require.NoError(t, err)
	ch.AssertNotCalled(t, "Send")
}

func TestBlastConsume_HappyPath(t *testing.T) {
	ch := newMockChannel(dto.ChannelEmail)
	c := &BlastConsumer{channels: asChannels(ch)}

	event := makeTestBlastEvent("evt-002", dto.ChannelEmail)
	ch.On("Send", mock.Anything, mock.Anything).Return(nil)

	err := c.Consume(newCtx(), marshalEvent(event))

	require.NoError(t, err)
	ch.AssertCalled(t, "Send", mock.Anything, mock.Anything)
}

func TestBlastConsume_ChannelNotTargeted(t *testing.T) {
	// Consumer has email, event only targets push
	emailCh := newMockChannel(dto.ChannelEmail)
	c := &BlastConsumer{channels: asChannels(emailCh)}

	event := makeTestBlastEvent("evt-003", dto.ChannelPush)
	err := c.Consume(newCtx(), marshalEvent(event))

	require.NoError(t, err)
	emailCh.AssertNotCalled(t, "Send")
}

func TestBlastConsume_ChannelSendFails(t *testing.T) {
	ch := newMockChannel(dto.ChannelEmail)
	c := &BlastConsumer{channels: asChannels(ch)}

	event := makeTestBlastEvent("evt-004", dto.ChannelEmail)
	ch.On("Send", mock.Anything, mock.Anything).Return(errors.New("smtp error"))

	err := c.Consume(newCtx(), marshalEvent(event))

	require.Error(t, err) // all channels failed → requeue
}

func TestBlastConsume_PartialSuccess(t *testing.T) {
	// Two channels: email fails, push succeeds → partial success → nil (ack)
	emailCh := newMockChannel(dto.ChannelEmail)
	pushCh := newMockChannel(dto.ChannelPush)
	c := &BlastConsumer{channels: asChannels(emailCh, pushCh)}

	event := makeTestBlastEvent("evt-005", dto.ChannelEmail, dto.ChannelPush)
	emailCh.On("Send", mock.Anything, mock.Anything).Return(errors.New("smtp error"))
	pushCh.On("Send", mock.Anything, mock.Anything).Return(nil)

	err := c.Consume(newCtx(), marshalEvent(event))

	require.NoError(t, err) // at least one succeeded — ack
	emailCh.AssertCalled(t, "Send", mock.Anything, mock.Anything)
	pushCh.AssertCalled(t, "Send", mock.Anything, mock.Anything)
}

func TestBlastConsume_AllChannelsFail(t *testing.T) {
	emailCh := newMockChannel(dto.ChannelEmail)
	pushCh := newMockChannel(dto.ChannelPush)
	c := &BlastConsumer{channels: asChannels(emailCh, pushCh)}

	event := makeTestBlastEvent("evt-006", dto.ChannelEmail, dto.ChannelPush)
	emailCh.On("Send", mock.Anything, mock.Anything).Return(errors.New("email down"))
	pushCh.On("Send", mock.Anything, mock.Anything).Return(errors.New("fcm down"))

	err := c.Consume(newCtx(), marshalEvent(event))

	require.Error(t, err) // all failed → requeue
}

func TestBlastConsume_NilRenderer(t *testing.T) {
	// dispatch() skips rendering when renderer is nil — channel.Send still called
	ch := newMockChannel(dto.ChannelEmail)
	c := &BlastConsumer{channels: asChannels(ch), renderer: nil}

	event := makeTestBlastEvent("evt-007", dto.ChannelEmail)
	ch.On("Send", mock.Anything, mock.Anything).Return(nil)

	err := c.Consume(newCtx(), marshalEvent(event))

	require.NoError(t, err)
	ch.AssertCalled(t, "Send", mock.Anything, mock.Anything)
}

// ============================================================================
// dispatch() partial-success semantics (additional coverage)
// ============================================================================

func TestBlastConsume_EventWithMeta_ExtractsCampaignID(t *testing.T) {
	ch := newMockChannel(dto.ChannelEmail)
	c := &BlastConsumer{channels: asChannels(ch)}

	event := makeTestBlastEvent("evt-008", dto.ChannelEmail)
	event.Meta = map[string]string{"campaign_id": "99"}
	ch.On("Send", mock.Anything, mock.Anything).Return(nil)

	err := c.Consume(newCtx(), marshalEvent(event))
	require.NoError(t, err)
	// campaignID=99 is passed to dispatch — no logRepo means it's silently skipped
	ch.AssertCalled(t, "Send", mock.Anything, mock.Anything)
}
