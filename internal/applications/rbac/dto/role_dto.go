package dto

import "time"

// CreateRoleRequest represents a request to create a role
type CreateRoleRequest struct {
	Name        string  `json:"name" validate:"required,min=3,max=100"`
	Slug        string  `json:"slug" validate:"required,min=2,max=50,slug"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	TenantID    *string `json:"tenant_id,omitempty"`
	Permissions []int64 `json:"permissions,omitempty" validate:"omitempty,dive,min=1"`
}

// UpdateRoleRequest represents a request to update a role
type UpdateRoleRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Permissions []int64 `json:"permissions,omitempty" validate:"omitempty,dive,min=1"`
}

// RoleResponse represents a role in API responses
type RoleResponse struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Description  *string   `json:"description,omitempty"`
	TenantID     *string   `json:"tenant_id,omitempty"`
	IsSystemRole bool      `json:"is_system_role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RoleWithPermissionsResponse represents a role with its permissions
type RoleWithPermissionsResponse struct {
	RoleResponse
	Permissions []PermissionResponse `json:"permissions"`
}

// PermissionResponse represents a permission in API responses
type PermissionResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Module      string  `json:"module"`
	Resource    string  `json:"resource"`
	Action      string  `json:"action"`
	Description *string `json:"description,omitempty"`
}

// GetRolesRequest represents a request to get roles
type GetRolesRequest struct {
	TenantID   *string `json:"tenant_id,omitempty"`
	GlobalOnly bool    `json:"global_only,omitempty"`
	Page       int     `json:"page" validate:"min=1"`
	PageSize   int     `json:"page_size" validate:"min=1,max=100"`
}

// GetRolesResponse represents the response for role list
type GetRolesResponse struct {
	Roles []RoleResponse `json:"roles"`
	Total int            `json:"total"`
}
