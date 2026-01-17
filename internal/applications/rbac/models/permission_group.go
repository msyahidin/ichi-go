package models

import (
	"time"

	"github.com/uptrace/bun"
)

// PermissionGroup represents a group of permissions for UI organization
type PermissionGroup struct {
	bun.BaseModel `bun:"table:rbac_permission_groups,alias:rpg"`

	ID          int64   `bun:"id,pk,autoincrement" json:"id"`
	Name        string  `bun:"name,notnull" json:"name"`
	Slug        string  `bun:"slug,notnull,unique" json:"slug"`
	Description *string `bun:"description" json:"description,omitempty"`
	SortOrder   int     `bun:"sort_order,notnull,default:0" json:"sort_order"`

	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`

	// Relations
	Permissions []*Permission `bun:"m2m:rbac_permission_group_items,join:PermissionGroup=Permission" json:"permissions,omitempty"`
}

// TableName returns the table name
func (PermissionGroup) TableName() string {
	return "rbac_permission_groups"
}
