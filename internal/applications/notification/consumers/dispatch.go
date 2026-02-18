package consumers

import (
	"context"
	"strconv"
	"time"

	"ichi-go/internal/applications/notification/channels"
	"ichi-go/internal/applications/notification/dto"
	"ichi-go/internal/applications/notification/models"
	"ichi-go/internal/applications/notification/repositories"
	"ichi-go/internal/applications/notification/services"
	"ichi-go/pkg/logger"
)

// dispatch iterates registered channels and calls Send on each one the event targets.
//
// Before calling each channel's Send(), TemplateRenderer.Render() is invoked to inject
// __title__ and __body__ into event.Data. If rendering fails (bad DB override syntax),
// the channel is skipped (permanent — do not requeue).
//
// Channel failures are logged but never block other channels — a broken push
// provider must not prevent email from being sent.
//
// Returns a non-nil error only when ALL targeted channels fail with transient
// errors, so the message is requeued for a full retry. If at least one channel
// succeeded, nil is returned.
func dispatch(
	ctx context.Context,
	event dto.NotificationEvent,
	chs []channels.NotificationChannel,
	renderer *services.TemplateRenderer,
	logRepo *repositories.NotificationLogRepository,
	campaignID int64,
) error {
	var lastTransientErr error
	successCount := 0
	targetCount := 0

	userID, _ := strconv.ParseInt(event.UserID, 10, 64)

	for _, ch := range chs {
		if !event.HasChannel(ch.Name()) {
			continue
		}
		targetCount++

		// Create pending log entry.
		log := &models.NotificationLog{
			CampaignID: campaignID,
			UserID:     userID,
			Channel:    string(ch.Name()),
			Status:     models.LogStatusPending,
		}
		if logRepo != nil {
			if created, err := logRepo.CreateLog(ctx, log); err == nil {
				log = created
			}
		}

		// Render template — inject __title__ and __body__ into event.Data.
		// Make a per-channel copy of Data so different channels get independent rendering.
		eventCopy := event
		eventCopy.Data = copyData(event.Data)

		if renderer != nil {
			locale := event.Locale
			if locale == "" {
				locale = "en"
			}
			rendered, err := renderer.Render(ctx, event.EventType, string(ch.Name()), locale, eventCopy.Data)
			if err != nil {
				logger.Warnf("[dispatch] template render failed channel=%s event_id=%s: %v — skipping channel",
					ch.Name(), event.EventID, err)
				if logRepo != nil && log.ID > 0 {
					_ = logRepo.UpdateStatus(ctx, log.ID, models.LogStatusSkipped, err.Error(), nil)
				}
				continue // permanent skip — bad DB override syntax
			}
			if eventCopy.Data == nil {
				eventCopy.Data = make(map[string]any)
			}
			eventCopy.Data["__title__"] = rendered.Title
			eventCopy.Data["__body__"] = rendered.Body
		}

		// Send via channel.
		if err := ch.Send(ctx, eventCopy); err != nil {
			logger.Errorf("[dispatch] channel=%s event_id=%s failed: %v",
				ch.Name(), event.EventID, err)
			if logRepo != nil && log.ID > 0 {
				_ = logRepo.UpdateStatus(ctx, log.ID, models.LogStatusFailed, err.Error(), nil)
			}
			lastTransientErr = err
			continue
		}

		successCount++
		now := time.Now()
		if logRepo != nil && log.ID > 0 {
			_ = logRepo.UpdateStatus(ctx, log.ID, models.LogStatusSent, "", &now)
		}
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

// copyData returns a shallow copy of the data map so per-channel rendering
// doesn't cross-contaminate template variables across channels.
func copyData(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
