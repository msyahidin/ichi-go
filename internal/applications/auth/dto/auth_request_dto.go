package auth

// LoginRequest represents login credentials
// @Description Login credentials for user authentication
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=6" example:"Password123!"`
}

// RegisterRequest represents user registration data
// @Description User registration information
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100" example:"John Doe"`
	Email    string `json:"email" validate:"required,email" example:"john.doe@example.com"`
	Password string `json:"password" validate:"required,strong_password" example:"SecurePass123!@#"`
}

// RefreshTokenRequest represents refresh token data
// @Description Refresh token for obtaining new access token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// ChangePasswordRequest represents password change data
// @Description Password change information
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required,min=6" example:"OldPass123!"`
	NewPassword string `json:"new_password" validate:"required,strong_password" example:"NewSecurePass123!@#"`
}
