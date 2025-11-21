package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the auth service
// ALL labels use low-cardinality enums to prevent cardinality explosion
type Metrics struct {
	// HTTP request metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestSize     *prometheus.HistogramVec
	HTTPResponseSize    *prometheus.HistogramVec

	// Authentication metrics
	AuthAttemptsTotal       *prometheus.CounterVec
	AuthSuccessTotal        *prometheus.CounterVec
	AuthFailuresTotal       *prometheus.CounterVec
	RegistrationsTotal      *prometheus.CounterVec
	PasswordResetsTotal     *prometheus.CounterVec
	EmailVerificationsTotal *prometheus.CounterVec

	// Token metrics
	TokensIssuedTotal     *prometheus.CounterVec
	TokenRefreshesTotal   *prometheus.CounterVec
	TokenRevocationsTotal *prometheus.CounterVec
	TokenValidationsTotal *prometheus.CounterVec
	AccessTokensActive    prometheus.Gauge
	RefreshTokensActive   prometheus.Gauge

	// OAuth2 metrics
	OAuth2AuthorizationsTotal *prometheus.CounterVec
	OAuth2TokenGrantsTotal    *prometheus.CounterVec
	OAuth2TokenRefreshesTotal *prometheus.CounterVec
	OAuth2FlowDuration        *prometheus.HistogramVec

	// RBAC metrics
	PermissionChecksTotal     *prometheus.CounterVec
	RoleOperationsTotal       *prometheus.CounterVec
	PermissionOperationsTotal *prometheus.CounterVec
	AuthorizationDenials      *prometheus.CounterVec

	// Session metrics
	ActiveSessions        prometheus.Gauge
	SessionCreationsTotal *prometheus.CounterVec
	SessionDestroyedTotal *prometheus.CounterVec

	// Rate limiting metrics
	RateLimitHitsTotal   *prometheus.CounterVec
	RateLimitBlocksTotal *prometheus.CounterVec

	// Database metrics
	DBQueriesTotal       *prometheus.CounterVec
	DBQueryDuration      *prometheus.HistogramVec
	DBConnectionsActive  prometheus.Gauge
	DBConnectionsIdle    prometheus.Gauge
	DBConnectionsWaiting prometheus.Gauge

	// Redis metrics
	RedisOperationsTotal   *prometheus.CounterVec
	RedisOperationDuration *prometheus.HistogramVec
	RedisConnectionsIdle   prometheus.Gauge

	// Audit logging metrics
	AuditLogsTotal        *prometheus.CounterVec
	AuditLogWriteDuration *prometheus.HistogramVec

	// Organization metrics
	OrganizationsTotal  prometheus.Gauge
	OrgMembersTotal     prometheus.Gauge
	OrgInvitationsTotal *prometheus.CounterVec

	// API Key metrics
	APIKeysActive          prometheus.Gauge
	APIKeyValidationsTotal *prometheus.CounterVec

	// Error metrics
	ErrorsTotal          *prometheus.CounterVec
	PanicsRecoveredTotal prometheus.Counter

	// SLI/SLO metrics for business monitoring
	SLI *SLIMetrics
}

var (
	instance *Metrics
)

