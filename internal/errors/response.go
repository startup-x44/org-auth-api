package errors

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SendErrorResponse sends a structured error response with proper HTTP status
func SendErrorResponse(c *gin.Context, code ErrorCode, message string, details interface{}) {
	requestID := getRequestID(c)
	errorResp := NewErrorResponse(code, message, details, requestID)

	httpStatus := code.GetHTTPStatus()
	c.JSON(httpStatus, errorResp)
}

// SendSuccessResponse sends a structured success response
func SendSuccessResponse(c *gin.Context, message string, data interface{}) {
	requestID := getRequestID(c)
	c.JSON(200, gin.H{
		"success":    true,
		"message":    message,
		"data":       data,
		"request_id": requestID,
		"timestamp":  getTimestamp(),
	})
}

// getRequestID extracts or generates a request ID
func getRequestID(c *gin.Context) string {
	// Try to get from context first (set by middleware)
	if requestID, exists := c.Get("request_id"); exists {
		if rid, ok := requestID.(string); ok {
			return rid
		}
	}

	// Try from header
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}

	// Generate new one
	return uuid.New().String()
}

// getTimestamp returns current timestamp in RFC3339 format
func getTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
