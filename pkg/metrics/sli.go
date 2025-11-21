package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// SLIMetrics contains business-focused Service Level Indicator metrics
type SLIMetrics struct {
	// Authentication Success Rate SLI
	// Target: 99.9% of authentication attempts should succeed within 500ms
	AuthenticationSuccessRate *prometheus.CounterVec
	AuthenticationLatency     *prometheus.HistogramVec

	// Session Creation Success Rate SLI
	// Target: 99.95% of session creations should succeed within 200ms
	SessionCreationSuccessRate *prometheus.CounterVec
	SessionCreationLatency     *prometheus.HistogramVec

	// Token Validation Success Rate SLI
	// Target: 99.99% of token validations should succeed within 100ms
	TokenValidationSuccessRate *prometheus.CounterVec
	TokenValidationLatency     *prometheus.HistogramVec

	// OAuth Flow Completion Rate SLI
	// Target: 99.5% of OAuth flows should complete successfully within 2s
	OAuthFlowCompletionRate *prometheus.CounterVec
	OAuthFlowLatency        *prometheus.HistogramVec

	// User Registration Success Rate SLI
	// Target: 99.8% of user registrations should succeed within 1s
	UserRegistrationSuccessRate *prometheus.CounterVec
	UserRegistrationLatency     *prometheus.HistogramVec

	// Permission Check Success Rate SLI
	// Target: 99.99% of permission checks should succeed within 50ms
	PermissionCheckSuccessRate *prometheus.CounterVec
	PermissionCheckLatency     *prometheus.HistogramVec

	// Email Verification Success Rate SLI
	// Target: 99.9% of email verifications should succeed within 300ms
	EmailVerificationSuccessRate *prometheus.CounterVec
	EmailVerificationLatency     *prometheus.HistogramVec

	// Password Reset Success Rate SLI
	// Target: 99.9% of password resets should succeed within 500ms
	PasswordResetSuccessRate *prometheus.CounterVec
	PasswordResetLatency     *prometheus.HistogramVec

	// Business Availability SLIs
	ServiceAvailability  *prometheus.GaugeVec // 1 = available, 0 = unavailable
	DatabaseAvailability *prometheus.GaugeVec // 1 = available, 0 = unavailable
	RedisAvailability    *prometheus.GaugeVec // 1 = available, 0 = unavailable

	// Business Volume SLIs
	DailyActiveUsers    prometheus.Gauge
	MonthlyActiveUsers  prometheus.Gauge
	ConcurrentUsers     prometheus.Gauge
	PeakConcurrentUsers prometheus.Gauge

	// Security SLIs
	SuspiciousActivityRate *prometheus.CounterVec
	FailedLoginAttempts    *prometheus.CounterVec
	RateLimitViolations    *prometheus.CounterVec
	CSRFTokenFailures      *prometheus.CounterVec

	// Business Error Rates
	ClientErrorRate        *prometheus.CounterVec // 4xx errors
	ServerErrorRate        *prometheus.CounterVec // 5xx errors
	BusinessLogicErrorRate *prometheus.CounterVec // Domain-specific errors
}

