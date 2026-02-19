package models

import "time"

// LogStatus represents the delivery status of a single notification attempt.
type LogStatus string

const (
	LogStatusPending LogStatus = "pending"
	LogStatusSent    LogStatus = "sent"
	LogStatusFailed  LogStatus = "failed"
	LogStatusSkipped LogStatus = "skipped"
)

// NotificationLog records a single per-user, per-channel delivery attempt.
// Written by dispatch() after each channel Send() call.
//
// Note: This model does NOT embed CoreModel — it uses a simpler schema
// (no soft-delete, no audit fields) because logs are append-only and immutable.
type NotificationLog struct {
	ID         int64      `bun:"id,pk,autoincrement"                         json:"id"`
	CreatedAt  time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt  *time.Time `bun:"updated_at"                                  json:"updated_at,omitempty"`
	CampaignID int64      `bun:"campaign_id,notnull"                         json:"campaign_id"`
	UserID     int64      `bun:"user_id,notnull,default:0"                   json:"user_id"`
	Channel    string     `bun:"channel,notnull"                             json:"channel"`
	Status     LogStatus  `bun:"status,notnull,default:'pending'"            json:"status"`
	Error      string     `bun:"error"                                       json:"error,omitempty"`
	SentAt     *time.Time `bun:"sent_at"                                     json:"sent_at,omitempty"`

	_ struct{} `bun:"table:notification_logs,alias:nl"`
}
