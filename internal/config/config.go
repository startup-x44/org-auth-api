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

type CORSConfig struct {
	AllowedOrigins []string // List of allowed origins or patterns like "*.sprout.com"
}

type RateLimitConfig struct {
	LoginAttempts int // per 15 minutes per IP
	PasswordReset int // per hour per email
	APICalls      int // per minute per user
	Registration  int // per hour per IP
	MaxSessions   int // per user
}

type EmailConfig struct {
	Provider  string
	APIKey    string
	FromEmail string
	FromName  string
	ResetURL  string
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
			LoginAttempts: getEnvAsInt("RATE_LIMIT_LOGIN_ATTEMPTS", 5), // 5 per 15 min per IP
			PasswordReset: getEnvAsInt("RATE_LIMIT_PASSWORD_RESET", 3), // 3 per hour per email
			APICalls:      getEnvAsInt("RATE_LIMIT_API_CALLS", 1000),   // 1000 per minute per user
			Registration:  getEnvAsInt("RATE_LIMIT_REGISTRATION", 10),  // 10 per hour per IP
			MaxSessions:   getEnvAsInt("MAX_CONCURRENT_SESSIONS", 5),   // 5 per user
		},
		Email: EmailConfig{
			Provider:  getEnv("EMAIL_PROVIDER", "sendgrid"),
			APIKey:    getEnv("EMAIL_API_KEY", ""),
			FromEmail: getEnv("EMAIL_FROM_EMAIL", "noreply@example.com"),
			FromName:  getEnv("EMAIL_FROM_NAME", "Auth Service"),
			ResetURL:  getEnv("PASSWORD_RESET_URL", "https://app.example.com/reset-password"),
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
	if cfg.Environment == "production" {
		if cfg.Email.APIKey == "" {
			return errors.New("EMAIL_API_KEY must be set in production")
		}
		if cfg.Email.FromEmail == "" {
			return errors.New("EMAIL_FROM_EMAIL must be set in production")
		}
		if cfg.Email.ResetURL == "" {
			return errors.New("PASSWORD_RESET_URL must be set in production")
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
