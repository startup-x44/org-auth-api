package config

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	CORS        CORSConfig
	RateLimit   RateLimitConfig
	Email       EmailConfig
	Logging     LoggingConfig
	Tracing     TracingConfig
	Environment string
}

type ServerConfig struct {
	Port int
	Host string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  int // in minutes
	RefreshTokenTTL int // in days
	Issuer          string
	SigningMethod   string
}

type LoggingConfig struct {
	Level  string // debug, info, warn, error
	Format string // json, console
}

type TracingConfig struct {
	Enabled      bool    // Enable/disable distributed tracing
	ServiceName  string  // Service name for traces
	ExporterType string  // "otlp", "stdout", "jaeger"
	OTLPEndpoint string  // OTLP gRPC endpoint (e.g., "localhost:4317")
	OTLPInsecure bool    // Use insecure connection (disable TLS)
	SamplingRate float64 // Trace sampling rate (0.0 to 1.0)
}

type CORSConfig struct {
	AllowedOrigins []string // List of allowed origins or patterns like "*.sprout.com"
}

type RateLimitConfig struct {
	Enabled             bool // Enable/disable rate limiting globally
	LoginAttempts       int  // Max attempts per LoginWindow
	LoginWindow         int  // Window in seconds (default: 900 = 15 min)
	PasswordReset       int  // Max attempts per PasswordResetWindow
	PasswordResetWindow int  // Window in seconds (default: 3600 = 1 hour)
	TokenRefresh        int  // Max token refresh attempts per TokenRefreshWindow
	TokenRefreshWindow  int  // Window in seconds (default: 60 = 1 min)
	Registration        int  // Max registrations per RegistrationWindow
	RegistrationWindow  int  // Window in seconds (default: 3600 = 1 hour)
	OAuth2Token         int  // Max OAuth2 token requests per OAuth2TokenWindow
	OAuth2TokenWindow   int  // Window in seconds (default: 60 = 1 min)
	APICalls            int  // General API calls per APICallsWindow
	APICallsWindow      int  // Window in seconds (default: 60 = 1 min)
	MaxSessions         int  // Max concurrent sessions per user
}

