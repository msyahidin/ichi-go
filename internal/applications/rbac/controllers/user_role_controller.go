package controllers

import (
	"net/http"
	"strconv"
	"time"

	"ichi-go/internal/applications/rbac/dto"
	"ichi-go/internal/applications/rbac/models"
	"ichi-go/internal/applications/rbac/services"
	"ichi-go/pkg/requestctx"

	"ichi-go/pkg/utils/response"

	"github.com/labstack/echo/v4"
)

// UserRoleController handles user role assignment endpoints
type UserRoleController struct {
	userRoleService *services.UserRoleService
}

// NewUserRoleController creates a new user role controller
func NewUserRoleController(userRoleService *services.UserRoleService) *UserRoleController {
	return &UserRoleController{
		userRoleService: userRoleService,
	}
}

// GetUserRoles godoc
//
//	@Summary		Get user roles
//	@Description	Retrieve all roles assigned to a specific user in a tenant
//	@Tags			RBAC - User Roles
//	@Accept			json
//	@Produce		json
//	@Param			userId		path		int		true	"User ID"
//	@Param			tenant_id	query		string	false	"Tenant ID (uses request context if not provided)"
//	@Success		200			{object}	response.SuccessResponse{data=dto.GetUserRolesResponse}
//	@Failure		400			{object}	response.ErrorResponse
//	@Failure		500			{object}	response.ErrorResponse
//	@Security		BearerAuth
//	@Router			/v1/rbac/users/{userId}/roles [get]
func (c *UserRoleController) GetUserRoles(ctx echo.Context) error {
	userID, err := strconv.ParseInt(ctx.Param("userId"), 10, 64)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID"))
	}

	tenantID := ctx.QueryParam("tenant_id")
	if tenantID == "" {
		rc := requestctx.FromRequest(ctx.Request())
		if rc.TenantID != "" {
			tenantID = rc.TenantID
		} else {
			return response.Error(ctx, http.StatusBadRequest, echo.NewHTTPError(http.StatusBadRequest, "tenant_id is required"))
		}
	}

	userRoles, err := c.userRoleService.GetUserRoles(ctx.Request().Context(), userID, tenantID)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	// Convert to DTO
	roleResponses := make([]dto.UserRoleResponse, 0, len(userRoles))
	for _, ur := range userRoles {
		roleResponses = append(roleResponses, c.toUserRoleResponse(&ur))
	}

	resp := dto.GetUserRolesResponse{
		Roles: roleResponses,
		Total: len(roleResponses),
	}

	return response.Success(ctx, resp)
}

// GetActiveUserRoles godoc
//
//	@Summary		Get active user roles
//	@Description	Retrieve only active (non-expired) roles for a specific user in a tenant
//	@Tags			RBAC - User Roles
//	@Accept			json
//	@Produce		json
//	@Param			userId		path		int		true	"User ID"
//	@Param			tenant_id	query		string	false	"Tenant ID (uses request context if not provided)"
//	@Success		200			{object}	response.SuccessResponse{data=dto.GetUserRolesResponse}
//	@Failure		400			{object}	response.ErrorResponse
//	@Failure		500			{object}	response.ErrorResponse
//	@Security		BearerAuth
//	@Router			/v1/rbac/users/{userId}/roles/active [get]
func (c *UserRoleController) GetActiveUserRoles(ctx echo.Context) error {
	userID, err := strconv.ParseInt(ctx.Param("userId"), 10, 64)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID"))
	}

	tenantID := ctx.QueryParam("tenant_id")
	if tenantID == "" {
		rc := requestctx.FromRequest(ctx.Request())
		if rc.TenantID != "" {
			tenantID = rc.TenantID
		} else {
			return response.Error(ctx, http.StatusBadRequest, echo.NewHTTPError(http.StatusBadRequest, "tenant_id is required"))
		}
	}

	userRoles, err := c.userRoleService.GetActiveUserRoles(ctx.Request().Context(), userID, tenantID)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	// Convert to DTO
	roleResponses := make([]dto.UserRoleResponse, 0, len(userRoles))
	for _, ur := range userRoles {
		roleResponses = append(roleResponses, c.toUserRoleResponse(&ur))
	}

	resp := dto.GetUserRolesResponse{
		Roles: roleResponses,
		Total: len(roleResponses),
	}

	return response.Success(ctx, resp)
}

// AssignRole godoc
//
//	@Summary		Assign role to user
//	@Description	Assign a role to a user in a specific tenant
//	@Tags			RBAC - User Roles
//	@Accept			json
//	@Produce		json
//	@Param			userId	path		int						true	"User ID"
//	@Param			request	body		dto.AssignRoleRequest	true	"Assign role request"
//	@Success		200		{object}	response.SuccessResponse{data=dto.MessageResponse}
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		401		{object}	response.ErrorResponse
//	@Failure		500		{object}	response.ErrorResponse
//	@Security		BearerAuth
//	@Router			/v1/rbac/users/{userId}/roles [post]
func (c *UserRoleController) AssignRole(ctx echo.Context) error {
	userID, err := strconv.ParseInt(ctx.Param("userId"), 10, 64)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID"))
	}

	var req dto.AssignRoleRequest

	if err := ctx.Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	if err := ctx.Validate(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	// Override user ID from path parameter
	req.UserID = userID

	// Get actor ID from context
	actorID := requestctx.GetUserIDAsInt64(ctx.Request().Context())
	if actorID == 0 {
		return response.Error(ctx, http.StatusUnauthorized, echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated"))
	}

	err = c.userRoleService.AssignRole(
		ctx.Request().Context(),
		req.UserID,
		req.RoleSlug,
		req.TenantID,
		actorID,
		req.Reason,
	)

	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, dto.NewMessageResponse("Role assigned successfully"))
}

