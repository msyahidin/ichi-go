package dto

// CheckPermissionRequest represents a single permission check request
type CheckPermissionRequest struct {
	UserID   int64  `json:"user_id" validate:"required,min=1"`
	TenantID string `json:"tenant_id" validate:"required"`
	Resource string `json:"resource" validate:"required"`
	Action   string `json:"action" validate:"required"`
}

// CheckPermissionResponse represents the permission check result
type CheckPermissionResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// BatchPermissionCheck represents a single check in a batch request
type BatchPermissionCheck struct {
	Resource string `json:"resource" validate:"required"`
	Action   string `json:"action" validate:"required"`
}

// BatchCheckPermissionRequest represents a batch permission check request
type BatchCheckPermissionRequest struct {
	UserID   int64                  `json:"user_id" validate:"required,min=1"`
	TenantID string                 `json:"tenant_id" validate:"required"`
	Checks   []BatchPermissionCheck `json:"checks" validate:"required,min=1,dive"`
}

// BatchCheckPermissionResponse represents batch check results
type BatchCheckPermissionResponse struct {
	Results map[string]bool `json:"results"`
}

// GetUserPermissionsResponse represents user's effective permissions
type GetUserPermissionsResponse struct {
	Permissions []string `json:"permissions"`
}
