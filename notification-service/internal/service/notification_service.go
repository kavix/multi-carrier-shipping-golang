package service

import (
	"fmt"

	"github.com/shipping/shared/pkg/logger"
)

type NotificationService struct{}

func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

func (s *NotificationService) SendEmail(to, subject, body string) error {
	logger.Info("sending email",
		logger.String("to", to),
		logger.String("subject", subject),
	)
	// In production: integrate with SendGrid, AWS SES, or SMTP
	fmt.Printf("[EMAIL] To: %s | Subject: %s | Body: %s\n", to, subject, body)
	return nil
}

func (s *NotificationService) SendSMS(phone, message string) error {
	logger.Info("sending SMS",
		logger.String("phone", phone),
	)
	// In production: integrate with Twilio
	fmt.Printf("[SMS] To: %s | Message: %s\n", phone, message)
	return nil
}