// RevokeRole godoc
//
//	@Summary		Revoke role from user
//	@Description	Revoke a role assignment from a user in a specific tenant
//	@Tags			RBAC - User Roles
//	@Accept			json
//	@Produce		json
//	@Param			userId		path		int						true	"User ID"
//	@Param			roleSlug	path		string					true	"Role slug"
//	@Param			request		body		dto.RevokeRoleRequest	true	"Revoke role request (optional tenant_id and reason)"
//	@Success		200			{object}	response.SuccessResponse{data=dto.MessageResponse}
//	@Failure		400			{object}	response.ErrorResponse
//	@Failure		401			{object}	response.ErrorResponse
//	@Failure		500			{object}	response.ErrorResponse
//	@Security		BearerAuth
//	@Router			/v1/rbac/users/{userId}/roles/{roleSlug} [delete]
func (c *UserRoleController) RevokeRole(ctx echo.Context) error {
	userID, err := strconv.ParseInt(ctx.Param("userId"), 10, 64)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID"))
	}

	roleSlug := ctx.Param("roleSlug")
	if roleSlug == "" {
		return response.Error(ctx, http.StatusBadRequest, echo.NewHTTPError(http.StatusBadRequest, "Role slug is required"))
	}

	var req dto.RevokeRoleRequest

	if err := ctx.Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	// Override from path parameters
	req.UserID = userID
	req.RoleSlug = roleSlug

	if req.TenantID == "" {
		rc := requestctx.FromRequest(ctx.Request())
		if rc.TenantID != "" {
			req.TenantID = rc.TenantID
		} else {
			return response.Error(ctx, http.StatusBadRequest, echo.NewHTTPError(http.StatusBadRequest, "tenant_id is required"))
		}
	}

	// Get actor ID from context
	actorID := requestctx.GetUserIDAsInt64(ctx.Request().Context())
	if actorID == 0 {
		return response.Error(ctx, http.StatusUnauthorized, echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated"))
	}

	err = c.userRoleService.RevokeRole(
		ctx.Request().Context(),
		req.UserID,
		req.RoleSlug,
		req.TenantID,
		actorID,
		req.Reason,
	)

	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, dto.NewMessageResponse("Role revoked successfully"))
}

// GetUsersWithRole godoc
//
//	@Summary		Get users with role
//	@Description	Retrieve all users that have been assigned a specific role in a tenant
//	@Tags			RBAC - User Roles
//	@Accept			json
//	@Produce		json
//	@Param			roleId		path		int		true	"Role ID"
//	@Param			tenant_id	query		string	false	"Tenant ID (uses request context if not provided)"
//	@Success		200			{object}	response.SuccessResponse{data=dto.GetUsersWithRoleResponse}
//	@Failure		400			{object}	response.ErrorResponse
//	@Failure		500			{object}	response.ErrorResponse
//	@Security		BearerAuth
//	@Router			/v1/rbac/roles/{roleId}/users [get]
func (c *UserRoleController) GetUsersWithRole(ctx echo.Context) error {
	roleID, err := strconv.ParseInt(ctx.Param("roleId"), 10, 64)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, echo.NewHTTPError(http.StatusBadRequest, "Invalid role ID"))
	}

	tenantID := ctx.QueryParam("tenant_id")
	if tenantID == "" {
		rc := requestctx.FromRequest(ctx.Request())
		if rc.TenantID != "" {
			tenantID = rc.TenantID
		} else {
			return response.Error(ctx, http.StatusBadRequest, echo.NewHTTPError(http.StatusBadRequest, "tenant_id is required"))
		}
	}

	userRoles, err := c.userRoleService.GetUsersWithRole(ctx.Request().Context(), roleID, tenantID)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	// Convert to DTO
	userResponses := make([]dto.UserRoleResponse, 0, len(userRoles))
	for _, ur := range userRoles {
		userResponses = append(userResponses, c.toUserRoleResponse(&ur))
	}

	resp := dto.GetUsersWithRoleResponse{
		Users: userResponses,
		Total: len(userResponses),
	}

	return response.Success(ctx, resp)
}

// Helper methods

func (c *UserRoleController) toUserRoleResponse(userRole *models.UserRole) dto.UserRoleResponse {
	var assignedAt time.Time
	if userRole.AssignedAt != nil {
		assignedAt = *userRole.AssignedAt
	}

	resp := dto.UserRoleResponse{
		ID:         userRole.ID,
		UserID:     userRole.UserID,
		RoleID:     userRole.RoleID,
		TenantID:   userRole.TenantID,
		AssignedBy: userRole.AssignedBy,
		AssignedAt: assignedAt,
		ExpiresAt:  userRole.ExpiresAt,
		IsActive:   userRole.IsActive(),
	}

	// Add role details if loaded
	if userRole.Role != nil {
		resp.RoleName = userRole.Role.Name
		resp.RoleSlug = userRole.Role.Slug
	}

	return resp
}
