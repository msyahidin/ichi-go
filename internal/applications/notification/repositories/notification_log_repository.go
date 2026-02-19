package repositories

import (
	"context"
	"time"

	"github.com/uptrace/bun"

	"ichi-go/internal/applications/notification/models"
)

// NotificationLogRepository manages per-user, per-channel delivery attempt records.
// Logs are append-only — no updates to existing rows except status + error via UpdateStatus.
type NotificationLogRepository struct {
	db *bun.DB
}

func NewNotificationLogRepository(db *bun.DB) *NotificationLogRepository {
	return &NotificationLogRepository{db: db}
}

// CreateLog inserts a single delivery attempt log.
func (r *NotificationLogRepository) CreateLog(ctx context.Context, log *models.NotificationLog) (*models.NotificationLog, error) {
	_, err := r.db.NewInsert().Model(log).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return log, nil
}

// CreateBatch inserts multiple delivery attempt logs in one statement.
// Efficient for campaigns with many target users.
func (r *NotificationLogRepository) CreateBatch(ctx context.Context, logs []*models.NotificationLog) error {
	if len(logs) == 0 {
		return nil
	}
	_, err := r.db.NewInsert().Model(&logs).Exec(ctx)
	return err
}

// UpdateStatus updates the status, error, and sent_at timestamp of a log entry.
// Uses raw SQL to avoid OmitZero skipping the "sent" status string.
func (r *NotificationLogRepository) UpdateStatus(
	ctx context.Context,
	id int64,
	status models.LogStatus,
	errMsg string,
	sentAt *time.Time,
) error {
	q := r.db.NewUpdate().
		TableExpr("notification_logs").
		Set("status = ?", status).
		Set("updated_at = ?", time.Now())

	if errMsg != "" {
		q = q.Set("error = ?", errMsg)
	}
	if sentAt != nil {
		q = q.Set("sent_at = ?", *sentAt)
	}

	_, err := q.Where("id = ?", id).Exec(ctx)
	return err
}
