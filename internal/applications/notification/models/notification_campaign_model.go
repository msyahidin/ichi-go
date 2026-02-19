package models

import (
	"ichi-go/pkg/db/model"
	"time"

	"github.com/uptrace/bun"
)

// CampaignStatus represents the lifecycle status of a notification campaign.
type CampaignStatus string

const (
	CampaignStatusPending    CampaignStatus = "pending"
	CampaignStatusProcessing CampaignStatus = "processing"
	CampaignStatusPublished  CampaignStatus = "published"
	CampaignStatusFailed     CampaignStatus = "failed"
)

// NotificationCampaign records one POST /api/notifications/send call.
// Tracks the full lifecycle: pending → published | failed.
type NotificationCampaign struct {
	model.CoreModel `bun:"table:notification_campaigns,alias:nc"`

	// DeliveryMode is "blast" or "user".
	DeliveryMode string `bun:"delivery_mode,notnull" json:"delivery_mode"`

	// EventSlug identifies the notification type. Must be registered in the Go TemplateRegistry.
	EventSlug string `bun:"event_slug,notnull" json:"event_slug"`

	// Channels is the list of delivery channels requested.
	// Stored as JSON array: ["email","push"]
	Channels []string `bun:"channels,type:json,notnull" json:"channels"`

	// UserTargetIDs is the list of user IDs to notify (delivery_mode=user only).
	// Stored as JSON array.
	UserTargetIDs []int64 `bun:"user_target_ids,type:json" json:"user_target_ids,omitempty"`

	// UserExcludeIDs is the list of user IDs always excluded from delivery.
	// Applied after UserTargetIDs — excluded IDs are removed before publishing.
	UserExcludeIDs []int64 `bun:"user_exclude_ids,type:json" json:"user_exclude_ids,omitempty"`

	// Locale is the BCP-47 language tag for template selection. Default: "en".
	Locale string `bun:"locale,notnull,default:'en'" json:"locale"`

	// Data holds the template variable map passed to text/template rendering.
	// Stored as JSON object.
	Data map[string]any `bun:"data,type:json" json:"data,omitempty"`

	// Meta holds operational metadata: trace_id, source service, tenant_id, etc.
	// Not shown to end users.
	Meta map[string]string `bun:"meta,type:json" json:"meta,omitempty"`

	// ScheduledAt is the absolute delivery time in UTC (mutually exclusive with DelaySeconds).
	ScheduledAt bun.NullTime `bun:"scheduled_at" json:"scheduled_at,omitempty"`

	// DelaySeconds is the relative delivery delay in seconds (mutually exclusive with ScheduledAt).
	// Max value: 2,147,483 (~24.8 days) due to RabbitMQ x-delay int32 limit.
	DelaySeconds *uint32 `bun:"delay_seconds" json:"delay_seconds,omitempty"`

	// Status is the current lifecycle status.
	Status CampaignStatus `bun:"status,notnull,default:'pending'" json:"status"`

	// ErrorMessage holds the error details when Status is "failed".
	ErrorMessage string `bun:"error_message" json:"error_message,omitempty"`

	// PublishedAt is set when the message was successfully queued in RabbitMQ.
	PublishedAt *time.Time `bun:"published_at" json:"published_at,omitempty"`
}
