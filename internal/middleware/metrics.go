package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"auth-service/pkg/metrics"
)

// MetricsMiddleware collects HTTP request metrics
func MetricsMiddleware() gin.HandlerFunc {
	m := metrics.GetMetrics()

	return func(c *gin.Context) {
		start := time.Now()

		// Get request size
		reqSize := c.Request.ContentLength

		// Create a custom response writer to capture response size
		body := make([]byte, 0)
		blw := &bodyLogWriter{body: &body, ResponseWriter: c.Writer}
		c.Writer = blw

		// Process request
		c.Next()

		// Calculate metrics
		duration := time.Since(start)
		status := strconv.Itoa(c.Writer.Status())

		// CRITICAL: Use normalized path only - NEVER c.Request.URL.Path
		// This prevents cardinality explosion from dynamic route parameters
		path := c.FullPath()
		if path == "" {
			path = "unknown" // Fallback for 404s or unmatched routes
		}

		// Get response size
		respSize := int64(len(*blw.body))

		// Record HTTP request metrics
		m.RecordHTTPRequest(c.Request.Method, path, status, duration, reqSize, respSize)
	}
}

// bodyLogWriter wraps gin.ResponseWriter to capture response body size
type bodyLogWriter struct {
	gin.ResponseWriter
	body *[]byte
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	*w.body = append(*w.body, b...)
	return w.ResponseWriter.Write(b)
}

func (w *bodyLogWriter) WriteString(s string) (int, error) {
	*w.body = append(*w.body, []byte(s)...)
	return w.ResponseWriter.WriteString(s)
}
