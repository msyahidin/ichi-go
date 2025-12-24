package authenticator

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// LoadPrivateKey loads a private key from a file path based on the signing method
// Supports RSA and ECDSA signing methods
func LoadPrivateKey(path string, method jwt.SigningMethod) (interface{}, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	return ParsePrivateKey(keyData, method)
}

// LoadPublicKey loads a public key from a file path based on the signing method
// Supports RSA and ECDSA signing methods
func LoadPublicKey(path string, method jwt.SigningMethod) (interface{}, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	return ParsePublicKey(keyData, method)
}

// ParsePrivateKey parses private key bytes based on the signing method
func ParsePrivateKey(keyData []byte, method jwt.SigningMethod) (interface{}, error) {
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	switch method.(type) {
	case *jwt.SigningMethodRSA, *jwt.SigningMethodRSAPSS:
		return parseRSAPrivateKey(block.Bytes)
	case *jwt.SigningMethodECDSA:
		return parseECDSAPrivateKey(block.Bytes)
	case *jwt.SigningMethodHMAC:
		// For HMAC, the key is just the raw bytes (no PEM encoding needed)
		return keyData, nil
	default:
		return nil, fmt.Errorf("unsupported signing method: %s", method.Alg())
	}
}

// ParsePublicKey parses public key bytes based on the signing method
func ParsePublicKey(keyData []byte, method jwt.SigningMethod) (interface{}, error) {
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	switch method.(type) {
	case *jwt.SigningMethodRSA, *jwt.SigningMethodRSAPSS:
		return parseRSAPublicKey(block.Bytes)
	case *jwt.SigningMethodECDSA:
		return parseECDSAPublicKey(block.Bytes)
	case *jwt.SigningMethodHMAC:
		// For HMAC, the key is just the raw bytes
		return keyData, nil
	default:
		return nil, fmt.Errorf("unsupported signing method: %s", method.Alg())
	}
}

// parseRSAPrivateKey parses RSA private key from DER encoded bytes
func parseRSAPrivateKey(der []byte) (*rsa.PrivateKey, error) {
	// Try PKCS1 format first
	key, err := x509.ParsePKCS1PrivateKey(der)
	if err == nil {
		return key, nil
	}

	// Try PKCS8 format
	keyInterface, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
	}

	rsaKey, ok := keyInterface.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not RSA private key")
	}

	return rsaKey, nil
}

// parseRSAPublicKey parses RSA public key from DER encoded bytes
func parseRSAPublicKey(der []byte) (*rsa.PublicKey, error) {
	// Try PKIX format first (most common)
	keyInterface, err := x509.ParsePKIXPublicKey(der)
	if err == nil {
		rsaKey, ok := keyInterface.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("key is not RSA public key")
		}
		return rsaKey, nil
	}

	// Try PKCS1 format
	rsaKey, err := x509.ParsePKCS1PublicKey(der)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
	}

	return rsaKey, nil
}

// parseECDSAPrivateKey parses ECDSA private key from DER encoded bytes
func parseECDSAPrivateKey(der []byte) (*ecdsa.PrivateKey, error) {
	// Try SEC1 format first
	key, err := x509.ParseECPrivateKey(der)
	if err == nil {
		return key, nil
	}

	// Try PKCS8 format
	keyInterface, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ECDSA private key: %w", err)
	}

	ecdsaKey, ok := keyInterface.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not ECDSA private key")
	}

	return ecdsaKey, nil
}

// parseECDSAPublicKey parses ECDSA public key from DER encoded bytes
func parseECDSAPublicKey(der []byte) (*ecdsa.PublicKey, error) {
	keyInterface, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ECDSA public key: %w", err)
	}

	ecdsaKey, ok := keyInterface.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key is not ECDSA public key")
	}

	return ecdsaKey, nil
}
