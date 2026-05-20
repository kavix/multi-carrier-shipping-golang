package domain

import (
	"context"
	"time"
)

// Shipment represents the core domain model for a shipment.
type Shipment struct {
	ID             string    `json:"id"`
	Carrier        string    `json:"carrier"`
	TrackingNumber string    `json:"tracking_number"`
	Weight         float64   `json:"weight"`
	Origin         string    `json:"origin"`
	Destination    string    `json:"destination"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ShipmentRepository defines the data access contract for Shipments.
type ShipmentRepository interface {
	Create(ctx context.Context, shipment *Shipment) error
	GetByID(ctx context.Context, id string) (*Shipment, error)
	GetByTracking(ctx context.Context, trackingNum string) (*Shipment, error)
	List(ctx context.Context) ([]*Shipment, error)
	Update(ctx context.Context, shipment *Shipment) error
	Delete(ctx context.Context, id string) error
}

// ShipmentService defines the business logic operations for Shipments.
type ShipmentService interface {
	CreateShipment(ctx context.Context, carrier, trackingNumber string, weight float64, origin, destination string) (*Shipment, error)
	GetShipment(ctx context.Context, id string) (*Shipment, error)
	ListShipments(ctx context.Context) ([]*Shipment, error)
}
