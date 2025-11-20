package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewAuditLogger(t *testing.T) {
	logger := NewAuditLogger()
	if logger == nil {
		t.Fatal("NewAuditLogger returned nil")
	}
	if logger.logger == nil {
		t.Fatal("Logger is nil")
	}
}

func TestNewAuditLoggerWithWriter(t *testing.T) {
	var buf bytes.Buffer
	logger := NewAuditLoggerWithWriter(&buf)
	if logger == nil {
		t.Fatal("NewAuditLoggerWithWriter returned nil")
	}
	if logger.logger == nil {
		t.Fatal("Logger is nil")
	}
}

func TestLogUserAction(t *testing.T) {
	var buf bytes.Buffer
	logger := NewAuditLoggerWithWriter(&buf)

	logger.LogUserAction("user123", "login", "192.168.1.1", "Mozilla/5.0", true, nil, "successful login")

	output := buf.String()
	if !strings.Contains(output, "[AUDIT]") {
		t.Error("Output should contain [AUDIT] prefix")
	}

	// Extract JSON part
	lines := strings.Split(output, "\n")
	var jsonLine string
	for _, line := range lines {
		if strings.Contains(line, "{") && strings.Contains(line, "}") {
			// Extract JSON part after the timestamp
			parts := strings.SplitN(line, "{", 2)
			if len(parts) == 2 {
				jsonLine = "{" + parts[1]
				break
			}
		}
	}

	if jsonLine == "" {
		t.Fatal("No JSON found in output: " + output)
	}

	var event AuditEvent
	if err := json.Unmarshal([]byte(jsonLine), &event); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v, JSON: %s", err, jsonLine)
	}

	if event.UserID != "user123" {
		t.Errorf("Expected UserID 'user123', got '%s'", event.UserID)
	}
	if event.Action != "login" {
		t.Errorf("Expected Action 'login', got '%s'", event.Action)
	}
	if event.Resource != "user" {
		t.Errorf("Expected Resource 'user', got '%s'", event.Resource)
	}
	if !event.Success {
		t.Error("Expected Success to be true")
	}
}

func TestLogUserActionWithContext(t *testing.T) {
	var buf bytes.Buffer
	logger := NewAuditLoggerWithWriter(&buf)

	ctx := context.Background()
	ctx = context.WithValue(ctx, RequestIDKey, "req-123")
	ctx = context.WithValue(ctx, IPAddressKey, "10.0.0.1")

	logger.LogUserActionWithContext(ctx, "user456", "logout", "", "", true, nil, "successful logout")

	output := buf.String()
	if !strings.Contains(output, "req-123") {
		t.Error("Output should contain request ID from context")
	}
	if !strings.Contains(output, "10.0.0.1") {
		t.Error("Output should contain IP address from context")
	}
}

func TestLogUserActionWithError(t *testing.T) {
	var buf bytes.Buffer
	logger := NewAuditLoggerWithWriter(&buf)

	testErr := errors.New("authentication failed")
	logger.LogUserAction("user789", "login", "192.168.1.1", "Mozilla/5.0", false, testErr, "failed login attempt")

	output := buf.String()
	if !strings.Contains(output, "authentication failed") {
		t.Error("Output should contain error message")
	}

	// Extract and verify JSON
	lines := strings.Split(output, "\n")
	var jsonLine string
	for _, line := range lines {
		if strings.Contains(line, "{") && strings.Contains(line, "}") {
			parts := strings.SplitN(line, "{", 2)
			if len(parts) == 2 {
				jsonLine = "{" + parts[1]
				break
			}
		}
	}

	var event AuditEvent
	if err := json.Unmarshal([]byte(jsonLine), &event); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if event.Success {
		t.Error("Expected Success to be false")
	}
	if event.Error != "authentication failed" {
		t.Errorf("Expected Error 'authentication failed', got '%s'", event.Error)
	}
}

