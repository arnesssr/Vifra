package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	ServerAddress string
	DatabaseURL   string
	JWTSecret     string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		ServerAddress: getEnv("SERVER_ADDRESS", ":8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://user:password@localhost/vpsmonitor?sslmode=disable"),
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key"),
	}
}

// getEnv returns the value of an environment variable or a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}