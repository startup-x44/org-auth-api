package test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"

	"auth-service/pkg/metrics"
)

// TestMetricsNoPII validates that the /metrics endpoint NEVER exposes PII
// This test prevents cardinality explosion and data leaks in production
func TestMetricsNoPII(t *testing.T) {
	// Initialize metrics
	metrics.Initialize()

	// Create test router with metrics endpoint
	router := gin.New()
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Create test server
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(w, req)

	// Get response body
	body, err := io.ReadAll(w.Body)
	assert.NoError(t, err)
	metricsOutput := string(body)

	// Define PII patterns that should NEVER appear
	piiPatterns := []struct {
		name    string
		pattern string
		regex   *regexp.Regexp
	}{
		{
			name:    "UUID",
			pattern: `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`,
			regex:   regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`),
		},
		{
			name:    "Email",
			pattern: `[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`,
			regex:   regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`),
		},
		{
			name:    "IPv4",
			pattern: `\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`,
			regex:   regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`),
		},
		{
			name:    "Long Hash (Base64-like)",
			pattern: `[A-Za-z0-9+/]{40,}`,
			regex:   regexp.MustCompile(`[A-Za-z0-9+/]{40,}`),
		},
	}

	// Check for PII patterns
	for _, pii := range piiPatterns {
		matches := pii.regex.FindAllString(metricsOutput, -1)
		if len(matches) > 0 {
			// Filter out false positives
			realMatches := filterFalsePositives(pii.name, matches)
			if len(realMatches) > 0 {
				t.Errorf("PII VIOLATION: Found %s pattern in metrics output!\nMatches: %v\nThis violates Prometheus best practices and creates cardinality explosion.",
					pii.name, realMatches)
			}
		}
	}

	// Verify only low-cardinality labels are used
	t.Run("VerifyLowCardinalityLabels", func(t *testing.T) {
		lines := strings.Split(metricsOutput, "\n")
		for _, line := range lines {
			// Skip comments and empty lines
			if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
				continue
			}

			// Check for forbidden label values
			forbiddenPatterns := []string{
				`organization="[0-9a-fA-F-]+"`, // UUID in organization label
				`client_id="[0-9a-fA-F-]+"`,    // UUID in client_id label
				`permission="[^"]{20,}"`,       // Long permission strings
				`identifier="[^"]+"`,           // Any identifier label (should be removed)
				`user_id="`,                    // user_id should never be in labels
				`email="`,                      // email should never be in labels
				`ip="`,                         // IP should never be in labels
			}

			for _, pattern := range forbiddenPatterns {
				matched, _ := regexp.MatchString(pattern, line)
				if matched {
					t.Errorf("High-cardinality label found in metrics:\n%s\nPattern: %s", line, pattern)
				}
			}
		}
	})

	// Verify expected low-cardinality labels ARE present
	t.Run("VerifyExpectedLabels", func(t *testing.T) {
		expectedLabels := []string{
			`method=`,          // HTTP method
			`path=`,            // Normalized path
			`status=`,          // HTTP status
			`type=`,            // Auth type
			`reason=`,          // Failure reason
			`scope=`,           // Rate limit scope
			`operation=`,       // DB/Redis operation
			`table=`,           // DB table
			`result=`,          // Permission check result
			`category=`,        // Permission category
			`grant_type=`,      // OAuth2 grant type
			`flow_type=`,       // OAuth2 flow type
			`action_category=`, // Audit log category
			`resource=`,        // Resource type
			`destination=`,     // Audit destination
			`component=`,       // Error component
		}

		for _, label := range expectedLabels {
			if !strings.Contains(metricsOutput, label) {
				t.Logf("Warning: Expected label '%s' not found (may be zero-valued)", label)
			}
		}
	})

	// Verify metrics use enums, not dynamic values
	t.Run("VerifyEnumValues", func(t *testing.T) {
		// Extract all label values and check they're from expected enums
		labelValueRegex := regexp.MustCompile(`(\w+)="([^"]+)"`)
		matches := labelValueRegex.FindAllStringSubmatch(metricsOutput, -1)

		allowedValues := map[string][]string{
			"type":            {"login", "oauth2", "api_key", "refresh", "access", "authorization_code"},
			"status":          {"success", "error", "valid", "invalid", "expired", "revoked", "sent", "accepted", "cancelled", "200", "201", "400", "401", "403", "404", "500"},
			"method":          {"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
			"result":          {"granted", "denied", "allowed", "blocked"},
			"reason":          {"invalid_credentials", "account_disabled", "email_not_verified", "rate_limited", "insufficient_permissions", "not_member", "invalid_token", "expired", "revoked", "malformed", "signature_invalid", "logout", "timeout", "admin_action"},
			"scope":           {"login", "registration", "password_reset", "token_refresh", "oauth2_token", "api_calls", "user", "organization", "single"},
			"operation":       {"SELECT", "INSERT", "UPDATE", "DELETE", "GET", "SET", "DEL", "EXPIRE", "INCR", "ZADD", "create", "update", "delete", "assign", "revoke", "grant"},
			"category":        {"org", "role", "member", "permission", "user", "session", "oauth", "api_key", "invitation", "other", "unknown"},
			"action_category": {"auth", "role", "permission", "org", "session", "oauth", "api_key", "other", "unknown"},
			"grant_type":      {"authorization_code", "refresh_token", "client_credentials"},
			"flow_type":       {"authorization_code", "token_exchange"},
			"resource":        {"user", "role", "permission", "organization", "session", "oauth", "api_key"},
			"destination":     {"database", "structured_log"},
			"component":       {"auth_handler", "role_handler", "oauth_handler", "user_service", "database", "redis", "validation", "authentication", "authorization", "internal"},
		}

		for _, match := range matches {
			if len(match) < 3 {
				continue
			}
			labelName := match[1]
			labelValue := match[2]

			// Skip path and table - they can have varied but still safe values
			if labelName == "path" || labelName == "table" {
				// Verify path doesn't contain UUIDs
				uuidRegex := regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}`)
				if uuidRegex.MatchString(labelValue) {
					t.Errorf("Path contains UUID: %s", labelValue)
				}
				continue
			}

			// Check if label value is in allowed list
			if allowed, exists := allowedValues[labelName]; exists {
				found := false
				for _, allowedVal := range allowed {
					if labelValue == allowedVal {
						found = true
						break
					}
				}
				if !found && labelValue != "0" && !strings.HasPrefix(labelValue, "/") {
					// Allow numeric values and paths
					if _, err := fmt.Sscanf(labelValue, "%d", new(int)); err != nil {
						t.Errorf("Label '%s' has unexpected value '%s'. Expected one of: %v", labelName, labelValue, allowed)
					}
				}
			}
		}
	})
}

// filterFalsePositives removes known false positives from PII matches
func filterFalsePositives(piiType string, matches []string) []string {
	var realMatches []string
	for _, match := range matches {
		// Filter out known safe patterns
		switch piiType {
		case "IPv4":
			// Bucket boundaries like "0.001" are not IPs
			if !strings.Contains(match, "0.0") && !regexp.MustCompile(`^0\.\d+$`).MatchString(match) {
				realMatches = append(realMatches, match)
			}
		case "UUID":
			// No false positives for UUIDs - any match is a violation
			realMatches = append(realMatches, match)
		case "Email":
			// No false positives for emails - any match is a violation
			realMatches = append(realMatches, match)
		case "Long Hash (Base64-like)":
			// Filter out Prometheus metadata (starts with HELP, TYPE, etc.)
			if !strings.HasPrefix(match, "HELP") && !strings.HasPrefix(match, "TYPE") {
				realMatches = append(realMatches, match)
			}
		default:
			realMatches = append(realMatches, match)
		}
	}
	return realMatches
}

// TestPermissionCategorization tests that permissions are properly categorized
func TestPermissionCategorization(t *testing.T) {
	testCases := []struct {
		permission string
		expected   string
	}{
		{"org:create", "org"},
		{"organization:update", "org"},
		{"role:assign", "role"},
		{"member:invite", "member"},
		{"permission:grant", "permission"},
		{"user:delete", "user"},
		{"session:revoke", "session"},
		{"oauth2:authorize", "oauth"},
		{"oauth:token", "oauth"},
		{"api_key:create", "api_key"},
		{"invitation:send", "invitation"},
		{"custom:action", "other"},
		{"", "unknown"},
	}

	for _, tc := range testCases {
		result := metrics.CategorizePermission(tc.permission)
		assert.Equal(t, tc.expected, result, "Permission '%s' should categorize to '%s', got '%s'", tc.permission, tc.expected, result)
	}
}

// TestActionCategorization tests that audit actions are properly categorized
func TestActionCategorization(t *testing.T) {
	testCases := []struct {
		action   string
		expected string
	}{
		{"Login", "auth"},
		{"Register", "auth"},
		{"Logout", "auth"},
		{"PasswordChange", "auth"},
		{"RoleCreate", "role"},
		{"PermissionGrant", "permission"},
		{"OrganizationUpdate", "org"},
		{"OrgDelete", "org"},
		{"SessionRevoke", "session"},
		{"OAuth2Authorize", "oauth"},
		{"APIKeyCreate", "api_key"},
		{"CustomAction", "other"},
		{"", "unknown"},
	}

	for _, tc := range testCases {
		result := metrics.CategorizeAction(tc.action)
		assert.Equal(t, tc.expected, result, "Action '%s' should categorize to '%s', got '%s'", tc.action, tc.expected, result)
	}
}

// BenchmarkMetricsCollection benchmarks metrics collection performance
func BenchmarkMetricsCollection(t *testing.B) {
	m := metrics.Initialize()

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		m.RecordAuthAttempt("login", true, "")
		m.RecordTokenIssuance("access")
		m.RecordPermissionCheck("role:create", true)
		m.RecordRateLimit("login", false)
	}
}
