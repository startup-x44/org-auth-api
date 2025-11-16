package pkce

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
)

const (
	// CodeVerifierMinLength minimum length for code verifier
	CodeVerifierMinLength = 43
	// CodeVerifierMaxLength maximum length for code verifier
	CodeVerifierMaxLength = 128
)

// GenerateCodeVerifier generates a cryptographically secure random code verifier
func GenerateCodeVerifier() (string, error) {
	// Generate 32 random bytes (will produce 43 base64url characters)
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode as base64url without padding
	verifier := base64.RawURLEncoding.EncodeToString(randomBytes)
	return verifier, nil
}

// GenerateCodeChallenge generates S256 code challenge from verifier
func GenerateCodeChallenge(verifier string) (string, error) {
	if len(verifier) < CodeVerifierMinLength || len(verifier) > CodeVerifierMaxLength {
		return "", fmt.Errorf("code verifier length must be between %d and %d characters", CodeVerifierMinLength, CodeVerifierMaxLength)
	}

	// SHA256 hash of the verifier
	hash := sha256.Sum256([]byte(verifier))

	// Base64url encode without padding
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])
	return challenge, nil
}

// VerifyCodeChallenge verifies that the code verifier matches the code challenge using S256 method
func VerifyCodeChallenge(verifier, challenge, method string) error {
	if method != "S256" {
		return errors.New("only S256 code challenge method is supported")
	}

	if len(verifier) < CodeVerifierMinLength || len(verifier) > CodeVerifierMaxLength {
		return fmt.Errorf("code verifier length must be between %d and %d characters", CodeVerifierMinLength, CodeVerifierMaxLength)
	}

	// Generate challenge from verifier
	computedChallenge, err := GenerateCodeChallenge(verifier)
	if err != nil {
		return err
	}

	// Constant-time comparison
	if computedChallenge != challenge {
		return errors.New("code verifier does not match code challenge")
	}

	return nil
}

// GeneratePKCEPair generates both verifier and challenge for PKCE flow
func GeneratePKCEPair() (verifier, challenge string, err error) {
	verifier, err = GenerateCodeVerifier()
	if err != nil {
		return "", "", err
	}

	challenge, err = GenerateCodeChallenge(verifier)
	if err != nil {
		return "", "", err
	}

	return verifier, challenge, nil
}
