package dto

import "time"

// AuditQueryRequest represents a request to query audit logs
type AuditQueryRequest struct {
	TenantID      *string    `json:"tenant_id,omitempty"`
	ActorID       *string    `json:"actor_id,omitempty"`
	SubjectID     *string    `json:"subject_id,omitempty"`
	Action        *string    `json:"action,omitempty"`
	ResourceType  *string    `json:"resource_type,omitempty"`
	ResourceID    *string    `json:"resource_id,omitempty"`
	Decision      *string    `json:"decision,omitempty" validate:"omitempty,oneof=allow deny"`
	StartDate     *time.Time `json:"start_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	Page          int        `json:"page" validate:"min=1"`
	PageSize      int        `json:"page_size" validate:"min=1,max=100"`
	SortBy        string     `json:"sort_by,omitempty" validate:"omitempty,oneof=timestamp action actor_id"`
	SortDirection string     `json:"sort_direction,omitempty" validate:"omitempty,oneof=asc desc"`
}

// AuditLogResponse represents an audit log entry
type AuditLogResponse struct {
	EventID        string                 `json:"event_id"`
	Timestamp      time.Time              `json:"timestamp"`
	ActorID        string                 `json:"actor_id"`
	ActorType      string                 `json:"actor_type"`
	Action         string                 `json:"action"`
	ResourceType   *string                `json:"resource_type,omitempty"`
	ResourceID     *string                `json:"resource_id,omitempty"`
	SubjectID      *string                `json:"subject_id,omitempty"`
	TenantID       string                 `json:"tenant_id"`
	Decision       *string                `json:"decision,omitempty"`
	DecisionReason *string                `json:"decision_reason,omitempty"`
	PolicyBefore   map[string]interface{} `json:"policy_before,omitempty"`
	PolicyAfter    map[string]interface{} `json:"policy_after,omitempty"`
	Reason         *string                `json:"reason,omitempty"`
	LatencyMs      *int                   `json:"latency_ms,omitempty"`
}

// GetAuditLogsResponse represents the response for audit log queries
type GetAuditLogsResponse struct {
	Logs  []AuditLogResponse `json:"logs"`
	Total int                `json:"total"`
}

// AuditStatsRequest represents a request for audit statistics
type AuditStatsRequest struct {
	TenantID  string    `json:"tenant_id" validate:"required"`
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date" validate:"required,gtfield=StartDate"`
}

// AuditStatsResponse represents audit statistics
type AuditStatsResponse struct {
	TotalEvents      int            `json:"total_events"`
	EventsByAction   map[string]int `json:"events_by_action"`
	DecisionStats    DecisionStats  `json:"decision_stats"`
	TopActors        []ActorStat    `json:"top_actors"`
	AverageLatencyMs float64        `json:"average_latency_ms"`
}

// DecisionStats represents permission decision statistics
type DecisionStats struct {
	TotalDecisions int `json:"total_decisions"`
	Allowed        int `json:"allowed"`
	Denied         int `json:"denied"`
}

// ActorStat represents statistics for an actor
type ActorStat struct {
	ActorID     string `json:"actor_id"`
	EventCount  int    `json:"event_count"`
	LastActivity time.Time `json:"last_activity"`
}

// ExportAuditLogsRequest represents a request to export audit logs
type ExportAuditLogsRequest struct {
	AuditQueryRequest
	Format   string `json:"format" validate:"required,oneof=csv json"`
	FileName string `json:"file_name,omitempty"`
}

// ExportAuditLogsResponse represents the response for export request
type ExportAuditLogsResponse struct {
	FilePath     string    `json:"file_path"`
	RecordCount  int       `json:"record_count"`
	ExportedAt   time.Time `json:"exported_at"`
	Format       string    `json:"format"`
}
