package dto

import "time"

// AssignRoleRequest represents a request to assign a role to a user
type AssignRoleRequest struct {
	UserID    int64      `json:"user_id" validate:"required,min=1"`
	RoleSlug  string     `json:"role_slug" validate:"required"`
	TenantID  string     `json:"tenant_id" validate:"required"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Reason    string     `json:"reason,omitempty" validate:"omitempty,max=500"`
}

// RevokeRoleRequest represents a request to revoke a role from a user
type RevokeRoleRequest struct {
	UserID   int64  `json:"user_id" validate:"required,min=1"`
	RoleSlug string `json:"role_slug" validate:"required"`
	TenantID string `json:"tenant_id" validate:"required"`
	Reason   string `json:"reason,omitempty" validate:"omitempty,max=500"`
}

// UserRoleResponse represents a user-role assignment
type UserRoleResponse struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"user_id"`
	RoleID     int64      `json:"role_id"`
	RoleName   string     `json:"role_name"`
	RoleSlug   string     `json:"role_slug"`
	TenantID   string     `json:"tenant_id"`
	AssignedBy *int64     `json:"assigned_by,omitempty"`
	AssignedAt time.Time  `json:"assigned_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	IsActive   bool       `json:"is_active"`
}

// GetUserRolesRequest represents a request to get user roles
type GetUserRolesRequest struct {
	UserID   int64  `json:"user_id" validate:"required,min=1"`
	TenantID string `json:"tenant_id" validate:"required"`
}

// GetUserRolesResponse represents the response for user roles
type GetUserRolesResponse struct {
	Roles []UserRoleResponse `json:"roles"`
	Total int                `json:"total"`
}

// GetUsersWithRoleRequest represents a request to get users with a role
type GetUsersWithRoleRequest struct {
	RoleID   int64  `json:"role_id" validate:"required,min=1"`
	TenantID string `json:"tenant_id" validate:"required"`
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
}

// GetUsersWithRoleResponse represents the response for users with role
type GetUsersWithRoleResponse struct {
	Users []UserRoleResponse `json:"users"`
	Total int                `json:"total"`
}
