package services

import (
	"context"
	"fmt"
	"time"

	"ichi-go/internal/applications/rbac/models"
	"ichi-go/internal/applications/rbac/repositories"
	"ichi-go/internal/infra/authz/cache"
	"ichi-go/internal/infra/authz/enforcer"
	"ichi-go/pkg/logger"
	"ichi-go/pkg/rbac"
)

// EnforcementService provides the main RBAC permission checking API
// This is the CRITICAL service that applications use to check permissions
type EnforcementService struct {
	enforcer      *enforcer.Enforcer
	decisionCache *cache.DecisionCache
	platformRepo  *repositories.PlatformRepository
	auditRepo     *repositories.AuditRepository
	config        *rbac.Config
}

// PermissionCheck represents a single permission check request
type PermissionCheck struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// PermissionResult represents the result of a permission check
type PermissionResult struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// NewEnforcementService creates a new enforcement service
func NewEnforcementService(
	enforcer *enforcer.Enforcer,
	decisionCache *cache.DecisionCache,
	platformRepo *repositories.PlatformRepository,
	auditRepo *repositories.AuditRepository,
	config *rbac.Config,
) *EnforcementService {
	return &EnforcementService{
		enforcer:      enforcer,
		decisionCache: decisionCache,
		platformRepo:  platformRepo,
		auditRepo:     auditRepo,
		config:        config,
	}
}

// CheckPermission is the main permission checking method
// Returns (allowed bool, error)
func (s *EnforcementService) CheckPermission(
	ctx context.Context,
	userID int64,
	tenantID string,
	resource string,
	action string,
) (bool, error) {
	startTime := time.Now()

	// 1. Check platform permissions first (Layer 1)
	isPlatformAdmin, err := s.platformRepo.IsPlatformAdmin(ctx, userID)
	if err != nil {
		logger.WithContext(ctx).Errorf("Failed to check platform admin: %v", err)
	} else if isPlatformAdmin {
		// Platform admins bypass all checks
		s.auditDecision(ctx, userID, tenantID, resource, action, true, "platform_admin", startTime)
		return true, nil
	}

	// 2. Check decision cache (Layer 2)
	cacheKey := cache.MakeCacheKey(tenantID, fmt.Sprintf("%d", userID), resource, action)
	if s.decisionCache != nil && s.config.Cache.Enabled {
		if cached, found, err := s.decisionCache.Get(ctx, cacheKey); err == nil && found {
			s.auditDecision(ctx, userID, tenantID, resource, action, cached, "cache_hit", startTime)
			return cached, nil
		}
	}

	// 3. Check via Casbin enforcer (Layer 3)
	allowed, err := s.enforcer.CheckPermission(
		ctx,
		fmt.Sprintf("%d", userID),
		tenantID,
		resource,
		action,
	)

	if err != nil {
		logger.WithContext(ctx).Errorf(
			"Permission check failed: user=%d tenant=%s resource=%s action=%s error=%v",
			userID, tenantID, resource, action, err,
		)
		return false, fmt.Errorf("permission check failed: %w", err)
	}

	// 4. Cache the result
	if s.decisionCache != nil && s.config.Cache.Enabled {
		if err := s.decisionCache.Set(ctx, cacheKey, allowed); err != nil {
			logger.WithContext(ctx).Errorf("Failed to cache decision: %v", err)
		}
	}

	// 5. Audit the decision
	reason := "enforcer_check"
	if !allowed {
		reason = "permission_denied"
	}
	s.auditDecision(ctx, userID, tenantID, resource, action, allowed, reason, startTime)

	return allowed, nil
}

// CheckBatch checks multiple permissions in a single call (for UI)
func (s *EnforcementService) CheckBatch(
	ctx context.Context,
	userID int64,
	tenantID string,
	checks []PermissionCheck,
) (map[string]bool, error) {
	results := make(map[string]bool, len(checks))

	// Check if platform admin (bypass all checks)
	isPlatformAdmin, _ := s.platformRepo.IsPlatformAdmin(ctx, userID)
	if isPlatformAdmin {
		for _, check := range checks {
			key := fmt.Sprintf("%s:%s", check.Resource, check.Action)
			results[key] = true
		}
		return results, nil
	}

	// Check each permission
	for _, check := range checks {
		key := fmt.Sprintf("%s:%s", check.Resource, check.Action)
		allowed, err := s.CheckPermission(ctx, userID, tenantID, check.Resource, check.Action)
		if err != nil {
			return nil, err
		}
		results[key] = allowed
	}

	return results, nil
}

