package tracing

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"
)

// Config holds tracing configuration
type Config struct {
	Enabled      bool
	ServiceName  string
	Environment  string
	ExporterType string // "otlp", "stdout", "jaeger"
	OTLPEndpoint string // e.g., "localhost:4317" for OTLP gRPC
	OTLPInsecure bool   // Use insecure connection (disable TLS)
	SamplingRate float64
}

// TracerProvider wraps the OpenTelemetry tracer provider
type TracerProvider struct {
	provider *sdktrace.TracerProvider
	config   *Config
}

// Initialize sets up OpenTelemetry tracing
func Initialize(cfg *Config) (*TracerProvider, error) {
	if !cfg.Enabled {
		// Use noop tracer when disabled
		otel.SetTracerProvider(trace.NewNoopTracerProvider())
		return &TracerProvider{
			provider: nil,
			config:   cfg,
		}, nil
	}

	// SECURITY: Prevent insecure OTLP in production
	if cfg.Environment == "production" && cfg.ExporterType == "otlp" && cfg.OTLPInsecure {
		return nil, fmt.Errorf("CRITICAL SECURITY ERROR: insecure OTLP is not allowed in production environment")
	}

	// Create exporter based on configuration
	var exporter sdktrace.SpanExporter
	var err error

	switch cfg.ExporterType {
	case "otlp":
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		}

		if cfg.OTLPInsecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		} else {
			opts = append(opts, otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
		}

		exporter, err = otlptracegrpc.New(context.Background(), opts...)
	case "stdout":
		exporter, err = stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
			stdouttrace.WithWriter(os.Stdout),
		)
	default:
		return nil, fmt.Errorf("unsupported exporter type: %s", cfg.ExporterType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(getServiceVersion()),
			semconv.DeploymentEnvironment(cfg.Environment),
			attribute.String("service.namespace", "NiloAuth"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider with sampling
	var sampler sdktrace.Sampler
	if cfg.SamplingRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if cfg.SamplingRate <= 0.0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(cfg.SamplingRate)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	return &TracerProvider{
		provider: tp,
		config:   cfg,
	}, nil
}

// Shutdown gracefully shuts down the tracer provider
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if tp.provider != nil {
		return tp.provider.Shutdown(ctx)
	}
	return nil
}

// Tracer returns a tracer for the given name
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// StartSpan starts a new span with the given name
// Uses proper namespace format: "component.operation"
func StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	// Get tracer name from context or use default
	// This allows proper namespace segmentation
	tracer := otel.Tracer("auth-service")
	return tracer.Start(ctx, spanName, opts...)
}

// getServiceVersion returns the service version
// This should be set via build flags in production
func getServiceVersion() string {
	version := os.Getenv("SERVICE_VERSION")
	if version == "" {
		return "1.0.0-dev"
	}
	return version
}
