package shipment

import (
	"context"
	"sync"
	"time"
)

type MemoryShipmentRepository struct {
	mu        sync.RWMutex
	shipments map[string]*Shipment
}

// NewMemoryShipmentRepository instantiates an isolated in-memory database for Shipments.
func NewMemoryShipmentRepository() *MemoryShipmentRepository {
	return &MemoryShipmentRepository{
		shipments: make(map[string]*Shipment),
	}
}

func (r *MemoryShipmentRepository) Create(ctx context.Context, shipment *Shipment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already exists by ID
	if _, exists := r.shipments[shipment.ID]; exists {
		return ErrShipmentAlreadyExists
	}

	// Check if already exists by tracking number (if assigned)
	if shipment.TrackingNumber != "" && shipment.TrackingNumber != "PENDING" {
		for _, s := range r.shipments {
			if s.TrackingNumber == shipment.TrackingNumber {
				return ErrShipmentAlreadyExists
			}
		}
	}

	r.shipments[shipment.ID] = shipment
	return nil
}

func (r *MemoryShipmentRepository) GetByID(ctx context.Context, id string) (*Shipment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shipment, exists := r.shipments[id]
	if !exists {
		return nil, ErrShipmentNotFound
	}

	// Return copy
	cp := *shipment
	return &cp, nil
}

func (r *MemoryShipmentRepository) GetByTracking(ctx context.Context, trackingNum string) (*Shipment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, s := range r.shipments {
		if s.TrackingNumber == trackingNum {
			cp := *s
			return &cp, nil
		}
	}

	return nil, ErrShipmentNotFound
}

func (r *MemoryShipmentRepository) List(ctx context.Context) ([]*Shipment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*Shipment, 0, len(r.shipments))
	for _, s := range r.shipments {
		cp := *s
		list = append(list, &cp)
	}

	return list, nil
}

func (r *MemoryShipmentRepository) Update(ctx context.Context, shipment *Shipment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.shipments[shipment.ID]; !exists {
		return ErrShipmentNotFound
	}

	shipment.UpdatedAt = time.Now()
	r.shipments[shipment.ID] = shipment
	return nil
}

func (r *MemoryShipmentRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.shipments[id]; !exists {
		return ErrShipmentNotFound
	}

	delete(r.shipments, id)
	return nil
}
