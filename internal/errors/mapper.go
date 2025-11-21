package errors

import (
	"errors"
	"strings"

	"auth-service/internal/service"
)

// ErrorMapper maps internal service errors to structured error codes
type ErrorMapper struct{}

// NewErrorMapper creates a new error mapper
func NewErrorMapper() *ErrorMapper {
	return &ErrorMapper{}
}

// MapServiceError maps service layer errors to structured error codes
func (em *ErrorMapper) MapServiceError(err error) (ErrorCode, string) {
	if err == nil {
		return ErrCodeInternalError, "Unknown error occurred"
	}

	errMsg := err.Error()

	// Authentication errors
	if errors.Is(err, service.ErrInvalidCredentials) {
		return ErrCodeInvalidCredentials, "Invalid username or password"
	}

	// Role-related errors
	if errors.Is(err, service.ErrRoleNotFound) {
		return ErrCodeRoleNotFound, "Role not found"
	}
	if errors.Is(err, service.ErrRoleNotFoundInOrg) {
		return ErrCodeRoleNotFound, "Role not found in this organization"
	}
	if errors.Is(err, service.ErrCannotUpdateSystemRole) {
		return ErrCodeInsufficientPermissions, "Cannot modify system roles"
	}
	if errors.Is(err, service.ErrCannotDeleteSystemRole) {
		return ErrCodeInsufficientPermissions, "Cannot delete system roles"
	}
	if errors.Is(err, service.ErrRoleHasActiveMembers) {
		return ErrCodeValidationFailed, "Cannot delete role with active members"
	}
	if errors.Is(err, service.ErrRoleNameAlreadyExists) {
		return ErrCodeValidationFailed, "Role name already exists in organization"
	}

	// Permission-related errors
	if errors.Is(err, service.ErrPermissionNotFound) {
		return ErrCodePermissionNotFound, "Permission not found"
	}
	if errors.Is(err, service.ErrPermissionAlreadyExists) {
		return ErrCodeValidationFailed, "Permission already exists"
	}
	if errors.Is(err, service.ErrCannotModifySystemPerms) {
		return ErrCodeInsufficientPermissions, "Cannot modify system permissions"
	}
	if errors.Is(err, service.ErrCannotUpdateSystemPerm) {
		return ErrCodeInsufficientPermissions, "Cannot update system permissions"
	}
	if errors.Is(err, service.ErrCannotDeleteSystemPerm) {
		return ErrCodeInsufficientPermissions, "Cannot delete system permissions"
	}
	if errors.Is(err, service.ErrPermissionInUse) {
		return ErrCodeValidationFailed, "Permission is currently in use and cannot be deleted"
	}
	if errors.Is(err, service.ErrSomePermissionsNotFound) {
		return ErrCodePermissionNotFound, "Some permissions were not found"
	}

	// Organization-related errors
	if errors.Is(err, service.ErrInsufficientPermission) {
		return ErrCodeInsufficientPermissions, "Insufficient permissions to perform this action"
	}

	// General errors
	if errors.Is(err, service.ErrInvalidUUID) {
		return ErrCodeInvalidFormat, "Invalid UUID format"
	}
	if errors.Is(err, service.ErrInvalidData) {
		return ErrCodeValidationFailed, "Invalid request data"
	}

	// String-based error matching for other common errors
	errLower := strings.ToLower(errMsg)

	// Database errors
	if strings.Contains(errLower, "database") || strings.Contains(errLower, "sql") {
		return ErrCodeDatabaseError, "Database operation failed"
	}

	// Token errors
	if strings.Contains(errLower, "token expired") {
		return ErrCodeTokenExpired, "Token has expired"
	}
	if strings.Contains(errLower, "invalid token") {
		return ErrCodeTokenInvalid, "Invalid token provided"
	}
	if strings.Contains(errLower, "token revoked") {
		return ErrCodeTokenRevoked, "Token has been revoked"
	}

	// User errors
	if strings.Contains(errLower, "user not found") {
		return ErrCodeUserNotFound, "User not found"
	}
	if strings.Contains(errLower, "user already exists") {
		return ErrCodeUserAlreadyExists, "User already exists"
	}
	if strings.Contains(errLower, "email not verified") {
		return ErrCodeEmailNotVerified, "Email address not verified"
	}

	// Rate limiting errors
	if strings.Contains(errLower, "rate limit") || strings.Contains(errLower, "too many requests") {
		return ErrCodeRateLimitExceeded, "Rate limit exceeded. Please try again later"
	}

	// CSRF errors
	if strings.Contains(errLower, "csrf") {
		if strings.Contains(errLower, "missing") {
			return ErrCodeCSRFTokenMissing, "CSRF token is required"
		}
		return ErrCodeCSRFTokenInvalid, "Invalid CSRF token"
	}

	// OAuth2 errors
	if strings.Contains(errLower, "invalid_client") {
		return ErrCodeOAuthInvalidClient, "Invalid client credentials"
	}
	if strings.Contains(errLower, "invalid_grant") {
		return ErrCodeOAuthInvalidGrant, "Invalid authorization grant"
	}
	if strings.Contains(errLower, "invalid_scope") {
		return ErrCodeOAuthInvalidScope, "Invalid or unsupported scope"
	}
	if strings.Contains(errLower, "unsupported_grant_type") {
		return ErrCodeOAuthUnsupportedGrant, "Unsupported grant type"
	}

	// Default to internal error
	return ErrCodeInternalError, "An internal error occurred"
}

// MapValidationError maps validation errors to structured error codes
func (em *ErrorMapper) MapValidationError(field, tag, value string) (ErrorCode, string, map[string]interface{}) {
	details := map[string]interface{}{
		"field": field,
		"tag":   tag,
		"value": value,
	}

	switch tag {
	case "required":
		return ErrCodeMissingField, field + " is required", details
	case "email":
		return ErrCodeInvalidFormat, "Invalid email format", details
	case "min":
		return ErrCodeValidationFailed, field + " is too short", details
	case "max":
		return ErrCodeValidationFailed, field + " is too long", details
	case "uuid":
		return ErrCodeInvalidFormat, "Invalid UUID format", details
	default:
		return ErrCodeValidationFailed, "Validation failed for " + field, details
	}
}
