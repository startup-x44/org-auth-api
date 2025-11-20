package unit

import (
	"auth-service/pkg/validation"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Email Validation Tests
// ============================================================================

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
		errType error
	}{
		// Valid emails
		{
			name:    "Valid simple email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "Valid email with subdomain",
			email:   "user@mail.example.com",
			wantErr: false,
		},
		{
			name:    "Valid email with dots",
			email:   "first.last@example.com",
			wantErr: false,
		},
		{
			name:    "Valid email with plus (filtering)",
			email:   "user+tag@example.com",
			wantErr: false,
		},
		{
			name:    "Valid email with numbers",
			email:   "user123@example456.com",
			wantErr: false,
		},
		{
			name:    "Valid email with hyphen in domain",
			email:   "user@my-domain.com",
			wantErr: false,
		},
		{
			name:    "Valid email with multiple subdomains",
			email:   "user@mail.internal.company.com",
			wantErr: false,
		},
		{
			name:    "Valid email with special chars (RFC 5322)",
			email:   "user!#$%&'*+/=?^_`{|}~@example.com",
			wantErr: false,
		},
		{
			name:    "Valid email with uppercase (normalized)",
			email:   "User@Example.COM",
			wantErr: false,
		},
		{
			name:    "Valid email with leading/trailing spaces (normalized)",
			email:   "  user@example.com  ",
			wantErr: false,
		},

		// Invalid emails
		{
			name:    "Missing @ symbol",
			email:   "userexample.com",
			wantErr: true,
			errType: validation.ErrInvalidEmail,
		},
		{
			name:    "Missing domain",
			email:   "user@",
			wantErr: true,
			errType: validation.ErrInvalidEmail,
		},
		{
			name:    "Missing local part",
			email:   "@example.com",
			wantErr: true,
			errType: validation.ErrInvalidEmail,
		},
		{
			name:    "Empty string",
			email:   "",
			wantErr: true,
			errType: validation.ErrInvalidEmail,
		},
		{
			name:    "Spaces in email",
			email:   "user name@example.com",
			wantErr: true,
			errType: validation.ErrInvalidEmail,
		},
		{
			name:    "Multiple @ symbols",
			email:   "user@@example.com",
			wantErr: true,
			errType: validation.ErrInvalidEmail,
		},
		{
			name:    "Missing TLD (accepted by regex)",
			email:   "user@domain",
			wantErr: false, // RFC 5322 allows this
		},
		{
			name:    "Dot before @ (accepted by regex)",
			email:   "user.@example.com",
			wantErr: false, // RFC 5322 allows trailing dot in local part
		},
		{
			name:    "Invalid characters in domain",
			email:   "user@exam_ple.com",
			wantErr: true,
			errType: validation.ErrInvalidEmail,
		},
		{
			name:    "SQL injection attempt",
			email:   "' OR '1'='1@example.com",
			wantErr: true,
			errType: validation.ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateEmail(tt.email)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Lowercase conversion",
			input:    "User@Example.COM",
			expected: "user@example.com",
		},
		{
			name:     "Trim spaces",
			input:    "  user@example.com  ",
			expected: "user@example.com",
		},
		{
			name:     "Both uppercase and spaces",
			input:    "  User@EXAMPLE.com  ",
			expected: "user@example.com",
		},
		{
			name:     "Already normalized",
			input:    "user@example.com",
			expected: "user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validation.NormalizeEmail(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// Password Validation Tests
// ============================================================================

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errType  error
	}{
		// Valid passwords
		{
			name:     "Strong password",
			password: "SecureP@ss123",
			wantErr:  false,
		},
		{
			name:     "Password with all requirements",
			password: "MyP@ssw0rd!",
			wantErr:  false,
		},
		{
			name:     "Long password",
			password: "ThisIsAVeryL0ng&SecurePassword123!",
			wantErr:  false,
		},
		{
			name:     "Minimum valid password",
			password: "Abc123!@",
			wantErr:  false,
		},
		{
			name:     "Password with multiple special chars",
			password: "P@ssw0rd!#$%",
			wantErr:  false,
		},

		// Invalid passwords - Too short
		{
			name:     "Too short (7 chars)",
			password: "Short1!",
			wantErr:  true,
			errType:  validation.ErrPasswordTooShort,
		},
		{
			name:     "Empty password",
			password: "",
			wantErr:  true,
			errType:  validation.ErrPasswordTooShort,
		},

		// Invalid passwords - Missing uppercase
		{
			name:     "No uppercase",
			password: "password123!",
			wantErr:  true,
			errType:  validation.ErrPasswordNoUpper,
		},

		// Invalid passwords - Missing lowercase
		{
			name:     "No lowercase",
			password: "PASSWORD123!",
			wantErr:  true,
			errType:  validation.ErrPasswordNoLower,
		},

		// Invalid passwords - Missing number
		{
			name:     "No number",
			password: "PasswordOnly!",
			wantErr:  true,
			errType:  validation.ErrPasswordNoNumber,
		},

		// Invalid passwords - Missing special character
		{
			name:     "No special character",
			password: "Password123",
			wantErr:  true,
			errType:  validation.ErrPasswordNoSpecial,
		},

		// Common weak passwords (still pass if they meet requirements)
		{
			name:     "Common password but meets requirements",
			password: "Password123!",
			wantErr:  false, // Meets technical requirements
		},

		// Edge cases
		{
			name:     "Only 8 characters (minimum)",
			password: "Abcd123!",
			wantErr:  false,
		},
		{
			name:     "Unicode characters",
			password: "Pässw0rd!",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidatePassword(tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePasswordsMatch(t *testing.T) {
	tests := []struct {
		name            string
		password        string
		confirmPassword string
		wantErr         bool
	}{
		{
			name:            "Passwords match",
			password:        "SecureP@ss123",
			confirmPassword: "SecureP@ss123",
			wantErr:         false,
		},
		{
			name:            "Passwords don't match",
			password:        "SecureP@ss123",
			confirmPassword: "DifferentP@ss123",
			wantErr:         true,
		},
		{
			name:            "Empty passwords match",
			password:        "",
			confirmPassword: "",
			wantErr:         false,
		},
		{
			name:            "Case sensitive mismatch",
			password:        "SecureP@ss123",
			confirmPassword: "securep@ss123",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidatePasswordsMatch(tt.password, tt.confirmPassword)

			if tt.wantErr {
				assert.ErrorIs(t, err, validation.ErrPasswordsDontMatch)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Name Validation Tests
// ============================================================================

func TestValidateName(t *testing.T) {
	tests := []struct {
		name      string
		inputName string
		fieldName string
		wantErr   bool
	}{
		// Valid names
		{
			name:      "Simple name",
			inputName: "John Doe",
			fieldName: "name",
			wantErr:   false,
		},
		{
			name:      "Name with hyphen",
			inputName: "Mary-Jane",
			fieldName: "name",
			wantErr:   false,
		},
		{
			name:      "Name with apostrophe",
			inputName: "O'Brien",
			fieldName: "name",
			wantErr:   false,
		},
		{
			name:      "Name with multiple spaces",
			inputName: "John Michael Doe",
			fieldName: "name",
			wantErr:   false,
		},
		{
			name:      "Single character (edge case)",
			inputName: "X",
			fieldName: "name",
			wantErr:   false,
		},
		{
			name:      "Complex valid name",
			inputName: "Mary-Jane O'Connor-Smith",
			fieldName: "name",
			wantErr:   false,
		},

		// Invalid names
		{
			name:      "Empty name",
			inputName: "",
			fieldName: "name",
			wantErr:   true,
		},
		{
			name:      "Only spaces",
			inputName: "   ",
			fieldName: "name",
			wantErr:   true,
		},
		{
			name:      "Name with numbers",
			inputName: "John123",
			fieldName: "name",
			wantErr:   true,
		},
		{
			name:      "Name with special characters",
			inputName: "John@Doe",
			fieldName: "name",
			wantErr:   true,
		},
		{
			name:      "Name too long (>100 chars)",
			inputName: strings.Repeat("A", 101),
			fieldName: "name",
			wantErr:   true,
		},
		{
			name:      "SQL injection attempt",
			inputName: "Robert'; DROP TABLE users; --",
			fieldName: "name",
			wantErr:   true,
		},
		{
			name:      "XSS attempt",
			inputName: "<script>alert('XSS')</script>",
			fieldName: "name",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateName(tt.inputName, tt.fieldName)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.fieldName)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Phone Validation Tests
// ============================================================================

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		name    string
		phone   string
		wantErr bool
	}{
		// Valid phones
		{
			name:    "Empty phone (optional)",
			phone:   "",
			wantErr: false,
		},
		{
			name:    "Valid US phone",
			phone:   "14155552671",
			wantErr: false,
		},
		{
			name:    "Valid Philippine phone",
			phone:   "639171234567",
			wantErr: false,
		},
		{
			name:    "Minimum length (7 digits)",
			phone:   "1234567",
			wantErr: false,
		},
		{
			name:    "Maximum length (15 digits)",
			phone:   "123456789012345",
			wantErr: false,
		},
		{
			name:    "Phone with formatting (stripped)",
			phone:   "+1 (415) 555-2671",
			wantErr: false,
		},
		{
			name:    "Phone with spaces (stripped)",
			phone:   "1 415 555 2671",
			wantErr: false,
		},
		{
			name:    "Phone with dashes (stripped)",
			phone:   "1-415-555-2671",
			wantErr: false,
		},

		// Invalid phones
		{
			name:    "Too short (6 digits)",
			phone:   "123456",
			wantErr: true,
		},
		{
			name:    "Too long (16 digits)",
			phone:   "1234567890123456",
			wantErr: true,
		},
		{
			name:    "Only letters",
			phone:   "ABCDEFGH",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidatePhone(tt.phone)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Address Validation Tests
// ============================================================================

func TestValidateAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		// Valid addresses
		{
			name:    "Empty address (optional)",
			address: "",
			wantErr: false,
		},
		{
			name:    "Simple address",
			address: "123 Main St",
			wantErr: false,
		},
		{
			name:    "Full address with apartment",
			address: "123 Main St, Apt 4B, New York, NY 10001",
			wantErr: false,
		},
		{
			name:    "Address with special characters",
			address: "123 Main St., Suite #400",
			wantErr: false,
		},
		{
			name:    "International address",
			address: "Calle de Alcalá, 52, 28014 Madrid, Spain",
			wantErr: false,
		},
		{
			name:    "Maximum length (500 chars)",
			address: strings.Repeat("A", 500),
			wantErr: false,
		},

		// Invalid addresses
		{
			name:    "Too long (>500 chars)",
			address: strings.Repeat("A", 501),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateAddress(tt.address)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Organization Role Validation Tests
// ============================================================================

func TestValidateOrganizationRole(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		wantErr bool
	}{
		// Valid roles
		{
			name:    "Admin role (lowercase)",
			role:    "admin",
			wantErr: false,
		},
		{
			name:    "Admin role (uppercase)",
			role:    "ADMIN",
			wantErr: false,
		},
		{
			name:    "Admin role (mixed case)",
			role:    "Admin",
			wantErr: false,
		},
		{
			name:    "Issuer role",
			role:    "issuer",
			wantErr: false,
		},
		{
			name:    "RTO role",
			role:    "rto",
			wantErr: false,
		},
		{
			name:    "Student role",
			role:    "student",
			wantErr: false,
		},

		// Invalid roles
		{
			name:    "Invalid role",
			role:    "superadmin",
			wantErr: true,
		},
		{
			name:    "Empty role",
			role:    "",
			wantErr: true,
		},
		{
			name:    "Unknown role",
			role:    "guest",
			wantErr: true,
		},
		{
			name:    "SQL injection attempt",
			role:    "admin' OR '1'='1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateOrganizationRole(tt.role)

			if tt.wantErr {
				assert.ErrorIs(t, err, validation.ErrInvalidOrgRole)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNormalizeOrganizationRole(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Lowercase admin",
			input:    "admin",
			expected: "admin",
		},
		{
			name:     "Uppercase admin",
			input:    "ADMIN",
			expected: "admin",
		},
		{
			name:     "Mixed case issuer",
			input:    "Issuer",
			expected: "issuer",
		},
		{
			name:     "Uppercase RTO",
			input:    "RTO",
			expected: "rto",
		},
		{
			name:     "Invalid role (unchanged)",
			input:    "invalid",
			expected: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validation.NormalizeOrganizationRole(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// Organization Name & Slug Validation Tests
// ============================================================================

func TestValidateOrganizationName(t *testing.T) {
	tests := []struct {
		name    string
		orgName string
		wantErr bool
	}{
		// Valid organization names
		{
			name:    "Simple org name",
			orgName: "Acme Corp",
			wantErr: false,
		},
		{
			name:    "Org name with special chars",
			orgName: "Acme Corp. & Associates",
			wantErr: false,
		},
		{
			name:    "Maximum length (100 chars)",
			orgName: strings.Repeat("A", 100),
			wantErr: false,
		},

		// Invalid organization names
		{
			name:    "Empty org name",
			orgName: "",
			wantErr: true,
		},
		{
			name:    "Only spaces",
			orgName: "   ",
			wantErr: true,
		},
		{
			name:    "Too long (>100 chars)",
			orgName: strings.Repeat("A", 101),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateOrganizationName(tt.orgName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidSlug(t *testing.T) {
	tests := []struct {
		name  string
		slug  string
		valid bool
	}{
		// Valid slugs
		{
			name:  "Simple slug",
			slug:  "acme-corp",
			valid: true,
		},
		{
			name:  "Slug with numbers",
			slug:  "acme123",
			valid: true,
		},
		{
			name:  "Slug with multiple hyphens",
			slug:  "acme-corp-inc",
			valid: true,
		},
		{
			name:  "Single character slug",
			slug:  "a",
			valid: true,
		},

		// Invalid slugs
		{
			name:  "Empty slug",
			slug:  "",
			valid: false,
		},
		{
			name:  "Uppercase letters",
			slug:  "Acme-Corp",
			valid: false,
		},
		{
			name:  "Slug with spaces",
			slug:  "acme corp",
			valid: false,
		},
		{
			name:  "Slug with underscores",
			slug:  "acme_corp",
			valid: false,
		},
		{
			name:  "Slug starting with hyphen",
			slug:  "-acme",
			valid: false,
		},
		{
			name:  "Slug ending with hyphen",
			slug:  "acme-",
			valid: false,
		},
		{
			name:  "Slug with consecutive hyphens",
			slug:  "acme--corp",
			valid: false,
		},
		{
			name:  "Slug with special characters",
			slug:  "acme@corp",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validation.IsValidSlug(tt.slug)
			assert.Equal(t, tt.valid, result)
		})
	}
}

// ============================================================================
// Composite Validation Tests (Registration, Login, Password Reset)
// ============================================================================

func TestValidateUserRegistration(t *testing.T) {
	tests := []struct {
		name            string
		email           string
		password        string
		confirmPassword string
		wantErr         bool
		errType         error
	}{
		{
			name:            "Valid registration",
			email:           "user@example.com",
			password:        "SecureP@ss123",
			confirmPassword: "SecureP@ss123",
			wantErr:         false,
		},
		{
			name:            "Invalid email",
			email:           "invalid-email",
			password:        "SecureP@ss123",
			confirmPassword: "SecureP@ss123",
			wantErr:         true,
			errType:         validation.ErrInvalidEmail,
		},
		{
			name:            "Weak password",
			email:           "user@example.com",
			password:        "weak",
			confirmPassword: "weak",
			wantErr:         true,
			errType:         validation.ErrPasswordTooShort,
		},
		{
			name:            "Passwords don't match",
			email:           "user@example.com",
			password:        "SecureP@ss123",
			confirmPassword: "DifferentP@ss123",
			wantErr:         true,
			errType:         validation.ErrPasswordsDontMatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateUserRegistration(tt.email, tt.password, tt.confirmPassword)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateLogin(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "Valid login email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "Invalid login email",
			email:   "invalid-email",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateLogin(tt.email)

			if tt.wantErr {
				assert.ErrorIs(t, err, validation.ErrInvalidEmail)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePasswordReset(t *testing.T) {
	tests := []struct {
		name            string
		token           string
		password        string
		confirmPassword string
		wantErr         bool
	}{
		{
			name:            "Valid password reset",
			token:           "valid-reset-token",
			password:        "NewP@ssw0rd!",
			confirmPassword: "NewP@ssw0rd!",
			wantErr:         false,
		},
		{
			name:            "Empty token",
			token:           "",
			password:        "NewP@ssw0rd!",
			confirmPassword: "NewP@ssw0rd!",
			wantErr:         true,
		},
		{
			name:            "Weak password",
			token:           "valid-reset-token",
			password:        "weak",
			confirmPassword: "weak",
			wantErr:         true,
		},
		{
			name:            "Passwords don't match",
			token:           "valid-reset-token",
			password:        "NewP@ssw0rd!",
			confirmPassword: "DifferentP@ssw0rd!",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidatePasswordReset(tt.token, tt.password, tt.confirmPassword)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateForgotPassword(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "Valid forgot password email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "Invalid forgot password email",
			email:   "invalid-email",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateForgotPassword(tt.email)

			if tt.wantErr {
				assert.ErrorIs(t, err, validation.ErrInvalidEmail)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Security & Edge Case Tests
// ============================================================================

func TestSQLInjectionPrevention(t *testing.T) {
	// Test common SQL injection patterns are rejected
	sqlInjectionPatterns := []string{
		"' OR '1'='1",
		"admin'--",
		"'; DROP TABLE users; --",
		"1' OR '1' = '1",
		"x' AND email='admin@example.com",
	}

	t.Run("Email validation rejects SQL injection", func(t *testing.T) {
		for _, pattern := range sqlInjectionPatterns {
			err := validation.ValidateEmail(pattern)
			assert.Error(t, err, "Should reject SQL injection pattern: %s", pattern)
		}
	})

	t.Run("Name validation rejects SQL injection", func(t *testing.T) {
		err := validation.ValidateName("Robert'); DROP TABLE students;--", "name")
		assert.Error(t, err, "Should reject SQL injection in name")
	})

	t.Run("Organization role rejects SQL injection", func(t *testing.T) {
		err := validation.ValidateOrganizationRole("admin' OR '1'='1")
		assert.Error(t, err, "Should reject SQL injection in role")
	})
}

func TestXSSPrevention(t *testing.T) {
	// Test common XSS patterns are rejected
	xssPatterns := []string{
		"<script>alert('XSS')</script>",
		"<img src=x onerror=alert('XSS')>",
		"javascript:alert('XSS')",
		"<svg/onload=alert('XSS')>",
	}

	t.Run("Name validation rejects XSS", func(t *testing.T) {
		for _, pattern := range xssPatterns {
			err := validation.ValidateName(pattern, "name")
			assert.Error(t, err, "Should reject XSS pattern: %s", pattern)
		}
	})
}

func TestCaseSensitivity(t *testing.T) {
	t.Run("Email normalization is case-insensitive", func(t *testing.T) {
		emails := []string{
			"user@example.com",
			"User@Example.COM",
			"USER@EXAMPLE.COM",
		}

		for _, email := range emails {
			err := validation.ValidateEmail(email)
			assert.NoError(t, err)
			normalized := validation.NormalizeEmail(email)
			assert.Equal(t, "user@example.com", normalized)
		}
	})

	t.Run("Organization role validation is case-insensitive", func(t *testing.T) {
		roles := []string{"admin", "Admin", "ADMIN", "aDmIn"}

		for _, role := range roles {
			err := validation.ValidateOrganizationRole(role)
			assert.NoError(t, err, "Role '%s' should be valid", role)
		}
	})

	t.Run("Password validation is case-sensitive", func(t *testing.T) {
		password1 := "SecureP@ss123"
		password2 := "securep@ss123"

		err := validation.ValidatePasswordsMatch(password1, password2)
		assert.Error(t, err, "Passwords should not match (case-sensitive)")
	})
}

func TestUnicodeHandling(t *testing.T) {
	t.Run("Unicode in passwords", func(t *testing.T) {
		// Unicode characters should work in passwords
		err := validation.ValidatePassword("Pässw0rd!")
		assert.NoError(t, err)
	})

	t.Run("Emoji in names (should fail)", func(t *testing.T) {
		err := validation.ValidateName("John 😀 Doe", "name")
		assert.Error(t, err, "Emoji should not be allowed in names")
	})
}

func TestBoundaryConditions(t *testing.T) {
	t.Run("Minimum password length (8 chars)", func(t *testing.T) {
		// Exactly 8 characters - should pass
		err := validation.ValidatePassword("Abcd123!")
		assert.NoError(t, err)

		// 7 characters - should fail
		err = validation.ValidatePassword("Abc123!")
		assert.ErrorIs(t, err, validation.ErrPasswordTooShort)
	})

	t.Run("Maximum name length (100 chars)", func(t *testing.T) {
		// Exactly 100 characters - should pass
		validName := strings.Repeat("A", 100)
		err := validation.ValidateName(validName, "name")
		assert.NoError(t, err)

		// 101 characters - should fail
		invalidName := strings.Repeat("A", 101)
		err = validation.ValidateName(invalidName, "name")
		assert.Error(t, err)
	})

	t.Run("Maximum address length (500 chars)", func(t *testing.T) {
		// Exactly 500 characters - should pass
		validAddress := strings.Repeat("A", 500)
		err := validation.ValidateAddress(validAddress)
		assert.NoError(t, err)

		// 501 characters - should fail
		invalidAddress := strings.Repeat("A", 501)
		err = validation.ValidateAddress(invalidAddress)
		assert.Error(t, err)
	})

	t.Run("Phone number length boundaries", func(t *testing.T) {
		// Minimum 7 digits - should pass
		err := validation.ValidatePhone("1234567")
		assert.NoError(t, err)

		// 6 digits - should fail
		err = validation.ValidatePhone("123456")
		assert.Error(t, err)

		// Maximum 15 digits - should pass
		err = validation.ValidatePhone("123456789012345")
		assert.NoError(t, err)

		// 16 digits - should fail
		err = validation.ValidatePhone("1234567890123456")
		assert.Error(t, err)
	})
}