// Initialize creates and registers all Prometheus metrics
// ALL LABELS USE LOW-CARDINALITY ENUMS - NO UUIDs, IPs, emails, or user input
func Initialize() *Metrics {
	if instance != nil {
		return instance
	}

	instance = &Metrics{
		// HTTP request metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request latency in seconds",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_size_bytes",
				Help:    "HTTP request size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "path"},
		),
		HTTPResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "path"},
		),

		// Authentication metrics - NO organization UUIDs
		AuthAttemptsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_attempts_total",
				Help: "Total number of authentication attempts",
			},
			[]string{"type"}, // login, oauth2, api_key, refresh
		),
		AuthSuccessTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_success_total",
				Help: "Total number of successful authentications",
			},
			[]string{"type"},
		),
		AuthFailuresTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_failures_total",
				Help: "Total number of failed authentications",
			},
			[]string{"type", "reason"}, // invalid_credentials, account_disabled, email_not_verified, rate_limited
		),
		RegistrationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "registrations_total",
				Help: "Total number of user registrations",
			},
			[]string{"status"}, // success, error
		),
		PasswordResetsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "password_resets_total",
				Help: "Total number of password reset requests",
			},
			[]string{"status"},
		),
		EmailVerificationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "email_verifications_total",
				Help: "Total number of email verification attempts",
			},
			[]string{"status"},
		),

		// Token metrics - NO organization UUIDs
		TokensIssuedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tokens_issued_total",
				Help: "Total number of tokens issued",
			},
			[]string{"type"}, // access, refresh, authorization_code
		),
		TokenRefreshesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "token_refreshes_total",
				Help: "Total number of token refresh operations",
			},
			[]string{"status"}, // success, error, revoked, expired
		),
		TokenRevocationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "token_revocations_total",
				Help: "Total number of token revocations",
			},
			[]string{"type", "scope"}, // type: access/refresh, scope: user/organization/single
		),
		TokenValidationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "token_validations_total",
				Help: "Total number of token validation checks",
			},
			[]string{"status", "reason"}, // valid/invalid, expired/revoked/malformed/signature_invalid
		),
		AccessTokensActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "access_tokens_active",
				Help: "Current number of active access tokens (estimated)",
			},
		),
		RefreshTokensActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "refresh_tokens_active",
				Help: "Current number of active refresh tokens",
			},
		),

		// OAuth2 metrics - NO client_id UUIDs
		OAuth2AuthorizationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "oauth2_authorizations_total",
				Help: "Total number of OAuth2 authorization requests",
			},
			[]string{"status"}, // success, error
		),
		OAuth2TokenGrantsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "oauth2_token_grants_total",
				Help: "Total number of OAuth2 token grants",
			},
			[]string{"grant_type", "status"}, // authorization_code/refresh_token/client_credentials
		),
		OAuth2TokenRefreshesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "oauth2_token_refreshes_total",
				Help: "Total number of OAuth2 token refreshes",
			},
			[]string{"status"},
		),
		OAuth2FlowDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "oauth2_flow_duration_seconds",
				Help:    "OAuth2 flow duration in seconds",
				Buckets: []float64{.1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"flow_type", "status"}, // authorization_code/token_exchange
		),

		// RBAC metrics - NO permission names, role UUIDs, or org UUIDs
		PermissionChecksTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "permission_checks_total",
				Help: "Total number of permission checks performed",
			},
			[]string{"category", "result"}, // category: org/role/member/permission/user, result: granted/denied
		),
		RoleOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "role_operations_total",
				Help: "Total number of role operations",
			},
			[]string{"operation", "status"}, // create/update/delete/assign/revoke
		),
		PermissionOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "permission_operations_total",
				Help: "Total number of permission operations",
			},
			[]string{"operation", "status"}, // create/update/delete/grant/revoke
		),
		AuthorizationDenials: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "authorization_denials_total",
				Help: "Total number of authorization denials",
			},
			[]string{"category", "reason"}, // insufficient_permissions/not_member/invalid_token
		),

		// Session metrics - NO organization UUIDs
		ActiveSessions: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "sessions_active",
				Help: "Current number of active sessions",
			},
		),
		SessionCreationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "session_creations_total",
				Help: "Total number of sessions created",
			},
			[]string{}, // No labels - just count
		),
		SessionDestroyedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "session_destroyed_total",
				Help: "Total number of sessions destroyed",
			},
			[]string{"reason"}, // logout/timeout/revocation/admin_action
		),

		// Rate limiting metrics - NO identifiers (IPs, user IDs)
		RateLimitHitsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rate_limit_hits_total",
				Help: "Total number of rate limit checks",
			},
			[]string{"scope", "result"}, // login/registration/password_reset/token_refresh/oauth2_token/api_calls, allowed/blocked
		),
		RateLimitBlocksTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rate_limit_blocks_total",
				Help: "Total number of requests blocked by rate limiting",
			},
			[]string{"scope"}, // NO identifier label - removed for cardinality
		),

		// Database metrics
		DBQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "table", "status"}, // SELECT/INSERT/UPDATE/DELETE
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: []float64{.0001, .0005, .001, .005, .01, .025, .05, .1, .25, .5, 1},
			},
			[]string{"operation", "table"},
		),
		DBConnectionsActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_active",
				Help: "Current number of active database connections",
			},
		),
		DBConnectionsIdle: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_idle",
				Help: "Current number of idle database connections",
			},
		),
		DBConnectionsWaiting: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_waiting",
				Help: "Current number of connections waiting for a database connection",
			},
		),

		// Redis metrics
		RedisOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "redis_operations_total",
				Help: "Total number of Redis operations",
			},
			[]string{"operation", "status"}, // GET/SET/DEL/EXPIRE/INCR/ZADD
		),
		RedisOperationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "redis_operation_duration_seconds",
				Help:    "Redis operation duration in seconds",
				Buckets: []float64{.0001, .0005, .001, .005, .01, .025, .05, .1},
			},
			[]string{"operation"},
		),
		RedisConnectionsIdle: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "redis_connections_idle",
				Help: "Current number of idle Redis connections (redis v8)",
			},
		),

		// Audit logging metrics
		AuditLogsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "audit_logs_total",
				Help: "Total number of audit log entries",
			},
			[]string{"action_category", "resource", "status"}, // action_category: auth/role/permission/org/session/oauth/api_key
		),
		AuditLogWriteDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "audit_log_write_duration_seconds",
				Help:    "Audit log write operation duration in seconds",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5},
			},
			[]string{"destination"}, // database/structured_log
		),

		// Organization metrics - NO organization UUIDs
		OrganizationsTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "organizations_total",
				Help: "Current total number of organizations",
			},
		),
		OrgMembersTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "org_members_total",
				Help: "Current total number of organization members",
			},
		),
		OrgInvitationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "org_invitations_total",
				Help: "Total number of organization invitations",
			},
			[]string{"status"}, // sent/accepted/cancelled/expired
		),

		// API Key metrics
		APIKeysActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "api_keys_active",
				Help: "Current number of active API keys",
			},
		),
		APIKeyValidationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_key_validations_total",
				Help: "Total number of API key validation attempts",
			},
			[]string{"status"}, // valid/invalid/expired/revoked
		),

		// Error metrics
		ErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "errors_total",
				Help: "Total number of errors by type",
			},
			[]string{"type", "component"}, // database/redis/validation/authentication/authorization/internal
		),
		PanicsRecoveredTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "panics_recovered_total",
				Help: "Total number of panics recovered",
			},
		),

		// Initialize SLI/SLO metrics
		SLI: NewSLIMetrics(),
	}

	return instance
}