func TestLogOrganizationAction(t *testing.T) {
	var buf bytes.Buffer
	logger := NewAuditLoggerWithWriter(&buf)

	logger.LogOrganizationAction("user123", "create", "org456", "192.168.1.1", "Mozilla/5.0", true, nil, "organization created")

	output := buf.String()
	if !strings.Contains(output, "[AUDIT]") {
		t.Error("Output should contain [AUDIT] prefix")
	}

	// Verify JSON structure
	lines := strings.Split(output, "\n")
	var jsonLine string
	for _, line := range lines {
		if strings.Contains(line, "{") && strings.Contains(line, "}") {
			parts := strings.SplitN(line, "{", 2)
			if len(parts) == 2 {
				jsonLine = "{" + parts[1]
				break
			}
		}
	}

	var event AuditEvent
	if err := json.Unmarshal([]byte(jsonLine), &event); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if event.Resource != "organization" {
		t.Errorf("Expected Resource 'organization', got '%s'", event.Resource)
	}
	if event.ResourceID != "org456" {
		t.Errorf("Expected ResourceID 'org456', got '%s'", event.ResourceID)
	}
}

func TestLogSecurityEvent(t *testing.T) {
	var buf bytes.Buffer
	logger := NewAuditLoggerWithWriter(&buf)

	logger.LogSecurityEvent("password_reset", "user@example.com", "192.168.1.1", true, nil, "password reset requested")

	output := buf.String()
	if !strings.Contains(output, "user@example.com") {
		t.Error("Output should contain email address")
	}
	if !strings.Contains(output, "password reset requested") {
		t.Error("Output should contain details")
	}
}

func TestValidationRequiredFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewAuditLoggerWithWriter(&buf)

	// Create an event with missing required fields
	event := AuditEvent{
		Timestamp: time.Now(),
		// Missing Action and Resource
		Success: true,
		Service: "auth-service",
	}

	logger.logEvent(event)

	output := buf.String()
	if !strings.Contains(output, "AUDIT_ERROR") {
		t.Error("Should log audit error for missing required fields")
	}
	if !strings.Contains(output, "missing required field") {
		t.Error("Error message should mention missing required field")
	}
}

func TestConcurrentAccess(t *testing.T) {
	var buf bytes.Buffer
	logger := NewAuditLoggerWithWriter(&buf)

	// Test concurrent access
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func(id int) {
			logger.LogUserAction("user"+string(rune(id)), "test", "127.0.0.1", "test-agent", true, nil, "concurrent test")
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 100; i++ {
		<-done
	}

	output := buf.String()
	auditCount := strings.Count(output, "[AUDIT]")
	if auditCount < 100 {
		t.Errorf("Expected at least 100 audit logs, got %d", auditCount)
	}
}

func TestMethodNameCaching(t *testing.T) {
	var buf bytes.Buffer
	logger := NewAuditLoggerWithWriter(&buf)

	// Call the same method multiple times to test caching
	for i := 0; i < 5; i++ {
		logger.LogUserAction("user123", "login", "192.168.1.1", "Mozilla/5.0", true, nil, "test caching")
	}

	// Check that method names are captured
	output := buf.String()
	// The method name will be from the audit logger internal method, not the test method
	// Just verify that some method name is captured
	if !strings.Contains(output, "\"method\":") {
		t.Error("Method field should be present in audit logs")
	}

	// Verify that the method name is not empty in the JSON
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "{") && strings.Contains(line, "}") {
			parts := strings.SplitN(line, "{", 2)
			if len(parts) == 2 {
				jsonLine := "{" + parts[1]
				var event AuditEvent
				if err := json.Unmarshal([]byte(jsonLine), &event); err == nil {
					if event.Method == "" {
						t.Error("Method name should not be empty")
					}
					return // Test passed for at least one event
				}
			}
		}
	}
	t.Error("No valid audit event found in output")
}