// NewSLIMetrics creates a new SLI metrics instance
func NewSLIMetrics() *SLIMetrics {
	return &SLIMetrics{
		// Authentication SLIs
		AuthenticationSuccessRate: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_authentication_attempts_total",
				Help: "Total authentication attempts by result and latency bucket",
			},
			[]string{"result", "latency_bucket"}, // result: success|failure, latency_bucket: fast|slow|timeout
		),
		AuthenticationLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "auth_authentication_duration_seconds",
				Help:    "Time spent on authentication requests",
				Buckets: prometheus.ExponentialBuckets(0.01, 2, 10), // 10ms to ~10s
			},
			[]string{"result"},
		),

		// Session Creation SLIs
		SessionCreationSuccessRate: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_session_creation_attempts_total",
				Help: "Total session creation attempts by result and latency bucket",
			},
			[]string{"result", "latency_bucket"},
		),
		SessionCreationLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "auth_session_creation_duration_seconds",
				Help:    "Time spent creating sessions",
				Buckets: prometheus.ExponentialBuckets(0.005, 2, 10), // 5ms to ~5s
			},
			[]string{"result"},
		),

		// Token Validation SLIs
		TokenValidationSuccessRate: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_token_validation_attempts_total",
				Help: "Total token validation attempts by result and latency bucket",
			},
			[]string{"result", "latency_bucket"},
		),
		TokenValidationLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "auth_token_validation_duration_seconds",
				Help:    "Time spent validating tokens",
				Buckets: prometheus.ExponentialBuckets(0.001, 2, 10), // 1ms to ~1s
			},
			[]string{"result"},
		),

		// OAuth Flow SLIs
		OAuthFlowCompletionRate: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_oauth_flow_attempts_total",
				Help: "Total OAuth flow attempts by result and latency bucket",
			},
			[]string{"result", "latency_bucket", "grant_type"},
		),
		OAuthFlowLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "auth_oauth_flow_duration_seconds",
				Help:    "Time spent completing OAuth flows",
				Buckets: prometheus.ExponentialBuckets(0.1, 2, 10), // 100ms to ~100s
			},
			[]string{"result", "grant_type"},
		),

		// User Registration SLIs
		UserRegistrationSuccessRate: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_user_registration_attempts_total",
				Help: "Total user registration attempts by result and latency bucket",
			},
			[]string{"result", "latency_bucket"},
		),
		UserRegistrationLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "auth_user_registration_duration_seconds",
				Help:    "Time spent on user registrations",
				Buckets: prometheus.ExponentialBuckets(0.05, 2, 10), // 50ms to ~50s
			},
			[]string{"result"},
		),

		// Permission Check SLIs
		PermissionCheckSuccessRate: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_permission_check_attempts_total",
				Help: "Total permission check attempts by result and latency bucket",
			},
			[]string{"result", "latency_bucket"},
		),
		PermissionCheckLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "auth_permission_check_duration_seconds",
				Help:    "Time spent checking permissions",
				Buckets: prometheus.ExponentialBuckets(0.001, 2, 8), // 1ms to ~256ms
			},
			[]string{"result"},
		),

		// Email Verification SLIs
		EmailVerificationSuccessRate: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_email_verification_attempts_total",
				Help: "Total email verification attempts by result and latency bucket",
			},
			[]string{"result", "latency_bucket"},
		),
		EmailVerificationLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "auth_email_verification_duration_seconds",
				Help:    "Time spent on email verifications",
				Buckets: prometheus.ExponentialBuckets(0.01, 2, 10), // 10ms to ~10s
			},
			[]string{"result"},
		),

		// Password Reset SLIs
		PasswordResetSuccessRate: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_password_reset_attempts_total",
				Help: "Total password reset attempts by result and latency bucket",
			},
			[]string{"result", "latency_bucket"},
		),
		PasswordResetLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "auth_password_reset_duration_seconds",
				Help:    "Time spent on password resets",
				Buckets: prometheus.ExponentialBuckets(0.01, 2, 10), // 10ms to ~10s
			},
			[]string{"result"},
		),

		// Availability SLIs
		ServiceAvailability: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "auth_service_availability",
				Help: "Service availability (1 = available, 0 = unavailable)",
			},
			[]string{"component"}, // component: api|database|redis
		),
		DatabaseAvailability: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "auth_database_availability",
				Help: "Database availability (1 = available, 0 = unavailable)",
			},
			[]string{"database"}, // database: primary|replica
		),
		RedisAvailability: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "auth_redis_availability",
				Help: "Redis availability (1 = available, 0 = unavailable)",
			},
			[]string{"role"}, // role: master|slave
		),

		// Business Volume SLIs
		DailyActiveUsers: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "auth_daily_active_users",
				Help: "Number of daily active users",
			},
		),
		MonthlyActiveUsers: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "auth_monthly_active_users",
				Help: "Number of monthly active users",
			},
		),
		ConcurrentUsers: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "auth_concurrent_users",
				Help: "Number of currently active concurrent users",
			},
		),
		PeakConcurrentUsers: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "auth_peak_concurrent_users",
				Help: "Peak concurrent users in the last 24 hours",
			},
		),

		// Security SLIs
		SuspiciousActivityRate: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_suspicious_activity_total",
				Help: "Total suspicious activity events by type",
			},
			[]string{"activity_type"}, // activity_type: multiple_failed_logins|unusual_location|brute_force
		),
		FailedLoginAttempts: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_failed_login_attempts_total",
				Help: "Total failed login attempts by reason",
			},
			[]string{"reason"}, // reason: invalid_credentials|account_locked|rate_limited
		),
		RateLimitViolations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_rate_limit_violations_total",
				Help: "Total rate limit violations by scope",
			},
			[]string{"scope"}, // scope: ip|user|email
		),
		CSRFTokenFailures: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_csrf_token_failures_total",
				Help: "Total CSRF token validation failures by reason",
			},
			[]string{"reason"}, // reason: missing|invalid|expired
		),

		// Business Error Rates
		ClientErrorRate: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_client_errors_total",
				Help: "Total client errors (4xx) by error code",
			},
			[]string{"error_code"},
		),
		ServerErrorRate: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_server_errors_total",
				Help: "Total server errors (5xx) by error code",
			},
			[]string{"error_code"},
		),
		BusinessLogicErrorRate: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_business_logic_errors_total",
				Help: "Total business logic errors by domain",
			},
			[]string{"domain"}, // domain: user_management|rbac|oauth|session
		),
	}
}