// GetMetrics returns the singleton metrics instance
func GetMetrics() *Metrics {
	if instance == nil {
		return Initialize()
	}
	return instance
}

// CategorizePermission converts permission strings to low-cardinality categories
// NEVER use the actual permission string in labels - causes cardinality explosion
func CategorizePermission(permission string) string {
	// Extract category from permission prefix
	if len(permission) == 0 {
		return "unknown"
	}

	// Common prefixes - use strings.HasPrefix for safety
	if len(permission) >= 3 && permission[:3] == "org" {
		return "org"
	}
	if len(permission) >= 12 && permission[:12] == "organization" {
		return "org"
	}
	if len(permission) >= 4 && permission[:4] == "role" {
		return "role"
	}
	if len(permission) >= 6 && permission[:6] == "member" {
		return "member"
	}
	if len(permission) >= 10 && permission[:10] == "permission" {
		return "permission"
	}
	if len(permission) >= 4 && permission[:4] == "user" {
		return "user"
	}
	if len(permission) >= 7 && permission[:7] == "session" {
		return "session"
	}
	if len(permission) >= 6 && permission[:6] == "oauth2" {
		return "oauth"
	}
	if len(permission) >= 5 && permission[:5] == "oauth" {
		return "oauth"
	}
	if len(permission) >= 7 && permission[:7] == "api_key" {
		return "api_key"
	}
	if len(permission) >= 10 && permission[:10] == "invitation" {
		return "invitation"
	}
	return "other"
}

