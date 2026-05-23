package notification

import (
	"context"
	"time"
)

type NotificationLog struct {
	ID        int64     `json:"id"`
	Recipient string    `json:"recipient"` // Email or Phone number
	Method    string    `json:"method"`    // "EMAIL" or "TELEGRAM"
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	Status    string    `json:"status"` // "SENT", "FAILED"
	CreatedAt time.Time `json:"created_at"`
}

type NotificationRepository interface {
	Create(ctx context.Context, log *NotificationLog) error
	List(ctx context.Context) ([]*NotificationLog, error)
	Close() error
}

type NotificationService interface {
	SendEmailNotification(ctx context.Context, recipient, subject, body string) (*NotificationLog, error)
	SendTelegramNotification(ctx context.Context, recipient, message string) (*NotificationLog, error)
	ListLogs(ctx context.Context) ([]*NotificationLog, error)
}
