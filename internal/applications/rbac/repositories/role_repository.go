package repositories

import (
	"context"
	"fmt"

	"ichi-go/internal/applications/rbac/models"

	"github.com/uptrace/bun"
)

// RoleRepository handles role database operations
type RoleRepository struct {
	db *bun.DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *bun.DB) *RoleRepository {
	return &RoleRepository{
		db: db,
	}
}

// FindAll retrieves all roles
func (r *RoleRepository) FindAll(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role

	err := r.db.NewSelect().
		Model(&roles).
		Where("deleted_at IS NULL").
		Order("level DESC", "name ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find all roles: %w", err)
	}

	return roles, nil
}

// FindByID retrieves a role by ID
func (r *RoleRepository) FindByID(ctx context.Context, id int64) (*models.Role, error) {
	role := new(models.Role)

	err := r.db.NewSelect().
		Model(role).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find role by ID: %w", err)
	}

	return role, nil
}

// FindBySlug retrieves a role by slug
func (r *RoleRepository) FindBySlug(ctx context.Context, slug string, tenantID *string) (*models.Role, error) {
	role := new(models.Role)

	query := r.db.NewSelect().
		Model(role).
		Where("slug = ?", slug).
		Where("deleted_at IS NULL")

	if tenantID != nil {
		query = query.Where("tenant_id = ?", *tenantID)
	} else {
		query = query.Where("tenant_id IS NULL")
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find role by slug: %w", err)
	}

	return role, nil
}

// FindGlobalRoles retrieves all global roles (tenant_id IS NULL)
func (r *RoleRepository) FindGlobalRoles(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role

	err := r.db.NewSelect().
		Model(&roles).
		Where("tenant_id IS NULL").
		Where("deleted_at IS NULL").
		Order("level DESC", "name ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find global roles: %w", err)
	}

	return roles, nil
}

// FindTenantRoles retrieves all roles for a specific tenant
func (r *RoleRepository) FindTenantRoles(ctx context.Context, tenantID string) ([]models.Role, error) {
	var roles []models.Role

	err := r.db.NewSelect().
		Model(&roles).
		Where("tenant_id = ?", tenantID).
		Where("deleted_at IS NULL").
		Order("level DESC", "name ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find tenant roles: %w", err)
	}

	return roles, nil
}

// FindWithPermissions retrieves a role with its permissions
func (r *RoleRepository) FindWithPermissions(ctx context.Context, id int64) (*models.Role, error) {
	role := new(models.Role)

	err := r.db.NewSelect().
		Model(role).
		Relation("Permissions").
		Where("role.id = ?", id).
		Where("role.deleted_at IS NULL").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find role with permissions: %w", err)
	}

	return role, nil
}

// Create creates a new role
func (r *RoleRepository) Create(ctx context.Context, role *models.Role) error {
	_, err := r.db.NewInsert().
		Model(role).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}

// Update updates an existing role
func (r *RoleRepository) Update(ctx context.Context, role *models.Role) error {
	_, err := r.db.NewUpdate().
		Model(role).
		WherePK().
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	return nil
}

// Delete soft deletes a role
func (r *RoleRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewUpdate().
		Model((*models.Role)(nil)).
		Set("deleted_at = NOW()").
		Where("id = ?", id).
		Where("is_system_role = ?", false). // Prevent deletion of system roles
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// Exists checks if a role exists by slug
func (r *RoleRepository) Exists(ctx context.Context, slug string, tenantID *string) (bool, error) {
	query := r.db.NewSelect().
		Model((*models.Role)(nil)).
		Where("slug = ?", slug).
		Where("deleted_at IS NULL")

	if tenantID != nil {
		query = query.Where("tenant_id = ?", *tenantID)
	} else {
		query = query.Where("tenant_id IS NULL")
	}

	exists, err := query.Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check role existence: %w", err)
	}

	return exists, nil
}
