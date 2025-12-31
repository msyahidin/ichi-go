package auth

import "time"

// LoginResponse represents successful login response with tokens
// @Description Successful authentication response containing user info and JWT tokens
type LoginResponse struct {
	User         UserInfo `json:"user" description:"User profile information"`
	AccessToken  string   `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
	RefreshToken string   `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
	TokenType    string   `json:"token_type" example:"Bearer"`
	ExpiresIn    int64    `json:"expires_in" example:"3600" description:"Seconds until access token expires"`
}

// RegisterResponse represents successful registration response
// @Description Successful user registration response
type RegisterResponse struct {
	User         UserInfo `json:"user" description:"Newly created user information"`
	AccessToken  string   `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string   `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType    string   `json:"token_type" example:"Bearer"`
	ExpiresIn    int64    `json:"expires_in" example:"3600" description:"Seconds until access token expires"`
}

// RefreshTokenResponse represents successful token refresh
// @Description New JWT token pair
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType    string `json:"token_type" example:"Bearer"`
	ExpiresIn    int64  `json:"expires_in" example:"3600" description:"Seconds until access token expires"`
}

// UserInfo represents basic user information in auth responses
// @Description User profile information
type UserInfo struct {
	ID        uint64    `json:"id" example:"1"`
	Name      string    `json:"name" example:"John Doe"`
	Email     string    `json:"email" example:"john.doe@example.com"`
	CreatedAt time.Time `json:"created_at" example:"2025-01-01T00:00:00Z"`
}
