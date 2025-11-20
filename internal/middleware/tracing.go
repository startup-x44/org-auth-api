package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TracingConfig holds configuration for tracing middleware
type TracingConfig struct {
	ExcludePaths []string
}

// TracingMiddleware returns OpenTelemetry tracing middleware for Gin
func TracingMiddleware(serviceName string, configs ...TracingConfig) gin.HandlerFunc {
	var excludePaths []string
	if len(configs) > 0 {
		excludePaths = configs[0].ExcludePaths
	}

	otelMiddleware := otelgin.Middleware(serviceName)

	return func(c *gin.Context) {
		// Skip tracing for excluded paths (health checks, metrics)
		for _, path := range excludePaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		otelMiddleware(c)
	}
}

// AddSpanAttributes adds custom attributes to the current span
func AddSpanAttributes(c *gin.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(c.Request.Context())
	if span.IsRecording() {
		span.SetAttributes(attrs...)
	}
}

// AddSpanEvent adds an event to the current span
func AddSpanEvent(c *gin.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(c.Request.Context())
	if span.IsRecording() {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// SetSpanStatus sets the status of the current span
func SetSpanStatus(c *gin.Context, code codes.Code, description string) {
	span := trace.SpanFromContext(c.Request.Context())
	if span.IsRecording() {
		span.SetStatus(code, description)
	}
}
