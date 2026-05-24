package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	KafkaBrokers   []string
	SMTPHost       string
	SMTPPort       int
	SMTPFrom       string
	SMTPUser       string
	SMTPPassword   string
	SendGridAPIKey string
}

func Load() *Config {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}
	// SMTP defaults (can be overridden via env)
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		smtpHost = "smtp.mail.yahoo.com"
	}
	smtpPort := 587
	if p := os.Getenv("SMTP_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			smtpPort = v
		}
	}
	smtpFrom := os.Getenv("SMTP_FROM")
	smtpUser := os.Getenv("SMTP_USER")
	// Accept either SMTP_PASSWORD or SMTP_PASS
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	if smtpPassword == "" {
		smtpPassword = os.Getenv("SMTP_PASS")
	}
	sendgridAPIKey := os.Getenv("SENDGRID_API_KEY")

	return &Config{
		KafkaBrokers:   strings.Split(brokers, ","),
		SMTPHost:       smtpHost,
		SMTPPort:       smtpPort,
		SMTPFrom:       smtpFrom,
		SMTPUser:       smtpUser,
		SMTPPassword:   smtpPassword,
		SendGridAPIKey: sendgridAPIKey,
	}
}
