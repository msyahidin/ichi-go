package repositories

import (
	"context"
	"fmt"

	"ichi-go/internal/applications/rbac/models"

	"github.com/uptrace/bun"
)

// PlatformRepository handles platform permission database operations
type PlatformRepository struct {
	db *bun.DB
}

// NewPlatformRepository creates a new platform repository
func NewPlatformRepository(db *bun.DB) *PlatformRepository {
	return &PlatformRepository{
		db: db,
	}
}

// FindByUser retrieves all platform permissions for a user
func (r *PlatformRepository) FindByUser(ctx context.Context, userID int64) ([]models.PlatformPermission, error) {
	var permissions []models.PlatformPermission

	err := r.db.NewSelect().
		Model(&permissions).
		Where("user_id = ?", userID).
		Where("(expires_at IS NULL OR expires_at > NOW())").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find platform permissions: %w", err)
	}

	return permissions, nil
}

// FindByPermission retrieves all users with a specific platform permission
func (r *PlatformRepository) FindByPermission(ctx context.Context, permission string) ([]models.PlatformPermission, error) {
	var permissions []models.PlatformPermission

	err := r.db.NewSelect().
		Model(&permissions).
		Where("permission = ?", permission).
		Where("(expires_at IS NULL OR expires_at > NOW())").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find users with permission: %w", err)
	}

	return permissions, nil
}

// HasPermission checks if a user has a specific platform permission
func (r *PlatformRepository) HasPermission(ctx context.Context, userID int64, permission string) (bool, error) {
	exists, err := r.db.NewSelect().
		Model((*models.PlatformPermission)(nil)).
		Where("user_id = ?", userID).
		Where("permission = ?", permission).
		Where("(expires_at IS NULL OR expires_at > NOW())").
		Exists(ctx)

	if err != nil {
		return false, fmt.Errorf("failed to check platform permission: %w", err)
	}

	return exists, nil
}

// IsPlatformAdmin checks if a user has platform admin permission
func (r *PlatformRepository) IsPlatformAdmin(ctx context.Context, userID int64) (bool, error) {
	return r.HasPermission(ctx, userID, models.PlatformAdmin)
}

// Grant grants a platform permission to a user
func (r *PlatformRepository) Grant(ctx context.Context, permission *models.PlatformPermission) error {
	_, err := r.db.NewInsert().
		Model(permission).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to grant platform permission: %w", err)
	}

	return nil
}

// Revoke revokes a platform permission from a user
func (r *PlatformRepository) Revoke(ctx context.Context, userID int64, permission string) error {
	_, err := r.db.NewDelete().
		Model((*models.PlatformPermission)(nil)).
		Where("user_id = ?", userID).
		Where("permission = ?", permission).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to revoke platform permission: %w", err)
	}

	return nil
}

// RevokeAll revokes all platform permissions from a user
func (r *PlatformRepository) RevokeAll(ctx context.Context, userID int64) error {
	_, err := r.db.NewDelete().
		Model((*models.PlatformPermission)(nil)).
		Where("user_id = ?", userID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to revoke all platform permissions: %w", err)
	}

	return nil
}

// FindAll retrieves all platform permissions
func (r *PlatformRepository) FindAll(ctx context.Context) ([]models.PlatformPermission, error) {
	var permissions []models.PlatformPermission

	err := r.db.NewSelect().
		Model(&permissions).
		Order("user_id ASC", "permission ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find all platform permissions: %w", err)
	}

	return permissions, nil
}

// CountByPermission counts users with a specific platform permission
func (r *PlatformRepository) CountByPermission(ctx context.Context, permission string) (int, error) {
	count, err := r.db.NewSelect().
		Model((*models.PlatformPermission)(nil)).
		Where("permission = ?", permission).
		Where("(expires_at IS NULL OR expires_at > NOW())").
		Count(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to count platform permissions: %w", err)
	}

	return count, nil
}
