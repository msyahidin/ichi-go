package authenticator

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test Helper: Generate RSA key pair for testing
func generateTestRSAKeys(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return privateKey, &privateKey.PublicKey
}

// Test Helper: Save key to temporary file
func saveKeyToTempFile(t *testing.T, keyPEM []byte) string {
	tmpfile, err := os.CreateTemp("", "test-key-*.pem")
	require.NoError(t, err)
	t.Cleanup(func() { os.Remove(tmpfile.Name()) })

	_, err = tmpfile.Write(keyPEM)
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())

	return tmpfile.Name()
}

// Test 1: Key Loading - RSA
func TestLoadRSAKeys(t *testing.T) {
	privateKey, publicKey := generateTestRSAKeys(t)

	// Encode private key
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	privateKeyPath := saveKeyToTempFile(t, privateKeyPEM)

	// Encode public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	require.NoError(t, err)
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	publicKeyPath := saveKeyToTempFile(t, publicKeyPEM)

	// Test loading
	loadedPrivateKey, err := LoadPrivateKey(privateKeyPath, jwt.SigningMethodRS256)
	assert.NoError(t, err)
	assert.NotNil(t, loadedPrivateKey)

	loadedPublicKey, err := LoadPublicKey(publicKeyPath, jwt.SigningMethodRS256)
	assert.NoError(t, err)
	assert.NotNil(t, loadedPublicKey)
}

// Test 2: Token Generation - HMAC
func TestGenerateTokenHMAC(t *testing.T) {
	config := JWTConfig{
		SigningMethod:  jwt.SigningMethodHS256,
		SecretKey:      []byte("test-secret-key-minimum-32-chars-long"),
		AccessTokenTTL: 15 * time.Minute,
		Issuer:         "test-app",
		Audience:       []string{"test-users"},
	}

	// Generate token
	token, err := GenerateAccessToken(12345, config)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token can be parsed
	parsedToken, err := ParseToken(token, config)
	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)
}

// Test 3: Token Generation - RSA
func TestGenerateTokenRSA(t *testing.T) {
	privateKey, publicKey := generateTestRSAKeys(t)

	config := JWTConfig{
		SigningMethod:  jwt.SigningMethodRS256,
		PrivateKey:     privateKey,
		PublicKey:      publicKey,
		AccessTokenTTL: 15 * time.Minute,
		Issuer:         "test-app",
		Audience:       []string{"test-users"},
	}

	// Generate token
	token, err := GenerateAccessToken(12345, config)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token can be parsed
	parsedToken, err := ParseToken(token, config)
	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)
}

// Test 4: Custom Claims
func TestCreateCustomClaims(t *testing.T) {
	input := StandardClaimsInput{
		UserID:    12345,
		Issuer:    "test-app",
		Audience:  []string{"test-users"},
		ExpiresIn: 15 * time.Minute,
		CustomClaims: map[string]interface{}{
			"email": "test@example.com",
			"role":  "admin",
		},
	}

	claims := CreateCustomClaims(input)

	// Verify standard claims
	assert.Equal(t, "12345", claims["sub"])
	assert.Equal(t, "test-app", claims["iss"])
	assert.Equal(t, "test-users", claims["aud"])
	assert.NotNil(t, claims["exp"])
	assert.NotNil(t, claims["iat"])

	// Verify custom claims
	assert.Equal(t, "test@example.com", claims["email"])
	assert.Equal(t, "admin", claims["role"])
}

// Test 5: Token Validation - Success
func TestValidateToken_Success(t *testing.T) {
	config := JWTConfig{
		SigningMethod:  jwt.SigningMethodHS256,
		SecretKey:      []byte("test-secret-key-minimum-32-chars-long"),
		AccessTokenTTL: 15 * time.Minute,
		Issuer:         "test-app",
		Audience:       []string{"test-users"},
		RequireClaims:  []string{"sub", "exp"},
	}

	// Generate valid token
	token, err := GenerateAccessToken(12345, config)
	require.NoError(t, err)

	// Parse token
	parsedToken, err := ParseToken(token, config)
	require.NoError(t, err)

	// Validate token
	claims, err := ValidateToken(parsedToken, config)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
}

// Test 6: Token Validation - Expired
func TestValidateToken_Expired(t *testing.T) {
	config := JWTConfig{
		SigningMethod:  jwt.SigningMethodHS256,
		SecretKey:      []byte("test-secret-key-minimum-32-chars-long"),
		AccessTokenTTL: -1 * time.Hour, // Already expired
		Issuer:         "test-app",
	}

	// Generate expired token
	token, err := GenerateAccessToken(12345, config)
	require.NoError(t, err)

	// Try to parse - should fail
	_, err = ParseToken(token, config)
	assert.Error(t, err)
}

