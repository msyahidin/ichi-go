// Package testutil provides test fixtures and builders for common test data
package testutil

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ============================================================================
// Time Fixtures
// ============================================================================

// FixedTime returns a fixed time for consistent testing
var FixedTime = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

// FixedTimePtr returns a pointer to a fixed time
func FixedTimePtr() *time.Time {
	t := FixedTime
	return &t
}

// NowUTC returns current time in UTC (wrapper for test consistency)
func NowUTC() time.Time {
	return time.Now().UTC()
}

// ============================================================================
// ID Generators
// ============================================================================

// RandomUUID generates a random UUID string
func RandomUUID() string {
	return uuid.New().String()
}

// RandomUint64 generates a random uint64 ID
func RandomUint64() uint64 {
	return uint64(time.Now().UnixNano())
}

// ============================================================================
// String Builders
// ============================================================================

// RandomEmail generates a random email address for testing
func RandomEmail() string {
	return "user_" + RandomUUID() + "@example.com"
}

// RandomUsername generates a random username for testing
func RandomUsername() string {
	return "user_" + uuid.New().String()[:8]
}

// RandomString generates a random string of specified length
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// ============================================================================
// Password Helpers
// ============================================================================

// DefaultTestPassword is a common password used in tests
const DefaultTestPassword = "Password123!"

// HashPassword creates a bcrypt hash of a password
func HashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err) // This should never happen in tests
	}
	return string(hash)
}

// ValidPassword returns a valid test password
func ValidPassword() string {
	return DefaultTestPassword
}

// ============================================================================
// User Fixtures
// ============================================================================

// UserFixture represents a test user
type UserFixture struct {
	ID        uint64
	Email     string
	Username  string
	Password  string
	HashedPwd string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUserFixture creates a new user fixture with random data
func NewUserFixture() *UserFixture {
	password := DefaultTestPassword
	return &UserFixture{
		ID:        RandomUint64(),
		Email:     RandomEmail(),
		Username:  RandomUsername(),
		Password:  password,
		HashedPwd: HashPassword(password),
		CreatedAt: FixedTime,
		UpdatedAt: FixedTime,
	}
}

// WithID sets the user ID
func (u *UserFixture) WithID(id uint64) *UserFixture {
	u.ID = id
	return u
}

// WithEmail sets the user email
func (u *UserFixture) WithEmail(email string) *UserFixture {
	u.Email = email
	return u
}

// WithUsername sets the username
func (u *UserFixture) WithUsername(username string) *UserFixture {
	u.Username = username
	return u
}

// WithPassword sets the password and generates hash
func (u *UserFixture) WithPassword(password string) *UserFixture {
	u.Password = password
	u.HashedPwd = HashPassword(password)
	return u
}

// ============================================================================
// Token Fixtures
// ============================================================================

// TokenFixture represents a test JWT token pair
type TokenFixture struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	UserID       uint64
}

// NewTokenFixture creates a new token fixture
func NewTokenFixture() *TokenFixture {
	return &TokenFixture{
		AccessToken:  "test_access_token_" + RandomUUID(),
		RefreshToken: "test_refresh_token_" + RandomUUID(),
		ExpiresIn:    3600,
		UserID:       RandomUint64(),
	}
}

// WithAccessToken sets the access token
func (t *TokenFixture) WithAccessToken(token string) *TokenFixture {
	t.AccessToken = token
	return t
}

// WithRefreshToken sets the refresh token
func (t *TokenFixture) WithRefreshToken(token string) *TokenFixture {
	t.RefreshToken = token
	return t
}

// WithUserID sets the user ID
func (t *TokenFixture) WithUserID(userID uint64) *TokenFixture {
	t.UserID = userID
	return t
}

// ============================================================================
// Order Fixtures (Example for domain testing)
// ============================================================================

