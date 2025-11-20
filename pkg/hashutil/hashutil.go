package hashutil

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
)

var (
	// Secret key for HMAC-SHA256 - loaded from environment
	hmacSecret []byte
)

// InitializeHMACSecret initializes the HMAC secret from environment variable
// This should be called once during application startup
func InitializeHMACSecret() error {
	secret := os.Getenv("HMAC_SECRET")
	if secret == "" {
		return fmt.Errorf("HMAC_SECRET environment variable not set")
	}
	hmacSecret = []byte(secret)
	return nil
}

// SetHMACSecret sets the HMAC secret (for testing purposes)
func SetHMACSecret(secret string) {
	hmacSecret = []byte(secret)
}

// HMACHash creates a deterministic HMAC-SHA256 hash of the input
// This is suitable for token and code hashing where lookup is required
func HMACHash(data string) (string, error) {
	if len(hmacSecret) == 0 {
		return "", fmt.Errorf("HMAC secret not initialized - call InitializeHMACSecret first")
	}

	h := hmac.New(sha256.New, hmacSecret)
	h.Write([]byte(data))
	hash := h.Sum(nil)
	return hex.EncodeToString(hash), nil
}

// VerifyHMACHash verifies that the input matches the hash
func VerifyHMACHash(input, expectedHash string) (bool, error) {
	actualHash, err := HMACHash(input)
	if err != nil {
		return false, err
	}
	return hmac.Equal([]byte(actualHash), []byte(expectedHash)), nil
}

// SHA256Hash creates a simple SHA256 hash (for non-secret data like user agent, IP)
// This doesn't require a secret key
func SHA256Hash(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	hash := h.Sum(nil)
	return hex.EncodeToString(hash)
}
