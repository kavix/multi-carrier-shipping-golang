package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port string
	DB   string
}

func Load() *Config {
	return &Config{
		Port: getEnv("PORT", "8082"),
		DB: fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "5432"),
			getEnv("DB_USER", "postgres"),
			getEnv("DB_PASS", "postgres"),
			getEnv("DB_NAME", "carriers")),
	}
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
