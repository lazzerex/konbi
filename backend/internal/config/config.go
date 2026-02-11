package config

import (
	"os"
	"strconv"
	"time"
)

// config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Storage  StorageConfig
	Security SecurityConfig
}

// server configuration
type ServerConfig struct {
	Port           string
	AllowedOrigins string
	Environment    string
	BaseURL        string // base url for shortened urls
}

// database configuration
type DatabaseConfig struct {
	URL            string
	MaxConnections int
	MaxIdleConns   int
	ConnMaxLife    time.Duration
}

// storage configuration
type StorageConfig struct {
	UploadDir      string
	MaxFileSize    int64
	ExpirationDays int
}

// security configuration
type SecurityConfig struct {
	AdminSecret     string
	RateLimitPerSec int
	RateLimitBurst  int
}

// load reads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:           getEnv("PORT", "8080"),
			AllowedOrigins: getEnv("ALLOWED_ORIGINS", "*"),
			Environment:    getEnv("ENVIRONMENT", "development"),
			BaseURL:        getEnv("BASE_URL", "http://localhost:8080"),
		},
		Database: DatabaseConfig{
			URL:            getEnv("DATABASE_URL", ""),
			MaxConnections: getEnvAsInt("DB_MAX_CONNECTIONS", 25),
			MaxIdleConns:   getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLife:    time.Duration(getEnvAsInt("DB_CONN_MAX_LIFE_MINUTES", 5)) * time.Minute,
		},
		Storage: StorageConfig{
			UploadDir:      getEnv("UPLOAD_DIR", "uploads"),
			MaxFileSize:    int64(getEnvAsInt("MAX_FILE_SIZE_MB", 50)) * 1024 * 1024,
			ExpirationDays: getEnvAsInt("EXPIRATION_DAYS", 7),
		},
		Security: SecurityConfig{
			AdminSecret:     getEnv("ADMIN_SECRET", ""),
			RateLimitPerSec: getEnvAsInt("RATE_LIMIT_PER_SEC", 10),
			RateLimitBurst:  getEnvAsInt("RATE_LIMIT_BURST", 10),
		},
	}
}

// helper to get env variable with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// helper to get env variable as int with default
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
