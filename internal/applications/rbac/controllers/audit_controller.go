package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"ichi-go/internal/applications/rbac/dto"
	"ichi-go/internal/applications/rbac/repositories"
	"ichi-go/internal/applications/rbac/services"

	"ichi-go/pkg/utils/response"

	"github.com/labstack/echo/v4"
)

// AuditController handles audit log endpoints
type AuditController struct {
	auditService *services.AuditService
}

// NewAuditController creates a new audit controller
func NewAuditController(auditService *services.AuditService) *AuditController {
	return &AuditController{
		auditService: auditService,
	}
}

// QueryAuditLogs godoc
//
//	@Summary		Query audit logs
//	@Description	Retrieve audit logs with comprehensive filtering and pagination options
//	@Tags			RBAC - Audit
//	@Accept			json
//	@Produce		json
//	@Param			tenant_id		query		string	false	"Filter by tenant ID"
//	@Param			actor_id		query		string	false	"Filter by actor ID"
//	@Param			subject_id		query		string	false	"Filter by subject ID"
//	@Param			action			query		string	false	"Filter by action type"
//	@Param			decision		query		string	false	"Filter by decision (allowed/denied)"
//	@Param			start_date		query		string	false	"Filter by start date (RFC3339)"
//	@Param			end_date		query		string	false	"Filter by end date (RFC3339)"
//	@Param			page			query		int		false	"Page number"				default(1)
//	@Param			page_size		query		int		false	"Page size (max 100)"		default(20)
//	@Param			sort_by			query		string	false	"Sort field"				default("timestamp")
//	@Param			sort_direction	query		string	false	"Sort direction (asc/desc)"	default("desc")
//	@Success		200				{object}	response.SuccessResponse{data=dto.GetAuditLogsResponse}
//	@Failure		400				{object}	response.ErrorResponse
//	@Failure		500				{object}	response.ErrorResponse
//	@Security		BearerAuth
//	@Router			/v1/rbac/audit/logs [get]
func (c *AuditController) QueryAuditLogs(ctx echo.Context) error {
	var req dto.AuditQueryRequest

	if err := ctx.Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	if err := ctx.Validate(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	if req.SortBy == "" {
		req.SortBy = "timestamp"
	}
	if req.SortDirection == "" {
		req.SortDirection = "desc"
	}

	// Build query
	query := repositories.AuditQuery{
		TenantID:  req.TenantID,
		ActorID:   req.ActorID,
		SubjectID: req.SubjectID,
		Action:    req.Action,
		Decision:  req.Decision,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Limit:     req.PageSize,
		Offset:    (req.Page - 1) * req.PageSize,
	}

	logs, total, err := c.auditService.QueryAuditLogs(ctx.Request().Context(), query)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	// Convert to DTO
	logResponses := make([]dto.AuditLogResponse, 0, len(logs))
	for _, log := range logs {
		logResponses = append(logResponses, dto.AuditLogResponse{
			EventID:        log.EventID,
			Timestamp:      log.Timestamp,
			ActorID:        log.ActorID,
			ActorType:      log.ActorType,
			Action:         log.Action,
			ResourceType:   log.ResourceType,
			ResourceID:     log.ResourceID,
			SubjectID:      log.SubjectID,
			TenantID:       log.TenantID,
			Decision:       log.Decision,
			DecisionReason: log.DecisionReason,
			PolicyBefore:   log.PolicyBefore.(map[string]interface{}),
			PolicyAfter:    log.PolicyAfter.(map[string]interface{}),
			Reason:         log.Reason,
			LatencyMs:      log.LatencyMs,
		})
	}

	resp := dto.GetAuditLogsResponse{
		Logs:  logResponses,
		Total: total,
	}

	return response.Success(ctx, resp)
}

// GetAuditStats godoc
//
//	@Summary		Get audit statistics
//	@Description	Retrieve audit statistics including event counts by action, decision stats, and performance metrics
//	@Tags			RBAC - Audit
//	@Accept			json
//	@Produce		json
//	@Param			tenant_id	query		string	true	"Tenant ID"
//	@Param			start_date	query		string	false	"Start date (RFC3339)"
//	@Param			end_date	query		string	false	"End date (RFC3339)"
//	@Success		200			{object}	response.SuccessResponse{data=dto.AuditStatsResponse}
//	@Failure		400			{object}	response.ErrorResponse
//	@Failure		500			{object}	response.ErrorResponse
//	@Security		BearerAuth
//	@Router			/v1/rbac/audit/stats [get]
func (c *AuditController) GetAuditStats(ctx echo.Context) error {
	var req dto.AuditStatsRequest

	if err := ctx.Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	if err := ctx.Validate(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	stats, err := c.auditService.GetAuditStats(
		ctx.Request().Context(),
		req.TenantID,
		req.StartDate,
		req.EndDate,
	)

	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	// Build response
	resp := dto.AuditStatsResponse{
		EventsByAction: stats,
		TotalEvents:    0,
	}

	for _, count := range stats {
		resp.TotalEvents += count
	}

	// TODO: Add more detailed stats (decisions, top actors, latency)
	resp.DecisionStats = dto.DecisionStats{
		TotalDecisions: 0,
		Allowed:        0,
		Denied:         0,
	}
	resp.TopActors = []dto.ActorStat{}
	resp.AverageLatencyMs = 0.0

	return response.Success(ctx, resp)
}

// ExportAuditLogs godoc
//
//	@Summary		Export audit logs
//	@Description	Export audit logs to CSV or JSON format with optional filtering
//	@Tags			RBAC - Audit
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.ExportAuditLogsRequest	true	"Export audit logs request"
//	@Success		200		{object}	response.SuccessResponse{data=dto.ExportAuditLogsResponse}
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		500		{object}	response.ErrorResponse
//	@Security		BearerAuth
//	@Router			/v1/rbac/audit/export [post]
func (c *AuditController) ExportAuditLogs(ctx echo.Context) error {
	var req dto.ExportAuditLogsRequest

	if err := ctx.Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	if err := ctx.Validate(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	// Build query
	query := repositories.AuditQuery{
		TenantID:  req.TenantID,
		ActorID:   req.ActorID,
		SubjectID: req.SubjectID,
		Action:    req.Action,
		Decision:  req.Decision,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}

	// Generate filename
	fileName := req.FileName
	if fileName == "" {
		fileName = fmt.Sprintf("audit_logs_%s.%s", time.Now().Format("20060102_150405"), req.Format)
	}

	// Create export directory
	exportDir := "./exports/audit"
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	filePath := filepath.Join(exportDir, fileName)

	// Export based on format
	var err error
	var recordCount int

	switch req.Format {
	case "csv":
		err = c.auditService.ExportToCSV(ctx.Request().Context(), query, filePath)
	case "json":
		err = c.auditService.ExportToJSON(ctx.Request().Context(), query, filePath)
	default:
		return response.Error(ctx, http.StatusBadRequest, echo.NewHTTPError(http.StatusBadRequest, "Invalid format. Use 'csv' or 'json'"))
	}

	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	// Get record count (re-query with no limit)
	logs, _, err := c.auditService.QueryAuditLogs(ctx.Request().Context(), query)
	if err == nil {
		recordCount = len(logs)
	}

	resp := dto.ExportAuditLogsResponse{
		FilePath:    filePath,
		RecordCount: recordCount,
		ExportedAt:  time.Now(),
		Format:      req.Format,
	}

	return response.Success(ctx, resp)
}

// GetRecentMutations godoc
//
//	@Summary		Get recent mutations
//	@Description	Retrieve recent policy and role mutations (additions, removals, assignments, revocations)
//	@Tags			RBAC - Audit
//	@Accept			json
//	@Produce		json
//	@Param			tenant_id	query		string	false	"Filter by tenant ID"
//	@Success		200			{object}	response.SuccessResponse{data=dto.GetAuditLogsResponse}
//	@Failure		500			{object}	response.ErrorResponse
//	@Security		BearerAuth
//	@Router			/v1/rbac/audit/mutations [get]
func (c *AuditController) GetRecentMutations(ctx echo.Context) error {
	tenantID := ctx.QueryParam("tenant_id")
	limit := 50 // Default limit

	var tenantIDPtr *string
	if tenantID != "" {
		tenantIDPtr = &tenantID
	}

	logs, err := c.auditService.GetRecentMutations(ctx.Request().Context(), tenantIDPtr, limit)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	// Convert to DTO
	logResponses := make([]dto.AuditLogResponse, 0, len(logs))
	for _, log := range logs {
		logResponses = append(logResponses, dto.AuditLogResponse{
			EventID:      log.EventID,
			Timestamp:    log.Timestamp,
			ActorID:      log.ActorID,
			ActorType:    log.ActorType,
			Action:       log.Action,
			ResourceType: log.ResourceType,
			ResourceID:   log.ResourceID,
			SubjectID:    log.SubjectID,
			TenantID:     log.TenantID,
			PolicyBefore: log.PolicyBefore.(map[string]interface{}),
			PolicyAfter:  log.PolicyAfter.(map[string]interface{}),
			Reason:       log.Reason,
		})
	}

	resp := dto.GetAuditLogsResponse{
		Logs:  logResponses,
		Total: len(logResponses),
	}

	return response.Success(ctx, resp)
}

// GetRecentDecisions godoc
//
//	@Summary		Get recent permission decisions
//	@Description	Retrieve recent permission check decisions (allowed/denied) with performance metrics
//	@Tags			RBAC - Audit
//	@Accept			json
//	@Produce		json
//	@Param			tenant_id	query		string	false	"Filter by tenant ID"
//	@Success		200			{object}	response.SuccessResponse{data=dto.GetAuditLogsResponse}
//	@Failure		500			{object}	response.ErrorResponse
//	@Security		BearerAuth
//	@Router			/v1/rbac/audit/decisions [get]
func (c *AuditController) GetRecentDecisions(ctx echo.Context) error {
	tenantID := ctx.QueryParam("tenant_id")
	limit := 50 // Default limit

	var tenantIDPtr *string
	if tenantID != "" {
		tenantIDPtr = &tenantID
	}

	logs, err := c.auditService.GetRecentDecisions(ctx.Request().Context(), tenantIDPtr, limit)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	// Convert to DTO
	logResponses := make([]dto.AuditLogResponse, 0, len(logs))
	for _, log := range logs {
		logResponses = append(logResponses, dto.AuditLogResponse{
			EventID:        log.EventID,
			Timestamp:      log.Timestamp,
			ActorID:        log.ActorID,
			ActorType:      log.ActorType,
			Action:         log.Action,
			SubjectID:      log.SubjectID,
			TenantID:       log.TenantID,
			Decision:       log.Decision,
			DecisionReason: log.DecisionReason,
			LatencyMs:      log.LatencyMs,
		})
	}

	resp := dto.GetAuditLogsResponse{
		Logs:  logResponses,
		Total: len(logResponses),
	}

	return response.Success(ctx, resp)
}
