package service

import (
	"context"
	"errors"
	"testing"

	"github.com/kavindus/multi-carrier-shipping-golang/internal/domain"
	"github.com/kavindus/multi-carrier-shipping-golang/internal/repository"
)

func TestCreateShipment(t *testing.T) {
	repo := repository.NewMemoryShipmentRepository()
	svc := NewShipmentService(repo)
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		shipment, err := svc.CreateShipment(ctx, "DHL", "TRK-987654", 5.5, "London", "Paris")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if shipment.ID == "" {
			t.Errorf("expected generated ID to be populated, got empty string")
		}
		if shipment.Carrier != "DHL" {
			t.Errorf("expected carrier 'DHL', got %s", shipment.Carrier)
		}
		if shipment.Status != "CREATED" {
			t.Errorf("expected status 'CREATED', got %s", shipment.Status)
		}
	})

	t.Run("missing carrier validation error", func(t *testing.T) {
		_, err := svc.CreateShipment(ctx, "", "TRK-000", 2.0, "Origin", "Destination")
		if !errors.Is(err, domain.ErrCarrierRequired) {
			t.Errorf("expected ErrCarrierRequired, got %v", err)
		}
	})

	t.Run("missing tracking number validation error", func(t *testing.T) {
		_, err := svc.CreateShipment(ctx, "FedEx", "", 2.0, "Origin", "Destination")
		if !errors.Is(err, domain.ErrTrackingNumberRequired) {
			t.Errorf("expected ErrTrackingNumberRequired, got %v", err)
		}
	})

	t.Run("invalid weight validation error", func(t *testing.T) {
		_, err := svc.CreateShipment(ctx, "FedEx", "TRK-001", -0.5, "Origin", "Destination")
		if !errors.Is(err, domain.ErrInvalidWeight) {
			t.Errorf("expected ErrInvalidWeight, got %v", err)
		}
	})

	t.Run("already exists validation error", func(t *testing.T) {
		// Create the first shipment
		_, err := svc.CreateShipment(ctx, "FedEx", "TRK-DUP", 10.0, "Origin", "Destination")
		if err != nil {
			t.Fatalf("first creation expected no error, got %v", err)
		}

		// Try creating a second shipment with the same tracking number
		_, err = svc.CreateShipment(ctx, "UPS", "TRK-DUP", 5.0, "Origin", "Destination")
		if !errors.Is(err, domain.ErrShipmentAlreadyExists) {
			t.Errorf("expected ErrShipmentAlreadyExists, got %v", err)
		}
	})
}
