package rbac

import (
	"ichi-go/config"
	"ichi-go/internal/applications/rbac/controllers"
	"ichi-go/internal/applications/rbac/repositories"
	"ichi-go/internal/applications/rbac/services"
	"ichi-go/internal/infra/authz/cache"
	"ichi-go/internal/infra/authz/enforcer"
	"ichi-go/internal/infra/queue/rabbitmq"
	"ichi-go/pkg/logger"

	"github.com/redis/go-redis/v9"

	"github.com/samber/do/v2"
	"github.com/uptrace/bun"
)

// RegisterProviders registers all RBAC domain dependencies
func RegisterProviders(injector do.Injector) {
	// Infrastructure
	do.Provide(injector, ProvideDecisionCache)

	// Repositories
	do.Provide(injector, ProvidePolicyRepository)
	do.Provide(injector, ProvideRoleRepository)
	do.Provide(injector, ProvidePermissionRepository)
	do.Provide(injector, ProvideUserRoleRepository)
	do.Provide(injector, ProvideAuditRepository)
	do.Provide(injector, ProvidePlatformRepository)

	// Services
	do.Provide(injector, ProvideEnforcementService)
	do.Provide(injector, ProvidePolicyService)
	do.Provide(injector, ProvideRoleService)
	do.Provide(injector, ProvideUserRoleService)
	do.Provide(injector, ProvideAuditService)

	// Controllers
	do.Provide(injector, ProvideEnforcementController)
	do.Provide(injector, ProvidePolicyController)
	do.Provide(injector, ProvideRoleController)
	do.Provide(injector, ProvideUserRoleController)
	do.Provide(injector, ProvideAuditController)
}

// Repository Providers

func ProvidePolicyRepository(i do.Injector) (*repositories.PolicyRepository, error) {
	db := do.MustInvoke[*bun.DB](i)
	return repositories.NewPolicyRepository(db), nil
}

func ProvideRoleRepository(i do.Injector) (*repositories.RoleRepository, error) {
	db := do.MustInvoke[*bun.DB](i)
	return repositories.NewRoleRepository(db), nil
}

func ProvidePermissionRepository(i do.Injector) (*repositories.PermissionRepository, error) {
	db := do.MustInvoke[*bun.DB](i)
	return repositories.NewPermissionRepository(db), nil
}

func ProvideUserRoleRepository(i do.Injector) (*repositories.UserRoleRepository, error) {
	db := do.MustInvoke[*bun.DB](i)
	return repositories.NewUserRoleRepository(db), nil
}

func ProvideAuditRepository(i do.Injector) (*repositories.AuditRepository, error) {
	db := do.MustInvoke[*bun.DB](i)
	return repositories.NewAuditRepository(db), nil
}

func ProvidePlatformRepository(i do.Injector) (*repositories.PlatformRepository, error) {
	db := do.MustInvoke[*bun.DB](i)
	return repositories.NewPlatformRepository(db), nil
}

func ProvideDecisionCache(i do.Injector) (*cache.DecisionCache, error) {
	cfg := do.MustInvoke[*config.Config](i)
	cacheClient := do.MustInvoke[*redis.Client](i)
	rbacRedis, err := cache.NewRedisCache(cacheClient, true)
	if err != nil {
		return nil, err
	}
	return cache.NewDecisionCache(rbacRedis, &cfg.RBAC().Cache)
}

// Service Providers

func ProvideEnforcementService(i do.Injector) (*services.EnforcementService, error) {
	cfg := do.MustInvoke[*config.Config](i)
	enf := do.MustInvoke[*enforcer.Enforcer](i)
	platformRepo := do.MustInvoke[*repositories.PlatformRepository](i)
	auditRepo := do.MustInvoke[*repositories.AuditRepository](i)
	decisionCache := do.MustInvoke[*cache.DecisionCache](i)

	return services.NewEnforcementService(enf, decisionCache, platformRepo, auditRepo, cfg.RBAC()), nil
}

func ProvidePolicyService(i do.Injector) (*services.PolicyService, error) {
	policyRepo := do.MustInvoke[*repositories.PolicyRepository](i)
	auditRepo := do.MustInvoke[*repositories.AuditRepository](i)
	enf := do.MustInvoke[*enforcer.Enforcer](i)

	// Queue producer is optional
	var producer rabbitmq.MessageProducer
	if p, err := do.Invoke[rabbitmq.MessageProducer](i); err == nil {
		producer = p
		if producer != nil {
			logger.Infof("✅ Policy service using message producer")
		} else {
			logger.Warnf("⚠️  Message producer is nil")
		}
	} else {
		logger.Warnf("⚠️  Queue not available for policy service: %v", err)
	}

	return services.NewPolicyService(enf, policyRepo, auditRepo, producer), nil
}

func ProvideRoleService(i do.Injector) (*services.RoleService, error) {
	roleRepo := do.MustInvoke[*repositories.RoleRepository](i)
	permissionRepo := do.MustInvoke[*repositories.PermissionRepository](i)

	return services.NewRoleService(roleRepo, permissionRepo), nil
}

func ProvideUserRoleService(i do.Injector) (*services.UserRoleService, error) {
	userRoleRepo := do.MustInvoke[*repositories.UserRoleRepository](i)
	roleRepo := do.MustInvoke[*repositories.RoleRepository](i)
	auditRepo := do.MustInvoke[*repositories.AuditRepository](i)
	enforcer := do.MustInvoke[*enforcer.Enforcer](i)

	// Queue producer is optional
	var producer rabbitmq.MessageProducer
	if p, err := do.Invoke[rabbitmq.MessageProducer](i); err == nil {
		producer = p
		if producer != nil {
			logger.Infof("✅ User role service using message producer")
		} else {
			logger.Warnf("⚠️  Message producer is nil")
		}
	} else {
		logger.Warnf("⚠️  Queue not available for user role service: %v", err)
	}

	return services.NewUserRoleService(userRoleRepo, roleRepo, auditRepo, enforcer, producer), nil
}

func ProvideAuditService(i do.Injector) (*services.AuditService, error) {
	auditRepo := do.MustInvoke[*repositories.AuditRepository](i)
	return services.NewAuditService(auditRepo), nil
}

// Controller Providers

func ProvideEnforcementController(i do.Injector) (*controllers.EnforcementController, error) {
	svc := do.MustInvoke[*services.EnforcementService](i)
	return controllers.NewEnforcementController(svc), nil
}

func ProvidePolicyController(i do.Injector) (*controllers.PolicyController, error) {
	svc := do.MustInvoke[*services.PolicyService](i)
	return controllers.NewPolicyController(svc), nil
}

func ProvideRoleController(i do.Injector) (*controllers.RoleController, error) {
	svc := do.MustInvoke[*services.RoleService](i)
	return controllers.NewRoleController(svc), nil
}

func ProvideUserRoleController(i do.Injector) (*controllers.UserRoleController, error) {
	svc := do.MustInvoke[*services.UserRoleService](i)
	return controllers.NewUserRoleController(svc), nil
}

func ProvideAuditController(i do.Injector) (*controllers.AuditController, error) {
	svc := do.MustInvoke[*services.AuditService](i)
	return controllers.NewAuditController(svc), nil
}

// GetDB is a helper to get the Bun DB instance
func GetDB(i do.Injector) *bun.DB {
	db := do.MustInvoke[*bun.DB](i)
	return db
}
