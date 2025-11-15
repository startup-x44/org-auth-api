package validation

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"auth-service/internal/models"
)

// Validation errors
var (
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters")
	ErrPasswordNoUpper    = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLower    = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoNumber   = errors.New("password must contain at least one number")
	ErrPasswordNoSpecial  = errors.New("password must contain at least one special character")
	ErrInvalidOrgRole     = errors.New("invalid organization role")
	ErrPasswordsDontMatch = errors.New("passwords do not match")
)

// NormalizeEmail lowercases and trims email BEFORE validation
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// ValidateEmail validates email format (after normalization)
// Supports + character for email filtering (e.g., user+tag@domain.com)
func ValidateEmail(email string) error {
	email = NormalizeEmail(email)

	// Use regex that allows + character in local part
	// RFC 5322 compliant pattern for email validation
	emailRegex := regexp.MustCompile(`^[a-z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)*$`)

	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	return nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return ErrPasswordNoUpper
	}
	if !hasLower {
		return ErrPasswordNoLower
	}
	if !hasNumber {
		return ErrPasswordNoNumber
	}
	if !hasSpecial {
		return ErrPasswordNoSpecial
	}

	return nil
}

// Fast organization role validation (O(1) lookup)
var validOrgRoles = map[string]bool{
	strings.ToLower(models.OrganizationRoleAdmin):   true,
	strings.ToLower(models.OrganizationRoleIssuer):  true,
	strings.ToLower(models.OrganizationRoleRTO):     true,
	strings.ToLower(models.OrganizationRoleStudent): true,
}

// ValidateOrganizationRole validates organization role
func ValidateOrganizationRole(role string) error {
	if validOrgRoles[strings.ToLower(role)] {
		return nil
	}
	return ErrInvalidOrgRole
}

// NormalizeOrganizationRole normalizes org role to proper case
func NormalizeOrganizationRole(role string) string {
	roleLower := strings.ToLower(role)

	for _, v := range []string{
		models.OrganizationRoleAdmin,
		models.OrganizationRoleIssuer,
		models.OrganizationRoleRTO,
		models.OrganizationRoleStudent,
	} {
		if roleLower == strings.ToLower(v) {
			return v
		}
	}

	return role
}

// ValidatePasswordsMatch checks if passwords match
func ValidatePasswordsMatch(password, confirmPassword string) error {
	if password != confirmPassword {
		return ErrPasswordsDontMatch
	}
	return nil
}

// ValidatePhone validates phone number (optional)
func ValidatePhone(phone string) error {
	if phone == "" {
		return nil
	}

	phone = regexp.MustCompile(`\D`).ReplaceAllString(phone, "")

	if len(phone) < 7 || len(phone) > 15 {
		return errors.New("phone number must be between 7 and 15 digits")
	}

	return nil
}

// ValidateName ensures name fields are valid
func ValidateName(name string, fieldName string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}

	if len(name) > 100 {
		return fmt.Errorf("%s cannot be longer than 100 characters", fieldName)
	}

	validName := regexp.MustCompile(`^[a-zA-Z\s\-']+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("%s contains invalid characters", fieldName)
	}

	return nil
}

// ValidateAddress validates address field
func ValidateAddress(address string) error {
	if address == "" {
		return nil
	}

	if len(address) > 500 {
		return errors.New("address cannot be longer than 500 characters")
	}

	return nil
}

// ValidateOrganizationName ensures organization name is valid
func ValidateOrganizationName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("organization name cannot be empty")
	}

	if len(name) > 100 {
		return errors.New("organization name cannot be longer than 100 characters")
	}

	return nil
}

// IsValidSlug validates organization slug format
func IsValidSlug(slug string) bool {
	if slug == "" {
		return false
	}

	slugRegex := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	return slugRegex.MatchString(slug)
}

// ValidateUserRegistration validates global user registration data
func ValidateUserRegistration(email, password, confirmPassword string) error {
	if err := ValidateEmail(email); err != nil {
		return err
	}

	if err := ValidatePassword(password); err != nil {
		return err
	}

	if err := ValidatePasswordsMatch(password, confirmPassword); err != nil {
		return err
	}

	return nil
}

// ValidateLogin validates global login data
func ValidateLogin(email string) error {
	if err := ValidateEmail(email); err != nil {
		return err
	}

	return nil
}

// ValidatePasswordReset validates password reset data
func ValidatePasswordReset(token, password, confirmPassword string) error {
	if token == "" {
		return errors.New("reset token is required")
	}

	if err := ValidatePassword(password); err != nil {
		return err
	}

	if err := ValidatePasswordsMatch(password, confirmPassword); err != nil {
		return err
	}

	return nil
}

// ValidateForgotPassword validates forgot password request
func ValidateForgotPassword(email string) error {
	if err := ValidateEmail(email); err != nil {
		return err
	}

	return nil
}
