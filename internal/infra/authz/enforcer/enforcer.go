package enforcer

import (
	"context"
	"errors"
	"fmt"
	"ichi-go/internal/infra/authz/adapter"
	"ichi-go/pkg/logger"
	"ichi-go/pkg/rbac"
	"sync"
	"time"

	"github.com/uptrace/bun"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
)

// Enforcer wraps Casbin enforcer with additional functionality
type Enforcer struct {
	enforcer      *casbin.Enforcer
	adapter       *adapter.BunAdapter
	config        *rbac.Config
	mu            sync.RWMutex
	lastReload    time.Time
	isFiltered    bool
	currentTenant string
}

var (
	// ErrEnforcerNotInitialized is returned when enforcer is not initialized
	ErrEnforcerNotInitialized = errors.New("enforcer not initialized")

	// ErrInvalidPermissionCheck is returned when permission check parameters are invalid
	ErrInvalidPermissionCheck = errors.New("invalid permission check parameters")

	// ErrPolicyNotFound is returned when a policy doesn't exist
	ErrPolicyNotFound = errors.New("policy not found")
)

// New creates a new Enforcer instance
func New(db *bun.DB, config *rbac.Config) (*Enforcer, error) {
	if db == nil {
		return nil, errors.New("database connection is required")
	}
	if config == nil {
		return nil, errors.New("RBAC config is required")
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid RBAC config: %w", err)
	}

	// Create Bun adapter
	bunAdapter, err := adapter.NewBunAdapter(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	// Load Casbin model
	m, err := model.NewModelFromFile(config.ModelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load Casbin model: %w", err)
	}

	// Create Casbin enforcer
	casbinEnforcer, err := casbin.NewEnforcer(m, bunAdapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}

	// Enable auto-save (persist changes to database)
	casbinEnforcer.EnableAutoSave(true)

	// Enable logging
	//casbinEnforcer.EnableLog(true)

	e := &Enforcer{
		enforcer:   casbinEnforcer,
		adapter:    bunAdapter,
		config:     config,
		lastReload: time.Now(),
	}

	// Load policies based on strategy
	if err := e.loadPolicies(); err != nil {
		return nil, fmt.Errorf("failed to load policies: %w", err)
	}

	logger.Infof("RBAC Enforcer initialized successfully")
	return e, nil
}

// CheckPermission checks if a user has permission to perform an action on a resource
// Returns (allowed bool, error)
func (e *Enforcer) CheckPermission(ctx context.Context, userID, tenantID, resource, action string) (bool, error) {
	if userID == "" || tenantID == "" || resource == "" || action == "" {
		return false, ErrInvalidPermissionCheck
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	// Format subject as "user:{id}"
	subject := fmt.Sprintf("user:%s", userID)

	// Enforce: (subject, domain, object, action)
	allowed, err := e.enforcer.Enforce(subject, tenantID, resource, action)
	if err != nil {
		logger.WithContext(ctx).Errorf(
			"Permission check failed: user=%s tenant=%s resource=%s action=%s error=%v",
			userID, tenantID, resource, action, err)
		return false, fmt.Errorf("permission check failed: %w", err)
	}

	logger.WithContext(ctx).Debugf(
		"Permission check: user=%s tenant=%s resource=%s action=%s allowed=%v",
		userID, tenantID, resource, action, allowed)

	return allowed, nil
}

// CheckBatch checks multiple permissions in a single call
func (e *Enforcer) CheckBatch(ctx context.Context, userID, tenantID string, checks []PermissionCheck) (map[string]bool, error) {
	if userID == "" || tenantID == "" {
		return nil, ErrInvalidPermissionCheck
	}

	results := make(map[string]bool, len(checks))

	for _, check := range checks {
		key := fmt.Sprintf("%s:%s", check.Resource, check.Action)
		allowed, err := e.CheckPermission(ctx, userID, tenantID, check.Resource, check.Action)
		if err != nil {
			return nil, err
		}
		results[key] = allowed
	}

	return results, nil
}

// PermissionCheck represents a single permission check request
type PermissionCheck struct {
	Resource string
	Action   string
}

// AddPolicy adds a new policy rule
func (e *Enforcer) AddPolicy(role, tenantID, resource, action string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	added, err := e.enforcer.AddPolicy(role, tenantID, resource, action)
	if err != nil {
		return fmt.Errorf("failed to add policy: %w", err)
	}

	if !added {
		return errors.New("policy already exists")
	}

	logger.Infof("Added policy: role=%s tenant=%s resource=%s action=%s",
		role, tenantID, resource, action)

	return nil
}

// RemovePolicy removes a policy rule
func (e *Enforcer) RemovePolicy(role, tenantID, resource, action string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	removed, err := e.enforcer.RemovePolicy(role, tenantID, resource, action)
	if err != nil {
		return fmt.Errorf("failed to remove policy: %w", err)
	}

	if !removed {
		return ErrPolicyNotFound
	}

	logger.Infof("Removed policy: role=%s tenant=%s resource=%s action=%s",
		role, tenantID, resource, action)

	return nil
}

// AssignRoleToUser assigns a role to a user in a specific tenant
func (e *Enforcer) AssignRoleToUser(userID, role, tenantID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	subject := fmt.Sprintf("user:%s", userID)

	added, err := e.enforcer.AddGroupingPolicy(subject, role, tenantID)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	if !added {
		return errors.New("role already assigned")
	}

	logger.Infof("Assigned role: user=%s role=%s tenant=%s", userID, role, tenantID)

	return nil
}

// RevokeRoleFromUser removes a role from a user in a specific tenant
func (e *Enforcer) RevokeRoleFromUser(userID, role, tenantID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	subject := fmt.Sprintf("user:%s", userID)

	removed, err := e.enforcer.RemoveGroupingPolicy(subject, role, tenantID)
	if err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}

	if !removed {
		return errors.New("role not assigned")
	}

	logger.Infof("Revoked role: user=%s role=%s tenant=%s", userID, role, tenantID)

	return nil
}

// GetUserRoles returns all roles assigned to a user in a specific tenant
func (e *Enforcer) GetUserRoles(userID, tenantID string) ([]string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	subject := fmt.Sprintf("user:%s", userID)

	// Get all roles for user (returns all tenants)
	allRoles, err := e.enforcer.GetImplicitRolesForUser(subject)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Filter by tenant
	var roles []string
	for _, role := range allRoles {
		// Check if role is assigned in this tenant
		hasRole, _ := e.enforcer.HasGroupingPolicy(subject, role, tenantID)
		if hasRole {
			roles = append(roles, role)
		}
	}

	return roles, nil
}

// GetRolePermissions returns all permissions for a role in a specific tenant
func (e *Enforcer) GetRolePermissions(role, tenantID string) ([]Permission, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Get filtered policies for this role and tenant
	policies, err := e.enforcer.GetFilteredPolicy(0, role, tenantID)
	if err != nil {
		return nil, err
	}

	permissions := make([]Permission, 0, len(policies))
	for _, policy := range policies {
		if len(policy) >= 4 {
			permissions = append(permissions, Permission{
				Role:     policy[0],
				TenantID: policy[1],
				Resource: policy[2],
				Action:   policy[3],
			})
		}
	}

	return permissions, nil
}

// Permission represents a policy rule
type Permission struct {
	Role     string
	TenantID string
	Resource string
	Action   string
}

// LoadFilteredPolicy loads policies for a specific tenant
func (e *Enforcer) LoadFilteredPolicy(tenantID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	filter := &adapter.Filter{
		TenantID: tenantID,
	}

	if err := e.enforcer.LoadFilteredPolicy(filter); err != nil {
		return fmt.Errorf("failed to load filtered policy: %w", err)
	}

	e.isFiltered = true
	e.currentTenant = tenantID
	e.lastReload = time.Now()

	logger.Infof("Loaded filtered policies for tenant: %s", tenantID)

	return nil
}

// ReloadPolicy reloads all policies from database
func (e *Enforcer) ReloadPolicy() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := e.enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("failed to reload policy: %w", err)
	}

	e.isFiltered = false
	e.currentTenant = ""
	e.lastReload = time.Now()

	logger.Infof("Reloaded all policies from database")

	return nil
}

