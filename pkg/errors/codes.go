package errors

// Auth domain error codes
const (
	ErrCodeInvalidCredentials = "AUTH_INVALID_CREDENTIALS"
	ErrCodeUserExists         = "AUTH_USER_EXISTS"
	ErrCodeUserNotFound       = "AUTH_USER_NOT_FOUND"
	ErrCodeInvalidToken       = "AUTH_INVALID_TOKEN"
	ErrCodeTokenExpired       = "AUTH_TOKEN_EXPIRED"
	ErrCodePasswordWeak       = "AUTH_PASSWORD_WEAK"
)

// User domain error codes
const (
	ErrCodeUserCreateFailed = "USER_CREATE_FAILED"
	ErrCodeUserUpdateFailed = "USER_UPDATE_FAILED"
	ErrCodeUserDeleteFailed = "USER_DELETE_FAILED"
)

// Infrastructure error codes
const (
	ErrCodeDatabase     = "DATABASE_ERROR"
	ErrCodeCache        = "CACHE_ERROR"
	ErrCodeQueue        = "QUEUE_ERROR"
	ErrCodeInternal     = "INTERNAL_ERROR"
	ErrCodeValidation   = "VALIDATION_ERROR"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeUnauthorized = "UNAUTHORIZED"
)
