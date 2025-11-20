package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ContextKey type for context keys
type ContextKey string

const (
	// Context keys for structured logging
	RequestIDKey ContextKey = "request_id"
	TraceIDKey   ContextKey = "trace_id"
	UserIDKey    ContextKey = "user_id"
	OrgIDKey     ContextKey = "org_id"
	SessionIDKey ContextKey = "session_id"
	ClientIDKey  ContextKey = "client_id"
	IPAddressKey ContextKey = "ip_address"
	UserAgentKey ContextKey = "user_agent"
)

// Config represents logger configuration
type Config struct {
	Level        string // debug, info, warn, error
	Format       string // json, console
	Output       io.Writer
	EnableCaller bool
	TimeFormat   string
}

// DefaultConfig returns default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:        "info",
		Format:       "json",
		Output:       os.Stdout,
		EnableCaller: true,
		TimeFormat:   time.RFC3339,
	}
}

// Initialize sets up the global logger
func Initialize(cfg *Config) {
	// Set log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure output format
	var output io.Writer = cfg.Output
	if cfg.Format == "console" {
		output = zerolog.ConsoleWriter{
			Out:        cfg.Output,
			TimeFormat: cfg.TimeFormat,
			NoColor:    false,
		}
	}

	// Configure logger
	logger := zerolog.New(output).With().Timestamp()

	if cfg.EnableCaller {
		logger = logger.Caller()
	}

	log.Logger = logger.Logger()
}

// FromContext extracts a logger with context fields from the given context
func FromContext(ctx context.Context) *zerolog.Logger {
	logger := log.With()

	// Add request_id if present
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		logger = logger.Str("request_id", requestID)
	}

	// Add trace_id if present
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
		logger = logger.Str("trace_id", traceID)
	}

	// Add user_id if present
	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		logger = logger.Str("user_id", userID)
	}

	// Add org_id if present
	if orgID, ok := ctx.Value(OrgIDKey).(string); ok && orgID != "" {
		logger = logger.Str("org_id", orgID)
	}

	// Add session_id if present
	if sessionID, ok := ctx.Value(SessionIDKey).(string); ok && sessionID != "" {
		logger = logger.Str("session_id", sessionID)
	}

	// Add client_id if present
	if clientID, ok := ctx.Value(ClientIDKey).(string); ok && clientID != "" {
		logger = logger.Str("client_id", clientID)
	}

	// Add ip_address if present
	if ipAddress, ok := ctx.Value(IPAddressKey).(string); ok && ipAddress != "" {
		logger = logger.Str("ip_address", ipAddress)
	}

	// Add user_agent if present
	if userAgent, ok := ctx.Value(UserAgentKey).(string); ok && userAgent != "" {
		logger = logger.Str("user_agent", userAgent)
	}

	l := logger.Logger()
	return &l
}

// WithContext adds context fields to the logger
func WithContext(ctx context.Context, key ContextKey, value string) context.Context {
	return context.WithValue(ctx, key, value)
}

// Info logs an info message with context
func Info(ctx context.Context) *zerolog.Event {
	return FromContext(ctx).Info()
}

// Debug logs a debug message with context
func Debug(ctx context.Context) *zerolog.Event {
	return FromContext(ctx).Debug()
}

// Warn logs a warning message with context
func Warn(ctx context.Context) *zerolog.Event {
	return FromContext(ctx).Warn()
}

// Error logs an error message with context
func Error(ctx context.Context) *zerolog.Event {
	return FromContext(ctx).Error()
}

// Fatal logs a fatal message with context and exits
func Fatal(ctx context.Context) *zerolog.Event {
	return FromContext(ctx).Fatal()
}

// Panic logs a panic message with context and panics
func Panic(ctx context.Context) *zerolog.Event {
	return FromContext(ctx).Panic()
}

// WithFields adds multiple fields to a log event
func WithFields(event *zerolog.Event, fields map[string]interface{}) *zerolog.Event {
	for key, value := range fields {
		switch v := value.(type) {
		case string:
			event = event.Str(key, v)
		case int:
			event = event.Int(key, v)
		case int64:
			event = event.Int64(key, v)
		case float64:
			event = event.Float64(key, v)
		case bool:
			event = event.Bool(key, v)
		case error:
			event = event.Err(v)
		case time.Duration:
			event = event.Dur(key, v)
		default:
			event = event.Interface(key, v)
		}
	}
	return event
}

// Global convenience functions (without context)
func InfoMsg(msg string, fields ...map[string]interface{}) {
	event := log.Info()
	if len(fields) > 0 {
		event = WithFields(event, fields[0])
	}
	event.Msg(msg)
}

func DebugMsg(msg string, fields ...map[string]interface{}) {
	event := log.Debug()
	if len(fields) > 0 {
		event = WithFields(event, fields[0])
	}
	event.Msg(msg)
}

func WarnMsg(msg string, fields ...map[string]interface{}) {
	event := log.Warn()
	if len(fields) > 0 {
		event = WithFields(event, fields[0])
	}
	event.Msg(msg)
}

func ErrorMsg(msg string, err error, fields ...map[string]interface{}) {
	event := log.Error().Err(err)
	if len(fields) > 0 {
		event = WithFields(event, fields[0])
	}
	event.Msg(msg)
}

func FatalMsg(msg string, err error, fields ...map[string]interface{}) {
	event := log.Fatal().Err(err)
	if len(fields) > 0 {
		event = WithFields(event, fields[0])
	}
	event.Msg(msg)
}