// CategorizeAction converts action strings to categories for audit logs
func CategorizeAction(action string) string {
	// Extract category from action prefix
	if len(action) == 0 {
		return "unknown"
	}

	if len(action) >= 5 && action[:5] == "Login" {
		return "auth"
	}
	if len(action) >= 8 && action[:8] == "Register" {
		return "auth"
	}
	if len(action) >= 6 && action[:6] == "Logout" {
		return "auth"
	}
	if len(action) >= 14 && action[:14] == "PasswordChange" {
		return "auth"
	}
	if len(action) >= 4 && action[:4] == "Role" {
		return "role"
	}
	if len(action) >= 10 && action[:10] == "Permission" {
		return "permission"
	}
	if len(action) >= 12 && action[:12] == "Organization" {
		return "org"
	}
	if len(action) >= 3 && action[:3] == "Org" {
		return "org"
	}
	if len(action) >= 7 && action[:7] == "Session" {
		return "session"
	}
	if len(action) >= 6 && action[:6] == "OAuth2" {
		return "oauth"
	}
	if len(action) >= 6 && action[:6] == "APIKey" {
		return "api_key"
	}
	return "other"
}

// RecordHTTPRequest records an HTTP request with all relevant metrics
// path MUST be normalized (c.FullPath() with "unknown" fallback) - never use raw URL
func (m *Metrics) RecordHTTPRequest(method, path, status string, duration time.Duration, reqSize, respSize int64) {
	m.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path, status).Observe(duration.Seconds())
	if reqSize > 0 {
		m.HTTPRequestSize.WithLabelValues(method, path).Observe(float64(reqSize))
	}
	if respSize > 0 {
		m.HTTPResponseSize.WithLabelValues(method, path).Observe(float64(respSize))
	}
}

// RecordAuthAttempt records an authentication attempt
// NO organization parameter - removed for low cardinality
func (m *Metrics) RecordAuthAttempt(authType string, success bool, reason string) {
	m.AuthAttemptsTotal.WithLabelValues(authType).Inc()
	if success {
		m.AuthSuccessTotal.WithLabelValues(authType).Inc()
	} else {
		m.AuthFailuresTotal.WithLabelValues(authType, reason).Inc()
	}
}

// RecordTokenIssuance records a token being issued
// NO organization parameter - removed for low cardinality
func (m *Metrics) RecordTokenIssuance(tokenType string) {
	m.TokensIssuedTotal.WithLabelValues(tokenType).Inc()
	if tokenType == "access" {
		m.AccessTokensActive.Inc()
	} else if tokenType == "refresh" {
		m.RefreshTokensActive.Inc()
	}
}

// RecordTokenRefresh records a token refresh operation
func (m *Metrics) RecordTokenRefresh(status string) {
	m.TokenRefreshesTotal.WithLabelValues(status).Inc()
}

// RecordTokenRevocation records a token revocation
func (m *Metrics) RecordTokenRevocation(tokenType, scope string) {
	m.TokenRevocationsTotal.WithLabelValues(tokenType, scope).Inc()
	if tokenType == "access" {
		m.AccessTokensActive.Dec()
	} else if tokenType == "refresh" {
		m.RefreshTokensActive.Dec()
	}
}

// RecordTokenValidation records a token validation check
func (m *Metrics) RecordTokenValidation(status, reason string) {
	m.TokenValidationsTotal.WithLabelValues(status, reason).Inc()
}

// RecordOAuth2Authorization records an OAuth2 authorization request
// NO client_id - removed for low cardinality
func (m *Metrics) RecordOAuth2Authorization(status string) {
	m.OAuth2AuthorizationsTotal.WithLabelValues(status).Inc()
}

// RecordOAuth2TokenGrant records an OAuth2 token grant
// NO client_id - removed for low cardinality
func (m *Metrics) RecordOAuth2TokenGrant(grantType, status string) {
	m.OAuth2TokenGrantsTotal.WithLabelValues(grantType, status).Inc()
}

// RecordOAuth2TokenRefresh records an OAuth2 token refresh
// NO client_id - removed for low cardinality
func (m *Metrics) RecordOAuth2TokenRefresh(status string) {
	m.OAuth2TokenRefreshesTotal.WithLabelValues(status).Inc()
}

// RecordOAuth2FlowDuration records OAuth2 flow completion time
func (m *Metrics) RecordOAuth2FlowDuration(flowType, status string, duration time.Duration) {
	m.OAuth2FlowDuration.WithLabelValues(flowType, status).Observe(duration.Seconds())
}

