package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/uptrace/bun"

	notiftemplate "ichi-go/pkg/notification/template"

	"ichi-go/internal/applications/notification/dto"
	"ichi-go/internal/applications/notification/models"
	"ichi-go/internal/applications/notification/repositories"
	"ichi-go/internal/infra/queue/rabbitmq"
	"ichi-go/pkg/logger"
)

const (
	// dispatchRoutingKey is the routing key used on the app.events (x-delayed-message) exchange.
	// The DispatcherConsumer is bound to this key and re-routes to blast/user exchanges.
	dispatchRoutingKey = "notification.dispatch"

	// maxDelaySeconds is the maximum allowed delay to prevent int32 overflow in x-delay header.
	// 2,147,483 seconds ≈ 24.8 days.
	maxDelaySeconds = 2_147_483
)

// CampaignService orchestrates the full notification send flow:
//  1. Validate event slug against Go TemplateRegistry
//  2. Validate channels against event's SupportedChannels()
//  3. Validate schedule/delay constraints
//  4. Persist campaign record (status=pending)
//  5. Compute effective delay and apply user exclusions
//  6. Publish NotificationEvent(s) to app.events exchange with x-delay header
//  7. Update campaign status to published | failed
type CampaignService struct {
	registry     *notiftemplate.Registry
	campaignRepo *repositories.NotificationCampaignRepository
	producer     rabbitmq.MessageProducer // app.events (x-delayed-message) exchange
}

func NewCampaignService(
	registry *notiftemplate.Registry,
	campaignRepo *repositories.NotificationCampaignRepository,
	producer rabbitmq.MessageProducer,
) *CampaignService {
	return &CampaignService{
		registry:     registry,
		campaignRepo: campaignRepo,
		producer:     producer,
	}
}

// Send processes a SendNotificationRequest end-to-end.
// Returns the campaign record with final status and published_at.
func (s *CampaignService) Send(ctx context.Context, req dto.SendNotificationRequest) (*models.NotificationCampaign, error) {
	// --- Step 1: Validate event slug against Go registry ---
	goTmpl, err := s.registry.MustGet(req.EventSlug)
	if err != nil {
		return nil, fmt.Errorf("event_not_registered: %w", err)
	}

	// --- Step 2: Validate channels against template's SupportedChannels() ---
	if err := validateChannels(req.Channels, goTmpl.SupportedChannels()); err != nil {
		return nil, err
	}

	// --- Step 3: Validate schedule/delay ---
	effectiveDelay, err := resolveDelay(req.ScheduledAt, req.DelaySeconds)
	if err != nil {
		return nil, err
	}

	// Default locale to "en".
	locale := req.Locale
	if locale == "" {
		locale = "en"
	}

	// Convert dto.Channel slice to strings for the model.
	channelStrings := make([]string, len(req.Channels))
	for i, ch := range req.Channels {
		channelStrings[i] = string(ch)
	}

	// --- Step 4: Persist campaign record with status=pending ---
	campaign := &models.NotificationCampaign{
		DeliveryMode:   string(req.DeliveryMode),
		EventSlug:      req.EventSlug,
		Channels:       channelStrings,
		UserTargetIDs:  req.UserTargetIDs,
		UserExcludeIDs: req.UserExcludeIDs,
		Locale:         locale,
		Data:           req.Data,
		Meta:           req.Meta,
		DelaySeconds:   req.DelaySeconds,
		Status:         models.CampaignStatusPending,
	}
	if req.ScheduledAt != nil {
		campaign.ScheduledAt = bun.NullTime{Time: *req.ScheduledAt}
	}

	campaign, err = s.campaignRepo.CreateCampaign(ctx, campaign)
	if err != nil {
		return nil, fmt.Errorf("campaign_service: failed to persist campaign: %w", err)
	}

	// --- Step 5: Apply user exclusions ---
	filteredUserIDs := applyExclusions(req.UserTargetIDs, req.UserExcludeIDs)

	// --- Step 6: Publish to app.events (x-delayed-message) ---
	publishErr := s.publish(ctx, campaign, filteredUserIDs, effectiveDelay, locale, req.Data, req.Meta)

	// --- Step 7: Update campaign status ---
	now := time.Now()
	if publishErr != nil {
		logger.Errorf("[campaign] publish failed campaign_id=%d: %v", campaign.ID, publishErr)
		_ = s.campaignRepo.UpdateStatus(ctx, campaign.ID, models.CampaignStatusFailed, publishErr.Error(), nil)
		campaign.Status = models.CampaignStatusFailed
		campaign.ErrorMessage = publishErr.Error()
		return campaign, publishErr
	}

	_ = s.campaignRepo.UpdateStatus(ctx, campaign.ID, models.CampaignStatusPublished, "", &now)
	campaign.Status = models.CampaignStatusPublished
	campaign.PublishedAt = &now
	return campaign, nil
}

