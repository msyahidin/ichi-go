package models

import (
	"time"

	"github.com/uptrace/bun"
)

// AuditLog represents an RBAC audit event (SOC2/GDPR compliant)
type AuditLog struct {
	bun.BaseModel `bun:"table:rbac_audit_log,alias:ral"`

	ID        int64     `bun:"id,pk,autoincrement" json:"id"`
	EventID   string    `bun:"event_id,notnull,unique" json:"event_id"`
	Timestamp time.Time `bun:"timestamp,notnull" json:"timestamp"`

	// Actor context
	ActorID        string  `bun:"actor_id,notnull" json:"actor_id"`
	ActorType      string  `bun:"actor_type,notnull" json:"actor_type"` // user, system, platform_admin
	ActorEmailHash *string `bun:"actor_email_hash" json:"actor_email_hash,omitempty"`

	// Action details
	Action       string  `bun:"action,notnull" json:"action"`
	ResourceType *string `bun:"resource_type" json:"resource_type,omitempty"`
	ResourceID   *string `bun:"resource_id" json:"resource_id,omitempty"`

	// Target context
	SubjectID        *string `bun:"subject_id" json:"subject_id,omitempty"`
	SubjectEmailHash *string `bun:"subject_email_hash" json:"subject_email_hash,omitempty"`
	TenantID         string  `bun:"tenant_id,notnull" json:"tenant_id"`

	// Policy details (JSON)
	PolicyBefore interface{} `bun:"policy_before,type:json" json:"policy_before,omitempty"`
	PolicyAfter  interface{} `bun:"policy_after,type:json" json:"policy_after,omitempty"`

	// Authorization decision
	Decision       *string `bun:"decision" json:"decision,omitempty"`           // allow, deny
	DecisionReason *string `bun:"decision_reason" json:"decision_reason,omitempty"`

	// Metadata
	Reason    *string `bun:"reason,type:text" json:"reason,omitempty"`
	IPAddress *string `bun:"ip_address" json:"ip_address,omitempty"`
	UserAgent *string `bun:"user_agent,type:text" json:"user_agent,omitempty"`
	RequestID *string `bun:"request_id" json:"request_id,omitempty"`

	// Performance metrics
	LatencyMs *int `bun:"latency_ms" json:"latency_ms,omitempty"`
}

// TableName returns the table name
func (AuditLog) TableName() string {
	return "rbac_audit_log"
}

// Action types
const (
	ActionPolicyAdded        = "policy_added"
	ActionPolicyRemoved      = "policy_removed"
	ActionRoleAssigned       = "role_assigned"
	ActionRoleRevoked        = "role_revoked"
	ActionPermissionChecked  = "permission_checked"
	ActionPermissionDenied   = "permission_denied"
)

// Actor types
const (
	ActorTypeUser          = "user"
	ActorTypeSystem        = "system"
	ActorTypePlatformAdmin = "platform_admin"
)

// Decision types
const (
	DecisionAllow = "allow"
	DecisionDeny  = "deny"
)

// IsMutation returns true if this is a mutating action
func (al *AuditLog) IsMutation() bool {
	return al.Action == ActionPolicyAdded ||
		al.Action == ActionPolicyRemoved ||
		al.Action == ActionRoleAssigned ||
		al.Action == ActionRoleRevoked
}

// IsDecision returns true if this is a permission check decision
func (al *AuditLog) IsDecision() bool {
	return al.Action == ActionPermissionChecked ||
		al.Action == ActionPermissionDenied
}

// IsAllowed returns true if the decision was allow
func (al *AuditLog) IsAllowed() bool {
	return al.Decision != nil && *al.Decision == DecisionAllow
}

// IsDenied returns true if the decision was deny
func (al *AuditLog) IsDenied() bool {
	return al.Decision != nil && *al.Decision == DecisionDeny
}
