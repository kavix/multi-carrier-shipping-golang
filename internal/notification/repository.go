package notification

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteNotificationRepository struct {
	db *sql.DB
}

// NewSQLiteNotificationRepository instantiates a new SQLite database connection for Notifications.
func NewSQLiteNotificationRepository(dbPath string) (*SQLiteNotificationRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS notification_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		recipient TEXT NOT NULL,
		method TEXT NOT NULL,
		subject TEXT NOT NULL,
		body TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at INTEGER NOT NULL
	);`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &SQLiteNotificationRepository{db: db}, nil
}

func (r *SQLiteNotificationRepository) Create(ctx context.Context, log *NotificationLog) error {
	query := `INSERT INTO notification_logs (recipient, method, subject, body, status, created_at) VALUES (?, ?, ?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, query, log.Recipient, log.Method, log.Subject, log.Body, log.Status, log.CreatedAt.Unix())
	if err != nil {
		return fmt.Errorf("failed to insert notification log: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		log.ID = id
	}
	return nil
}

func (r *SQLiteNotificationRepository) List(ctx context.Context) ([]*NotificationLog, error) {
	query := `SELECT id, recipient, method, subject, body, status, created_at FROM notification_logs ORDER BY id DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query notification logs: %w", err)
	}
	defer rows.Close()

	var logs []*NotificationLog
	for rows.Next() {
		var l NotificationLog
		var createdUnix int64
		if err := rows.Scan(&l.ID, &l.Recipient, &l.Method, &l.Subject, &l.Body, &l.Status, &createdUnix); err != nil {
			return nil, fmt.Errorf("failed to scan notification log: %w", err)
		}
		l.CreatedAt = time.Unix(createdUnix, 0)
		logs = append(logs, &l)
	}
	return logs, nil
}

func (r *SQLiteNotificationRepository) Close() error {
	return r.db.Close()
}
