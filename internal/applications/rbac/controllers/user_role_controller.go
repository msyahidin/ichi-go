package controllers

import (
	"net/http"
	"strconv"
	"time"

	"ichi-go/internal/applications/rbac/dto"
	"ichi-go/internal/applications/rbac/models"
	"ichi-go/internal/applications/rbac/services"
	"ichi-go/pkg/requestctx"

	"github.com/labstack/echo/v4"
	"ichi-go/pkg/utils/response"
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

// GetUserRoles retrieves all roles for a user
// GET /api/v1/rbac/users/:userId/roles
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

// GetActiveUserRoles retrieves active (non-expired) roles for a user
// GET /api/v1/rbac/users/:userId/roles/active
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

// AssignRole assigns a role to a user
// POST /api/v1/rbac/users/:userId/roles
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

// RevokeRole revokes a role from a user
// DELETE /api/v1/rbac/users/:userId/roles/:roleSlug
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

// GetUsersWithRole retrieves all users that have a specific role
// GET /api/v1/rbac/roles/:roleId/users
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