// OrderFixture represents a test order
type OrderFixture struct {
	ID         uint64
	UserID     uint64
	Status     string
	TotalPrice float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewOrderFixture creates a new order fixture
func NewOrderFixture() *OrderFixture {
	return &OrderFixture{
		ID:         RandomUint64(),
		UserID:     RandomUint64(),
		Status:     "pending",
		TotalPrice: 100.00,
		CreatedAt:  FixedTime,
		UpdatedAt:  FixedTime,
	}
}

// WithID sets the order ID
func (o *OrderFixture) WithID(id uint64) *OrderFixture {
	o.ID = id
	return o
}

// WithUserID sets the user ID
func (o *OrderFixture) WithUserID(userID uint64) *OrderFixture {
	o.UserID = userID
	return o
}

// WithStatus sets the order status
func (o *OrderFixture) WithStatus(status string) *OrderFixture {
	o.Status = status
	return o
}

// WithTotalPrice sets the total price
func (o *OrderFixture) WithTotalPrice(price float64) *OrderFixture {
	o.TotalPrice = price
	return o
}

// ============================================================================
// DTO Builders (for testing DTOs)
// ============================================================================

// LoginDTOBuilder helps build login DTOs for testing
type LoginDTOBuilder struct {
	email    string
	password string
}

// NewLoginDTOBuilder creates a new login DTO builder
func NewLoginDTOBuilder() *LoginDTOBuilder {
	return &LoginDTOBuilder{
		email:    RandomEmail(),
		password: DefaultTestPassword,
	}
}

// WithEmail sets the email
func (b *LoginDTOBuilder) WithEmail(email string) *LoginDTOBuilder {
	b.email = email
	return b
}

// WithPassword sets the password
func (b *LoginDTOBuilder) WithPassword(password string) *LoginDTOBuilder {
	b.password = password
	return b
}

// Build returns the email and password
func (b *LoginDTOBuilder) Build() (string, string) {
	return b.email, b.password
}

// ============================================================================
// Register DTO Builder
// ============================================================================

// RegisterDTOBuilder helps build registration DTOs for testing
type RegisterDTOBuilder struct {
	email    string
	username string
	password string
}

// NewRegisterDTOBuilder creates a new register DTO builder
func NewRegisterDTOBuilder() *RegisterDTOBuilder {
	return &RegisterDTOBuilder{
		email:    RandomEmail(),
		username: RandomUsername(),
		password: DefaultTestPassword,
	}
}

// WithEmail sets the email
func (b *RegisterDTOBuilder) WithEmail(email string) *RegisterDTOBuilder {
	b.email = email
	return b
}

// WithUsername sets the username
func (b *RegisterDTOBuilder) WithUsername(username string) *RegisterDTOBuilder {
	b.username = username
	return b
}

// WithPassword sets the password
func (b *RegisterDTOBuilder) WithPassword(password string) *RegisterDTOBuilder {
	b.password = password
	return b
}

// Build returns the email, username, and password
func (b *RegisterDTOBuilder) Build() (string, string, string) {
	return b.email, b.username, b.password
}

// ============================================================================
// Validation Test Cases
// ============================================================================

// ValidationTestCase represents a validation test case
type ValidationTestCase struct {
	Name        string
	Input       interface{}
	ExpectError bool
	ErrorField  string // Expected field that should have error
	ErrorTag    string // Expected validation tag
}

// InvalidEmailCases returns common invalid email test cases
func InvalidEmailCases() []string {
	return []string{
		"",
		"invalid",
		"@example.com",
		"user@",
		"user space@example.com",
		"user@.com",
	}
}

// InvalidPasswordCases returns common invalid password test cases
func InvalidPasswordCases() []string {
	return []string{
		"",           // empty
		"short",      // too short
		"password",   // no uppercase
		"PASSWORD",   // no lowercase
		"Password",   // no number
		"Pass123",    // no special char (depending on validation rules)
	}
}

// ============================================================================
// Concurrency Test Helpers
// ============================================================================

// ConcurrentRunner helps run functions concurrently for testing
type ConcurrentRunner struct {
	workers int
	fn      func(workerID int)
}

// NewConcurrentRunner creates a new concurrent runner
func NewConcurrentRunner(workers int, fn func(workerID int)) *ConcurrentRunner {
	return &ConcurrentRunner{
		workers: workers,
		fn:      fn,
	}
}

// Run executes the function concurrently
func (r *ConcurrentRunner) Run() {
	done := make(chan bool, r.workers)

	for i := 0; i < r.workers; i++ {
		go func(workerID int) {
			r.fn(workerID)
			done <- true
		}(i)
	}

	for i := 0; i < r.workers; i++ {
		<-done
	}
}

// ============================================================================
// Mock Response Builders
// ============================================================================

// MockHTTPResponse represents a mock HTTP response for testing
type MockHTTPResponse struct {
	StatusCode int
	Body       []byte
	Headers    map[string]string
	Error      error
}

// NewMockHTTPResponse creates a new mock HTTP response
func NewMockHTTPResponse(statusCode int, body []byte) *MockHTTPResponse {
	return &MockHTTPResponse{
		StatusCode: statusCode,
		Body:       body,
		Headers:    make(map[string]string),
	}
}

// WithHeader adds a header to the response
func (r *MockHTTPResponse) WithHeader(key, value string) *MockHTTPResponse {
	r.Headers[key] = value
	return r
}

// WithError sets an error
func (r *MockHTTPResponse) WithError(err error) *MockHTTPResponse {
	r.Error = err
	return r
}
