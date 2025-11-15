package validation

import (
	"errors"
	"fmt"
	"net/mail"
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
	ErrInvalidUserType    = errors.New("invalid user type")
	ErrInvalidTenantID    = errors.New("invalid tenant ID")
	ErrPasswordsDontMatch = errors.New("passwords do not match")
)

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
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

// ValidateUserType validates user type
func ValidateUserType(userType string) error {
	validTypes := []string{
		models.UserTypeAdmin,
		models.UserTypeStudent,
		models.UserTypeRTO,
		models.UserTypeIssuer,
		models.UserTypeValidator,
		models.UserTypeBadger,
		models.UserTypeNonPartner,
		models.UserTypePartner,
	}

	userTypeLower := strings.ToLower(userType)
	for _, validType := range validTypes {
		if userTypeLower == strings.ToLower(validType) {
			return nil
		}
	}

	return ErrInvalidUserType
}

// NormalizeUserType normalizes user type to proper case
func NormalizeUserType(userType string) string {
	validTypes := []string{
		models.UserTypeAdmin,
		models.UserTypeStudent,
		models.UserTypeRTO,
		models.UserTypeIssuer,
		models.UserTypeValidator,
		models.UserTypeBadger,
		models.UserTypeNonPartner,
		models.UserTypePartner,
	}

	userTypeLower := strings.ToLower(userType)
	for _, validType := range validTypes {
		if userTypeLower == strings.ToLower(validType) {
			return validType
		}
	}

	return userType // Return as-is if not found (shouldn't happen after validation)
}

// ValidatePasswordsMatch checks if passwords match
func ValidatePasswordsMatch(password, confirmPassword string) error {
	if password != confirmPassword {
		return ErrPasswordsDontMatch
	}
	return nil
}

// ValidatePhone validates phone number format (basic validation)
func ValidatePhone(phone string) error {
	if phone == "" {
		return nil // Phone is optional
	}

	// Remove all non-digit characters for validation
	phone = regexp.MustCompile(`\D`).ReplaceAllString(phone, "")

	// Check if it's a reasonable length (7-15 digits)
	if len(phone) < 7 || len(phone) > 15 {
		return errors.New("phone number must be between 7 and 15 digits")
	}

	return nil
}

// ValidateName validates name fields
func ValidateName(name string, fieldName string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}

	if len(name) > 100 {
		return fmt.Errorf("%s cannot be longer than 100 characters", fieldName)
	}

	// Check for valid characters (letters, spaces, hyphens, apostrophes)
	validName := regexp.MustCompile(`^[a-zA-Z\s\-']+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("%s contains invalid characters", fieldName)
	}

	return nil
}

// ValidateAddress validates address field
func ValidateAddress(address string) error {
	if address == "" {
		return nil // Address is optional
	}

	if len(address) > 500 {
		return errors.New("address cannot be longer than 500 characters")
	}

	return nil
}

// ValidateTenantDomain validates tenant domain
func ValidateTenantDomain(domain string) error {
	if strings.TrimSpace(domain) == "" {
		return errors.New("domain cannot be empty")
	}

	// Basic domain validation
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	if !domainRegex.MatchString(domain) {
		return errors.New("invalid domain format")
	}

	return nil
}

// ValidateTenantName validates tenant name
func ValidateTenantName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("tenant name cannot be empty")
	}

	if len(name) > 100 {
		return errors.New("tenant name cannot be longer than 100 characters")
	}

	return nil
}

// IsValidSlug validates organization slug format
func IsValidSlug(slug string) bool {
	if slug == "" {
		return false
	}

	// Slug should be lowercase, contain only letters, numbers, and hyphens
	slugRegex := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	return slugRegex.MatchString(slug)
}

// ValidateUserRegistration validates user registration data
func ValidateUserRegistration(email, password, confirmPassword, userType, tenantID string) error {
	if err := ValidateEmail(email); err != nil {
		return err
	}

	if err := ValidatePassword(password); err != nil {
		return err
	}

	if err := ValidatePasswordsMatch(password, confirmPassword); err != nil {
		return err
	}

	if err := ValidateUserType(userType); err != nil {
		return err
	}

	// Tenant ID is now optional - will be auto-assigned if not provided
	return nil
}

// ValidateLogin validates login data
func ValidateLogin(email, tenantID string) error {
	if err := ValidateEmail(email); err != nil {
		return err
	}

	if tenantID == "" {
		return ErrInvalidTenantID
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
func ValidateForgotPassword(email, tenantID string) error {
	if err := ValidateEmail(email); err != nil {
		return err
	}

	if tenantID == "" {
		return ErrInvalidTenantID
	}

	return nil
}
