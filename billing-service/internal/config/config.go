package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port            string
	DB              string
	KafkaBrokers    []string
	StripeSecretKey string
}

func Load() *Config {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}
	return &Config{
		Port: getEnv("PORT", "8087"),
		DB: fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "5432"),
			getEnv("DB_USER", "postgres"),
			getEnv("DB_PASS", "postgres"),
			getEnv("DB_NAME", "billing")),
		KafkaBrokers:    strings.Split(brokers, ","),
		StripeSecretKey: os.Getenv("STRIPE_SECRET_KEY"),
	}
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
