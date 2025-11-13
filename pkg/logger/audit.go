package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"
)

// AuditEvent represents an audit log event
type AuditEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	UserID      string    `json:"user_id,omitempty"`
	Action      string    `json:"action"`
	Resource    string    `json:"resource"`
	ResourceID  string    `json:"resource_id,omitempty"`
	IPAddress   string    `json:"ip_address,omitempty"`
	UserAgent   string    `json:"user_agent,omitempty"`
	Details     string    `json:"details,omitempty"`
	Success     bool      `json:"success"`
	Error       string    `json:"error,omitempty"`
	Service     string    `json:"service"`
	Method      string    `json:"method,omitempty"`
}

// AuditLogger handles structured audit logging
type AuditLogger struct {
	logger *log.Logger
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		logger: log.New(os.Stdout, "[AUDIT] ", log.LstdFlags),
	}
}

// NewAuditLoggerWithWriter creates a new audit logger with custom writer
func NewAuditLoggerWithWriter(writer io.Writer) *AuditLogger {
	return &AuditLogger{
		logger: log.New(writer, "[AUDIT] ", log.LstdFlags),
	}
}

// LogAdminAction logs an admin action
func (a *AuditLogger) LogAdminAction(userID, action, resource, resourceID, ipAddress, userAgent string, success bool, err error, details string) {
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

	// Get calling method name
	if pc, _, _, ok := runtime.Caller(1); ok {
		if fn := runtime.FuncForPC(pc); fn != nil {
			event.Method = fn.Name()
		}
	}

	a.logEvent(event)
}

// LogUserAction logs a user action
func (a *AuditLogger) LogUserAction(userID, action, ipAddress, userAgent string, success bool, err error, details string) {
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

	// Get calling method name
	if pc, _, _, ok := runtime.Caller(1); ok {
		if fn := runtime.FuncForPC(pc); fn != nil {
			event.Method = fn.Name()
		}
	}

	a.logEvent(event)
}

// LogTenantAction logs a tenant management action
func (a *AuditLogger) LogTenantAction(userID, action, tenantID, ipAddress, userAgent string, success bool, err error, details string) {
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

	// Get calling method name
	if pc, _, _, ok := runtime.Caller(1); ok {
		if fn := runtime.FuncForPC(pc); fn != nil {
			event.Method = fn.Name()
		}
	}

	a.logEvent(event)
}

// LogSecurityEvent logs a security-related event
func (a *AuditLogger) LogSecurityEvent(action, email, ipAddress string, success bool, err error, details string) {
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

	// Get calling method name
	if pc, _, _, ok := runtime.Caller(1); ok {
		if fn := runtime.FuncForPC(pc); fn != nil {
			event.Method = fn.Name()
		}
	}

	a.logEvent(event)
}

// LogSystemEvent logs a system-level event
func (a *AuditLogger) LogSystemEvent(userID, action, resource, resourceID, ipAddress, userAgent string, success bool, err error, details string) {
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

	// Get calling method name
	if pc, _, _, ok := runtime.Caller(1); ok {
		if fn := runtime.FuncForPC(pc); fn != nil {
			event.Method = fn.Name()
		}
	}

	a.logEvent(event)
}

// logEvent logs the audit event as JSON
func (a *AuditLogger) logEvent(event AuditEvent) {
	jsonData, err := json.Marshal(event)
	if err != nil {
		// Fallback to basic logging if JSON marshaling fails
		a.logger.Printf("ERROR: Failed to marshal audit event: %v", err)
		return
	}

	a.logger.Println(string(jsonData))
}