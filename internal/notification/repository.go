package notification

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type PostgresNotificationRepository struct {
	db *sql.DB
}

// NewPostgresNotificationRepository instantiates a new PostgreSQL database connection for Notifications.
func NewPostgresNotificationRepository(dsn string) (*PostgresNotificationRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres db: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS notification_logs (
		id SERIAL PRIMARY KEY,
		recipient TEXT NOT NULL,
		method TEXT NOT NULL,
		subject TEXT NOT NULL,
		body TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL
	);`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &PostgresNotificationRepository{db: db}, nil
}

func (r *PostgresNotificationRepository) Create(ctx context.Context, log *NotificationLog) error {
	query := `INSERT INTO notification_logs (recipient, method, subject, body, status, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	err := r.db.QueryRowContext(ctx, query, log.Recipient, log.Method, log.Subject, log.Body, log.Status, log.CreatedAt).Scan(&log.ID)
	if err != nil {
		return fmt.Errorf("failed to insert notification log: %w", err)
	}
	return nil
}

func (r *PostgresNotificationRepository) List(ctx context.Context) ([]*NotificationLog, error) {
	query := `SELECT id, recipient, method, subject, body, status, created_at FROM notification_logs ORDER BY id DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query notification logs: %w", err)
	}
	defer rows.Close()

	var logs []*NotificationLog
	for rows.Next() {
		var l NotificationLog
		if err := rows.Scan(&l.ID, &l.Recipient, &l.Method, &l.Subject, &l.Body, &l.Status, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan notification log: %w", err)
		}
		logs = append(logs, &l)
	}
	return logs, nil
}

func (r *PostgresNotificationRepository) Close() error {
	return r.db.Close()
}
