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

// PolicyService handles Casbin policy management with audit trails
type PolicyService struct {
	enforcer   *enforcer.Enforcer
	policyRepo *repositories.PolicyRepository
	auditRepo  *repositories.AuditRepository
	publisher  rabbitmq.MessageProducer
}

// NewPolicyService creates a new policy service
func NewPolicyService(
	enforcer *enforcer.Enforcer,
	policyRepo *repositories.PolicyRepository,
	auditRepo *repositories.AuditRepository,
	publisher rabbitmq.MessageProducer,
) *PolicyService {
	return &PolicyService{
		enforcer:   enforcer,
		policyRepo: policyRepo,
		auditRepo:  auditRepo,
		publisher:  publisher,
	}
}

// AddPolicy adds a new policy rule with audit trail
func (s *PolicyService) AddPolicy(
	ctx context.Context,
	role string,
	tenantID string,
	resource string,
	action string,
	actorID int64,
	reason string,
) error {
	// Add policy via enforcer
	if err := s.enforcer.AddPolicy(role, tenantID, resource, action); err != nil {
		return fmt.Errorf("failed to add policy: %w", err)
	}

	// Audit the change
	s.auditMutation(ctx, actorID, tenantID, models.ActionPolicyAdded, reason, map[string]interface{}{
		"role":     role,
		"resource": resource,
		"action":   action,
	})

	// Publish cache invalidation event
	s.publishInvalidationEvent(ctx, tenantID, models.ActionPolicyAdded, role)

	logger.WithContext(ctx).Infof(
		"Policy added: role=%s tenant=%s resource=%s action=%s by user=%d",
		role, tenantID, resource, action, actorID,
	)

	return nil
}

// RemovePolicy removes a policy rule with audit trail
func (s *PolicyService) RemovePolicy(
	ctx context.Context,
	role string,
	tenantID string,
	resource string,
	action string,
	actorID int64,
	reason string,
) error {
	// Remove policy via enforcer
	if err := s.enforcer.RemovePolicy(role, tenantID, resource, action); err != nil {
		return fmt.Errorf("failed to remove policy: %w", err)
	}

	// Audit the change
	s.auditMutation(ctx, actorID, tenantID, models.ActionPolicyRemoved, reason, map[string]interface{}{
		"role":     role,
		"resource": resource,
		"action":   action,
	})

	// Publish cache invalidation event
	s.publishInvalidationEvent(ctx, tenantID, models.ActionPolicyRemoved, role)

	logger.WithContext(ctx).Infof(
		"Policy removed: role=%s tenant=%s resource=%s action=%s by user=%d",
		role, tenantID, resource, action, actorID,
	)

	return nil
}

// GetPoliciesByTenant retrieves all policies for a tenant
func (s *PolicyService) GetPoliciesByTenant(ctx context.Context, tenantID string) ([]models.CasbinRule, error) {
	return s.policyRepo.FindByTenant(ctx, tenantID)
}

// GetPoliciesByRole retrieves all policies for a specific role
func (s *PolicyService) GetPoliciesByRole(ctx context.Context, role string, tenantID string) ([]models.CasbinRule, error) {
	return s.policyRepo.FindBySubject(ctx, role, tenantID)
}

// CountPolicies counts all policies in the system
func (s *PolicyService) CountPolicies(ctx context.Context) (int, error) {
	return s.policyRepo.CountAll(ctx)
}

// CountPoliciesByTenant counts policies for a specific tenant
func (s *PolicyService) CountPoliciesByTenant(ctx context.Context, tenantID string) (int, error) {
	return s.policyRepo.CountByTenant(ctx, tenantID)
}

// ReloadPolicies reloads all policies from database
func (s *PolicyService) ReloadPolicies(ctx context.Context) error {
	return s.enforcer.ReloadPolicy()
}

// auditMutation creates an audit log for policy mutations
func (s *PolicyService) auditMutation(
	ctx context.Context,
	actorID int64,
	tenantID string,
	action string,
	reason string,
	policyDetails map[string]interface{},
) {
	log := &models.AuditLog{
		EventID:      fmt.Sprintf("policy_%d_%d", actorID, time.Now().UnixNano()),
		Timestamp:    time.Now(),
		ActorID:      fmt.Sprintf("%d", actorID),
		ActorType:    models.ActorTypeUser,
		Action:       action,
		ResourceType: strPtr("policy"),
		TenantID:     tenantID,
		PolicyAfter:  policyDetails,
		Reason:       &reason,
	}

	// Save audit log (async)
	go func() {
		if err := s.auditRepo.Create(context.Background(), log); err != nil {
			logger.Errorf("Failed to save audit log: %v", err)
		}
	}()
}

// publishInvalidationEvent publishes a cache invalidation event to RabbitMQ
func (s *PolicyService) publishInvalidationEvent(
	ctx context.Context,
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
		Details: watcher.EventDetails{
			Role:         role,
			ReloadPolicy: true,
		},
	}

	// Publish event (async)
	go func() {
		if err := watcher.PublishEvent(context.Background(), s.publisher, event); err != nil {
			logger.Errorf("Failed to publish invalidation event: %v", err)
		}
	}()
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}
