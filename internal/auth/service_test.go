package auth

import (
	"context"
	"errors"
	"os"
	"testing"
)

func TestAuthService(t *testing.T) {
	dbFile := "test_auth.db"
	defer os.Remove(dbFile)

	repo, err := NewSQLiteAuthRepository(dbFile)
	if err != nil {
		t.Fatalf("failed to initialize test repository: %v", err)
	}
	defer repo.Close()

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
