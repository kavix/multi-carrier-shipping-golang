package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port         string
	DB           string
	KafkaBrokers []string
	S3BucketARN  string
	AWSRegion    string
	AWSAccessKey string
	AWSSecretKey string
}

func Load() *Config {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}
	return &Config{
		Port: getEnv("PORT", "8084"),
		DB: fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "5432"),
			getEnv("DB_USER", "postgres"),
			getEnv("DB_PASS", "postgres"),
			getEnv("DB_NAME", "labels")),
		KafkaBrokers: strings.Split(brokers, ","),
		S3BucketARN:  os.Getenv("S3_BUCKET_ARN"),
		AWSRegion:    getEnv("AWS_REGION", "us-east-1"),
		AWSAccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
