package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	RateLimit   RateLimitConfig
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
	Secret           string
	AccessTokenTTL   int // in minutes
	RefreshTokenTTL  int // in days
	Issuer           string
	SigningMethod    string
}

type RateLimitConfig struct {
	LoginAttempts    int // per 15 minutes per IP
	PasswordReset    int // per hour per email
	APICalls         int // per minute per user
	Registration     int // per hour per IP
	MaxSessions      int // per user
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnvAsInt("SERVER_PORT", 8080),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "auth_service"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
			AccessTokenTTL:  getEnvAsInt("JWT_ACCESS_TTL_MINUTES", 60), // 1 hour
			RefreshTokenTTL: getEnvAsInt("JWT_REFRESH_TTL_DAYS", 30),   // 30 days
			Issuer:          getEnv("JWT_ISSUER", "auth-service"),
			SigningMethod:   getEnv("JWT_SIGNING_METHOD", "HS256"),
		},
		RateLimit: RateLimitConfig{
			LoginAttempts: getEnvAsInt("RATE_LIMIT_LOGIN_ATTEMPTS", 5),    // 5 per 15 min per IP
			PasswordReset: getEnvAsInt("RATE_LIMIT_PASSWORD_RESET", 3),     // 3 per hour per email
			APICalls:      getEnvAsInt("RATE_LIMIT_API_CALLS", 1000),       // 1000 per minute per user
			Registration:  getEnvAsInt("RATE_LIMIT_REGISTRATION", 10),      // 10 per hour per IP
			MaxSessions:   getEnvAsInt("MAX_CONCURRENT_SESSIONS", 5),       // 5 per user
		},
		Environment: getEnv("ENVIRONMENT", "development"),
	}
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