// loadPolicies loads policies based on configured strategy
func (e *Enforcer) loadPolicies() error {
	strategy := e.config.Performance.LoadingStrategy

	switch strategy {
	case "full":
		return e.enforcer.LoadPolicy()

	case "filtered":
		// For filtered mode in hybrid/multi-tenant, we'll load dynamically per request
		// For now, just load for default tenant
		if e.config.IsSingleTenant() || e.config.IsHybrid() {
			return e.LoadFilteredPolicy(e.config.DefaultTenant)
		}
		return e.enforcer.LoadPolicy()

	case "adaptive":
		// Check policy count
		count, err := e.adapter.CountPolicies()
		if err != nil {
			return fmt.Errorf("failed to count policies: %w", err)
		}

		threshold := e.config.Performance.AdaptiveThreshold
		if count < threshold {
			logger.Infof("Policy count (%d) below threshold (%d), using full loading", count, threshold)
			return e.enforcer.LoadPolicy()
		}

		logger.Infof("Policy count (%d) above threshold (%d), using filtered loading", count, threshold)
		if e.config.IsSingleTenant() || e.config.IsHybrid() {
			return e.LoadFilteredPolicy(e.config.DefaultTenant)
		}
		return e.enforcer.LoadPolicy()

	default:
		return fmt.Errorf("invalid loading strategy: %s", strategy)
	}
}

// GetPolicyCount returns the number of policies currently loaded
func (e *Enforcer) GetPolicyCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	policy, err := e.enforcer.GetPolicy()
	if err != nil {
		return 0
	}
	return len(policy)
}

// IsFiltered returns true if enforcer is using filtered policy loading
func (e *Enforcer) IsFiltered() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.isFiltered
}

// GetLastReloadTime returns the last time policies were reloaded
func (e *Enforcer) GetLastReloadTime() time.Time {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.lastReload
}

// ClearPolicy removes all policies from enforcer (not from database)
func (e *Enforcer) ClearPolicy() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.enforcer.ClearPolicy()
	logger.Warnf("Cleared all policies from enforcer memory")

	return nil
}
