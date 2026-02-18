package dto

import "time"

// SendNotificationRequest is the API request body for POST /api/notifications/send.
type SendNotificationRequest struct {
	// EventSlug identifies the notification type.
	// Must be registered in the Go TemplateRegistry (pkg/notification/template).
	EventSlug string `json:"event_slug" validate:"required"`

	// DeliveryMode controls routing: "blast" (all users) or "user" (specific users).
	DeliveryMode DeliveryMode `json:"delivery_mode" validate:"required,oneof=blast user"`

	// Channels lists the delivery channels. Each must be supported by the event's Go template.
	Channels []Channel `json:"channels" validate:"required,min=1"`

	// UserTargetIDs is the list of user IDs to notify.
	// Required and non-empty when DeliveryMode is "user". Ignored for "blast".
	UserTargetIDs []int64 `json:"user_target_ids,omitempty"`

	// UserExcludeIDs is the list of user IDs to always exclude from delivery.
	// Applied after UserTargetIDs — excluded IDs are removed before publishing.
	UserExcludeIDs []int64 `json:"user_exclude_ids,omitempty"`

	// Locale is the BCP-47 language tag for template selection.
	// Defaults to "en" when empty.
	Locale string `json:"locale,omitempty" validate:"omitempty,bcp47_language_tag"`

	// Data holds the template variables used during text/template rendering.
	// Keys must match the variables expected by the event's template.
	Data map[string]any `json:"data,omitempty"`

	// Meta holds operational metadata: trace_id, source service, correlation IDs.
	// Not shown to end users.
	Meta map[string]string `json:"meta,omitempty"`

	// ScheduledAt is the absolute delivery time in UTC (RFC3339).
	// Mutually exclusive with DelaySeconds.
	// Must be in the future.
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`

	// DelaySeconds is the relative delivery delay in seconds.
	// Mutually exclusive with ScheduledAt.
	// Maximum: 2,147,483 seconds (~24.8 days) due to RabbitMQ x-delay int32 limit.
	DelaySeconds *uint32 `json:"delay_seconds,omitempty"`
}

// SendNotificationResponse is returned on successful campaign creation.
type SendNotificationResponse struct {
	// CampaignID is the ID of the created NotificationCampaign record.
	CampaignID int64 `json:"campaign_id"`

	// Status is the campaign status after the API call: "published" or "failed".
	Status string `json:"status"`

	// PublishedAt is set when the campaign was successfully queued.
	PublishedAt *time.Time `json:"published_at,omitempty"`
}
