package services

import (
	"context"
	"fmt"

	"ichi-go/internal/applications/rbac/models"
	"ichi-go/internal/applications/rbac/repositories"
	"ichi-go/pkg/logger"
)

// RoleService handles role management operations
type RoleService struct {
	roleRepo       *repositories.RoleRepository
	permissionRepo *repositories.PermissionRepository
}

// NewRoleService creates a new role service
func NewRoleService(
	roleRepo *repositories.RoleRepository,
	permissionRepo *repositories.PermissionRepository,
) *RoleService {
	return &RoleService{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

// GetAllRoles retrieves all roles
func (s *RoleService) GetAllRoles(ctx context.Context) ([]models.Role, error) {
	return s.roleRepo.FindAll(ctx)
}

// GetRoleByID retrieves a role by ID
func (s *RoleService) GetRoleByID(ctx context.Context, id int64) (*models.Role, error) {
	return s.roleRepo.FindByID(ctx, id)
}

// GetRoleBySlug retrieves a role by slug
func (s *RoleService) GetRoleBySlug(ctx context.Context, slug string, tenantID *string) (*models.Role, error) {
	return s.roleRepo.FindBySlug(ctx, slug, tenantID)
}

// GetGlobalRoles retrieves all global roles
func (s *RoleService) GetGlobalRoles(ctx context.Context) ([]models.Role, error) {
	return s.roleRepo.FindGlobalRoles(ctx)
}

// GetTenantRoles retrieves all roles for a specific tenant
func (s *RoleService) GetTenantRoles(ctx context.Context, tenantID string) ([]models.Role, error) {
	return s.roleRepo.FindTenantRoles(ctx, tenantID)
}

// GetRoleWithPermissions retrieves a role with its permissions
func (s *RoleService) GetRoleWithPermissions(ctx context.Context, id int64) (*models.Role, error) {
	return s.roleRepo.FindWithPermissions(ctx, id)
}

// CreateRole creates a new role
func (s *RoleService) CreateRole(ctx context.Context, role *models.Role) error {
	// Validate slug uniqueness
	exists, err := s.roleRepo.Exists(ctx, role.Slug, role.TenantID)
	if err != nil {
		return fmt.Errorf("failed to check role existence: %w", err)
	}
	if exists {
		return fmt.Errorf("role with slug '%s' already exists", role.Slug)
	}

	// Create role
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	logger.WithContext(ctx).Infof("Role created: id=%d slug=%s", role.ID, role.Slug)

	return nil
}

// UpdateRole updates an existing role
func (s *RoleService) UpdateRole(ctx context.Context, role *models.Role) error {
	// Check if role exists
	existing, err := s.roleRepo.FindByID(ctx, role.ID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Prevent changes to system role slug
	if existing.IsSystemRole && existing.Slug != role.Slug {
		return fmt.Errorf("cannot change slug of system role")
	}

	// Update role
	if err := s.roleRepo.Update(ctx, role); err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	logger.WithContext(ctx).Infof("Role updated: id=%d slug=%s", role.ID, role.Slug)

	return nil
}

// DeleteRole deletes a role (only if not a system role)
func (s *RoleService) DeleteRole(ctx context.Context, id int64) error {
	// Check if role can be deleted
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	if !role.CanDelete() {
		return fmt.Errorf("cannot delete system role: %s", role.Slug)
	}

	// Delete role
	if err := s.roleRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	logger.WithContext(ctx).Infof("Role deleted: id=%d slug=%s", role.ID, role.Slug)

	return nil
}
