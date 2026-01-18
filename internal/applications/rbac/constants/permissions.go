package constants

// System role slugs
const (
	RoleSuperAdmin    = "super-admin"
	RolePlatformAdmin = "platform-admin"
	RoleTenantAdmin   = "tenant-admin"
	RoleTenantEditor  = "tenant-editor"
	RoleTenantViewer  = "tenant-viewer"
	RoleDeveloper     = "developer"
	RoleSupport       = "support"
	RoleAuditor       = "auditor"
	RoleGuest         = "guest"
)

// Permission modules
const (
	ModuleUsers       = "users"
	ModuleRoles       = "roles"
	ModulePermissions = "permissions"
	ModuleRBAC        = "rbac"
	ModuleAudit       = "audit"
	ModuleTenants     = "tenants"
	ModuleSystem      = "system"
	ModuleReports     = "reports"
)

// Permission actions
const (
	ActionView   = "view"
	ActionCreate = "create"
	ActionEdit   = "edit"
	ActionDelete = "delete"
	ActionManage = "manage"
	ActionExport = "export"
	ActionImport = "import"
)

// Permission resource patterns
const (
	ResourceWildcard = "*"
	ResourceAll      = "all"
)

// Platform permission slugs (Layer 1 - Global)
const (
	PlatformAdminAccess    = "platform:admin:access"
	PlatformSettingsManage = "platform:settings:manage"
	PlatformUsersManage    = "platform:users:manage"
	PlatformTenantsManage  = "platform:tenants:manage"
	PlatformSystemMonitor  = "platform:system:monitor"
	PlatformAuditView      = "platform:audit:view"
	PlatformReportsView    = "platform:reports:view"
)

// RBAC permission slugs (Layer 2 - Tenant-scoped)
const (
	// User permissions
	UsersView   = "users:view"
	UsersCreate = "users:create"
	UsersEdit   = "users:edit"
	UsersDelete = "users:delete"
	UsersManage = "users:manage"

	// Role permissions
	RolesView   = "roles:view"
	RolesCreate = "roles:create"
	RolesEdit   = "roles:edit"
	RolesDelete = "roles:delete"
	RolesManage = "roles:manage"

	// Permission permissions
	PermissionsView   = "permissions:view"
	PermissionsCreate = "permissions:create"
	PermissionsEdit   = "permissions:edit"
	PermissionsDelete = "permissions:delete"
	PermissionsManage = "permissions:manage"

	// RBAC management permissions
	RBACPoliciesView   = "rbac:policies:view"
	RBACPoliciesManage = "rbac:policies:manage"
	RBACRolesAssign    = "rbac:roles:assign"
	RBACRolesRevoke    = "rbac:roles:revoke"

	// Audit permissions
	AuditLogsView   = "audit:logs:view"
	AuditLogsExport = "audit:logs:export"
	AuditStatsView  = "audit:stats:view"

	// Tenant permissions
	TenantsView   = "tenants:view"
	TenantsCreate = "tenants:create"
	TenantsEdit   = "tenants:edit"
	TenantsDelete = "tenants:delete"
	TenantsManage = "tenants:manage"
)

// Special permission flags
const (
	// Bypass all tenant-level checks (platform admin)
	PermissionBypassTenantChecks = "platform:bypass:tenant-checks"

	// Allow cross-tenant operations
	PermissionCrossTenantAccess = "platform:cross-tenant:access"

	// System maintenance mode
	PermissionSystemMaintenance = "platform:system:maintenance"
)

// Permission groups for UI organization
const (
	GroupUserManagement       = "User Management"
	GroupRoleManagement       = "Role Management"
	GroupPermissionManagement = "Permission Management"
	GroupRBACManagement       = "RBAC Management"
	GroupAuditManagement      = "Audit Management"
	GroupTenantManagement     = "Tenant Management"
	GroupSystemManagement     = "System Management"
)

// IsSystemRole checks if a role slug is a system role
func IsSystemRole(slug string) bool {
	systemRoles := []string{
		RoleSuperAdmin,
		RolePlatformAdmin,
		RoleTenantAdmin,
		RoleTenantEditor,
		RoleTenantViewer,
		RoleDeveloper,
		RoleSupport,
		RoleAuditor,
		RoleGuest,
	}

	for _, sr := range systemRoles {
		if slug == sr {
			return true
		}
	}
	return false
}

// IsPlatformPermission checks if a permission is a platform-level permission
func IsPlatformPermission(permission string) bool {
	platformPermissions := []string{
		PlatformAdminAccess,
		PlatformSettingsManage,
		PlatformUsersManage,
		PlatformTenantsManage,
		PlatformSystemMonitor,
		PlatformAuditView,
		PlatformReportsView,
		PermissionBypassTenantChecks,
		PermissionCrossTenantAccess,
		PermissionSystemMaintenance,
	}

	for _, pp := range platformPermissions {
		if permission == pp {
			return true
		}
	}
	return false
}
