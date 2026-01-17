package models

import (
	"time"

	"github.com/uptrace/bun"
)

// Permission represents a permission definition
type Permission struct {
	bun.BaseModel `bun:"table:rbac_permissions,alias:rp"`

	ID          int64   `bun:"id,pk,autoincrement" json:"id"`
	Name        string  `bun:"name,notnull" json:"name"`
	Slug        string  `bun:"slug,notnull,unique" json:"slug"`
	Description *string `bun:"description" json:"description,omitempty"`
	Module      *string `bun:"module" json:"module,omitempty"`
	Resource    *string `bun:"resource" json:"resource,omitempty"`
	Action      *string `bun:"action" json:"action,omitempty"`

	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`

	// Relations
	Roles  []*Role            `bun:"m2m:rbac_role_permissions,join:Permission=Role" json:"roles,omitempty"`
	Groups []*PermissionGroup `bun:"m2m:rbac_permission_group_items,join:Permission=PermissionGroup" json:"groups,omitempty"`
}

// TableName returns the table name
func (Permission) TableName() string {
	return "rbac_permissions"
}

// GetFullName returns a human-readable full name
func (p *Permission) GetFullName() string {
	if p.Module != nil && *p.Module != "" {
		return *p.Module + "." + p.Slug
	}
	return p.Slug
}

// GetResourceAction returns the resource and action as a tuple
func (p *Permission) GetResourceAction() (string, string) {
	resource := ""
	action := ""

	if p.Resource != nil {
		resource = *p.Resource
	}
	if p.Action != nil {
		action = *p.Action
	}

	return resource, action
}
