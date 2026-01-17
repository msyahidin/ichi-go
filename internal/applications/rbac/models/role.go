package models

import (
	"time"

	"github.com/uptrace/bun"
)

// Role represents an application role
type Role struct {
	bun.BaseModel `bun:"table:rbac_roles,alias:rr"`

	ID          int64   `bun:"id,pk,autoincrement" json:"id"`
	Name        string  `bun:"name,notnull" json:"name"`
	Slug        string  `bun:"slug,notnull" json:"slug"`
	Description *string `bun:"description" json:"description,omitempty"`
	TenantID    *string `bun:"tenant_id" json:"tenant_id,omitempty"` // NULL for global roles
	Level       int     `bun:"level,notnull,default:0" json:"level"`
	IsSystemRole bool   `bun:"is_system_role,notnull,default:false" json:"is_system_role"`

	CreatedAt time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt *time.Time `bun:"deleted_at,soft_delete" json:"deleted_at,omitempty"`

	// Relations
	Permissions []*Permission `bun:"m2m:rbac_role_permissions,join:Role=Permission" json:"permissions,omitempty"`
}

// TableName returns the table name
func (Role) TableName() string {
	return "rbac_roles"
}

// IsGlobal returns true if this is a global role (tenant_id is NULL)
func (r *Role) IsGlobal() bool {
	return r.TenantID == nil
}

// IsTenantSpecific returns true if this role is specific to a tenant
func (r *Role) IsTenantSpecific() bool {
	return r.TenantID != nil
}

// CanDelete returns true if the role can be deleted
func (r *Role) CanDelete() bool {
	return !r.IsSystemRole
}

// CanEdit returns true if the role can be edited
func (r *Role) CanEdit() bool {
	// System roles can be edited but not deleted
	return true
}
