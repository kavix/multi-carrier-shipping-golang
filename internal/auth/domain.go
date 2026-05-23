package auth

import (
	"context"
	"errors"
	"time"
)

// User represents a user registration record in the SQLite database.
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// Session represents an active login token stored in SQLite database.
type Session struct {
	Token     string    `json:"token"`
	Username  string    `json:"username"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AuditLog represents an action taken by a logged-in user.
type AuditLog struct {
	ID             int64     `json:"id"`
	Username       string    `json:"username"`
	LoginTimestamp time.Time `json:"login_timestamp"`
	Action         string    `json:"action"`
	CreatedAt      time.Time `json:"created_at"`
}

var (
	ErrUserAlreadyExists  = errors.New("username is already taken")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrSessionNotFound    = errors.New("unauthorized: session not found")
	ErrSessionExpired     = errors.New("unauthorized: session has expired")
)

// AuthRepository defines the SQLite DB access layer.
type AuthRepository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	CreateSession(ctx context.Context, session *Session) error
	GetSession(ctx context.Context, token string) (*Session, error)
	DeleteSession(ctx context.Context, token string) error
	CreateAuditLog(ctx context.Context, log *AuditLog) error
	GetAuditLogs(ctx context.Context, username string) ([]*AuditLog, error)
}

// AuthService defines the business logic of Auth microservice.
type AuthService interface {
	Register(ctx context.Context, username, password string) error
	Login(ctx context.Context, username, password string) (*Session, error)
	VerifyToken(ctx context.Context, token string) (string, error)
	LogAction(ctx context.Context, username, action string) error
	GetAuditLogs(ctx context.Context, token string) ([]*AuditLog, error)
}
