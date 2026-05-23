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
	Status         string    `json:"status"`   // "PENDING", "CREATED", "CANCELLED"
	Username       string    `json:"username"` // Owner of this shipment record
	Email          string    `json:"email"`    // Recipient email for notifications
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
	ErrRateLimitExceeded     = errors.New("rate limit exceeded: you can only create one shipment every 5 seconds")
	ErrInvalidStatus         = errors.New("invalid status: must relate to standard shipping orders (CREATED, IN_TRANSIT, OUT_FOR_DELIVERY, DELIVERED, CANCELLED, RETURNED)")
)

// ShipmentRepository defines the SQLite DB access contract for Shipments.
type ShipmentRepository interface {
	Create(ctx context.Context, shipment *Shipment) error
	GetByID(ctx context.Context, id string) (*Shipment, error)
	GetByTracking(ctx context.Context, trackingNum string) (*Shipment, error)
	List(ctx context.Context) ([]*Shipment, error)
	ListByUsername(ctx context.Context, username string) ([]*Shipment, error)
	Update(ctx context.Context, shipment *Shipment) error
	Delete(ctx context.Context, id string) error
}

// ShipmentService defines the business logic operations for Shipments.
type ShipmentService interface {
	CreateShipment(ctx context.Context, token, carrier string, weight float64, origin, destination, email string) (*Shipment, *Label, error)
	GetShipment(ctx context.Context, id string) (*Shipment, error)
	ListShipments(ctx context.Context, token string) ([]*Shipment, error)
	UpdateShipment(ctx context.Context, token, id, carrier string, weight float64, origin, destination, status string) (*Shipment, error)
	DeleteShipment(ctx context.Context, token, id string) error
	CancelShipmentByTracking(ctx context.Context, trackingNumber string) error
}