type EmailConfig struct {
	Host        string
	Port        int
	Username    string
	Password    string
	FromEmail   string
	FromName    string
	Enabled     bool
	FrontendURL string
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnvAsInt("SERVER_PORT", 8080),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DATABASE_HOST", "localhost"),
			Port:     getEnvAsInt("DATABASE_PORT", 5432),
			User:     getEnv("DATABASE_USER", "postgres"),
			Password: getEnv("DATABASE_PASSWORD", "password"),
			Name:     getEnv("DATABASE_NAME", "auth_service"),
			SSLMode:  getEnv("DATABASE_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
			AccessTokenTTL:  getEnvAsInt("JWT_ACCESS_TOKEN_EXPIRY", 60),  // 1 hour
			RefreshTokenTTL: getEnvAsInt("JWT_REFRESH_TOKEN_EXPIRY", 30), // 30 days
			Issuer:          getEnv("JWT_ISSUER", "auth-service"),
			SigningMethod:   getEnv("JWT_SIGNING_METHOD", "HS256"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "*.localhost:3000"}),
		},
		RateLimit: RateLimitConfig{
			Enabled:             getEnv("RATE_LIMIT_ENABLED", "true") == "true",
			LoginAttempts:       getEnvAsInt("RATE_LIMIT_LOGIN_ATTEMPTS", 5),
			LoginWindow:         getEnvAsInt("RATE_LIMIT_LOGIN_WINDOW", 900), // 15 minutes
			PasswordReset:       getEnvAsInt("RATE_LIMIT_PASSWORD_RESET", 3),
			PasswordResetWindow: getEnvAsInt("RATE_LIMIT_PASSWORD_RESET_WINDOW", 3600), // 1 hour
			TokenRefresh:        getEnvAsInt("RATE_LIMIT_TOKEN_REFRESH", 10),
			TokenRefreshWindow:  getEnvAsInt("RATE_LIMIT_TOKEN_REFRESH_WINDOW", 60), // 1 minute
			Registration:        getEnvAsInt("RATE_LIMIT_REGISTRATION", 10),
			RegistrationWindow:  getEnvAsInt("RATE_LIMIT_REGISTRATION_WINDOW", 3600), // 1 hour
			OAuth2Token:         getEnvAsInt("RATE_LIMIT_OAUTH2_TOKEN", 20),
			OAuth2TokenWindow:   getEnvAsInt("RATE_LIMIT_OAUTH2_TOKEN_WINDOW", 60), // 1 minute
			APICalls:            getEnvAsInt("RATE_LIMIT_API_CALLS", 1000),
			APICallsWindow:      getEnvAsInt("RATE_LIMIT_API_CALLS_WINDOW", 60), // 1 minute
			MaxSessions:         getEnvAsInt("MAX_CONCURRENT_SESSIONS", 5),
		},
		Email: EmailConfig{
			Host:        getEnv("SMTP_HOST", "sandbox.smtp.mailtrap.io"),
			Port:        getEnvAsInt("SMTP_PORT", 587),
			Username:    getEnv("SMTP_USERNAME", ""),
			Password:    getEnv("SMTP_PASSWORD", ""),
			FromEmail:   getEnv("SMTP_FROM_EMAIL", "noreply@example.com"),
			FromName:    getEnv("SMTP_FROM_NAME", "Auth Service"),
			Enabled:     getEnv("EMAIL_ENABLED", "true") == "true",
			FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Tracing: TracingConfig{
			Enabled:      getEnv("TRACING_ENABLED", "false") == "true",
			ServiceName:  getEnv("TRACING_SERVICE_NAME", "auth-service"),
			ExporterType: getEnv("TRACING_EXPORTER", "stdout"), // otlp, stdout
			OTLPEndpoint: getEnv("TRACING_OTLP_ENDPOINT", "localhost:4317"),
			OTLPInsecure: getEnv("TRACING_OTLP_INSECURE", "false") == "true",
			SamplingRate: getEnvAsFloat("TRACING_SAMPLING_RATE", 1.0),
		},
		Environment: getEnv("ENVIRONMENT", "development"),
	}

	// Validate sensitive environment variables
	if err := validateConfig(cfg); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated parsing
		if value == "" {
			return defaultValue
		}
		// Split by comma and trim spaces
		parts := strings.Split(value, ",")
		for i, part := range parts {
			parts[i] = strings.TrimSpace(part)
		}
		return parts
	}
	return defaultValue
}

// validateConfig validates sensitive configuration values
func validateConfig(cfg *Config) error {
	// Validate JWT secret - only strict validation in production
	if cfg.Environment == "production" {
		if cfg.JWT.Secret == "" || cfg.JWT.Secret == "your-super-secret-jwt-key-change-in-production" {
			return errors.New("JWT_SECRET must be set to a secure value in production")
		}
		if len(cfg.JWT.Secret) < 32 {
			return errors.New("JWT_SECRET must be at least 32 characters long")
		}
	} else {
		// Development mode - just ensure it's not empty
		if cfg.JWT.Secret == "" {
			return errors.New("JWT_SECRET must be set")
		}
	}

	// Validate database password
	if cfg.Environment == "production" && cfg.Database.Password == "password" {
		return errors.New("DATABASE_PASSWORD must be set to a secure value in production")
	}

	// Validate Redis password in production
	if cfg.Environment == "production" && cfg.Redis.Password == "" {
		log.Println("Warning: REDIS_PASSWORD is not set in production environment")
	}

	// Validate email configuration in production
	if cfg.Environment == "production" && cfg.Email.Enabled {
		if cfg.Email.Username == "" || cfg.Email.Password == "" {
			return errors.New("SMTP credentials must be set when email is enabled in production")
		}
		if cfg.Email.FromEmail == "" {
			return errors.New("SMTP_FROM_EMAIL must be set in production")
		}
		if cfg.Email.FrontendURL == "" {
			return errors.New("FRONTEND_URL must be set in production")
		}
	}

	// Validate rate limiting settings
	if cfg.RateLimit.LoginAttempts < 1 {
		return errors.New("RATE_LIMIT_LOGIN_ATTEMPTS must be at least 1")
	}
	if cfg.RateLimit.APICalls < 1 {
		return errors.New("RATE_LIMIT_API_CALLS must be at least 1")
	}

	return nil
}