// GetUserPermissions returns all effective permissions for a user in a tenant
func (s *EnforcementService) GetUserPermissions(
	ctx context.Context,
	userID int64,
	tenantID string,
) ([]string, error) {
	// Check if platform admin
	isPlatformAdmin, _ := s.platformRepo.IsPlatformAdmin(ctx, userID)
	if isPlatformAdmin {
		return []string{"*.*"}, nil // Wildcard permission
	}

	// Get user's roles
	roles, err := s.enforcer.GetUserRoles(fmt.Sprintf("%d", userID), tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Collect all permissions from all roles
	permissionSet := make(map[string]bool)
	for _, role := range roles {
		permissions, err := s.enforcer.GetRolePermissions(role, tenantID)
		if err != nil {
			logger.WithContext(ctx).Errorf("Failed to get role permissions: %v", err)
			continue
		}

		for _, perm := range permissions {
			permKey := fmt.Sprintf("%s.%s", perm.Resource, perm.Action)
			permissionSet[permKey] = true
		}
	}

	// Convert to slice
	permissions := make([]string, 0, len(permissionSet))
	for perm := range permissionSet {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// RequirePermission is a helper that returns error if permission denied
func (s *EnforcementService) RequirePermission(
	ctx context.Context,
	userID int64,
	tenantID string,
	resource string,
	action string,
) error {
	allowed, err := s.CheckPermission(ctx, userID, tenantID, resource, action)
	if err != nil {
		return err
	}

	if !allowed {
		return fmt.Errorf("permission denied: user %d does not have %s.%s in tenant %s",
			userID, resource, action, tenantID)
	}

	return nil
}

// auditDecision logs a permission check decision
func (s *EnforcementService) auditDecision(
	ctx context.Context,
	userID int64,
	tenantID string,
	resource string,
	action string,
	allowed bool,
	reason string,
	startTime time.Time,
) {
	// Skip if decision logging is disabled
	if !s.config.Audit.Enabled || !s.config.Audit.LogDecisions {
		return
	}

	// Sample decisions if sample rate is set
	if s.config.Audit.SampleRate < 1.0 {
		// Simple sampling: skip based on user ID modulo
		if float64(userID%100)/100.0 > s.config.Audit.SampleRate {
			return
		}
	}

	// Calculate latency
	latencyMs := int(time.Since(startTime).Milliseconds())

	// Determine decision and action
	var auditAction, decision string
	if allowed {
		auditAction = models.ActionPermissionChecked
		decision = models.DecisionAllow
	} else {
		auditAction = models.ActionPermissionDenied
		decision = models.DecisionDeny
	}

	// Create audit log
	log := &models.AuditLog{
		EventID:        fmt.Sprintf("decision_%d_%d", userID, time.Now().UnixNano()),
		Timestamp:      time.Now(),
		ActorID:        fmt.Sprintf("%d", userID),
		ActorType:      models.ActorTypeUser,
		Action:         auditAction,
		ResourceType:   &resource,
		TenantID:       tenantID,
		Decision:       &decision,
		DecisionReason: &reason,
		LatencyMs:      &latencyMs,
	}

	// Save audit log (async, don't block on errors)
	go func() {
		if err := s.auditRepo.Create(context.Background(), log); err != nil {
			logger.Errorf("Failed to save audit log: %v", err)
		}
	}()
}

// GetCacheStats returns cache performance statistics
func (s *EnforcementService) GetCacheStats() cache.CacheStats {
	if s.decisionCache == nil {
		return cache.CacheStats{}
	}
	return s.decisionCache.GetStats()
}

// GetCacheHitRatio returns the cache hit ratio (0.0 - 1.0)
func (s *EnforcementService) GetCacheHitRatio() float64 {
	if s.decisionCache == nil {
		return 0.0
	}
	return s.decisionCache.GetHitRatio()
}

// InvalidateUserCache invalidates all cached decisions for a user in a tenant
func (s *EnforcementService) InvalidateUserCache(ctx context.Context, userID int64, tenantID string) error {
	if s.decisionCache == nil {
		return nil
	}

	pattern := cache.MakeUserPattern(tenantID, fmt.Sprintf("%d", userID))
	return s.decisionCache.DeletePattern(ctx, pattern)
}

// InvalidateTenantCache invalidates all cached decisions for a tenant
func (s *EnforcementService) InvalidateTenantCache(ctx context.Context, tenantID string) error {
	if s.decisionCache == nil {
		return nil
	}

	pattern := cache.MakeTenantPattern(tenantID)
	return s.decisionCache.DeletePattern(ctx, pattern)
}
