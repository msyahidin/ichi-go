package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"

	"ichi-go/internal/applications/notification/models"
	baseRepository "ichi-go/pkg/db/repository"
)

// NotificationTemplateOverrideRepository fetches optional DB copy overrides.
// All lookups return nil (not an error) when no active row is found —
// callers fall back to Go template defaults.
type NotificationTemplateOverrideRepository struct {
	*baseRepository.BaseRepository[models.NotificationTemplateOverride]
}

func NewNotificationTemplateOverrideRepository(db *bun.DB) *NotificationTemplateOverrideRepository {
	return &NotificationTemplateOverrideRepository{
		BaseRepository: baseRepository.NewRepository[models.NotificationTemplateOverride](db, &models.NotificationTemplateOverride{}),
	}
}

// FindOverride returns the active override row for (eventSlug, channel, locale).
// Locale fallback: tries exact locale first, then falls back to "en".
// Returns nil, nil when no active override exists — this is NOT an error.
func (r *NotificationTemplateOverrideRepository) FindOverride(ctx context.Context, eventSlug, channel, locale string) (*models.NotificationTemplateOverride, error) {
	// Try exact locale first.
	override, err := r.findExact(ctx, eventSlug, channel, locale)
	if err == nil {
		return override, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// Exact locale not found — fall back to "en" (unless we already tried en).
	if locale == "en" {
		return nil, nil // No override; caller uses Go default.
	}

	override, err = r.findExact(ctx, eventSlug, channel, "en")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Still no override; that's fine.
		}
		return nil, err
	}
	return override, nil
}

func (r *NotificationTemplateOverrideRepository) findExact(ctx context.Context, eventSlug, channel, locale string) (*models.NotificationTemplateOverride, error) {
	m := new(models.NotificationTemplateOverride)
	err := r.DB().NewSelect().
		Model(m).
		Where("event_slug = ?", eventSlug).
		Where("channel = ?", channel).
		Where("locale = ?", locale).
		Where("is_active = 1").
		Where("deleted_at IS NULL").
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return m, nil
}
