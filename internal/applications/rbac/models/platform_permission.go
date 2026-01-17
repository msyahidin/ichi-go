package models

import (
	"time"

	"github.com/uptrace/bun"
)

// PlatformPermission represents a global platform-level permission (Layer 1)
type PlatformPermission struct {
	bun.BaseModel `bun:"table:platform_permissions,alias:pp"`

	ID         int64      `bun:"id,pk,autoincrement" json:"id"`
	UserID     int64      `bun:"user_id,notnull" json:"user_id"`
	Permission string     `bun:"permission,notnull" json:"permission"`
	GrantedAt  time.Time  `bun:"granted_at,nullzero,default:current_timestamp" json:"granted_at"`
	GrantedBy  *int64     `bun:"granted_by" json:"granted_by,omitempty"`
	ExpiresAt  *time.Time `bun:"expires_at" json:"expires_at,omitempty"`
}

// TableName returns the table name
func (PlatformPermission) TableName() string {
	return "platform_permissions"
}

// Platform permission constants
const (
	PlatformAdmin       = "platform.admin"       // Full platform access
	PlatformImpersonate = "platform.impersonate" // Can impersonate users
	PlatformAudit       = "platform.audit"       // Can view all audit logs
	PlatformManageRoles = "platform.manage_roles" // Can manage roles globally
)

// IsExpired returns true if the permission has expired
func (pp *PlatformPermission) IsExpired() bool {
	if pp.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*pp.ExpiresAt)
}

// IsActive returns true if the permission is active (not expired)
func (pp *PlatformPermission) IsActive() bool {
	return !pp.IsExpired()
}

// IsPlatformAdmin returns true if this is the platform admin permission
func (pp *PlatformPermission) IsPlatformAdmin() bool {
	return pp.Permission == PlatformAdmin
}