// RecordPermissionCheck records a permission check
// permission is converted to category - NO raw permission strings or org UUIDs
func (m *Metrics) RecordPermissionCheck(permission string, granted bool) {
	category := CategorizePermission(permission)
	result := "granted"
	if !granted {
		result = "denied"
	}
	m.PermissionChecksTotal.WithLabelValues(category, result).Inc()
	if !granted {
		m.AuthorizationDenials.WithLabelValues(category, "insufficient_permissions").Inc()
	}
}

// RecordRoleOperation records a role operation
func (m *Metrics) RecordRoleOperation(operation, status string) {
	m.RoleOperationsTotal.WithLabelValues(operation, status).Inc()
}

// RecordPermissionOperation records a permission operation
func (m *Metrics) RecordPermissionOperation(operation, status string) {
	m.PermissionOperationsTotal.WithLabelValues(operation, status).Inc()
}

// RecordAuthorizationDenial records an authorization denial
func (m *Metrics) RecordAuthorizationDenial(permission, reason string) {
	category := CategorizePermission(permission)
	m.AuthorizationDenials.WithLabelValues(category, reason).Inc()
}

// RecordDBQuery records a database query
func (m *Metrics) RecordDBQuery(operation, table string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}
	m.DBQueriesTotal.WithLabelValues(operation, table, status).Inc()
	m.DBQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordRedisOperation records a Redis operation
func (m *Metrics) RecordRedisOperation(operation string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}
	m.RedisOperationsTotal.WithLabelValues(operation, status).Inc()
	m.RedisOperationDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordAuditLog records an audit log entry
// action is converted to category - NO raw action strings
func (m *Metrics) RecordAuditLog(action, resource string, duration time.Duration, destination string, err error) {
	actionCategory := CategorizeAction(action)
	status := "success"
	if err != nil {
		status = "error"
	}
	m.AuditLogsTotal.WithLabelValues(actionCategory, resource, status).Inc()
	m.AuditLogWriteDuration.WithLabelValues(destination).Observe(duration.Seconds())
}

// RecordRateLimit records a rate limit check
// NO identifier parameter - removed for cardinality safety
func (m *Metrics) RecordRateLimit(scope string, blocked bool) {
	result := "allowed"
	if blocked {
		result = "blocked"
		m.RateLimitBlocksTotal.WithLabelValues(scope).Inc()
	}
	m.RateLimitHitsTotal.WithLabelValues(scope, result).Inc()
}

// RecordSessionCreation records session creation
// NO organization parameter - removed for low cardinality
func (m *Metrics) RecordSessionCreation() {
	m.SessionCreationsTotal.WithLabelValues().Inc()
	m.ActiveSessions.Inc()
}

// RecordSessionDestruction records session destruction
// NO organization parameter - removed for low cardinality
func (m *Metrics) RecordSessionDestruction(reason string) {
	m.SessionDestroyedTotal.WithLabelValues(reason).Inc()
	m.ActiveSessions.Dec()
}

// RecordRegistration records a user registration
func (m *Metrics) RecordRegistration(status string) {
	m.RegistrationsTotal.WithLabelValues(status).Inc()
}

// RecordPasswordReset records a password reset request
func (m *Metrics) RecordPasswordReset(status string) {
	m.PasswordResetsTotal.WithLabelValues(status).Inc()
}

// RecordEmailVerification records an email verification attempt
func (m *Metrics) RecordEmailVerification(status string) {
	m.EmailVerificationsTotal.WithLabelValues(status).Inc()
}

// RecordOrgInvitation records an organization invitation
// NO organization parameter - removed for low cardinality
func (m *Metrics) RecordOrgInvitation(status string) {
	m.OrgInvitationsTotal.WithLabelValues(status).Inc()
}

// RecordAPIKeyValidation records an API key validation
func (m *Metrics) RecordAPIKeyValidation(status string) {
	m.APIKeyValidationsTotal.WithLabelValues(status).Inc()
}

// RecordError records an error
func (m *Metrics) RecordError(errorType, component string) {
	m.ErrorsTotal.WithLabelValues(errorType, component).Inc()
}

// RecordPanic records a recovered panic
func (m *Metrics) RecordPanic() {
	m.PanicsRecoveredTotal.Inc()
}
