package repositories

import (
	"context"
	"fmt"

	"ichi-go/internal/applications/rbac/models"

	"github.com/uptrace/bun"
)

// PermissionRepository handles permission database operations
type PermissionRepository struct {
	db *bun.DB
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *bun.DB) *PermissionRepository {
	return &PermissionRepository{
		db: db,
	}
}

// FindAll retrieves all permissions
func (r *PermissionRepository) FindAll(ctx context.Context) ([]models.Permission, error) {
	var permissions []models.Permission

	err := r.db.NewSelect().
		Model(&permissions).
		Order("module ASC", "name ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find all permissions: %w", err)
	}

	return permissions, nil
}

// FindByID retrieves a permission by ID
func (r *PermissionRepository) FindByID(ctx context.Context, id int64) (*models.Permission, error) {
	permission := new(models.Permission)

	err := r.db.NewSelect().
		Model(permission).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find permission by ID: %w", err)
	}

	return permission, nil
}

// FindBySlug retrieves a permission by slug
func (r *PermissionRepository) FindBySlug(ctx context.Context, slug string) (*models.Permission, error) {
	permission := new(models.Permission)

	err := r.db.NewSelect().
		Model(permission).
		Where("slug = ?", slug).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find permission by slug: %w", err)
	}

	return permission, nil
}

// FindByModule retrieves all permissions for a specific module
func (r *PermissionRepository) FindByModule(ctx context.Context, module string) ([]models.Permission, error) {
	var permissions []models.Permission

	err := r.db.NewSelect().
		Model(&permissions).
		Where("module = ?", module).
		Order("name ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find permissions by module: %w", err)
	}

	return permissions, nil
}

// FindByIDs retrieves multiple permissions by IDs
func (r *PermissionRepository) FindByIDs(ctx context.Context, ids []int64) ([]models.Permission, error) {
	var permissions []models.Permission

	err := r.db.NewSelect().
		Model(&permissions).
		Where("id IN (?)", bun.In(ids)).
		Order("module ASC", "name ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find permissions by IDs: %w", err)
	}

	return permissions, nil
}

// FindBySlugs retrieves multiple permissions by slugs
func (r *PermissionRepository) FindBySlugs(ctx context.Context, slugs []string) ([]models.Permission, error) {
	var permissions []models.Permission

	err := r.db.NewSelect().
		Model(&permissions).
		Where("slug IN (?)", bun.In(slugs)).
		Order("module ASC", "name ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find permissions by slugs: %w", err)
	}

	return permissions, nil
}

// FindWithGroups retrieves a permission with its groups
func (r *PermissionRepository) FindWithGroups(ctx context.Context, id int64) (*models.Permission, error) {
	permission := new(models.Permission)

	err := r.db.NewSelect().
		Model(permission).
		Relation("Groups").
		Where("permission.id = ?", id).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find permission with groups: %w", err)
	}

	return permission, nil
}

// GroupByModule retrieves permissions grouped by module
func (r *PermissionRepository) GroupByModule(ctx context.Context) (map[string][]models.Permission, error) {
	var permissions []models.Permission

	err := r.db.NewSelect().
		Model(&permissions).
		Order("module ASC", "name ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to group permissions by module: %w", err)
	}

	// Group by module
	grouped := make(map[string][]models.Permission)
	for _, perm := range permissions {
		module := "general"
		if perm.Module != nil {
			module = *perm.Module
		}

		grouped[module] = append(grouped[module], perm)
	}

	return grouped, nil
}

// Create creates a new permission
func (r *PermissionRepository) Create(ctx context.Context, permission *models.Permission) error {
	_, err := r.db.NewInsert().
		Model(permission).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}

	return nil
}

// Update updates an existing permission
func (r *PermissionRepository) Update(ctx context.Context, permission *models.Permission) error {
	_, err := r.db.NewUpdate().
		Model(permission).
		WherePK().
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update permission: %w", err)
	}

	return nil
}

// Delete deletes a permission
func (r *PermissionRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().
		Model((*models.Permission)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	return nil
}

// Exists checks if a permission exists by slug
func (r *PermissionRepository) Exists(ctx context.Context, slug string) (bool, error) {
	exists, err := r.db.NewSelect().
		Model((*models.Permission)(nil)).
		Where("slug = ?", slug).
		Exists(ctx)

	if err != nil {
		return false, fmt.Errorf("failed to check permission existence: %w", err)
	}

	return exists, nil
}