// publish sends the NotificationEvent(s) to RabbitMQ.
// Blast → ONE message. User → N messages (one per filtered user ID).
// Returns an error when the producer is nil (queue disabled).
func (s *CampaignService) publish(
	ctx context.Context,
	campaign *models.NotificationCampaign,
	filteredUserIDs []int64,
	delay time.Duration,
	locale string,
	data map[string]any,
	meta map[string]string,
) error {
	if s.producer == nil {
		return fmt.Errorf("campaign_service: message queue is unavailable (producer is nil)")
	}

	channels := make([]dto.Channel, len(campaign.Channels))
	for i, ch := range campaign.Channels {
		channels[i] = dto.Channel(ch)
	}

	baseHeaders := amqp.Table{
		"x-event-type":    campaign.EventSlug,
		"x-campaign-id":   strconv.FormatInt(campaign.ID, 10),
		"x-delivery-mode": campaign.DeliveryMode,
	}

	opts := rabbitmq.PublishOptions{
		Delay:   delay,
		Headers: baseHeaders,
	}

	switch dto.DeliveryMode(campaign.DeliveryMode) {
	case dto.DeliveryModeBlast:
		event := dto.NotificationEvent{
			EventID:      fmt.Sprintf("campaign-%d-blast", campaign.ID),
			EventType:    campaign.EventSlug,
			DeliveryMode: dto.DeliveryModeBlast,
			Channels:     channels,
			Locale:       locale,
			Data:         data,
			Meta:         meta,
		}
		return s.producer.Publish(ctx, dispatchRoutingKey, event, opts)

	case dto.DeliveryModeUser:
		if len(filteredUserIDs) == 0 {
			logger.Warnf("[campaign] delivery_mode=user but no users remain after exclusions, campaign_id=%d", campaign.ID)
			return nil
		}
		for _, userID := range filteredUserIDs {
			event := dto.NotificationEvent{
				EventID:      fmt.Sprintf("campaign-%d-user-%d", campaign.ID, userID),
				EventType:    campaign.EventSlug,
				DeliveryMode: dto.DeliveryModeUser,
				UserID:       strconv.FormatInt(userID, 10),
				Channels:     channels,
				Locale:       locale,
				Data:         data,
				Meta:         meta,
			}
			if err := s.producer.Publish(ctx, dispatchRoutingKey, event, opts); err != nil {
				return fmt.Errorf("failed to publish for user_id=%d: %w", userID, err)
			}
		}
		return nil

	default:
		return fmt.Errorf("unknown delivery_mode: %s", campaign.DeliveryMode)
	}
}

// --- helpers ---

// validateChannels checks that all requested channels are supported by the event template.
func validateChannels(requested []dto.Channel, supported []string) error {
	supportedSet := make(map[string]bool, len(supported))
	for _, s := range supported {
		supportedSet[s] = true
	}

	var unsupported []string
	for _, ch := range requested {
		if !supportedSet[string(ch)] {
			unsupported = append(unsupported, string(ch))
		}
	}

	if len(unsupported) > 0 {
		return fmt.Errorf("channels_not_supported: %s", strings.Join(unsupported, ", "))
	}
	return nil
}

// resolveDelay converts ScheduledAt / DelaySeconds into an effective time.Duration.
// Returns an error if constraints are violated.
func resolveDelay(scheduledAt *time.Time, delaySeconds *uint32) (time.Duration, error) {
	if scheduledAt != nil && delaySeconds != nil {
		return 0, fmt.Errorf("scheduled_at and delay_seconds are mutually exclusive")
	}

	if scheduledAt != nil {
		remaining := time.Until(*scheduledAt)
		if remaining <= 0 {
			return 0, fmt.Errorf("scheduled_at must be in the future")
		}
		if remaining > maxDelaySeconds*time.Second {
			return 0, fmt.Errorf("scheduled_at exceeds maximum delay of ~24.8 days")
		}
		return remaining, nil
	}

	if delaySeconds != nil {
		if *delaySeconds > maxDelaySeconds {
			return 0, fmt.Errorf("delay_seconds must not exceed %d (~24.8 days)", maxDelaySeconds)
		}
		return time.Duration(*delaySeconds) * time.Second, nil
	}

	return 0, nil // no delay
}

// applyExclusions removes excludeIDs from targetIDs and returns the filtered slice.
func applyExclusions(targetIDs, excludeIDs []int64) []int64 {
	if len(excludeIDs) == 0 {
		return targetIDs
	}

	excludeSet := make(map[int64]bool, len(excludeIDs))
	for _, id := range excludeIDs {
		excludeSet[id] = true
	}

	result := make([]int64, 0, len(targetIDs))
	for _, id := range targetIDs {
		if !excludeSet[id] {
			result = append(result, id)
		}
	}
	return result
}
