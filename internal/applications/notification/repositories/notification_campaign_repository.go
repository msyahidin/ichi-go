package repositories

import (
	"context"
	"time"

	"github.com/uptrace/bun"

	"ichi-go/internal/applications/notification/models"
	baseRepository "ichi-go/pkg/db/repository"
)

// NotificationCampaignRepository manages notification campaign lifecycle records.
type NotificationCampaignRepository struct {
	*baseRepository.BaseRepository[models.NotificationCampaign]
}

func NewNotificationCampaignRepository(db *bun.DB) *NotificationCampaignRepository {
	return &NotificationCampaignRepository{
		BaseRepository: baseRepository.NewRepository[models.NotificationCampaign](db, &models.NotificationCampaign{}),
	}
}

// CreateCampaign inserts a new campaign record and returns it with the generated ID.
func (r *NotificationCampaignRepository) CreateCampaign(ctx context.Context, campaign *models.NotificationCampaign) (*models.NotificationCampaign, error) {
	return r.Create(ctx, campaign)
}

// FindByID retrieves a campaign by primary key.
func (r *NotificationCampaignRepository) FindByID(ctx context.Context, id int64) (*models.NotificationCampaign, error) {
	return r.Find(ctx, id)
}

// UpdateStatus sets status, optional error message, and optional published_at timestamp.
// Uses raw SQL to avoid BaseRepository.Update's OmitZero skipping zero-value status fields.
func (r *NotificationCampaignRepository) UpdateStatus(
	ctx context.Context,
	id int64,
	status models.CampaignStatus,
	errMsg string,
	publishedAt *time.Time,
) error {
	q := r.DB().NewUpdate().
		TableExpr("notification_campaigns").
		Set("status = ?", status).
		Set("updated_at = ?", time.Now())

	if errMsg != "" {
		q = q.Set("error_message = ?", errMsg)
	}
	if publishedAt != nil {
		q = q.Set("published_at = ?", *publishedAt)
	}

	_, err := q.Where("id = ?", id).Exec(ctx)
	return err
}
