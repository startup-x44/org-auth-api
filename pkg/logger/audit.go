package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

// AuditEvent represents an audit log event
type AuditEvent struct {
	Timestamp  time.Time `json:"timestamp"`
	UserID     string    `json:"user_id,omitempty"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resource_id,omitempty"`
	IPAddress  string    `json:"ip_address,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	Details    string    `json:"details,omitempty"`
	Success    bool      `json:"success"`
	Error      string    `json:"error,omitempty"`
	Service    string    `json:"service"`
	Method     string    `json:"method,omitempty"`
}

// AuditLogger handles structured audit logging
type AuditLogger struct {
	logger      *log.Logger
	mu          sync.RWMutex // Protects concurrent access to logger
	methodCache sync.Map     // Cache for method names to improve performance
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		logger: log.New(os.Stdout, "[AUDIT] ", log.LstdFlags|log.Lmicroseconds),
	}
}

// NewAuditLoggerWithWriter creates a new audit logger with custom writer
func NewAuditLoggerWithWriter(writer io.Writer) *AuditLogger {
	return &AuditLogger{
		logger: log.New(writer, "[AUDIT] ", log.LstdFlags|log.Lmicroseconds),
	}
}

// LogAdminAction logs an admin action
func (a *AuditLogger) LogAdminAction(userID, action, resource, resourceID, ipAddress, userAgent string, success bool, err error, details string) {
	a.LogAdminActionWithContext(context.Background(), userID, action, resource, resourceID, ipAddress, userAgent, success, err, details)
}

// LogAdminActionWithContext logs an admin action with context
func (a *AuditLogger) LogAdminActionWithContext(ctx context.Context, userID, action, resource, resourceID, ipAddress, userAgent string, success bool, err error, details string) {
	event := AuditEvent{
		Timestamp:  time.Now(),
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Details:    details,
		Success:    success,
		Service:    "auth-service",
	}

	if err != nil {
		event.Error = err.Error()
	}

	// Extract additional context values
	a.extractContextValues(ctx, &event)

	// Get calling method name efficiently (check 2 levels up for the actual caller when using WithContext)
	event.Method = a.getMethodName(2)

	a.logEvent(event)
}

// LogUserAction logs a user action
func (a *AuditLogger) LogUserAction(userID, action, ipAddress, userAgent string, success bool, err error, details string) {
	a.LogUserActionWithContext(context.Background(), userID, action, ipAddress, userAgent, success, err, details)
}

// LogUserActionWithContext logs a user action with context
func (a *AuditLogger) LogUserActionWithContext(ctx context.Context, userID, action, ipAddress, userAgent string, success bool, err error, details string) {
	event := AuditEvent{
		Timestamp: time.Now(),
		UserID:    userID,
		Action:    action,
		Resource:  "user",
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Details:   details,
		Success:   success,
		Service:   "auth-service",
	}

	if err != nil {
		event.Error = err.Error()
	}

	// Extract additional context values
	a.extractContextValues(ctx, &event)

	// Get calling method name efficiently (check 2 levels up for the actual caller when using WithContext)
	event.Method = a.getMethodName(2)

	a.logEvent(event)
}

// LogTenantAction logs a tenant management action
func (a *AuditLogger) LogTenantAction(userID, action, tenantID, ipAddress, userAgent string, success bool, err error, details string) {
	a.LogTenantActionWithContext(context.Background(), userID, action, tenantID, ipAddress, userAgent, success, err, details)
}

// LogTenantActionWithContext logs a tenant management action with context
func (a *AuditLogger) LogTenantActionWithContext(ctx context.Context, userID, action, tenantID, ipAddress, userAgent string, success bool, err error, details string) {
	event := AuditEvent{
		Timestamp:  time.Now(),
		UserID:     userID,
		Action:     action,
		Resource:   "tenant",
		ResourceID: tenantID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Details:    details,
		Success:    success,
		Service:    "auth-service",
	}

	if err != nil {
		event.Error = err.Error()
	}

	// Extract additional context values
	a.extractContextValues(ctx, &event)

	// Get calling method name efficiently (check 2 levels up for the actual caller when using WithContext)
	event.Method = a.getMethodName(2)

	a.logEvent(event)
}

// LogOrganizationAction logs an organization management action
func (a *AuditLogger) LogOrganizationAction(userID, action, orgID, ipAddress, userAgent string, success bool, err error, details string) {
	a.LogOrganizationActionWithContext(context.Background(), userID, action, orgID, ipAddress, userAgent, success, err, details)
}

// LogOrganizationActionWithContext logs an organization management action with context
func (a *AuditLogger) LogOrganizationActionWithContext(ctx context.Context, userID, action, orgID, ipAddress, userAgent string, success bool, err error, details string) {
	event := AuditEvent{
		Timestamp:  time.Now(),
		UserID:     userID,
		Action:     action,
		Resource:   "organization",
		ResourceID: orgID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Details:    details,
		Success:    success,
		Service:    "auth-service",
	}

	if err != nil {
		event.Error = err.Error()
	}

	// Extract additional context values
	a.extractContextValues(ctx, &event)

	// Get calling method name efficiently (check 2 levels up for the actual caller when using WithContext)
	event.Method = a.getMethodName(2)

	a.logEvent(event)
}

// LogSecurityEvent logs a security-related event
func (a *AuditLogger) LogSecurityEvent(action, email, ipAddress string, success bool, err error, details string) {
	a.LogSecurityEventWithContext(context.Background(), action, email, ipAddress, success, err, details)
}

// LogSecurityEventWithContext logs a security-related event with context
func (a *AuditLogger) LogSecurityEventWithContext(ctx context.Context, action, email, ipAddress string, success bool, err error, details string) {
	event := AuditEvent{
		Timestamp: time.Now(),
		Action:    action,
		Resource:  "security",
		IPAddress: ipAddress,
		Details:   fmt.Sprintf("email=%s, %s", email, details),
		Success:   success,
		Service:   "auth-service",
	}

	if err != nil {
		event.Error = err.Error()
	}

	// Extract additional context values
	a.extractContextValues(ctx, &event)

	// Get calling method name efficiently (check 2 levels up for the actual caller when using WithContext)
	event.Method = a.getMethodName(2)

	a.logEvent(event)
}

// LogSystemEvent logs a system-level event
func (a *AuditLogger) LogSystemEvent(userID, action, resource, resourceID, ipAddress, userAgent string, success bool, err error, details string) {
	a.LogSystemEventWithContext(context.Background(), userID, action, resource, resourceID, ipAddress, userAgent, success, err, details)
}

// LogSystemEventWithContext logs a system-level event with context
func (a *AuditLogger) LogSystemEventWithContext(ctx context.Context, userID, action, resource, resourceID, ipAddress, userAgent string, success bool, err error, details string) {
	event := AuditEvent{
		Timestamp:  time.Now(),
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Details:    details,
		Success:    success,
		Service:    "auth-service",
	}

	if err != nil {
		event.Error = err.Error()
	}

	// Extract additional context values
	a.extractContextValues(ctx, &event)

	// Get calling method name efficiently (check 2 levels up for the actual caller when using WithContext)
	event.Method = a.getMethodName(2)

	a.logEvent(event)
}

// logEvent logs the audit event as JSON with thread safety
func (a *AuditLogger) logEvent(event AuditEvent) {
	// Validate required fields
	if event.Action == "" {
		a.logError("audit event missing required field: action")
		return
	}
	if event.Resource == "" {
		a.logError("audit event missing required field: resource")
		return
	}

	jsonData, err := json.Marshal(event)
	if err != nil {
		// Fallback to basic logging if JSON marshaling fails
		a.logError(fmt.Sprintf("Failed to marshal audit event: %v, event: %+v", err, event))
		return
	}

	// Thread-safe logging
	a.mu.Lock()
	defer a.mu.Unlock()
	a.logger.Println(string(jsonData))
}

// logError logs errors in audit logging itself
func (a *AuditLogger) logError(message string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.logger.Printf("AUDIT_ERROR: %s", message)
}

// getMethodName efficiently gets the calling method name with caching
func (a *AuditLogger) getMethodName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}

	// Check cache first
	if cached, exists := a.methodCache.Load(pc); exists {
		return cached.(string)
	}

	// Get method name and cache it
	methodName := ""
	if fn := runtime.FuncForPC(pc); fn != nil {
		methodName = fn.Name()
	}

	a.methodCache.Store(pc, methodName)
	return methodName
}

// extractContextValues extracts common values from context for audit logging
func (a *AuditLogger) extractContextValues(ctx context.Context, event *AuditEvent) {
	if ctx == nil {
		return
	}

	// Extract request ID if available
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		if event.Details == "" {
			event.Details = fmt.Sprintf("request_id=%s", requestID)
		} else {
			event.Details = fmt.Sprintf("request_id=%s, %s", requestID, event.Details)
		}
	}

	// Extract IP address if not already set
	if event.IPAddress == "" {
		if ip, ok := ctx.Value(IPAddressKey).(string); ok {
			event.IPAddress = ip
		}
	}

	// Extract user agent if not already set
	if event.UserAgent == "" {
		if ua, ok := ctx.Value(UserAgentKey).(string); ok {
			event.UserAgent = ua
		}
	}

	// Extract user ID if not already set
	if event.UserID == "" {
		if userID, ok := ctx.Value(UserIDKey).(string); ok {
			event.UserID = userID
		}
	}
}