// RecordAuthenticationAttempt records an authentication attempt with SLI metrics
func (m *SLIMetrics) RecordAuthenticationAttempt(success bool, duration time.Duration) {
	result := "failure"
	if success {
		result = "success"
	}

	// Record latency histogram
	m.AuthenticationLatency.WithLabelValues(result).Observe(duration.Seconds())

	// Record success rate with latency bucket
	latencyBucket := getLatencyBucket(duration, 500*time.Millisecond) // SLO: 500ms
	m.AuthenticationSuccessRate.WithLabelValues(result, latencyBucket).Inc()
}

// RecordSessionCreation records a session creation attempt with SLI metrics
func (m *SLIMetrics) RecordSessionCreation(success bool, duration time.Duration) {
	result := "failure"
	if success {
		result = "success"
	}

	m.SessionCreationLatency.WithLabelValues(result).Observe(duration.Seconds())
	latencyBucket := getLatencyBucket(duration, 200*time.Millisecond) // SLO: 200ms
	m.SessionCreationSuccessRate.WithLabelValues(result, latencyBucket).Inc()
}

// RecordTokenValidation records a token validation attempt with SLI metrics
func (m *SLIMetrics) RecordTokenValidation(success bool, duration time.Duration) {
	result := "failure"
	if success {
		result = "success"
	}

	m.TokenValidationLatency.WithLabelValues(result).Observe(duration.Seconds())
	latencyBucket := getLatencyBucket(duration, 100*time.Millisecond) // SLO: 100ms
	m.TokenValidationSuccessRate.WithLabelValues(result, latencyBucket).Inc()
}

// RecordOAuthFlow records an OAuth flow completion with SLI metrics
func (m *SLIMetrics) RecordOAuthFlow(success bool, duration time.Duration, grantType string) {
	result := "failure"
	if success {
		result = "success"
	}

	m.OAuthFlowLatency.WithLabelValues(result, grantType).Observe(duration.Seconds())
	latencyBucket := getLatencyBucket(duration, 2*time.Second) // SLO: 2s
	m.OAuthFlowCompletionRate.WithLabelValues(result, latencyBucket, grantType).Inc()
}

// RecordUserRegistration records a user registration attempt with SLI metrics
func (m *SLIMetrics) RecordUserRegistration(success bool, duration time.Duration) {
	result := "failure"
	if success {
		result = "success"
	}

	m.UserRegistrationLatency.WithLabelValues(result).Observe(duration.Seconds())
	latencyBucket := getLatencyBucket(duration, 1*time.Second) // SLO: 1s
	m.UserRegistrationSuccessRate.WithLabelValues(result, latencyBucket).Inc()
}

// RecordPermissionCheck records a permission check with SLI metrics
func (m *SLIMetrics) RecordPermissionCheck(success bool, duration time.Duration) {
	result := "failure"
	if success {
		result = "success"
	}

	m.PermissionCheckLatency.WithLabelValues(result).Observe(duration.Seconds())
	latencyBucket := getLatencyBucket(duration, 50*time.Millisecond) // SLO: 50ms
	m.PermissionCheckSuccessRate.WithLabelValues(result, latencyBucket).Inc()
}

// getLatencyBucket categorizes request duration against SLO target
func getLatencyBucket(duration, sloTarget time.Duration) string {
	if duration <= sloTarget {
		return "fast" // Within SLO
	} else if duration <= sloTarget*2 {
		return "slow" // Exceeds SLO but not critical
	} else {
		return "timeout" // Critical latency
	}
}

// UpdateServiceAvailability updates service availability status
func (m *SLIMetrics) UpdateServiceAvailability(component string, available bool) {
	value := 0.0
	if available {
		value = 1.0
	}
	m.ServiceAvailability.WithLabelValues(component).Set(value)
}

// RecordSuspiciousActivity records suspicious activity events
func (m *SLIMetrics) RecordSuspiciousActivity(activityType string) {
	m.SuspiciousActivityRate.WithLabelValues(activityType).Inc()
}

// RecordCSRFTokenFailure records CSRF token validation failures
func (m *SLIMetrics) RecordCSRFTokenFailure(reason string) {
	m.CSRFTokenFailures.WithLabelValues(reason).Inc()
}
