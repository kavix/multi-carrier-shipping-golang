package shipment

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateShipment(t *testing.T) {
	// Create mock Label Service HTTP server
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

	repo := NewMemoryShipmentRepository()
	svc := NewShipmentService(repo, mockLabelServer.URL)
	ctx := context.Background()

	t.Run("successful shipment creation and label assignment", func(t *testing.T) {
		shipment, label, err := svc.CreateShipment(ctx, "FedEx", 4.5, "Los Angeles", "New York")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if shipment.ID == "" {
			t.Errorf("expected generated shipment ID")
		}
		if shipment.Status != "CREATED" {
			t.Errorf("expected shipment status to be CREATED, got %s", shipment.Status)
		}
		if label.TrackingNumber != "FTX123456789" {
			t.Errorf("expected tracking number 'FTX123456789', got %s", label.TrackingNumber)
		}
	})

	t.Run("validation missing carrier", func(t *testing.T) {
		_, _, err := svc.CreateShipment(ctx, "", 1.0, "Origin", "Destination")
		if !errors.Is(err, ErrCarrierRequired) {
			t.Errorf("expected ErrCarrierRequired, got %v", err)
		}
	})
}
