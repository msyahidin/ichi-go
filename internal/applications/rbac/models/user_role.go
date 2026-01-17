package models

import (
	"time"

	"github.com/uptrace/bun"
)

// UserRole represents a user-role assignment in a specific tenant
type UserRole struct {
	bun.BaseModel `bun:"table:rbac_user_roles,alias:rur"`

	ID       int64  `bun:"id,pk,autoincrement" json:"id"`
	UserID   int64  `bun:"user_id,notnull" json:"user_id"`
	RoleID   int64  `bun:"role_id,notnull" json:"role_id"`
	TenantID string `bun:"tenant_id,notnull" json:"tenant_id"`

	AssignedAt *time.Time `bun:"assigned_at,nullzero,default:current_timestamp" json:"assigned_at"`
	AssignedBy *int64     `bun:"assigned_by" json:"assigned_by,omitempty"`
	ExpiresAt  *time.Time `bun:"expires_at" json:"expires_at,omitempty"`

	// Relations
	Role *Role `bun:"rel:belongs-to,join:role_id=id" json:"role,omitempty"`
}

// TableName returns the table name
func (UserRole) TableName() string {
	return "rbac_user_roles"
}

// IsExpired returns true if the role assignment has expired
func (ur *UserRole) IsExpired() bool {
	if ur.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*ur.ExpiresAt)
}

// IsActive returns true if the role assignment is active (not expired)
func (ur *UserRole) IsActive() bool {
	return !ur.IsExpired()
}

// GetRoleSlug returns the role slug (if role is loaded)
func (ur *UserRole) GetRoleSlug() string {
	if ur.Role != nil {
		return ur.Role.Slug
	}
	return ""
}
