package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/kavindus/multi-carrier-shipping-golang/internal/domain"
)

type shipmentService struct {
	repo domain.ShipmentRepository
}

// NewShipmentService instantiates a new ShipmentService implementation.
func NewShipmentService(repo domain.ShipmentRepository) domain.ShipmentService {
	return &shipmentService{
		repo: repo,
	}
}

func (s *shipmentService) CreateShipment(
	ctx context.Context,
	carrier, trackingNumber string,
	weight float64,
	origin, destination string,
) (*domain.Shipment, error) {
	// 1. Validation
	if carrier == "" {
		return nil, domain.ErrCarrierRequired
	}
	if trackingNumber == "" {
		return nil, domain.ErrTrackingNumberRequired
	}
	if weight <= 0 {
		return nil, domain.ErrInvalidWeight
	}

	// 2. Business rules / check existence
	_, err := s.repo.GetByTracking(ctx, trackingNumber)
	if err == nil {
		return nil, domain.ErrShipmentAlreadyExists
	}

	// 3. Entity construction with auto-generated ID
	id, err := generateUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate shipment id: %w", err)
	}

	now := time.Now()
	shipment := &domain.Shipment{
		ID:             id,
		Carrier:        carrier,
		TrackingNumber: trackingNumber,
		Weight:         weight,
		Origin:         origin,
		Destination:    destination,
		Status:         "CREATED",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// 4. Persistence
	if err := s.repo.Create(ctx, shipment); err != nil {
		return nil, err
	}

	return shipment, nil
}

func (s *shipmentService) GetShipment(ctx context.Context, id string) (*domain.Shipment, error) {
	if id == "" {
		return nil, domain.ErrShipmentNotFound
	}
	return s.repo.GetByID(ctx, id)
}

func (s *shipmentService) ListShipments(ctx context.Context) ([]*domain.Shipment, error) {
	return s.repo.List(ctx)
}

// Helper to generate a basic secure random UUIDv4 string without external dependencies
func generateUUID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// Variant and version bits (standard UUIDv4 formatting)
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant is 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
