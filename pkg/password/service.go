package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

// PasswordService defines the interface for password operations
type PasswordService interface {
	Hash(password string) (string, error)
	HashWithoutValidation(password string) (string, error)
	Verify(password, hash string) (bool, error)
	ValidatePassword(password string) error
}

// Service handles password operations
type Service struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
}

// NewService creates a new password service with Argon2id parameters
func NewService() *Service {
	return &Service{
		time:    1,         // Number of iterations
		memory:  64 * 1024, // 64 MB
		threads: 4,         // Number of threads
		keyLen:  32,        // Length of the hash
	}
}

// Hash generates a hash from a password
func (s *Service) Hash(password string) (string, error) {
	if err := s.ValidatePassword(password); err != nil {
		return "", err
	}

	return s.HashWithoutValidation(password)
}

// HashWithoutValidation generates a hash from a password without validation
// This is useful for API keys and other machine-generated secrets
func (s *Service) HashWithoutValidation(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, s.time, s.memory, s.threads, s.keyLen)

	// Format: $argon2id$v=19$m=65536,t=1,p=4$salt$hash
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		s.memory, s.time, s.threads, encodedSalt, encodedHash), nil
}

// Verify checks if a password matches a hash
func (s *Service) Verify(password, hash string) (bool, error) {
	// Parse the hash format
	parts := strings.Split(hash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, fmt.Errorf("invalid hash format")
	}

	// Extract parameters
	params := strings.Split(parts[3], ",")
	if len(params) != 3 {
		return false, fmt.Errorf("invalid parameters")
	}

	var memory, time uint32
	var threads uint8

	for _, param := range params {
		kv := strings.Split(param, "=")
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "m":
			if val, err := strconv.ParseUint(kv[1], 10, 32); err == nil {
				memory = uint32(val)
			}
		case "t":
			if val, err := strconv.ParseUint(kv[1], 10, 32); err == nil {
				time = uint32(val)
			}
		case "p":
			if val, err := strconv.ParseUint(kv[1], 10, 8); err == nil {
				threads = uint8(val)
			}
		}
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	computedHash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(expectedHash)))

	return subtle.ConstantTimeCompare(computedHash, expectedHash) == 1, nil
}

// ValidatePassword validates password strength requirements
func (s *Service) ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	if len(password) > 128 {
		return fmt.Errorf("password must be at most 128 characters long")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case (char >= 32 && char <= 47) || (char >= 58 && char <= 64) || (char >= 91 && char <= 96) || (char >= 123 && char <= 126):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// NeedsRehash checks if a hash needs to be rehashed with current parameters
func (s *Service) NeedsRehash(hash string) bool {
	parts := strings.Split(hash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return true
	}

	params := strings.Split(parts[3], ",")
	if len(params) != 3 {
		return true
	}

	var memory, time uint32
	var threads uint8

	for _, param := range params {
		kv := strings.Split(param, "=")
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "m":
			if val, err := strconv.ParseUint(kv[1], 10, 32); err == nil {
				memory = uint32(val)
			}
		case "t":
			if val, err := strconv.ParseUint(kv[1], 10, 32); err == nil {
				time = uint32(val)
			}
		case "p":
			if val, err := strconv.ParseUint(kv[1], 10, 8); err == nil {
				threads = uint8(val)
			}
		}
	}

	return memory != s.memory || time != s.time || threads != s.threads
}
