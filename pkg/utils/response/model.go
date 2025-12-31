package response

// SuccessResponse represents a successful API response
// This is the standard wrapper for all successful API responses
// @Description Standard success response wrapper
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data" swaggertype:"object" description:"Response payload - actual type varies by endpoint"`
	Message string      `json:"message,omitempty" example:"Operation completed successfully"`
}

// ErrorResponse represents an error API response
// This is the standard wrapper for all error responses
// @Description Standard error response wrapper
type ErrorResponse struct {
	Success bool        `json:"success" example:"false"`
	Error   ErrorDetail `json:"error" description:"Error details"`
}

// ErrorDetail contains detailed error information
// @Description Detailed error information
type ErrorDetail struct {
	Code    string                 `json:"code" example:"VALIDATION_ERROR"`
	Message string                 `json:"message" example:"Validation failed"`
	Details map[string]interface{} `json:"details,omitempty" swaggertype:"object" description:"Additional error context (e.g., field-level validation errors)"`
}

// ValidationErrorResponse represents validation error with field-specific details
// @Description Validation error response with field-level errors
type ValidationErrorResponse struct {
	Success bool            `json:"success" example:"false"`
	Error   ValidationError `json:"error"`
}

// ValidationError contains field-level validation errors
// @Description Field-level validation errors
type ValidationError struct {
	Code    string                 `json:"code" example:"VALIDATION_ERROR"`
	Message string                 `json:"message" example:"Validation failed"`
	Fields  map[string]interface{} `json:"fields" swaggertype:"object" description:"Map of field names to error messages" example:"email:must be a valid email,password:password too weak"`
}

// UnauthorizedErrorResponse represents authentication/authorization errors
// @Description Unauthorized access error
type UnauthorizedErrorResponse struct {
	Success bool              `json:"success" example:"false"`
	Error   UnauthorizedError `json:"error"`
}

// UnauthorizedError contains authentication error details
// @Description Authentication/authorization error details
type UnauthorizedError struct {
	Code    string `json:"code" example:"UNAUTHORIZED"`
	Message string `json:"message" example:"Invalid credentials"`
}

// NotFoundErrorResponse represents resource not found errors
// @Description Resource not found error
type NotFoundErrorResponse struct {
	Success bool          `json:"success" example:"false"`
	Error   NotFoundError `json:"error"`
}

// NotFoundError contains not found error details
// @Description Resource not found error details
type NotFoundError struct {
	Code    string `json:"code" example:"NOT_FOUND"`
	Message string `json:"message" example:"Resource not found"`
}

// InternalErrorResponse represents internal server errors
// @Description Internal server error
type InternalErrorResponse struct {
	Success bool          `json:"success" example:"false"`
	Error   InternalError `json:"error"`
}

// InternalError contains internal error details
// @Description Internal server error details
type InternalError struct {
	Code    string `json:"code" example:"INTERNAL_ERROR"`
	Message string `json:"message" example:"An internal error occurred"`
}

// Example responses for documentation purposes
// These won't be used in code, just for Swagger examples

// ExampleSuccessResponse shows a typical success response
var ExampleSuccessResponse = SuccessResponse{
	Success: true,
	Data: map[string]interface{}{
		"id":   1,
		"name": "Example",
	},
	Message: "Operation completed successfully",
}

// ExampleValidationError shows a typical validation error
var ExampleValidationError = ValidationErrorResponse{
	Success: false,
	Error: ValidationError{
		Code:    "VALIDATION_ERROR",
		Message: "Validation failed",
		Fields: map[string]interface{}{
			"email":    "Email must be a valid email address",
			"password": "Password must contain at least 8 characters, including uppercase, lowercase, number, and special character",
		},
	},
}

// ExampleUnauthorizedError shows a typical unauthorized error
var ExampleUnauthorizedError = UnauthorizedErrorResponse{
	Success: false,
	Error: UnauthorizedError{
		Code:    "UNAUTHORIZED",
		Message: "Invalid credentials",
	},
}

// ExampleNotFoundError shows a typical not found error
var ExampleNotFoundError = NotFoundErrorResponse{
	Success: false,
	Error: NotFoundError{
		Code:    "NOT_FOUND",
		Message: "User not found",
	},
}

// ExampleInternalError shows a typical internal error
var ExampleInternalError = InternalErrorResponse{
	Success: false,
	Error: InternalError{
		Code:    "INTERNAL_ERROR",
		Message: "An internal error occurred",
	},
}
