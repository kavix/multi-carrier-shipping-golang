package config

import (
	"os"
	"strings"
)

type Config struct {
	KafkaBrokers []string
}

func Load() *Config {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}
	return &Config{
		KafkaBrokers: strings.Split(brokers, ","),
	}
}
