package shipment

import (
	"context"
	"errors"
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
	Status         string    `json:"status"` // "PENDING", "CREATED", "CANCELLED"
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Label represents the label details received from the Label Service.
type Label struct {
	ID             string    `json:"id"`
	ShipmentID     string    `json:"shipment_id"`
	TrackingNumber string    `json:"tracking_number"`
	LabelURL       string    `json:"label_url"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// Sentinel errors representing domain-specific failure cases.
var (
	ErrShipmentNotFound      = errors.New("shipment not found")
	ErrShipmentAlreadyExists = errors.New("shipment already exists")
	ErrInvalidShipment       = errors.New("invalid shipment data")
	ErrCarrierRequired       = errors.New("carrier is required")
	ErrInvalidWeight         = errors.New("shipment weight must be greater than zero")
)

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
	CreateShipment(ctx context.Context, carrier string, weight float64, origin, destination string) (*Shipment, *Label, error)
	GetShipment(ctx context.Context, id string) (*Shipment, error)
	ListShipments(ctx context.Context) ([]*Shipment, error)
	CancelShipmentByTracking(ctx context.Context, trackingNumber string) error
}
