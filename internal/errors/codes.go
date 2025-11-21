package errors

import (
	"net/http"
	"time"
)

// ErrorCode represents a structured error code for client integration
type ErrorCode string

// Authentication Error Codes
const (
	// Authentication failures
	ErrCodeInvalidCredentials ErrorCode = "AUTH_INVALID_CREDENTIALS"
	ErrCodeAccountLocked      ErrorCode = "AUTH_ACCOUNT_LOCKED"
	ErrCodeAccountNotVerified ErrorCode = "AUTH_ACCOUNT_NOT_VERIFIED"
	ErrCodeTwoFactorRequired  ErrorCode = "AUTH_2FA_REQUIRED"
	ErrCodeTwoFactorInvalid   ErrorCode = "AUTH_2FA_INVALID"

	// Token errors
	ErrCodeTokenExpired        ErrorCode = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid        ErrorCode = "TOKEN_INVALID"
	ErrCodeTokenRevoked        ErrorCode = "TOKEN_REVOKED"
	ErrCodeRefreshTokenInvalid ErrorCode = "REFRESH_TOKEN_INVALID"

	// OAuth2 errors
	ErrCodeOAuthInvalidClient    ErrorCode = "OAUTH_INVALID_CLIENT"
	ErrCodeOAuthInvalidGrant     ErrorCode = "OAUTH_INVALID_GRANT"
	ErrCodeOAuthInvalidScope     ErrorCode = "OAUTH_INVALID_SCOPE"
	ErrCodeOAuthUnsupportedGrant ErrorCode = "OAUTH_UNSUPPORTED_GRANT"

	// Permission errors
	ErrCodeInsufficientPermissions ErrorCode = "PERMISSION_INSUFFICIENT"
	ErrCodeRoleNotFound            ErrorCode = "ROLE_NOT_FOUND"
	ErrCodePermissionNotFound      ErrorCode = "PERMISSION_NOT_FOUND"

	// Rate limiting
	ErrCodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"

	// Validation errors
	ErrCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrCodeInvalidFormat    ErrorCode = "INVALID_FORMAT"
	ErrCodeMissingField     ErrorCode = "MISSING_FIELD"

	// System errors
	ErrCodeInternalError      ErrorCode = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeDatabaseError      ErrorCode = "DATABASE_ERROR"

	// Security errors
	ErrCodeCSRFTokenMissing   ErrorCode = "CSRF_TOKEN_MISSING"
	ErrCodeCSRFTokenInvalid   ErrorCode = "CSRF_TOKEN_INVALID"
	ErrCodeSuspiciousActivity ErrorCode = "SUSPICIOUS_ACTIVITY"

	// User management errors
	ErrCodeUserNotFound      ErrorCode = "USER_NOT_FOUND"
	ErrCodeUserAlreadyExists ErrorCode = "USER_ALREADY_EXISTS"
	ErrCodeEmailNotVerified  ErrorCode = "EMAIL_NOT_VERIFIED"

	// Organization errors
	ErrCodeOrgNotFound     ErrorCode = "ORGANIZATION_NOT_FOUND"
	ErrCodeOrgAccessDenied ErrorCode = "ORGANIZATION_ACCESS_DENIED"
)

// ErrorResponse represents a structured error response for clients
type ErrorResponse struct {
	Success   bool        `json:"success"`
	ErrorCode ErrorCode   `json:"error_code"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// ErrorMapping maps error codes to HTTP status codes
var ErrorMapping = map[ErrorCode]int{
	// 400 Bad Request
	ErrCodeValidationFailed:      http.StatusBadRequest,
	ErrCodeInvalidFormat:         http.StatusBadRequest,
	ErrCodeMissingField:          http.StatusBadRequest,
	ErrCodeOAuthInvalidGrant:     http.StatusBadRequest,
	ErrCodeOAuthInvalidScope:     http.StatusBadRequest,
	ErrCodeOAuthUnsupportedGrant: http.StatusBadRequest,
	ErrCodeRefreshTokenInvalid:   http.StatusBadRequest,

	// 401 Unauthorized
	ErrCodeInvalidCredentials: http.StatusUnauthorized,
	ErrCodeTokenExpired:       http.StatusUnauthorized,
	ErrCodeTokenInvalid:       http.StatusUnauthorized,
	ErrCodeTokenRevoked:       http.StatusUnauthorized,
	ErrCodeOAuthInvalidClient: http.StatusUnauthorized,
	ErrCodeTwoFactorInvalid:   http.StatusUnauthorized,

	// 403 Forbidden
	ErrCodeInsufficientPermissions: http.StatusForbidden,
	ErrCodeAccountLocked:           http.StatusForbidden,
	ErrCodeAccountNotVerified:      http.StatusForbidden,
	ErrCodeCSRFTokenMissing:        http.StatusForbidden,
	ErrCodeCSRFTokenInvalid:        http.StatusForbidden,
	ErrCodeSuspiciousActivity:      http.StatusForbidden,
	ErrCodeEmailNotVerified:        http.StatusForbidden,
	ErrCodeOrgAccessDenied:         http.StatusForbidden,

	// 404 Not Found
	ErrCodeUserNotFound:       http.StatusNotFound,
	ErrCodeRoleNotFound:       http.StatusNotFound,
	ErrCodePermissionNotFound: http.StatusNotFound,
	ErrCodeOrgNotFound:        http.StatusNotFound,

	// 409 Conflict
	ErrCodeUserAlreadyExists: http.StatusConflict,

	// 422 Unprocessable Entity
	ErrCodeTwoFactorRequired: http.StatusUnprocessableEntity,

	// 429 Too Many Requests
	ErrCodeRateLimitExceeded: http.StatusTooManyRequests,

	// 500 Internal Server Error
	ErrCodeInternalError: http.StatusInternalServerError,
	ErrCodeDatabaseError: http.StatusInternalServerError,

	// 503 Service Unavailable
	ErrCodeServiceUnavailable: http.StatusServiceUnavailable,
}

// GetHTTPStatus returns the HTTP status code for an error code
func (e ErrorCode) GetHTTPStatus() int {
	if status, exists := ErrorMapping[e]; exists {
		return status
	}
	return http.StatusInternalServerError
}

// Error implements the error interface
func (e ErrorCode) Error() string {
	return string(e)
}

// NewErrorResponse creates a new structured error response
func NewErrorResponse(code ErrorCode, message string, details interface{}, requestID string) *ErrorResponse {
	return &ErrorResponse{
		Success:   false,
		ErrorCode: code,
		Message:   message,
		Details:   details,
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}
