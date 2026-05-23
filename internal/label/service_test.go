package label

import (
	"context"
	"errors"
	"testing"
)

func TestCreateLabel(t *testing.T) {
	repo := NewMemoryLabelRepository()
	svc := NewLabelService(repo, nil, "")
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		label, err := svc.CreateLabel(ctx, "shipment-123", "DHL", 5.5, "London", "Paris")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if label.ID == "" {
			t.Errorf("expected generated ID to be populated, got empty string")
		}
		if label.ShipmentID != "shipment-123" {
			t.Errorf("expected shipmentID 'shipment-123', got %s", label.ShipmentID)
		}
		if label.TrackingNumber == "" {
			t.Errorf("expected tracking number to be auto-assigned")
		}
		if label.Status != "ACTIVE" {
			t.Errorf("expected label status to be 'ACTIVE', got %s", label.Status)
		}
	})

	t.Run("missing carrier validation error", func(t *testing.T) {
		_, err := svc.CreateLabel(ctx, "shipment-123", "", 2.0, "Origin", "Destination")
		if !errors.Is(err, ErrCarrierRequired) {
			t.Errorf("expected ErrCarrierRequired, got %v", err)
		}
	})

	t.Run("invalid weight validation error", func(t *testing.T) {
		_, err := svc.CreateLabel(ctx, "shipment-123", "FedEx", -0.5, "Origin", "Destination")
		if !errors.Is(err, ErrInvalidWeight) {
			t.Errorf("expected ErrInvalidWeight, got %v", err)
		}
	})
}
