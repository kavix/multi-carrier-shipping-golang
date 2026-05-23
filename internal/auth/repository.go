package auth

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type PostgresAuthRepository struct {
	db *sql.DB
}

// NewPostgresAuthRepository creates a new AuthRepository utilizing a PostgreSQL database.
func NewPostgresAuthRepository(dsn string) (*PostgresAuthRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres db: %w", err)
	}

	// Auto-migrate schema DDL statements
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL
	);

	CREATE TABLE IF NOT EXISTS sessions (
		token TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		expires_at TIMESTAMP NOT NULL
	);

	CREATE TABLE IF NOT EXISTS audit_logs (
		id SERIAL PRIMARY KEY,
		username TEXT NOT NULL,
		login_timestamp TIMESTAMP NOT NULL,
		action TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL
	);`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &PostgresAuthRepository{db: db}, nil
}

func (r *PostgresAuthRepository) CreateUser(ctx context.Context, user *User) error {
	query := `INSERT INTO users (id, username, password_hash, created_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Username, user.PasswordHash, user.CreatedAt)
	if err != nil {
		return ErrUserAlreadyExists
	}
	return nil
}

func (r *PostgresAuthRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `SELECT id, username, password_hash, created_at FROM users WHERE username = $1`
	row := r.db.QueryRowContext(ctx, query, username)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrInvalidCredentials
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *PostgresAuthRepository) CreateSession(ctx context.Context, s *Session) error {
	query := `INSERT INTO sessions (token, username, expires_at) VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, s.Token, s.Username, s.ExpiresAt)
	return err
}

func (r *PostgresAuthRepository) GetSession(ctx context.Context, token string) (*Session, error) {
	query := `SELECT token, username, expires_at FROM sessions WHERE token = $1`
	row := r.db.QueryRowContext(ctx, query, token)

	var s Session
	err := row.Scan(&s.Token, &s.Username, &s.ExpiresAt)
	if err == sql.ErrNoRows {
		return nil, ErrSessionNotFound
	} else if err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *PostgresAuthRepository) DeleteSession(ctx context.Context, token string) error {
	query := `DELETE FROM sessions WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

func (r *PostgresAuthRepository) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	query := `INSERT INTO audit_logs (username, login_timestamp, action, created_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, log.Username, log.LoginTimestamp, log.Action, log.CreatedAt)
	return err
}

func (r *PostgresAuthRepository) GetAuditLogs(ctx context.Context, username string) ([]*AuditLog, error) {
	query := `SELECT id, username, login_timestamp, action, created_at FROM audit_logs WHERE username = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*AuditLog
	for rows.Next() {
		var log AuditLog
		if err := rows.Scan(&log.ID, &log.Username, &log.LoginTimestamp, &log.Action, &log.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}

	return logs, nil
}

func (r *PostgresAuthRepository) Close() error {
	return r.db.Close()
}