// Test 7: Token Validation - Missing Required Claims
func TestValidateToken_MissingRequiredClaims(t *testing.T) {
	config := JWTConfig{
		SigningMethod:  jwt.SigningMethodHS256,
		SecretKey:      []byte("test-secret-key-minimum-32-chars-long"),
		AccessTokenTTL: 15 * time.Minute,
		RequireClaims:  []string{"sub", "exp", "email"}, // email is required but not present
	}

	// Generate token without email claim
	token, err := GenerateAccessToken(12345, config)
	require.NoError(t, err)

	// Parse token (should work)
	parsedToken, err := ParseToken(token, config)
	require.NoError(t, err)

	// Validate token (should fail due to missing email claim)
	_, err = ValidateToken(parsedToken, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email")
}

// Test 8: Custom Claims Validator
func TestCustomClaimsValidator(t *testing.T) {
	config := JWTConfig{
		SigningMethod:  jwt.SigningMethodHS256,
		SecretKey:      []byte("test-secret-key-minimum-32-chars-long"),
		AccessTokenTTL: 15 * time.Minute,
		ClaimsValidator: func(claims jwt.MapClaims) error {
			role, ok := claims["role"].(string)
			if !ok || role != "admin" {
				return jwt.ErrTokenInvalidClaims
			}
			return nil
		},
	}

	// Test 1: Token without role claim (should fail)
	token1, err := GenerateAccessToken(12345, config)
	require.NoError(t, err)

	parsedToken1, err := ParseToken(token1, config)
	require.NoError(t, err)

	_, err = ValidateToken(parsedToken1, config)
	assert.Error(t, err)

	// Test 2: Token with admin role (should succeed)
	input := StandardClaimsInput{
		UserID:    12345,
		ExpiresIn: 15 * time.Minute,
		CustomClaims: map[string]interface{}{
			"role": "admin",
		},
	}
	claims := CreateCustomClaims(input)
	token2, err := GenerateToken(claims, config)
	require.NoError(t, err)

	parsedToken2, err := ParseToken(token2, config)
	require.NoError(t, err)

	validatedClaims, err := ValidateToken(parsedToken2, config)
	assert.NoError(t, err)
	assert.Equal(t, "admin", validatedClaims["role"])
}

// Test 9: Extract User ID
func TestGetUserId(t *testing.T) {
	config := JWTConfig{
		SigningMethod:  jwt.SigningMethodHS256,
		SecretKey:      []byte("test-secret-key-minimum-32-chars-long"),
		AccessTokenTTL: 15 * time.Minute,
	}

	// Generate token
	expectedUserID := uint64(12345)
	token, err := GenerateAccessToken(expectedUserID, config)
	require.NoError(t, err)

	// Parse and validate
	parsedToken, err := ParseToken(token, config)
	require.NoError(t, err)

	claims, err := ValidateToken(parsedToken, config)
	require.NoError(t, err)

	// Extract user ID
	user, err := GetUserIdFromMapClaims(claims)
	assert.NoError(t, err)
	assert.Equal(t, expectedUserID, user.ID)
}

// Test 10: Token Pair Generation
func TestGenerateTokenPair(t *testing.T) {
	config := JWTConfig{
		SigningMethod:   jwt.SigningMethodHS256,
		SecretKey:       []byte("test-secret-key-minimum-32-chars-long"),
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		AuthScheme:      "Bearer",
	}

	// Generate token pair
	tokenPair, err := GenerateTokenPair(12345, config)
	assert.NoError(t, err)
	assert.NotNil(t, tokenPair)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.Equal(t, "Bearer", tokenPair.TokenType)
	assert.Equal(t, int64(900), tokenPair.ExpiresIn) // 15 minutes = 900 seconds

	// Verify both tokens are valid
	accessToken, err := ParseToken(tokenPair.AccessToken, config)
	assert.NoError(t, err)
	assert.True(t, accessToken.Valid)

	refreshToken, err := ParseToken(tokenPair.RefreshToken, config)
	assert.NoError(t, err)
	assert.True(t, refreshToken.Valid)
}

// Test 11: Helper Functions - Get Claim String
func TestGetClaimString(t *testing.T) {
	claims := jwt.MapClaims{
		"email": "test@example.com",
		"role":  "admin",
	}

	// Test existing claim
	email, err := GetClaimString(claims, "email")
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", email)

	// Test missing claim
	_, err = GetClaimString(claims, "missing")
	assert.Error(t, err)
}

// Test 12: Helper Functions - Get Claim Int
func TestGetClaimInt(t *testing.T) {
	claims := jwt.MapClaims{
		"age":   float64(25),
		"count": int64(100),
	}

	// Test float64 (default JWT number type)
	age, err := GetClaimInt(claims, "age")
	assert.NoError(t, err)
	assert.Equal(t, int64(25), age)

	// Test int64
	count, err := GetClaimInt(claims, "count")
	assert.NoError(t, err)
	assert.Equal(t, int64(100), count)
}

// Test 13: Helper Functions - Get Audience
func TestGetAudience(t *testing.T) {
	// Test single string audience
	claims1 := jwt.MapClaims{
		"aud": "test-users",
	}
	aud1, err := GetAudience(claims1)
	assert.NoError(t, err)
	assert.Equal(t, []string{"test-users"}, aud1)

	// Test array audience
	claims2 := jwt.MapClaims{
		"aud": []interface{}{"test-users", "test-admins"},
	}
	aud2, err := GetAudience(claims2)
	assert.NoError(t, err)
	assert.Equal(t, []string{"test-users", "test-admins"}, aud2)
}

// Test 14: Config Validation
func TestConfigValidation(t *testing.T) {
	// Test valid HMAC config
	validHMACConfig := &JWTConfig{
		Enabled:        true,
		SigningMethod:  jwt.SigningMethodHS256,
		SecretKey:      []byte("valid-secret-key"),
		TokenLookup:    "header:Authorization",
		AccessTokenTTL: 15 * time.Minute,
	}
	assert.NoError(t, validHMACConfig.Validate())

	// Test invalid HMAC config (missing secret)
	invalidHMACConfig := &JWTConfig{
		Enabled:        true,
		SigningMethod:  jwt.SigningMethodHS256,
		TokenLookup:    "header:Authorization",
		AccessTokenTTL: 15 * time.Minute,
	}
	assert.Error(t, invalidHMACConfig.Validate())

	// Test invalid TTL
	invalidTTLConfig := &JWTConfig{
		Enabled:        true,
		SigningMethod:  jwt.SigningMethodHS256,
		SecretKey:      []byte("valid-secret-key"),
		TokenLookup:    "header:Authorization",
		AccessTokenTTL: 0, // Invalid
	}
	assert.Error(t, invalidTTLConfig.Validate())
}

// Test 15: Extract Token from Echo Context
func TestExtractToken(t *testing.T) {
	config := JWTConfig{
		TokenLookup: "header:Authorization",
		AuthScheme:  "Bearer",
	}

	// Create Echo context with token
	e := echo.New()
	req := echo.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer test-token-12345")
	rec := echo.NewResponseRecorder()
	c := e.NewContext(req, rec)

	// Extract token
	token, err := ExtractToken(c, config)
	assert.NoError(t, err)
	assert.Equal(t, "test-token-12345", token)
}

// Test 16: Parse Token Lookup Configuration
func TestParseTokenLookup(t *testing.T) {
	// Test single lookup
	lookups1 := parseTokenLookup("header:Authorization")
	assert.Len(t, lookups1, 1)
	assert.Equal(t, "header", lookups1[0].source)
	assert.Equal(t, "Authorization", lookups1[0].key)

	// Test multiple lookups
	lookups2 := parseTokenLookup("header:Authorization,query:token,cookie:jwt")
	assert.Len(t, lookups2, 3)
	assert.Equal(t, "header", lookups2[0].source)
	assert.Equal(t, "query", lookups2[1].source)
	assert.Equal(t, "cookie", lookups2[2].source)

	// Test empty string (should default)
	lookups3 := parseTokenLookup("")
	assert.Len(t, lookups3, 1)
	assert.Equal(t, "header", lookups3[0].source)
	assert.Equal(t, "Authorization", lookups3[0].key)
}

// Benchmark: Token Generation
func BenchmarkGenerateAccessToken(b *testing.B) {
	config := JWTConfig{
		SigningMethod:  jwt.SigningMethodHS256,
		SecretKey:      []byte("test-secret-key-minimum-32-chars-long"),
		AccessTokenTTL: 15 * time.Minute,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateAccessToken(12345, config)
	}
}

// Benchmark: Token Parsing
func BenchmarkParseToken(b *testing.B) {
	config := JWTConfig{
		SigningMethod:  jwt.SigningMethodHS256,
		SecretKey:      []byte("test-secret-key-minimum-32-chars-long"),
		AccessTokenTTL: 15 * time.Minute,
	}

	token, _ := GenerateAccessToken(12345, config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseToken(token, config)
	}
}
