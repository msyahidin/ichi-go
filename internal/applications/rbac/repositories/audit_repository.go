package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"ichi-go/internal/applications/rbac/models"
)

// AuditRepository handles audit log database operations
type AuditRepository struct {
	db *bun.DB
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *bun.DB) *AuditRepository {
	return &AuditRepository{
		db: db,
	}
}

// AuditQuery represents query parameters for audit logs
type AuditQuery struct {
	ActorID   *string
	SubjectID *string
	TenantID  *string
	Action    *string
	StartDate *time.Time
	EndDate   *time.Time
	Decision  *string
	Limit     int
	Offset    int
}

// Create creates a new audit log entry
func (r *AuditRepository) Create(ctx context.Context, log *models.AuditLog) error {
	_, err := r.db.NewInsert().
		Model(log).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// FindByQuery retrieves audit logs based on query parameters
func (r *AuditRepository) FindByQuery(ctx context.Context, query AuditQuery) ([]models.AuditLog, error) {
	var logs []models.AuditLog

	q := r.db.NewSelect().Model(&logs)

	// Apply filters
	if query.ActorID != nil {
		q = q.Where("actor_id = ?", *query.ActorID)
	}
	if query.SubjectID != nil {
		q = q.Where("subject_id = ?", *query.SubjectID)
	}
	if query.TenantID != nil {
		q = q.Where("tenant_id = ?", *query.TenantID)
	}
	if query.Action != nil {
		q = q.Where("action = ?", *query.Action)
	}
	if query.Decision != nil {
		q = q.Where("decision = ?", *query.Decision)
	}
	if query.StartDate != nil {
		q = q.Where("timestamp >= ?", *query.StartDate)
	}
	if query.EndDate != nil {
		q = q.Where("timestamp <= ?", *query.EndDate)
	}

	// Apply pagination
	if query.Limit > 0 {
		q = q.Limit(query.Limit)
	}
	if query.Offset > 0 {
		q = q.Offset(query.Offset)
	}

	// Order by timestamp descending (newest first)
	q = q.Order("timestamp DESC")

	err := q.Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find audit logs: %w", err)
	}

	return logs, nil
}

// FindByEventID retrieves an audit log by event ID
func (r *AuditRepository) FindByEventID(ctx context.Context, eventID string) (*models.AuditLog, error) {
	log := new(models.AuditLog)

	err := r.db.NewSelect().
		Model(log).
		Where("event_id = ?", eventID).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find audit log by event ID: %w", err)
	}

	return log, nil
}

// FindByRequestID retrieves all audit logs for a request
func (r *AuditRepository) FindByRequestID(ctx context.Context, requestID string) ([]models.AuditLog, error) {
	var logs []models.AuditLog

	err := r.db.NewSelect().
		Model(&logs).
		Where("request_id = ?", requestID).
		Order("timestamp ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find audit logs by request ID: %w", err)
	}

	return logs, nil
}

// FindMutations retrieves all mutation events (policy/role changes)
func (r *AuditRepository) FindMutations(ctx context.Context, tenantID *string, limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog

	q := r.db.NewSelect().
		Model(&logs).
		Where("action IN (?)", bun.In([]string{
			models.ActionPolicyAdded,
			models.ActionPolicyRemoved,
			models.ActionRoleAssigned,
			models.ActionRoleRevoked,
		}))

	if tenantID != nil {
		q = q.Where("tenant_id = ?", *tenantID)
	}

	if limit > 0 {
		q = q.Limit(limit)
	}

	q = q.Order("timestamp DESC")

	err := q.Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find mutations: %w", err)
	}

	return logs, nil
}

// FindDecisions retrieves all permission check decisions
func (r *AuditRepository) FindDecisions(ctx context.Context, tenantID *string, limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog

	q := r.db.NewSelect().
		Model(&logs).
		Where("action IN (?)", bun.In([]string{
			models.ActionPermissionChecked,
			models.ActionPermissionDenied,
		}))

	if tenantID != nil {
		q = q.Where("tenant_id = ?", *tenantID)
	}

	if limit > 0 {
		q = q.Limit(limit)
	}

	q = q.Order("timestamp DESC")

	err := q.Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find decisions: %w", err)
	}

	return logs, nil
}

// CountByQuery counts audit logs matching query
func (r *AuditRepository) CountByQuery(ctx context.Context, query AuditQuery) (int, error) {
	q := r.db.NewSelect().Model((*models.AuditLog)(nil))

	// Apply same filters as FindByQuery
	if query.ActorID != nil {
		q = q.Where("actor_id = ?", *query.ActorID)
	}
	if query.SubjectID != nil {
		q = q.Where("subject_id = ?", *query.SubjectID)
	}
	if query.TenantID != nil {
		q = q.Where("tenant_id = ?", *query.TenantID)
	}
	if query.Action != nil {
		q = q.Where("action = ?", *query.Action)
	}
	if query.Decision != nil {
		q = q.Where("decision = ?", *query.Decision)
	}
	if query.StartDate != nil {
		q = q.Where("timestamp >= ?", *query.StartDate)
	}
	if query.EndDate != nil {
		q = q.Where("timestamp <= ?", *query.EndDate)
	}

	count, err := q.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	return count, nil
}

// DeleteOldLogs deletes audit logs older than retention period (for cleanup jobs)
func (r *AuditRepository) DeleteOldLogs(ctx context.Context, retentionDays int) (int, error) {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	result, err := r.db.NewDelete().
		Model((*models.AuditLog)(nil)).
		Where("timestamp < ?", cutoffDate).
		Exec(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to delete old audit logs: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}

// GetStatsByTenant returns audit statistics for a tenant
func (r *AuditRepository) GetStatsByTenant(ctx context.Context, tenantID string, startDate, endDate time.Time) (map[string]int, error) {
	type ActionCount struct {
		Action string
		Count  int
	}

	var results []ActionCount

	err := r.db.NewSelect().
		Model((*models.AuditLog)(nil)).
		Column("action").
		ColumnExpr("COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Where("timestamp >= ?", startDate).
		Where("timestamp <= ?", endDate).
		Group("action").
		Scan(ctx, &results)

	if err != nil {
		return nil, fmt.Errorf("failed to get audit stats: %w", err)
	}

	// Convert to map
	stats := make(map[string]int)
	for _, result := range results {
		stats[result.Action] = result.Count
	}

	return stats, nil
}
