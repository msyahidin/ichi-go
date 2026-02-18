package consumers

import (
	"context"
	"ichi-go/internal/applications/notification/channels"
	"ichi-go/internal/applications/notification/dto"
	"ichi-go/pkg/logger"
)

// dispatch iterates registered channels and calls Send on each one the event targets.
//
// Channel failures are logged but never block other channels — a broken push
// provider must not prevent email from being sent.
//
// Returns a non-nil error only when ALL targeted channels fail with transient
// errors, so the message is requeued for a full retry. If at least one channel
// succeeded, nil is returned.
func dispatch(ctx context.Context, event dto.NotificationEvent, chs []channels.NotificationChannel) error {
	var lastTransientErr error
	successCount := 0
	targetCount := 0

	for _, ch := range chs {
		if !event.HasChannel(ch.Name()) {
			continue
		}
		targetCount++

		if err := ch.Send(ctx, event); err != nil {
			logger.Errorf("[dispatch] channel=%s event_id=%s failed: %v",
				ch.Name(), event.EventID, err)
			lastTransientErr = err
			continue
		}

		successCount++
		logger.Debugf("[dispatch] channel=%s event_id=%s ok", ch.Name(), event.EventID)
	}

	if targetCount == 0 {
		logger.Warnf("[dispatch] event_id=%s has no matching registered channels for %v",
			event.EventID, event.Channels)
		return nil
	}

	// Partial success: at least one channel delivered — ack the message.
	// Failed channels' logs are the signal for ops to investigate.
	if successCount > 0 {
		return nil
	}

	// All targeted channels failed — return error to trigger requeue.
	return lastTransientErr
}
