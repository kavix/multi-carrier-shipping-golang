package config

import (
	"os"
)

type Config struct {
	Port string
	Env  string // development or production

	// Database Connection Configurations (placed for future real database integrations)
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

// Load loads application configuration from environment variables with fallback defaults.
func Load() *Config {
	return &Config{
		Port:       getEnv("APP_PORT", "8080"),
		Env:        getEnv("APP_ENV", "development"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "shipping_db"),
	}
}

// Helper to extract an environment variable or use a fallback value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
