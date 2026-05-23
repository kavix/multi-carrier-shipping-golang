package auth

import (
	"context"
	"errors"
	"testing"
)

type mockAuthRepository struct {
	users     map[string]*User
	sessions  map[string]*Session
	auditLogs map[string][]*AuditLog
}

func newMockAuthRepository() *mockAuthRepository {
	return &mockAuthRepository{
		users:     make(map[string]*User),
		sessions:  make(map[string]*Session),
		auditLogs: make(map[string][]*AuditLog),
	}
}

func (m *mockAuthRepository) CreateUser(ctx context.Context, user *User) error {
	if _, exists := m.users[user.Username]; exists {
		return ErrUserAlreadyExists
	}
	m.users[user.Username] = user
	return nil
}

func (m *mockAuthRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	u, exists := m.users[username]
	if !exists {
		return nil, ErrInvalidCredentials
	}
	return u, nil
}

func (m *mockAuthRepository) CreateSession(ctx context.Context, session *Session) error {
	m.sessions[session.Token] = session
	return nil
}

func (m *mockAuthRepository) GetSession(ctx context.Context, token string) (*Session, error) {
	s, exists := m.sessions[token]
	if !exists {
		return nil, ErrSessionNotFound
	}
	return s, nil
}

func (m *mockAuthRepository) DeleteSession(ctx context.Context, token string) error {
	delete(m.sessions, token)
	return nil
}

func (m *mockAuthRepository) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	m.auditLogs[log.Username] = append(m.auditLogs[log.Username], log)
	return nil
}

func (m *mockAuthRepository) GetAuditLogs(ctx context.Context, username string) ([]*AuditLog, error) {
	return m.auditLogs[username], nil
}

func TestAuthService(t *testing.T) {
	repo := newMockAuthRepository()
	svc := NewAuthService(repo)
	ctx := context.Background()

	t.Run("successful register and login", func(t *testing.T) {
		err := svc.Register(ctx, "kavix", "mysecurepassword")
		if err != nil {
			t.Fatalf("expected no error during register, got %v", err)
		}

		// Trying duplicate username
		err = svc.Register(ctx, "kavix", "anotherpassword")
		if !errors.Is(err, ErrUserAlreadyExists) {
			t.Errorf("expected ErrUserAlreadyExists for duplicate username, got %v", err)
		}

		// Successful Login
		session, err := svc.Login(ctx, "kavix", "mysecurepassword")
		if err != nil {
			t.Fatalf("expected no error during login, got %v", err)
		}

		if session.Token == "" {
			t.Errorf("expected non-empty session token")
		}

		// Verify Token
		username, err := svc.VerifyToken(ctx, session.Token)
		if err != nil {
			t.Fatalf("expected token to be valid, got %v", err)
		}
		if username != "kavix" {
			t.Errorf("expected username 'kavix', got %s", username)
		}

		// Log Action
		err = svc.LogAction(ctx, "kavix", "Update Shipment")
		if err != nil {
			t.Fatalf("expected no error logging action, got %v", err)
		}

		// Retrieve logs
		logs, err := repo.GetAuditLogs(ctx, "kavix")
		if err != nil {
			t.Fatalf("expected no error fetching logs, got %v", err)
		}
		if len(logs) < 2 { // "Login" and "Update Shipment"
			t.Errorf("expected at least 2 audit logs, got %d", len(logs))
		}
	})

	t.Run("invalid credentials login", func(t *testing.T) {
		_, err := svc.Login(ctx, "kavix", "wrongpassword")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}

		_, err = svc.Login(ctx, "nonexistent", "somepassword")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials for non-existent user, got %v", err)
		}
	})
}
