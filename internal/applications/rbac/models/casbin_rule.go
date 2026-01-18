package models

import (
	"time"

	"github.com/uptrace/bun"
)

// CasbinRule represents a Casbin policy or grouping rule
type CasbinRule struct {
	bun.BaseModel `bun:"table:casbin_rule,alias:cr"`

	ID    int64  `bun:"id,pk,autoincrement" json:"id"`
	Ptype string `bun:"ptype,notnull" json:"ptype"` // p (policy) or g (grouping)
	V0    string `bun:"v0" json:"v0"`               // subject (user or role)
	V1    string `bun:"v1" json:"v1"`               // domain (tenant_id)
	V2    string `bun:"v2" json:"v2"`               // object (resource)
	V3    string `bun:"v3" json:"v3"`               // action
	V4    string `bun:"v4" json:"v4"`               // reserved
	V5    string `bun:"v5" json:"v5"`               // reserved

	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
}

// TableName returns the table name
func (CasbinRule) TableName() string {
	return "casbin_rule"
}

// IsPolicy returns true if this is a policy rule (not grouping)
func (c *CasbinRule) IsPolicy() bool {
	return c.Ptype == "p"
}

// IsGrouping returns true if this is a grouping rule
func (c *CasbinRule) IsGrouping() bool {
	return c.Ptype == "g" || c.Ptype == "g2"
}

// GetSubject returns the subject (v0)
func (c *CasbinRule) GetSubject() string {
	return c.V0
}

// GetDomain returns the domain/tenant (v1)
func (c *CasbinRule) GetDomain() string {
	return c.V1
}

// GetObject returns the object/resource (v2)
func (c *CasbinRule) GetObject() string {
	return c.V2
}

// GetAction returns the action (v3)
func (c *CasbinRule) GetAction() string {
	return c.V3
}
