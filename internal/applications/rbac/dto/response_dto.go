package dto

// APIResponse represents a standard API response wrapper
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool               `json:"success"`
	Message    string             `json:"message,omitempty"`
	Data       interface{}        `json:"data"`
	Pagination PaginationMetadata `json:"pagination"`
	Error      *ErrorInfo         `json:"error,omitempty"`
}

// ErrorInfo represents error details in API responses
type ErrorInfo struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// SuccessResponse creates a successful API response
func SuccessResponse(data interface{}, message string) APIResponse {
	return APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// ErrorResponse creates an error API response
func ErrorResponse(code, message string, details map[string]interface{}) APIResponse {
	return APIResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// PaginatedSuccessResponse creates a successful paginated response
func PaginatedSuccessResponse(data interface{}, pagination PaginationMetadata, message string) PaginatedResponse {
	return PaginatedResponse{
		Success:    true,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	}
}

// PaginatedErrorResponse creates an error paginated response
func PaginatedErrorResponse(code, message string, details map[string]interface{}) PaginatedResponse {
	return PaginatedResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}

// NewMessageResponse creates a new message response
func NewMessageResponse(message string) MessageResponse {
	return MessageResponse{
		Message: message,
	}
}
