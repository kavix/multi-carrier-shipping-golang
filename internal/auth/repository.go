package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteAuthRepository struct {
	db *sql.DB
}

// NewSQLiteAuthRepository creates a new AuthRepository utilizing an SQLite database.
func NewSQLiteAuthRepository(dbPath string) (*SQLiteAuthRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	// Auto-migrate schema DDL statements with INTEGER for dates
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS sessions (
		token TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		expires_at INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS audit_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		login_timestamp INTEGER NOT NULL,
		action TEXT NOT NULL,
		created_at INTEGER NOT NULL
	);`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &SQLiteAuthRepository{db: db}, nil
}

func (r *SQLiteAuthRepository) CreateUser(ctx context.Context, user *User) error {
	query := `INSERT INTO users (id, username, password_hash, created_at) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Username, user.PasswordHash, user.CreatedAt.Unix())
	if err != nil {
		return ErrUserAlreadyExists
	}
	return nil
}

func (r *SQLiteAuthRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `SELECT id, username, password_hash, created_at FROM users WHERE username = ?`
	row := r.db.QueryRowContext(ctx, query, username)

	var user User
	var createdAtUnix int64
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &createdAtUnix)
	if err == sql.ErrNoRows {
		return nil, ErrInvalidCredentials
	} else if err != nil {
		return nil, err
	}

	user.CreatedAt = time.Unix(createdAtUnix, 0)
	return &user, nil
}

func (r *SQLiteAuthRepository) CreateSession(ctx context.Context, s *Session) error {
	query := `INSERT INTO sessions (token, username, expires_at) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, s.Token, s.Username, s.ExpiresAt.Unix())
	return err
}

func (r *SQLiteAuthRepository) GetSession(ctx context.Context, token string) (*Session, error) {
	query := `SELECT token, username, expires_at FROM sessions WHERE token = ?`
	row := r.db.QueryRowContext(ctx, query, token)

	var s Session
	var expiresAtUnix int64
	err := row.Scan(&s.Token, &s.Username, &expiresAtUnix)
	if err == sql.ErrNoRows {
		return nil, ErrSessionNotFound
	} else if err != nil {
		return nil, err
	}

	s.ExpiresAt = time.Unix(expiresAtUnix, 0)
	return &s, nil
}

func (r *SQLiteAuthRepository) DeleteSession(ctx context.Context, token string) error {
	query := `DELETE FROM sessions WHERE token = ?`
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

func (r *SQLiteAuthRepository) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	query := `INSERT INTO audit_logs (username, login_timestamp, action, created_at) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, log.Username, log.LoginTimestamp.Unix(), log.Action, log.CreatedAt.Unix())
	return err
}

func (r *SQLiteAuthRepository) GetAuditLogs(ctx context.Context, username string) ([]*AuditLog, error) {
	query := `SELECT id, username, login_timestamp, action, created_at FROM audit_logs WHERE username = ? ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*AuditLog
	for rows.Next() {
		var log AuditLog
		var loginUnix, createdUnix int64
		if err := rows.Scan(&log.ID, &log.Username, &loginUnix, &log.Action, &createdUnix); err != nil {
			return nil, err
		}
		log.LoginTimestamp = time.Unix(loginUnix, 0)
		log.CreatedAt = time.Unix(createdUnix, 0)
		logs = append(logs, &log)
	}

	return logs, nil
}

func (r *SQLiteAuthRepository) Close() error {
	return r.db.Close()
}
