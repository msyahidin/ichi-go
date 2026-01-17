package services

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"ichi-go/internal/applications/rbac/models"
	"ichi-go/internal/applications/rbac/repositories"
	"ichi-go/pkg/logger"
)

// AuditService handles audit log queries and exports
type AuditService struct {
	auditRepo *repositories.AuditRepository
}

// NewAuditService creates a new audit service
func NewAuditService(auditRepo *repositories.AuditRepository) *AuditService {
	return &AuditService{
		auditRepo: auditRepo,
	}
}

// QueryAuditLogs retrieves audit logs based on query parameters
func (s *AuditService) QueryAuditLogs(
	ctx context.Context,
	query repositories.AuditQuery,
) ([]models.AuditLog, int, error) {
	// Get logs
	logs, err := s.auditRepo.FindByQuery(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit logs: %w", err)
	}

	// Get total count
	count, err := s.auditRepo.CountByQuery(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	return logs, count, nil
}

// GetAuditLogByEventID retrieves a specific audit log by event ID
func (s *AuditService) GetAuditLogByEventID(ctx context.Context, eventID string) (*models.AuditLog, error) {
	return s.auditRepo.FindByEventID(ctx, eventID)
}

// GetAuditLogsByRequest retrieves all audit logs for a request
func (s *AuditService) GetAuditLogsByRequest(ctx context.Context, requestID string) ([]models.AuditLog, error) {
	return s.auditRepo.FindByRequestID(ctx, requestID)
}

// GetRecentMutations retrieves recent policy/role changes
func (s *AuditService) GetRecentMutations(ctx context.Context, tenantID *string, limit int) ([]models.AuditLog, error) {
	return s.auditRepo.FindMutations(ctx, tenantID, limit)
}

// GetRecentDecisions retrieves recent permission check decisions
func (s *AuditService) GetRecentDecisions(ctx context.Context, tenantID *string, limit int) ([]models.AuditLog, error) {
	return s.auditRepo.FindDecisions(ctx, tenantID, limit)
}

// GetAuditStats returns audit statistics for a tenant
func (s *AuditService) GetAuditStats(
	ctx context.Context,
	tenantID string,
	startDate, endDate time.Time,
) (map[string]int, error) {
	return s.auditRepo.GetStatsByTenant(ctx, tenantID, startDate, endDate)
}

// ExportToCSV exports audit logs to CSV format (SOC2/GDPR compliance)
func (s *AuditService) ExportToCSV(
	ctx context.Context,
	query repositories.AuditQuery,
	filePath string,
) error {
	// Get all matching logs
	logs, err := s.auditRepo.FindByQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to fetch audit logs: %w", err)
	}

	// Create CSV file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"EventID", "Timestamp", "ActorID", "ActorType", "Action",
		"ResourceType", "ResourceID", "SubjectID", "TenantID",
		"Decision", "DecisionReason", "Reason", "LatencyMs",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write rows
	for _, log := range logs {
		row := []string{
			log.EventID,
			log.Timestamp.Format(time.RFC3339),
			log.ActorID,
			log.ActorType,
			log.Action,
			strVal(log.ResourceType),
			strVal(log.ResourceID),
			strVal(log.SubjectID),
			log.TenantID,
			strVal(log.Decision),
			strVal(log.DecisionReason),
			strVal(log.Reason),
			intVal(log.LatencyMs),
		}

		if err := writer.Write(row); err != nil {
			logger.WithContext(ctx).Errorf("Failed to write CSV row: %v", err)
		}
	}

	logger.WithContext(ctx).Infof("Exported %d audit logs to %s", len(logs), filePath)

	return nil
}

// ExportToJSON exports audit logs to JSON format (SOC2/GDPR compliance)
func (s *AuditService) ExportToJSON(
	ctx context.Context,
	query repositories.AuditQuery,
	filePath string,
) error {
	// Get all matching logs
	logs, err := s.auditRepo.FindByQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to fetch audit logs: %w", err)
	}

	// Create JSON file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	// Marshal to JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(logs); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}

	logger.WithContext(ctx).Infof("Exported %d audit logs to %s", len(logs), filePath)

	return nil
}

// CleanupOldLogs deletes audit logs older than retention period
func (s *AuditService) CleanupOldLogs(ctx context.Context, retentionDays int) (int, error) {
	deleted, err := s.auditRepo.DeleteOldLogs(ctx, retentionDays)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old logs: %w", err)
	}

	logger.WithContext(ctx).Infof("Deleted %d old audit logs (retention: %d days)", deleted, retentionDays)

	return deleted, nil
}

// Helper functions
func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func intVal(i *int) string {
	if i == nil {
		return ""
	}
	return fmt.Sprintf("%d", *i)
}
