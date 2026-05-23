package notification

import (
	"context"
	"testing"
)

type mockNotificationRepository struct {
	logs   []*NotificationLog
	nextID int64
}

func newMockNotificationRepository() *mockNotificationRepository {
	return &mockNotificationRepository{
		logs:   make([]*NotificationLog, 0),
		nextID: 1,
	}
}

func (m *mockNotificationRepository) Create(ctx context.Context, log *NotificationLog) error {
	log.ID = m.nextID
	m.nextID++
	m.logs = append(m.logs, log)
	return nil
}

func (m *mockNotificationRepository) List(ctx context.Context) ([]*NotificationLog, error) {
	// Return in reverse order (newest first) to mimic ORDER BY id DESC
	n := len(m.logs)
	reversed := make([]*NotificationLog, n)
	for i, l := range m.logs {
		reversed[n-1-i] = l
	}
	return reversed, nil
}

func (m *mockNotificationRepository) Close() error {
	return nil
}

func TestNotificationService(t *testing.T) {
	repo := newMockNotificationRepository()
	svc := NewNotificationService(repo)
	ctx := context.Background()

	t.Run("send email notification (test environment)", func(t *testing.T) {
		logRecord, err := svc.SendEmailNotification(ctx, "recipient@example.com", "Test Subject", "<p>Test Body</p>")
		if err != nil {
			t.Fatalf("expected no error in test environment, got %v", err)
		}

		if logRecord.Recipient != "recipient@example.com" {
			t.Errorf("expected recipient 'recipient@example.com', got '%s'", logRecord.Recipient)
		}
		if logRecord.Method != "EMAIL" {
			t.Errorf("expected method 'EMAIL', got '%s'", logRecord.Method)
		}
		if logRecord.Subject != "Test Subject" {
			t.Errorf("expected subject 'Test Subject', got '%s'", logRecord.Subject)
		}
		if logRecord.Body != "<p>Test Body</p>" {
			t.Errorf("expected body '<p>Test Body</p>', got '%s'", logRecord.Body)
		}
		if logRecord.Status != "SENT" {
			t.Errorf("expected status 'SENT', got '%s'", logRecord.Status)
		}
	})

	t.Run("send telegram notification (simulated)", func(t *testing.T) {
		logRecord, err := svc.SendTelegramNotification(ctx, "+15551234567", "Your shipment has been created!")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if logRecord.Recipient != "+15551234567" {
			t.Errorf("expected recipient '+15551234567', got '%s'", logRecord.Recipient)
		}
		if logRecord.Method != "TELEGRAM" {
			t.Errorf("expected method 'TELEGRAM', got '%s'", logRecord.Method)
		}
		if logRecord.Body != "Your shipment has been created!" {
			t.Errorf("expected body to match, got '%s'", logRecord.Body)
		}
		if logRecord.Status != "SENT" {
			t.Errorf("expected status 'SENT', got '%s'", logRecord.Status)
		}
	})

	t.Run("list logs returns all history logs", func(t *testing.T) {
		logs, err := svc.ListLogs(ctx)
		if err != nil {
			t.Fatalf("failed to list logs: %v", err)
		}

		// We sent 1 email and 1 telegram log
		if len(logs) != 2 {
			t.Errorf("expected 2 notification logs in database, got %d", len(logs))
		}

		// Ordered by created_at DESC / ID DESC, so the second one (TELEGRAM) is first
		if logs[0].Method != "TELEGRAM" {
			t.Errorf("expected first log to be 'TELEGRAM', got '%s'", logs[0].Method)
		}
		if logs[1].Method != "EMAIL" {
			t.Errorf("expected second log to be 'EMAIL', got '%s'", logs[1].Method)
		}
	})
}
