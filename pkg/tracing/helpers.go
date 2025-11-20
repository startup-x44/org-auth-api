package tracing

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Helper functions for adding traces to business logic

// StartDBSpan starts a span for a database operation
func StartDBSpan(ctx context.Context, operation string, table string) (context.Context, trace.Span) {
	ctx, span := StartSpan(ctx, "db."+operation)
	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", operation),
		attribute.String("db.table", table),
	)
	return ctx, span
}

// StartRedisSpan starts a span for a Redis operation
func StartRedisSpan(ctx context.Context, operation string, key string) (context.Context, trace.Span) {
	ctx, span := StartSpan(ctx, "redis."+operation)
	span.SetAttributes(
		attribute.String("db.system", "redis"),
		attribute.String("db.operation", operation),
		attribute.String("db.key", key),
	)
	return ctx, span
}

// StartServiceSpan starts a span for a service operation
func StartServiceSpan(ctx context.Context, service string, operation string) (context.Context, trace.Span) {
	ctx, span := StartSpan(ctx, service+"."+operation)
	span.SetAttributes(
		attribute.String("service", service),
		attribute.String("operation", operation),
	)
	return ctx, span
}

// RecordError records an error on the current span
func RecordError(span trace.Span, err error) {
	if err != nil && span.IsRecording() {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// RecordSuccess marks the span as successful
func RecordSuccess(span trace.Span) {
	if span.IsRecording() {
		span.SetStatus(codes.Ok, "")
	}
}

// AddAttributes adds attributes to the current span
func AddAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	if span.IsRecording() {
		span.SetAttributes(attrs...)
	}
}
