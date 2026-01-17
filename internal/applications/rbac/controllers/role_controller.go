package controllers

import (
	"net/http"
	"strconv"

	"ichi-go/internal/applications/rbac/dto"
	"ichi-go/internal/applications/rbac/models"
	"ichi-go/internal/applications/rbac/services"

	"github.com/labstack/echo/v4"
	"ichi-go/pkg/utils/response"
)

// RoleController handles role management endpoints
type RoleController struct {
	roleService *services.RoleService
}

// NewRoleController creates a new role controller
func NewRoleController(roleService *services.RoleService) *RoleController {
	return &RoleController{
		roleService: roleService,
	}
}

// GetRoles retrieves all roles with filtering
// GET /api/v1/rbac/roles
func (c *RoleController) GetRoles(ctx echo.Context) error {
	var req dto.GetRolesRequest

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

	var roles []models.Role
	var err error

	if req.GlobalOnly {
		roles, err = c.roleService.GetGlobalRoles(ctx.Request().Context())
	} else if req.TenantID != nil {
		roles, err = c.roleService.GetTenantRoles(ctx.Request().Context(), *req.TenantID)
	} else {
		roles, err = c.roleService.GetAllRoles(ctx.Request().Context())
	}

	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	// Convert to DTO
	roleResponses := make([]dto.RoleResponse, 0, len(roles))
	for _, role := range roles {
		roleResponses = append(roleResponses, c.toRoleResponse(&role))
	}

	// Apply pagination
	total := len(roleResponses)
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedRoles := roleResponses[start:end]

	resp := dto.GetRolesResponse{
		Roles: paginatedRoles,
		Total: total,
	}

	return response.Success(ctx, resp)
}

// GetRole retrieves a single role by ID
// GET /api/v1/rbac/roles/:id
func (c *RoleController) GetRole(ctx echo.Context) error {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, echo.NewHTTPError(http.StatusBadRequest, "Invalid role ID"))
	}

	role, err := c.roleService.GetRoleByID(ctx.Request().Context(), id)
	if err != nil {
		return response.Error(ctx, http.StatusNotFound, err)
	}

	return response.Success(ctx, c.toRoleResponse(role))
}

// GetRoleWithPermissions retrieves a role with its permissions
// GET /api/v1/rbac/roles/:id/permissions
func (c *RoleController) GetRoleWithPermissions(ctx echo.Context) error {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, echo.NewHTTPError(http.StatusBadRequest, "Invalid role ID"))
	}

	role, err := c.roleService.GetRoleWithPermissions(ctx.Request().Context(), id)
	if err != nil {
		return response.Error(ctx, http.StatusNotFound, err)
	}

	// Convert to DTO with permissions
	resp := dto.RoleWithPermissionsResponse{
		RoleResponse: c.toRoleResponse(role),
		Permissions:  c.toPermissionResponses(role.Permissions),
	}

	return response.Success(ctx, resp)
}

// CreateRole creates a new role
// POST /api/v1/rbac/roles
func (c *RoleController) CreateRole(ctx echo.Context) error {
	var req dto.CreateRoleRequest

	if err := ctx.Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	if err := ctx.Validate(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	// Convert DTO to model
	role := &models.Role{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		TenantID:    req.TenantID,
	}

	err := c.roleService.CreateRole(ctx.Request().Context(), role)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, c.toRoleResponse(role))
}

// UpdateRole updates an existing role
// PUT /api/v1/rbac/roles/:id
func (c *RoleController) UpdateRole(ctx echo.Context) error {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, echo.NewHTTPError(http.StatusBadRequest, "Invalid role ID"))
	}

	var req dto.UpdateRoleRequest

	if err := ctx.Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	if err := ctx.Validate(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	// Get existing role
	role, err := c.roleService.GetRoleByID(ctx.Request().Context(), id)
	if err != nil {
		return response.Error(ctx, http.StatusNotFound, err)
	}

	// Update fields
	if req.Name != nil {
		role.Name = *req.Name
	}
	if req.Description != nil {
		role.Description = req.Description
	}

	err = c.roleService.UpdateRole(ctx.Request().Context(), role)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, c.toRoleResponse(role))
}

// DeleteRole deletes a role
// DELETE /api/v1/rbac/roles/:id
func (c *RoleController) DeleteRole(ctx echo.Context) error {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, echo.NewHTTPError(http.StatusBadRequest, "Invalid role ID"))
	}

	err = c.roleService.DeleteRole(ctx.Request().Context(), id)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, dto.NewMessageResponse("Role deleted successfully"))
}

// Helper methods

func (c *RoleController) toRoleResponse(role *models.Role) dto.RoleResponse {
	return dto.RoleResponse{
		ID:           role.ID,
		Name:         role.Name,
		Slug:         role.Slug,
		Description:  role.Description,
		TenantID:     role.TenantID,
		IsSystemRole: role.IsSystemRole,
		IsActive:     role.DeletedAt == nil,
		CreatedAt:    role.CreatedAt,
		UpdatedAt:    role.UpdatedAt,
	}
}

func (c *RoleController) toPermissionResponses(permissions []*models.Permission) []dto.PermissionResponse {
	responses := make([]dto.PermissionResponse, 0, len(permissions))
	for _, p := range permissions {
		var module, resource, action string
		if p.Module != nil {
			module = *p.Module
		}
		if p.Resource != nil {
			resource = *p.Resource
		}
		if p.Action != nil {
			action = *p.Action
		}

		responses = append(responses, dto.PermissionResponse{
			ID:          p.ID,
			Name:        p.Name,
			Slug:        p.Slug,
			Module:      module,
			Resource:    resource,
			Action:      action,
			Description: p.Description,
		})
	}
	return responses
}
