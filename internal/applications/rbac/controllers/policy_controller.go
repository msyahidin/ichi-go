package controllers

import (
	"net/http"

	"ichi-go/internal/applications/rbac/dto"
	"ichi-go/internal/applications/rbac/services"
	"ichi-go/pkg/requestctx"

	"github.com/labstack/echo/v4"
	"ichi-go/pkg/utils/response"
)

// PolicyController handles policy management endpoints
type PolicyController struct {
	policyService *services.PolicyService
}

// NewPolicyController creates a new policy controller
func NewPolicyController(policyService *services.PolicyService) *PolicyController {
	return &PolicyController{
		policyService: policyService,
	}
}

// GetPolicies retrieves policies with filtering and pagination
// GET /api/v1/rbac/policies
func (c *PolicyController) GetPolicies(ctx echo.Context) error {
	var req dto.GetPoliciesRequest

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

	var policies []dto.PolicyResponse

	if req.TenantID != nil {
		casbinPolicies, err := c.policyService.GetPoliciesByTenant(ctx.Request().Context(), *req.TenantID)
		if err != nil {
			return response.Error(ctx, http.StatusInternalServerError, err)
		}
		// Convert to DTO
		for _, p := range casbinPolicies {
			policies = append(policies, dto.PolicyResponse{
				Role:     p.V0,
				TenantID: p.V1,
				Resource: p.V2,
				Action:   p.V3,
			})
		}
	} else if req.Role != nil {
		casbinPolicies, err := c.policyService.GetPoliciesByRole(ctx.Request().Context(), *req.Role, *req.TenantID)
		if err != nil {
			return response.Error(ctx, http.StatusInternalServerError, err)
		}
		// Convert to DTO
		for _, p := range casbinPolicies {
			policies = append(policies, dto.PolicyResponse{
				Role:     p.V0,
				TenantID: p.V1,
				Resource: p.V2,
				Action:   p.V3,
			})
		}
	}

	// Apply pagination
	total := len(policies)
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedPolicies := policies[start:end]

	resp := dto.GetPoliciesResponse{
		Policies: paginatedPolicies,
		Total:    total,
	}

	return response.Success(ctx, resp)
}

// AddPolicy adds a new policy
// POST /api/v1/rbac/policies
func (c *PolicyController) AddPolicy(ctx echo.Context) error {
	var req dto.AddPolicyRequest

	if err := ctx.Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	if err := ctx.Validate(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	// Get actor ID from context
	actorID := requestctx.GetUserIDAsInt64(ctx.Request().Context())
	if actorID == 0 {
		return response.Error(ctx, http.StatusUnauthorized, echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated"))
	}

	err := c.policyService.AddPolicy(
		ctx.Request().Context(),
		req.Role,
		req.TenantID,
		req.Resource,
		req.Action,
		actorID,
		req.Reason,
	)

	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, dto.NewMessageResponse("Policy added successfully"))
}

// RemovePolicy removes a policy
// DELETE /api/v1/rbac/policies
func (c *PolicyController) RemovePolicy(ctx echo.Context) error {
	var req dto.RemovePolicyRequest

	if err := ctx.Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	if err := ctx.Validate(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, err)
	}

	// Get actor ID from context
	actorID := requestctx.GetUserIDAsInt64(ctx.Request().Context())
	if actorID == 0 {
		return response.Error(ctx, http.StatusUnauthorized, echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated"))
	}

	err := c.policyService.RemovePolicy(
		ctx.Request().Context(),
		req.Role,
		req.TenantID,
		req.Resource,
		req.Action,
		actorID,
		req.Reason,
	)

	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, dto.NewMessageResponse("Policy removed successfully"))
}

// GetPolicyCount returns policy count statistics
// GET /api/v1/rbac/policies/count
func (c *PolicyController) GetPolicyCount(ctx echo.Context) error {
	tenantID := ctx.QueryParam("tenant_id")

	var count int
	var err error

	if tenantID != "" {
		count, err = c.policyService.CountPoliciesByTenant(ctx.Request().Context(), tenantID)
	} else {
		count, err = c.policyService.CountPolicies(ctx.Request().Context())
	}

	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	resp := dto.PolicyCountResponse{
		Total: count,
	}

	if tenantID != "" {
		resp.ByTenant = count
	} else {
		resp.GlobalPolicy = count
	}

	return response.Success(ctx, resp)
}

// ReloadPolicies reloads all policies from database
// POST /api/v1/rbac/policies/reload
func (c *PolicyController) ReloadPolicies(ctx echo.Context) error {
	// Get actor ID from context
	actorID := requestctx.GetUserIDAsInt64(ctx.Request().Context())
	if actorID == 0 {
		return response.Error(ctx, http.StatusUnauthorized, echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated"))
	}

	err := c.policyService.ReloadPolicies(ctx.Request().Context())
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, dto.NewMessageResponse("Policies reloaded successfully"))
}
