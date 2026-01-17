package services

import (
	"context"
	"fmt"
	"time"

	"ichi-go/internal/applications/rbac/models"
	"ichi-go/internal/applications/rbac/repositories"
	"ichi-go/internal/infra/authz/enforcer"
	"ichi-go/internal/infra/authz/watcher"
	"ichi-go/internal/infra/queue/rabbitmq"
	"ichi-go/pkg/logger"
)

// UserRoleService handles user-role assignment operations
type UserRoleService struct {
	userRoleRepo *repositories.UserRoleRepository
	roleRepo     *repositories.RoleRepository
	auditRepo    *repositories.AuditRepository
	enforcer     *enforcer.Enforcer
	publisher    rabbitmq.MessageProducer
}

// NewUserRoleService creates a new user role service
func NewUserRoleService(
	userRoleRepo *repositories.UserRoleRepository,
	roleRepo *repositories.RoleRepository,
	auditRepo *repositories.AuditRepository,
	enforcer *enforcer.Enforcer,
	publisher rabbitmq.MessageProducer,
) *UserRoleService {
	return &UserRoleService{
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
		auditRepo:    auditRepo,
		enforcer:     enforcer,
		publisher:    publisher,
	}
}

// AssignRole assigns a role to a user in a specific tenant
func (s *UserRoleService) AssignRole(
	ctx context.Context,
	userID int64,
	roleSlug string,
	tenantID string,
	assignedBy int64,
	reason string,
) error {
	// Get role by slug
	role, err := s.roleRepo.FindBySlug(ctx, roleSlug, nil)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Check if assignment already exists
	exists, err := s.userRoleRepo.Exists(ctx, userID, role.ID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to check existing assignment: %w", err)
	}
	if exists {
		return fmt.Errorf("user already has role '%s' in tenant '%s'", roleSlug, tenantID)
	}

	// Create user role assignment
	userRole := &models.UserRole{
		UserID:     userID,
		RoleID:     role.ID,
		TenantID:   tenantID,
		AssignedBy: &assignedBy,
	}

	if err := s.userRoleRepo.Create(ctx, userRole); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	// Add Casbin grouping policy
	if err := s.enforcer.AssignRoleToUser(fmt.Sprintf("%d", userID), roleSlug, tenantID); err != nil {
		return fmt.Errorf("failed to add Casbin policy: %w", err)
	}

	// Audit the assignment
	s.auditRoleChange(ctx, assignedBy, userID, tenantID, models.ActionRoleAssigned, roleSlug, reason)

	// Publish cache invalidation event
	s.publishInvalidationEvent(ctx, userID, tenantID, models.ActionRoleAssigned, roleSlug)

	logger.WithContext(ctx).Infof(
		"Role assigned: user=%d role=%s tenant=%s by=%d",
		userID, roleSlug, tenantID, assignedBy,
	)

	return nil
}

// RevokeRole revokes a role from a user in a specific tenant
func (s *UserRoleService) RevokeRole(
	ctx context.Context,
	userID int64,
	roleSlug string,
	tenantID string,
	revokedBy int64,
	reason string,
) error {
	// Get role by slug
	role, err := s.roleRepo.FindBySlug(ctx, roleSlug, nil)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Delete user role assignment
	if err := s.userRoleRepo.Delete(ctx, userID, role.ID, tenantID); err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}

	// Remove Casbin grouping policy
	if err := s.enforcer.RevokeRoleFromUser(fmt.Sprintf("%d", userID), roleSlug, tenantID); err != nil {
		return fmt.Errorf("failed to remove Casbin policy: %w", err)
	}

	// Audit the revocation
	s.auditRoleChange(ctx, revokedBy, userID, tenantID, models.ActionRoleRevoked, roleSlug, reason)

	// Publish cache invalidation event
	s.publishInvalidationEvent(ctx, userID, tenantID, models.ActionRoleRevoked, roleSlug)

	logger.WithContext(ctx).Infof(
		"Role revoked: user=%d role=%s tenant=%s by=%d",
		userID, roleSlug, tenantID, revokedBy,
	)

	return nil
}

// GetUserRoles retrieves all roles for a user in a tenant
func (s *UserRoleService) GetUserRoles(ctx context.Context, userID int64, tenantID string) ([]models.UserRole, error) {
	return s.userRoleRepo.FindByUserAndTenant(ctx, userID, tenantID)
}

// GetActiveUserRoles retrieves active (non-expired) roles for a user
func (s *UserRoleService) GetActiveUserRoles(ctx context.Context, userID int64, tenantID string) ([]models.UserRole, error) {
	return s.userRoleRepo.FindActiveRoles(ctx, userID, tenantID)
}

// GetUsersWithRole retrieves all users that have a specific role
func (s *UserRoleService) GetUsersWithRole(ctx context.Context, roleID int64, tenantID string) ([]models.UserRole, error) {
	return s.userRoleRepo.FindByRole(ctx, roleID, tenantID)
}

// auditRoleChange creates an audit log for role assignments/revocations
func (s *UserRoleService) auditRoleChange(
	ctx context.Context,
	actorID int64,
	subjectID int64,
	tenantID string,
	action string,
	role string,
	reason string,
) {
	subjectIDStr := fmt.Sprintf("%d", subjectID)
	log := &models.AuditLog{
		EventID:      fmt.Sprintf("role_%d_%d", actorID, time.Now().UnixNano()),
		Timestamp:    time.Now(),
		ActorID:      fmt.Sprintf("%d", actorID),
		ActorType:    models.ActorTypeUser,
		Action:       action,
		ResourceType: strPtr("role"),
		SubjectID:    &subjectIDStr,
		TenantID:     tenantID,
		PolicyAfter: map[string]interface{}{
			"role": role,
		},
		Reason: &reason,
	}

	// Save audit log (async)
	go func() {
		if err := s.auditRepo.Create(context.Background(), log); err != nil {
			logger.Errorf("Failed to save audit log: %v", err)
		}
	}()
}

// publishInvalidationEvent publishes a cache invalidation event for user role changes
func (s *UserRoleService) publishInvalidationEvent(
	ctx context.Context,
	userID int64,
	tenantID string,
	action string,
	role string,
) {
	if s.publisher == nil {
		return
	}

	event := &watcher.RBACEvent{
		EventID:   fmt.Sprintf("rbac_%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Action:    action,
		TenantID:  tenantID,
		SubjectID: fmt.Sprintf("%d", userID),
		Details: watcher.EventDetails{
			Role: role,
		},
	}

	// Publish event (async)
	go func() {
		if err := watcher.PublishEvent(context.Background(), s.publisher, event); err != nil {
			logger.Errorf("Failed to publish invalidation event: %v", err)
		}
	}()
}
