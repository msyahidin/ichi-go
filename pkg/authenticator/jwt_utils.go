package authenticator

import (
	"fmt"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

// GetUserId extracts user ID from JWT claims
// This is the existing function, kept for backward compatibility
func GetUserId(claims jwt.Claims) (UserSubject, error) {
	subj, er := claims.GetSubject()
	us := UserSubject{ID: 0}
	if er != nil {
		return us, jwt.ErrTokenInvalidSubject
	}
	userId, uError := strconv.ParseUint(subj, 10, 64)
	if uError != nil {
		return us, jwt.ErrTokenInvalidSubject
	}
	us.ID = userId
	return us, nil
}

// GetUserIdFromMapClaims extracts user ID from jwt.MapClaims
// This is useful when working with custom claims
func GetUserIdFromMapClaims(claims jwt.MapClaims) (UserSubject, error) {
	us := UserSubject{ID: 0}

	sub, ok := claims["sub"]
	if !ok {
		return us, jwt.ErrTokenInvalidSubject
	}

	// Handle both string and numeric subject
	var subStr string
	switch v := sub.(type) {
	case string:
		subStr = v
	case float64:
		subStr = strconv.FormatUint(uint64(v), 10)
	case int64:
		subStr = strconv.FormatInt(v, 10)
	case uint64:
		subStr = strconv.FormatUint(v, 10)
	default:
		return us, fmt.Errorf("invalid subject type: %T", sub)
	}

	userId, err := strconv.ParseUint(subStr, 10, 64)
	if err != nil {
		return us, jwt.ErrTokenInvalidSubject
	}

	us.ID = userId
	return us, nil
}

// GetClaimString extracts a string claim from MapClaims
func GetClaimString(claims jwt.MapClaims, key string) (string, error) {
	value, ok := claims[key]
	if !ok {
		return "", fmt.Errorf("claim '%s' not found", key)
	}

	strValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("claim '%s' is not a string", key)
	}

	return strValue, nil
}

// GetClaimInt extracts an integer claim from MapClaims
func GetClaimInt(claims jwt.MapClaims, key string) (int64, error) {
	value, ok := claims[key]
	if !ok {
		return 0, fmt.Errorf("claim '%s' not found", key)
	}

	switch v := value.(type) {
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("claim '%s' is not a number", key)
	}
}

// GetClaimBool extracts a boolean claim from MapClaims
func GetClaimBool(claims jwt.MapClaims, key string) (bool, error) {
	value, ok := claims[key]
	if !ok {
		return false, fmt.Errorf("claim '%s' not found", key)
	}

	boolValue, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("claim '%s' is not a boolean", key)
	}

	return boolValue, nil
}

// GetClaimStringSlice extracts a string slice claim from MapClaims
func GetClaimStringSlice(claims jwt.MapClaims, key string) ([]string, error) {
	value, ok := claims[key]
	if !ok {
		return nil, fmt.Errorf("claim '%s' not found", key)
	}

	// Handle single string
	if strValue, ok := value.(string); ok {
		return []string{strValue}, nil
	}

	// Handle array
	arrayValue, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("claim '%s' is not a string array", key)
	}

	result := make([]string, 0, len(arrayValue))
	for i, item := range arrayValue {
		strItem, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("claim '%s' array element at index %d is not a string", key, i)
		}
		result = append(result, strItem)
	}

	return result, nil
}

// HasClaim checks if a claim exists in MapClaims
func HasClaim(claims jwt.MapClaims, key string) bool {
	_, ok := claims[key]
	return ok
}

// GetAudience extracts audience claim (handles both string and []string)
func GetAudience(claims jwt.MapClaims) ([]string, error) {
	aud, ok := claims["aud"]
	if !ok {
		return nil, fmt.Errorf("audience claim not found")
	}

	// Handle single string
	if strValue, ok := aud.(string); ok {
		return []string{strValue}, nil
	}

	// Handle array
	arrayValue, ok := aud.([]interface{})
	if !ok {
		return nil, fmt.Errorf("audience claim is not a string or array")
	}

	result := make([]string, 0, len(arrayValue))
	for _, item := range arrayValue {
		strItem, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("audience array contains non-string element")
		}
		result = append(result, strItem)
	}

	return result, nil
}
