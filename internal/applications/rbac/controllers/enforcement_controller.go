package controllers

import (
	"net/http"

	"ichi-go/internal/applications/rbac/dto"
	"ichi-go/internal/applications/rbac/services"
	"ichi-go/pkg/requestctx"
	"ichi-go/pkg/utils/response"

	"github.com/labstack/echo/v4"
)

// EnforcementController handles permission check endpoints
type EnforcementController struct {
	enforcementService *services.EnforcementService
}

// NewEnforcementController creates a new enforcement controller
func NewEnforcementController(enforcementService *services.EnforcementService) *EnforcementController {
	return &EnforcementController{
		enforcementService: enforcementService,
	}
}

// CheckPermission godoc
//
//	@Summary		Check user permission
//	@Description	Check if a user has permission to perform an action on a resource
//	@Tags			RBAC - Enforcement
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.CheckPermissionRequest	true	"Permission check request"
//	@Success		200		{object}	response.SuccessResponse{data=dto.CheckPermissionResponse}
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		500		{object}	response.ErrorResponse
//	@Security		BearerAuth
//	@Router			/v1/rbac/enforce/check [post]
func (c *EnforcementController) CheckPermission(ctx echo.Context) error {
	var req dto.CheckPermissionRequest

	if err := ctx.Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	if err := ctx.Validate(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	allowed, err := c.enforcementService.CheckPermission(
		ctx.Request().Context(),
		req.UserID,
		req.TenantID,
		req.Resource,
		req.Action,
	)

	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	reason := "Permission granted"
	if !allowed {
		reason = "Permission denied"
	}

	resp := dto.CheckPermissionResponse{
		Allowed: allowed,
		Reason:  reason,
	}

	return response.Success(ctx, resp)
}

// CheckBatchPermissions godoc
//
//	@Summary		Check multiple permissions in batch
//	@Description	Check multiple resource-action permissions for a user in a single request
//	@Tags			RBAC - Enforcement
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.BatchCheckPermissionRequest	true	"Batch permission check request"
//	@Success		200		{object}	response.SuccessResponse{data=dto.BatchCheckPermissionResponse}
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		500		{object}	response.ErrorResponse
//	@Security		BearerAuth
//	@Router			/v1/rbac/enforce/batch [post]
func (c *EnforcementController) CheckBatchPermissions(ctx echo.Context) error {
	var req dto.BatchCheckPermissionRequest

	if err := ctx.Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	if err := ctx.Validate(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	// Convert DTO checks to service format
	checks := make([]services.PermissionCheck, len(req.Checks))
	for i, check := range req.Checks {
		checks[i] = services.PermissionCheck{
			Resource: check.Resource,
			Action:   check.Action,
		}
	}

	results, err := c.enforcementService.CheckBatch(
		ctx.Request().Context(),
		req.UserID,
		req.TenantID,
		checks,
	)

	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	resp := dto.BatchCheckPermissionResponse{
		Results: results,
	}

	return response.Success(ctx, resp)
}

// GetMyPermissions godoc
//
//	@Summary		Get current user's permissions
//	@Description	Retrieve all permissions for the authenticated user in a specific tenant
//	@Tags			RBAC - Enforcement
//	@Accept			json
//	@Produce		json
//	@Param			tenant_id	query		string	false	"Tenant ID (uses request context if not provided)"
//	@Success		200			{object}	response.SuccessResponse{data=dto.GetUserPermissionsResponse}
//	@Failure		400			{object}	response.ErrorResponse
//	@Failure		401			{object}	response.ErrorResponse
//	@Failure		500			{object}	response.ErrorResponse
//	@Security		BearerAuth
//	@Router			/v1/rbac/enforce/my-permissions [get]
func (c *EnforcementController) GetMyPermissions(ctx echo.Context) error {
	// Get user ID from request context
	userID := requestctx.GetUserIDAsInt64(ctx.Request().Context())
	if userID == 0 {
		return response.Error(ctx, http.StatusUnauthorized, echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated"))
	}

	// Get tenant ID from query parameter or request context
	tenantID := ctx.QueryParam("tenant_id")
	if tenantID == "" {
		rc := requestctx.FromRequest(ctx.Request())
		if rc.TenantID != "" {
			tenantID = rc.TenantID
		} else {
			return response.Error(ctx, http.StatusBadRequest, echo.NewHTTPError(http.StatusBadRequest, "tenant_id is required"))
		}
	}

	permissions, err := c.enforcementService.GetUserPermissions(
		ctx.Request().Context(),
		userID,
		tenantID,
	)

	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	resp := dto.GetUserPermissionsResponse{
		Permissions: permissions,
	}

	return response.Success(ctx, resp)
}

// RequirePermission is a helper method for middleware/handlers
// It checks permission and returns error if denied
func (c *EnforcementController) RequirePermission(ctx echo.Context, resource, action string) error {
	userID := requestctx.GetUserIDAsInt64(ctx.Request().Context())
	if userID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	rc := requestctx.FromRequest(ctx.Request())
	tenantID := rc.TenantID
	if tenantID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Tenant context not found")
	}

	allowed, err := c.enforcementService.CheckPermission(
		ctx.Request().Context(),
		userID,
		tenantID,
		resource,
		action,
	)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Permission check failed")
	}

	if !allowed {
		return echo.NewHTTPError(http.StatusForbidden, "Permission denied")
	}

	return nil
}
