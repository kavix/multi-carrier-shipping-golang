package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	repo AuthRepository
}

// NewAuthService instantiates a new AuthService.
func NewAuthService(repo AuthRepository) AuthService {
	return &authService{repo: repo}
}

func (s *authService) Register(ctx context.Context, username, password string) error {
	if username == "" || password == "" {
		return ErrInvalidCredentials
	}

	// Hash password using bcrypt
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate standard secure random user ID
	uid, err := generateRandomHex(16)
	if err != nil {
		return err
	}

	user := &User{
		ID:           uid,
		Username:     username,
		PasswordHash: string(hashed),
		CreatedAt:    time.Now(),
	}

	return s.repo.CreateUser(ctx, user)
}

func (s *authService) Login(ctx context.Context, username, password string) (*Session, error) {
	if username == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// Compare bcrypt password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate secure random session token
	token, err := generateRandomHex(32)
	if err != nil {
		return nil, err
	}

	session := &Session{
		Token:     token,
		Username:  username,
		ExpiresAt: time.Now().Add(24 * time.Hour), // Session active for 24h
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Write audit log entry for login action
	logEntry := &AuditLog{
		Username:       username,
		LoginTimestamp: time.Now(),
		Action:         "Login",
		CreatedAt:      time.Now(),
	}
	_ = s.repo.CreateAuditLog(ctx, logEntry)

	return session, nil
}

func (s *authService) VerifyToken(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", ErrSessionNotFound
	}

	session, err := s.repo.GetSession(ctx, token)
	if err != nil {
		return "", err
	}

	// Check if session token has expired
	if time.Now().After(session.ExpiresAt) {
		_ = s.repo.DeleteSession(ctx, token)
		return "", ErrSessionExpired
	}

	return session.Username, nil
}

func (s *authService) LogAction(ctx context.Context, username, action string) error {
	if username == "" || action == "" {
		return nil
	}

	// Record audit log action
	logEntry := &AuditLog{
		Username:       username,
		LoginTimestamp: time.Now(), // Defaulting timestamp of this action
		Action:         action,
		CreatedAt:      time.Now(),
	}

	return s.repo.CreateAuditLog(ctx, logEntry)
}

func (s *authService) GetAuditLogs(ctx context.Context, token string) ([]*AuditLog, error) {
	username, err := s.VerifyToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return s.repo.GetAuditLogs(ctx, username)
}

func generateRandomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}
