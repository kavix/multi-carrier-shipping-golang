package shipment

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCreateShipment(t *testing.T) {
	dbFile := "test_shipments.db"
	defer os.Remove(dbFile)

	// 1. Mock Auth Server
	mockAuthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/auth/verify" {
			token := r.URL.Query().Get("token")
			if token == "valid-token" {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"username": "kavix"})
			} else if token == "other-token" {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"username": "other_user"})
			} else if token == "admin-token" {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"username": "admin"})
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
		} else if r.URL.Path == "/api/v1/auth/logs" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"message": "logged"})
		}
	}))
	defer mockAuthServer.Close()

	// 2. Mock Label Server
	mockLabelServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":              "lbl-mock-123",
			"shipment_id":     "dummy-shipment-id",
			"tracking_number": "FTX123456789",
			"label_url":       "https://fedex-sandbox/labels/FTX123456789.pdf",
			"status":          "ACTIVE",
		})
	}))
	defer mockLabelServer.Close()

	// 3. Mock Notification Server
	mockNotificationServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/notifications" {
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         1,
				"recipient":  "mock-recipient",
				"method":     "EMAIL",
				"status":     "SENT",
				"created_at": time.Now().Unix(),
			})
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer mockNotificationServer.Close()

	repo, err := NewSQLiteShipmentRepository(dbFile)
	if err != nil {
		t.Fatalf("failed to open test sqlite db: %v", err)
	}
	defer repo.Close()

	svc := NewShipmentService(repo, mockLabelServer.URL, mockAuthServer.URL, mockNotificationServer.URL)
	svcImpl := svc.(*shipmentService)
	svcImpl.rateLimit = 0 // Disable rate limiting for standard sequential tests
	ctx := context.Background()

	t.Run("successful shipment creation and list isolation", func(t *testing.T) {
		shipment, _, err := svc.CreateShipment(ctx, "valid-token", "FedEx", 4.5, "Los Angeles", "New York", "recipient@example.com")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Create another shipment under other-token (other_user)
		_, _, err = svc.CreateShipment(ctx, "other-token", "DHL", 10.0, "Chicago", "Miami", "other@example.com")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// List shipments under valid-token (kavix)
		list, err := svc.ListShipments(ctx, "valid-token")
		if err != nil {
			t.Fatalf("expected no error listing, got %v", err)
		}

		if len(list) != 1 {
			t.Errorf("expected list to have 1 shipment owned by kavix, got %d", len(list))
		}
		if list[0].ID != shipment.ID {
			t.Errorf("expected shipment ID '%s', got '%s'", shipment.ID, list[0].ID)
		}
	})

	t.Run("forbidden modifications by other user", func(t *testing.T) {
		// Create a shipment under valid-token (kavix)
		shipment, _, err := svc.CreateShipment(ctx, "valid-token", "UPS", 8.2, "Boston", "Seattle", "recipient@example.com")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Attempt to update it with other-token (other_user)
		_, err = svc.UpdateShipment(ctx, "other-token", shipment.ID, "FedEx", 12.0, "Boston", "Seattle", "")
		if err == nil || !strings.Contains(err.Error(), "forbidden") {
			t.Errorf("expected forbidden error for cross-user update, got %v", err)
		}

		// Attempt to delete it with other-token (other_user)
		err = svc.DeleteShipment(ctx, "other-token", shipment.ID)
		if err == nil || !strings.Contains(err.Error(), "forbidden") {
			t.Errorf("expected forbidden error for cross-user delete, got %v", err)
		}

		// Successful update by the owner (valid-token)
		updated, err := svc.UpdateShipment(ctx, "valid-token", shipment.ID, "UPS", 15.0, "Boston", "Seattle", "")
		if err != nil {
			t.Fatalf("expected successful update by owner, got %v", err)
		}
		if updated.Weight != 15.0 {
			t.Errorf("expected weight to be updated to 15.0, got %.2f", updated.Weight)
		}
	})

	t.Run("unauthorized creation with invalid token", func(t *testing.T) {
		_, _, err := svc.CreateShipment(ctx, "invalid-token", "FedEx", 4.5, "Los Angeles", "New York", "invalid@example.com")
		if err == nil || !strings.Contains(err.Error(), "unauthorized") {
			t.Errorf("expected unauthorized error, got %v", err)
		}
	})

	t.Run("validation missing carrier", func(t *testing.T) {
		_, _, err := svc.CreateShipment(ctx, "valid-token", "", 1.0, "Origin", "Destination", "")
		if !errors.Is(err, ErrCarrierRequired) {
			t.Errorf("expected ErrCarrierRequired, got %v", err)
		}
	})

	t.Run("admin status modification and status validation", func(t *testing.T) {
		// Create a shipment under valid-token (kavix)
		shipment, _, err := svc.CreateShipment(ctx, "valid-token", "FedEx", 4.5, "Los Angeles", "New York", "recipient@example.com")
		if err != nil {
			t.Fatalf("failed to create: %v", err)
		}

		// Regular user other_user attempts to update status of kavix's shipment -> forbidden
		_, err = svc.UpdateShipment(ctx, "other-token", shipment.ID, "FedEx", 4.5, "Los Angeles", "New York", "IN_TRANSIT")
		if err == nil || !strings.Contains(err.Error(), "forbidden") {
			t.Errorf("expected forbidden error for other user trying to update, got %v", err)
		}

		// Admin attempts to update status of kavix's shipment -> success
		updated, err := svc.UpdateShipment(ctx, "admin-token", shipment.ID, "FedEx", 4.5, "Los Angeles", "New York", "IN_TRANSIT")
		if err != nil {
			t.Fatalf("expected admin to successfully update status, got %v", err)
		}
		if updated.Status != "IN_TRANSIT" {
			t.Errorf("expected status to be updated to IN_TRANSIT, got %s", updated.Status)
		}

		// Admin attempts to update status to an invalid value -> ErrInvalidStatus
		_, err = svc.UpdateShipment(ctx, "admin-token", shipment.ID, "FedEx", 4.5, "Los Angeles", "New York", "SHIPPED_OUT")
		if !errors.Is(err, ErrInvalidStatus) {
			t.Errorf("expected ErrInvalidStatus, got %v", err)
		}
	})

	t.Run("concurrent rate limiting safety check", func(t *testing.T) {
		svcImpl.rateLimit = 5 * time.Second
		svcImpl.lastCreated = make(map[string]time.Time) // Reset timestamps for clean slate
		defer func() { svcImpl.rateLimit = 0 }()

		var wg sync.WaitGroup
		numRequests := 10
		errs := make(chan error, numRequests)
		successCount := 0

		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _, err := svc.CreateShipment(ctx, "valid-token", "FedEx", 5.0, "Origin", "Destination", "test@example.com")
				errs <- err
			}()
		}
		wg.Wait()
		close(errs)

		rateLimitCount := 0
		for err := range errs {
			if err == nil {
				successCount++
			} else if errors.Is(err, ErrRateLimitExceeded) {
				rateLimitCount++
			}
		}

		if successCount != 1 {
			t.Errorf("expected exactly 1 successful shipment creation within 5 seconds, got %d", successCount)
		}
		if rateLimitCount != numRequests-1 {
			t.Errorf("expected %d rate limit violations, got %d", numRequests-1, rateLimitCount)
		}
	})
}
