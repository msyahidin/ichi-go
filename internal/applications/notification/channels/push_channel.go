package channels

import (
	"context"
	"encoding/json"
	"fmt"

	"ichi-go/internal/applications/notification/dto"
	"ichi-go/pkg/logger"
	"ichi-go/pkg/notification/fcm"
)

// PushChannel delivers notifications via Firebase Cloud Messaging (FCM).
//
// Device tokens are expected in event.Data["device_tokens"] as a JSON array of strings
// or a []string value. Future enhancement: inject a device token repository to look up
// tokens by event.UserID automatically.
//
// FCM pitfalls handled here:
//   - fcmClient is nil (FCM disabled in config): log and return nil (permanent skip)
//   - No device tokens in event data: log warning and return nil (no retry)
//   - Unregistered token: FCM returns ErrCodeUnregistered — log and return nil (no retry)
//   - Batch > 500 tokens: handled via SendToTokensBatch chunking
type PushChannel struct {
	fcmClient *fcm.Client // nil when FCM is disabled
}

// NewPushChannel creates a PushChannel. fcmClient may be nil when FCM is not configured.
func NewPushChannel(fcmClient *fcm.Client) *PushChannel {
	return &PushChannel{fcmClient: fcmClient}
}

func (c *PushChannel) Name() dto.Channel {
	return dto.ChannelPush
}

// Send delivers a push notification via FCM.
//
// Reads title and body from event.Data["__title__"] and event.Data["__body__"]
// (injected by TemplateRenderer in dispatch.go before Send is called).
//
// Reads device tokens from event.Data["device_tokens"] — a JSON array of strings
// or a []interface{} slice.
func (c *PushChannel) Send(ctx context.Context, event dto.NotificationEvent) error {
	// FCM disabled — skip silently.
	if c.fcmClient == nil {
		logger.Debugf("[push] FCM client not configured, skipping event_type=%s", event.EventType)
		return nil
	}

	title, _ := event.Data["__title__"].(string)
	body, _ := event.Data["__body__"].(string)

	// Extract device tokens from event data.
	tokens, err := extractDeviceTokens(event.Data)
	if err != nil {
		logger.Warnf("[push] failed to extract device_tokens event_type=%s user_id=%s: %v",
			event.EventType, event.UserID, err)
		return nil // permanent — bad data, do not retry
	}
	if len(tokens) == 0 {
		logger.Warnf("[push] no device_tokens in event data event_type=%s user_id=%s, skipping",
			event.EventType, event.UserID)
		return nil // permanent — no token = no retry
	}

	// Build FCM data payload from non-reserved event.Data keys.
	fcmData := toFCMData(event.Data)

	// Send via FCM — use batch for multiple tokens.
	logger.Infof("[push] sending to %d token(s) event_type=%s user_id=%s",
		len(tokens), event.EventType, event.UserID)

	failedTokens, err := c.fcmClient.SendToTokensBatch(ctx, tokens, title, body, fcmData)
	if err != nil {
		// Entire batch request failed (network/auth error) — transient, requeue.
		return fmt.Errorf("[push] FCM batch send failed: %w", err)
	}

	if len(failedTokens) > 0 {
		// Partial failure — some tokens failed. Log for token cleanup; don't requeue
		// (other tokens succeeded, and failed tokens may be stale/unregistered).
		logger.Warnf("[push] %d/%d tokens failed event_type=%s user_id=%s failed_tokens=%v",
			len(failedTokens), len(tokens), event.EventType, event.UserID, failedTokens)
	}

	logger.Infof("[push] sent event_type=%s user_id=%s success=%d failed=%d",
		event.EventType, event.UserID, len(tokens)-len(failedTokens), len(failedTokens))

	return nil // partial failure is not a requeue trigger — at least some tokens succeeded
}

// extractDeviceTokens reads device tokens from event.Data["device_tokens"].
// Accepts: []string, []interface{} (from JSON unmarshal), or JSON string array.
func extractDeviceTokens(data map[string]any) ([]string, error) {
	raw, ok := data["device_tokens"]
	if !ok || raw == nil {
		return nil, nil
	}

	switch v := raw.(type) {
	case []string:
		return v, nil
	case []interface{}:
		tokens := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				tokens = append(tokens, s)
			}
		}
		return tokens, nil
	case string:
		// Try to unmarshal a JSON string like "[\"token1\",\"token2\"]"
		var tokens []string
		if err := json.Unmarshal([]byte(v), &tokens); err != nil {
			return nil, fmt.Errorf("device_tokens string is not a valid JSON array: %w", err)
		}
		return tokens, nil
	default:
		return nil, fmt.Errorf("device_tokens has unsupported type %T", raw)
	}
}

// toFCMData converts event.Data to a map[string]string suitable for FCM's data payload.
// Skips reserved keys (prefixed with "__") and non-string values are JSON-encoded.
func toFCMData(data map[string]any) map[string]string {
	result := make(map[string]string, len(data))
	for k, v := range data {
		// Skip internal renderer keys.
		if len(k) >= 2 && k[0] == '_' && k[1] == '_' {
			continue
		}
		// Skip device_tokens — not a user-facing payload field.
		if k == "device_tokens" {
			continue
		}
		switch s := v.(type) {
		case string:
			result[k] = s
		default:
			if b, err := json.Marshal(v); err == nil {
				result[k] = string(b)
			}
		}
	}
	return result
}
