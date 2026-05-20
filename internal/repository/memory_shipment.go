package repository

import (
	"context"
	"sync"
	"time"

	"github.com/kavindus/multi-carrier-shipping-golang/internal/domain"
)

type MemoryShipmentRepository struct {
	mu        sync.RWMutex
	shipments map[string]*domain.Shipment
}

// NewMemoryShipmentRepository instantiates a new in-memory shipment store.
func NewMemoryShipmentRepository() *MemoryShipmentRepository {
	return &MemoryShipmentRepository{
		shipments: make(map[string]*domain.Shipment),
	}
}

func (r *MemoryShipmentRepository) Create(ctx context.Context, shipment *domain.Shipment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already exists by ID
	if _, exists := r.shipments[shipment.ID]; exists {
		return domain.ErrShipmentAlreadyExists
	}

	// Check if already exists by tracking number
	for _, s := range r.shipments {
		if s.TrackingNumber == shipment.TrackingNumber {
			return domain.ErrShipmentAlreadyExists
		}
	}

	r.shipments[shipment.ID] = shipment
	return nil
}

func (r *MemoryShipmentRepository) GetByID(ctx context.Context, id string) (*domain.Shipment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shipment, exists := r.shipments[id]
	if !exists {
		return nil, domain.ErrShipmentNotFound
	}

	// Return a copy to prevent concurrent modification of map entries
	cp := *shipment
	return &cp, nil
}

func (r *MemoryShipmentRepository) GetByTracking(ctx context.Context, trackingNum string) (*domain.Shipment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, s := range r.shipments {
		if s.TrackingNumber == trackingNum {
			cp := *s
			return &cp, nil
		}
	}

	return nil, domain.ErrShipmentNotFound
}

func (r *MemoryShipmentRepository) List(ctx context.Context) ([]*domain.Shipment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*domain.Shipment, 0, len(r.shipments))
	for _, s := range r.shipments {
		cp := *s
		list = append(list, &cp)
	}

	return list, nil
}

func (r *MemoryShipmentRepository) Update(ctx context.Context, shipment *domain.Shipment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.shipments[shipment.ID]; !exists {
		return domain.ErrShipmentNotFound
	}

	shipment.UpdatedAt = time.Now()
	r.shipments[shipment.ID] = shipment
	return nil
}

func (r *MemoryShipmentRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.shipments[id]; !exists {
		return domain.ErrShipmentNotFound
	}

	delete(r.shipments, id)
	return nil
}
