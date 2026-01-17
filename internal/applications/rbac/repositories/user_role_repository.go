package repositories

import (
	"context"
	"fmt"

	"ichi-go/internal/applications/rbac/models"

	"github.com/uptrace/bun"
)

// UserRoleRepository handles user-role assignment database operations
type UserRoleRepository struct {
	db *bun.DB
}

// NewUserRoleRepository creates a new user role repository
func NewUserRoleRepository(db *bun.DB) *UserRoleRepository {
	return &UserRoleRepository{
		db: db,
	}
}

// FindByUserAndTenant retrieves all role assignments for a user in a tenant
func (r *UserRoleRepository) FindByUserAndTenant(ctx context.Context, userID int64, tenantID string) ([]models.UserRole, error) {
	var userRoles []models.UserRole

	err := r.db.NewSelect().
		Model(&userRoles).
		Relation("Role").
		Where("user_id = ?", userID).
		Where("tenant_id = ?", tenantID).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find user roles: %w", err)
	}

	return userRoles, nil
}

// FindByUser retrieves all role assignments for a user (across all tenants)
func (r *UserRoleRepository) FindByUser(ctx context.Context, userID int64) ([]models.UserRole, error) {
	var userRoles []models.UserRole

	err := r.db.NewSelect().
		Model(&userRoles).
		Relation("Role").
		Where("user_id = ?", userID).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find user roles: %w", err)
	}

	return userRoles, nil
}

// FindByRole retrieves all user assignments for a role
func (r *UserRoleRepository) FindByRole(ctx context.Context, roleID int64, tenantID string) ([]models.UserRole, error) {
	var userRoles []models.UserRole

	query := r.db.NewSelect().
		Model(&userRoles).
		Where("role_id = ?", roleID)

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find role assignments: %w", err)
	}

	return userRoles, nil
}

// FindActiveRoles retrieves active (non-expired) roles for a user in a tenant
func (r *UserRoleRepository) FindActiveRoles(ctx context.Context, userID int64, tenantID string) ([]models.UserRole, error) {
	var userRoles []models.UserRole

	err := r.db.NewSelect().
		Model(&userRoles).
		Relation("Role").
		Where("user_id = ?", userID).
		Where("tenant_id = ?", tenantID).
		Where("(expires_at IS NULL OR expires_at > NOW())").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find active user roles: %w", err)
	}

	return userRoles, nil
}

// Create assigns a role to a user
func (r *UserRoleRepository) Create(ctx context.Context, userRole *models.UserRole) error {
	_, err := r.db.NewInsert().
		Model(userRole).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// Delete removes a role assignment
func (r *UserRoleRepository) Delete(ctx context.Context, userID int64, roleID int64, tenantID string) error {
	_, err := r.db.NewDelete().
		Model((*models.UserRole)(nil)).
		Where("user_id = ?", userID).
		Where("role_id = ?", roleID).
		Where("tenant_id = ?", tenantID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}

	return nil
}

// DeleteAllByUser removes all role assignments for a user in a tenant
func (r *UserRoleRepository) DeleteAllByUser(ctx context.Context, userID int64, tenantID string) error {
	_, err := r.db.NewDelete().
		Model((*models.UserRole)(nil)).
		Where("user_id = ?", userID).
		Where("tenant_id = ?", tenantID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete user roles: %w", err)
	}

	return nil
}

// Exists checks if a user has a specific role in a tenant
func (r *UserRoleRepository) Exists(ctx context.Context, userID int64, roleID int64, tenantID string) (bool, error) {
	exists, err := r.db.NewSelect().
		Model((*models.UserRole)(nil)).
		Where("user_id = ?", userID).
		Where("role_id = ?", roleID).
		Where("tenant_id = ?", tenantID).
		Exists(ctx)

	if err != nil {
		return false, fmt.Errorf("failed to check role assignment: %w", err)
	}

	return exists, nil
}

// CountByRole counts how many users have a specific role
func (r *UserRoleRepository) CountByRole(ctx context.Context, roleID int64, tenantID string) (int, error) {
	query := r.db.NewSelect().
		Model((*models.UserRole)(nil)).
		Where("role_id = ?", roleID)

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	count, err := query.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count role assignments: %w", err)
	}

	return count, nil
}
