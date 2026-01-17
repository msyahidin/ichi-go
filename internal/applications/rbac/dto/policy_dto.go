package dto

// AddPolicyRequest represents a request to add a policy
type AddPolicyRequest struct {
	Role     string `json:"role" validate:"required"`
	TenantID string `json:"tenant_id" validate:"required"`
	Resource string `json:"resource" validate:"required"`
	Action   string `json:"action" validate:"required"`
	Reason   string `json:"reason,omitempty"`
}

// RemovePolicyRequest represents a request to remove a policy
type RemovePolicyRequest struct {
	Role     string `json:"role" validate:"required"`
	TenantID string `json:"tenant_id" validate:"required"`
	Resource string `json:"resource" validate:"required"`
	Action   string `json:"action" validate:"required"`
	Reason   string `json:"reason,omitempty"`
}

// PolicyResponse represents a single policy
type PolicyResponse struct {
	Role     string `json:"role"`
	TenantID string `json:"tenant_id"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// GetPoliciesRequest represents a request to get policies
type GetPoliciesRequest struct {
	TenantID *string `json:"tenant_id,omitempty"`
	Role     *string `json:"role,omitempty"`
	Page     int     `json:"page" validate:"min=1"`
	PageSize int     `json:"page_size" validate:"min=1,max=100"`
}

// GetPoliciesResponse represents the response for policy list
type GetPoliciesResponse struct {
	Policies []PolicyResponse `json:"policies"`
	Total    int              `json:"total"`
}

// PolicyCountResponse represents policy count statistics
type PolicyCountResponse struct {
	Total        int `json:"total"`
	ByTenant     int `json:"by_tenant,omitempty"`
	GlobalPolicy int `json:"global_policy,omitempty"`
}
