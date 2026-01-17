package rbac

import (
	"ichi-go/internal/applications/rbac/controllers"
	"ichi-go/pkg/authenticator"

	"github.com/labstack/echo/v4"
	"github.com/samber/do/v2"
)

// Register registers all RBAC domain components
func Register(injector do.Injector, serviceName string, e *echo.Echo, auth *authenticator.Authenticator) {
	// Register all RBAC domain providers
	RegisterProviders(injector)

	// Get controllers from DI container
	enforcementCtrl := do.MustInvoke[*controllers.EnforcementController](injector)
	policyCtrl := do.MustInvoke[*controllers.PolicyController](injector)
	roleCtrl := do.MustInvoke[*controllers.RoleController](injector)
	userRoleCtrl := do.MustInvoke[*controllers.UserRoleController](injector)
	auditCtrl := do.MustInvoke[*controllers.AuditController](injector)

	// Register routes
	RegisterRoutes(serviceName, e, auth, enforcementCtrl, policyCtrl, roleCtrl, userRoleCtrl, auditCtrl)
}

// RegisterRoutes registers all RBAC routes
func RegisterRoutes(
	serviceName string,
	e *echo.Echo,
	auth *authenticator.Authenticator,
	enforcementCtrl *controllers.EnforcementController,
	policyCtrl *controllers.PolicyController,
	roleCtrl *controllers.RoleController,
	userRoleCtrl *controllers.UserRoleController,
	auditCtrl *controllers.AuditController,
) {
	// Base path for RBAC API
	basePath := serviceName + "/api/v1/rbac"

	// Enforcement routes (permission checking)
	enforcement := e.Group(basePath + "/enforce")
	enforcement.Use(auth.AuthenticateMiddleware()) // Require authentication
	{
		enforcement.POST("/check", enforcementCtrl.CheckPermission)
		enforcement.POST("/batch", enforcementCtrl.CheckBatchPermissions)
		enforcement.GET("/my-permissions", enforcementCtrl.GetMyPermissions)
	}

	// Policy routes (policy management)
	policies := e.Group(basePath + "/policies")
	policies.Use(auth.AuthenticateMiddleware()) // Require authentication
	{
		policies.GET("", policyCtrl.GetPolicies)
		policies.POST("", policyCtrl.AddPolicy)
		policies.DELETE("", policyCtrl.RemovePolicy)
		policies.GET("/count", policyCtrl.GetPolicyCount)
		policies.POST("/reload", policyCtrl.ReloadPolicies)
	}

	// Role routes (role management)
	roles := e.Group(basePath + "/roles")
	roles.Use(auth.AuthenticateMiddleware()) // Require authentication
	{
		roles.GET("", roleCtrl.GetRoles)
		roles.POST("", roleCtrl.CreateRole)
		roles.GET("/:id", roleCtrl.GetRole)
		roles.PUT("/:id", roleCtrl.UpdateRole)
		roles.DELETE("/:id", roleCtrl.DeleteRole)
		roles.GET("/:id/permissions", roleCtrl.GetRoleWithPermissions)

		// Users with role (nested route)
		roles.GET("/:roleId/users", userRoleCtrl.GetUsersWithRole)
	}

	// User role routes (user-role assignments)
	userRoles := e.Group(basePath + "/users")
	userRoles.Use(auth.AuthenticateMiddleware()) // Require authentication
	{
		userRoles.GET("/:userId/roles", userRoleCtrl.GetUserRoles)
		userRoles.GET("/:userId/roles/active", userRoleCtrl.GetActiveUserRoles)
		userRoles.POST("/:userId/roles", userRoleCtrl.AssignRole)
		userRoles.DELETE("/:userId/roles/:roleSlug", userRoleCtrl.RevokeRole)
	}

	// Audit routes (audit log querying)
	audit := e.Group(basePath + "/audit")
	audit.Use(auth.AuthenticateMiddleware()) // Require authentication
	{
		audit.GET("/logs", auditCtrl.QueryAuditLogs)
		audit.GET("/stats", auditCtrl.GetAuditStats)
		audit.POST("/export", auditCtrl.ExportAuditLogs)
		audit.GET("/mutations", auditCtrl.GetRecentMutations)
		audit.GET("/decisions", auditCtrl.GetRecentDecisions)
	}
}
